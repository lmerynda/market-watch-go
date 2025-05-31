package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"market-watch-go/internal/config"
	"market-watch-go/internal/database"
	"market-watch-go/internal/models"

	"github.com/robfig/cron/v3"
)

type CollectorService struct {
	db      *database.DB
	polygon *PolygonService
	cfg     *config.Config
	cron    *cron.Cron
	stats   *CollectionStats
	mutex   sync.RWMutex
}

type CollectionStats struct {
	LastRun        time.Time `json:"last_run"`
	NextRun        time.Time `json:"next_run"`
	SuccessfulRuns int       `json:"successful_runs"`
	FailedRuns     int       `json:"failed_runs"`
	LastError      string    `json:"last_error"`
	CollectedToday int       `json:"collected_today"`
	IsRunning      bool      `json:"is_running"`
	TotalCollected int64     `json:"total_collected"`
}

// NewCollectorService creates a new data collector service
func NewCollectorService(db *database.DB, polygon *PolygonService, cfg *config.Config) *CollectorService {
	cronInstance := cron.New(cron.WithLocation(time.UTC))

	return &CollectorService{
		db:      db,
		polygon: polygon,
		cfg:     cfg,
		cron:    cronInstance,
		stats: &CollectionStats{
			LastRun:        time.Time{},
			NextRun:        time.Time{},
			SuccessfulRuns: 0,
			FailedRuns:     0,
			LastError:      "",
			CollectedToday: 0,
			IsRunning:      false,
			TotalCollected: 0,
		},
	}
}

// Start begins the scheduled data collection
func (cs *CollectorService) Start() error {
	// Convert interval to cron expression
	cronExpr, err := cs.intervalToCron(cs.cfg.Collection.Interval)
	if err != nil {
		return fmt.Errorf("failed to convert interval to cron expression: %w", err)
	}

	// Schedule the collection job
	_, err = cs.cron.AddFunc(cronExpr, cs.collectData)
	if err != nil {
		return fmt.Errorf("failed to schedule collection job: %w", err)
	}

	// Schedule daily cleanup job (runs at 2 AM UTC)
	_, err = cs.cron.AddFunc("0 2 * * *", cs.cleanupOldData)
	if err != nil {
		return fmt.Errorf("failed to schedule cleanup job: %w", err)
	}

	// Start the cron scheduler
	cs.cron.Start()

	// Update next run time
	cs.updateNextRunTime()

	log.Printf("Data collector started with interval: %v", cs.cfg.Collection.Interval)

	// Run initial collection if market is open
	if cs.cfg.IsMarketHours() {
		go cs.collectData()
	}

	return nil
}

// Stop stops the scheduled data collection
func (cs *CollectorService) Stop() {
	if cs.cron != nil {
		cs.cron.Stop()
		log.Printf("Data collector stopped")
	}
}

// collectData performs the actual data collection
func (cs *CollectorService) collectData() {
	cs.mutex.Lock()
	if cs.stats.IsRunning {
		cs.mutex.Unlock()
		log.Printf("Collection already running, skipping...")
		return
	}
	cs.stats.IsRunning = true
	cs.stats.LastRun = time.Now()
	cs.mutex.Unlock()

	defer func() {
		cs.mutex.Lock()
		cs.stats.IsRunning = false
		cs.updateNextRunTime()
		cs.mutex.Unlock()
	}()

	log.Printf("Starting data collection for symbols: %v", cs.cfg.Collection.Symbols)

	// Check if market is open
	isOpen, err := cs.polygon.GetMarketStatus()
	if err != nil {
		log.Printf("Failed to get market status: %v, proceeding anyway", err)
	}

	if !isOpen {
		log.Printf("Market is closed, skipping collection")
		return
	}

	// Collect data for all symbols
	collectedCount := 0
	errorCount := 0

	for _, symbol := range cs.cfg.Collection.Symbols {
		count, err := cs.collectSymbolData(symbol)
		if err != nil {
			log.Printf("Failed to collect data for %s: %v", symbol, err)
			errorCount++
			continue
		}
		collectedCount += count
		log.Printf("Collected %d data points for %s", count, symbol)

		// Small delay between symbol requests to respect rate limits
		time.Sleep(100 * time.Millisecond)
	}

	// Update statistics
	cs.mutex.Lock()
	if errorCount == 0 {
		cs.stats.SuccessfulRuns++
		cs.stats.LastError = ""
	} else {
		cs.stats.FailedRuns++
		cs.stats.LastError = fmt.Sprintf("Failed to collect data for %d symbols", errorCount)
	}
	cs.stats.CollectedToday += collectedCount
	cs.stats.TotalCollected += int64(collectedCount)
	cs.mutex.Unlock()

	log.Printf("Data collection completed. Collected: %d, Errors: %d", collectedCount, errorCount)
}

// collectSymbolData collects data for a single symbol
func (cs *CollectorService) collectSymbolData(symbol string) (int, error) {
	// Get recent data (last 2 hours to ensure we don't miss anything)
	data, err := cs.polygon.GetLatestAggregates(symbol, 120)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest aggregates: %w", err)
	}

	if len(data) == 0 {
		log.Printf("No new data available for %s", symbol)
		return 0, nil
	}

	// Filter out data we might already have
	newData, err := cs.filterNewData(data)
	if err != nil {
		return 0, fmt.Errorf("failed to filter new data: %w", err)
	}

	if len(newData) == 0 {
		log.Printf("No new data points for %s", symbol)
		return 0, nil
	}

	// Insert new data into database
	err = cs.db.InsertVolumeDataBatch(newData)
	if err != nil {
		return 0, fmt.Errorf("failed to insert data batch: %w", err)
	}

	return len(newData), nil
}

// filterNewData filters out data points that already exist in the database
func (cs *CollectorService) filterNewData(data []*models.VolumeData) ([]*models.VolumeData, error) {
	if len(data) == 0 {
		return data, nil
	}

	symbol := data[0].Symbol

	// Get the latest timestamp we have for this symbol
	latest, err := cs.db.GetLatestVolumeData(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest volume data: %w", err)
	}

	// If we have no data, return all new data
	if latest == nil {
		return data, nil
	}

	// Filter out data points that are older or equal to our latest timestamp
	var newData []*models.VolumeData
	for _, vd := range data {
		if vd.Timestamp.After(latest.Timestamp) {
			newData = append(newData, vd)
		}
	}

	return newData, nil
}

// cleanupOldData removes old data based on retention policy
func (cs *CollectorService) cleanupOldData() {
	log.Printf("Starting data cleanup...")

	rowsDeleted, err := cs.db.CleanupOldData(cs.cfg.DataRetention.Days)
	if err != nil {
		log.Printf("Failed to cleanup old data: %v", err)
		return
	}

	log.Printf("Data cleanup completed. Deleted %d old records", rowsDeleted)
}

// intervalToCron converts a time.Duration to a cron expression
func (cs *CollectorService) intervalToCron(interval time.Duration) (string, error) {
	switch {
	case interval == time.Minute:
		return "* * * * *", nil
	case interval == 5*time.Minute:
		return "*/5 * * * *", nil
	case interval == 10*time.Minute:
		return "*/10 * * * *", nil
	case interval == 15*time.Minute:
		return "*/15 * * * *", nil
	case interval == 30*time.Minute:
		return "*/30 * * * *", nil
	case interval == time.Hour:
		return "0 * * * *", nil
	default:
		// For other intervals, use a minute-based approximation
		minutes := int(interval.Minutes())
		if minutes <= 0 {
			return "", fmt.Errorf("invalid interval: %v", interval)
		}
		if minutes >= 60 {
			return "0 * * * *", nil // Default to hourly for long intervals
		}
		return fmt.Sprintf("*/%d * * * *", minutes), nil
	}
}

// updateNextRunTime calculates and updates the next run time
func (cs *CollectorService) updateNextRunTime() {
	// This is an approximation since cron doesn't expose next run time easily
	cs.stats.NextRun = time.Now().Add(cs.cfg.Collection.Interval)
}

// GetStats returns the current collection statistics
func (cs *CollectorService) GetStats() *CollectionStats {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	// Create a copy to avoid race conditions
	statsCopy := *cs.stats
	return &statsCopy
}

// ForceCollection triggers an immediate data collection
func (cs *CollectorService) ForceCollection() error {
	log.Printf("Forcing immediate data collection...")
	go cs.collectData()
	return nil
}

// CollectHistoricalData collects historical data for all symbols
func (cs *CollectorService) CollectHistoricalData(days int) error {
	log.Printf("Starting historical data collection for %d days", days)

	for _, symbol := range cs.cfg.Collection.Symbols {
		log.Printf("Collecting historical data for %s", symbol)

		data, err := cs.polygon.GetHistoricalData(symbol, days)
		if err != nil {
			log.Printf("Failed to get historical data for %s: %v", symbol, err)
			continue
		}

		if len(data) == 0 {
			log.Printf("No historical data available for %s", symbol)
			continue
		}

		// Insert historical data
		err = cs.db.InsertVolumeDataBatch(data)
		if err != nil {
			log.Printf("Failed to insert historical data for %s: %v", symbol, err)
			continue
		}

		log.Printf("Inserted %d historical data points for %s", len(data), symbol)

		// Delay between symbols to respect rate limits
		time.Sleep(1 * time.Second)
	}

	log.Printf("Historical data collection completed")
	return nil
}

// HealthCheck performs a health check on the collector service
func (cs *CollectorService) HealthCheck() error {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	// Check if the service has been running recently
	if !cs.stats.LastRun.IsZero() {
		timeSinceLastRun := time.Since(cs.stats.LastRun)
		if timeSinceLastRun > 2*cs.cfg.Collection.Interval {
			return fmt.Errorf("collection hasn't run in %v (expected every %v)",
				timeSinceLastRun, cs.cfg.Collection.Interval)
		}
	}

	// Check error rate
	totalRuns := cs.stats.SuccessfulRuns + cs.stats.FailedRuns
	if totalRuns > 0 {
		errorRate := float64(cs.stats.FailedRuns) / float64(totalRuns)
		if errorRate > 0.5 { // More than 50% failure rate
			return fmt.Errorf("high error rate: %.1f%% (%d failed out of %d total runs)",
				errorRate*100, cs.stats.FailedRuns, totalRuns)
		}
	}

	return nil
}
