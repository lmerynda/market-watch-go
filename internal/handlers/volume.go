package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"
	"market-watch-go/internal/services"

	"github.com/gin-gonic/gin"
)

type VolumeHandler struct {
	db        *database.DB
	collector *services.CollectorService
	polygon   *services.PolygonService
}

// NewVolumeHandler creates a new volume handler
func NewVolumeHandler(db *database.DB, collector *services.CollectorService, polygon *services.PolygonService) *VolumeHandler {
	return &VolumeHandler{
		db:        db,
		collector: collector,
		polygon:   polygon,
	}
}

// GetVolumeData handles GET /api/volume/:symbol
func (vh *VolumeHandler) GetVolumeData(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Parse query parameters
	fromStr := c.DefaultQuery("from", "")
	toStr := c.DefaultQuery("to", "")
	intervalStr := c.DefaultQuery("interval", "5m")
	limitStr := c.DefaultQuery("limit", "1000")
	offsetStr := c.DefaultQuery("offset", "0")

	// Default time range (last 24 hours)
	to := time.Now()
	from := to.AddDate(0, 0, -1)

	// Parse from parameter if provided
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			parsedFrom, err = time.Parse("2006-01-02T15:04:05Z", fromStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error:   "invalid_from_date",
					Message: "Invalid from date format. Use YYYY-MM-DD or RFC3339",
				})
				return
			}
		}
		from = parsedFrom
	}

	// Parse to parameter if provided
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			parsedTo, err = time.Parse("2006-01-02T15:04:05Z", toStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error:   "invalid_to_date",
					Message: "Invalid to date format. Use YYYY-MM-DD or RFC3339",
				})
				return
			}
		}
		to = parsedTo
	}

	// Parse limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		limit = 1000
	}
	if limit > 10000 { // Max limit
		limit = 10000
	}

	// Parse offset
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Create filter
	filter := &models.VolumeDataFilter{
		Symbol:   symbol,
		From:     from,
		To:       to,
		Interval: intervalStr,
		Limit:    limit,
		Offset:   offset,
	}

	// Get data from database
	data, err := vh.db.GetVolumeData(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve volume data",
		})
		return
	}

	// Convert to slice of values for JSON response
	volumeDataValues := make([]models.VolumeData, len(data))
	for i, vd := range data {
		volumeDataValues[i] = *vd
	}

	// Create response
	response := models.VolumeDataResponse{
		Symbol:       symbol,
		Data:         volumeDataValues,
		TotalRecords: len(volumeDataValues),
		From:         from,
		To:           to,
		Interval:     intervalStr,
	}

	c.JSON(http.StatusOK, response)
}

// GetLatestVolumeData handles GET /api/volume/:symbol/latest
func (vh *VolumeHandler) GetLatestVolumeData(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Get latest data from database
	data, err := vh.db.GetLatestVolumeData(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve latest volume data",
		})
		return
	}

	if data == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "No volume data found for symbol",
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetDashboardSummary handles GET /api/dashboard/summary
func (vh *VolumeHandler) GetDashboardSummary(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 7
	}
	if days > 30 {
		days = 30
	}

	// Get watched symbols from database
	symbols, err := vh.db.GetWatchedSymbols()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve watched symbols",
		})
		return
	}

	var volumeStats []models.VolumeStats
	var lastUpdate time.Time

	// Get stats for each symbol
	for _, symbol := range symbols {
		stats, err := vh.db.GetVolumeStats(symbol, days)
		if err != nil {
			continue // Skip symbols with errors
		}
		if stats != nil {
			volumeStats = append(volumeStats, *stats)
			if stats.LastUpdate.After(lastUpdate) {
				lastUpdate = stats.LastUpdate
			}
		}
	}

	// Create summary
	summary := models.DashboardSummary{
		Symbols:        volumeStats,
		LastUpdate:     lastUpdate,
		CollectionMode: "24/7", // Indicates continuous data collection
	}

	c.JSON(http.StatusOK, summary)
}

// GetChartData handles GET /api/volume/:symbol/chart
func (vh *VolumeHandler) GetChartData(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Parse query parameters for time range
	rangeStr := c.DefaultQuery("range", "1W")

	var from, to time.Time
	now := time.Now()

	switch rangeStr {
	case "1D":
		from = now.AddDate(0, 0, -1)
		to = now
	case "1W":
		from = now.AddDate(0, 0, -7)
		to = now
	case "2W":
		from = now.AddDate(0, 0, -14)
		to = now
	case "1M":
		from = now.AddDate(0, -1, 0)
		to = now
	default:
		from = now.AddDate(0, 0, -7)
		to = now
	}

	// Get data from database
	filter := &models.VolumeDataFilter{
		Symbol: symbol,
		From:   from,
		To:     to,
		Limit:  1000,
	}

	data, err := vh.db.GetVolumeData(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve chart data",
		})
		return
	}

	// Convert to chart data format
	chartPoints := make([]models.ChartDataPoint, 0)
	for _, vd := range data {
		chartPoints = append(chartPoints, models.ChartDataPoint{
			X: vd.Timestamp.Format(time.RFC3339),
			Y: vd.Volume,
		})
	}

	// Get chart colors
	colors := models.GetChartColors()
	color, exists := colors[symbol]
	if !exists {
		color = models.ChartColors{
			Border:     "#1f77b4",
			Background: "rgba(31, 119, 180, 0.1)",
		}
	}

	// Create chart dataset
	dataset := models.ChartDataset{
		Label:           symbol + " Volume",
		Data:            chartPoints,
		BorderColor:     color.Border,
		BackgroundColor: color.Background,
		Fill:            true,
		Tension:         0.1,
	}

	// Create chart data
	chartData := models.ChartData{
		Labels:   []string{}, // Chart.js will use the X values from data points
		Datasets: []models.ChartDataset{dataset},
	}

	c.JSON(http.StatusOK, chartData)
}

// ForceCollection handles POST /api/collection/force
func (vh *VolumeHandler) ForceCollection(c *gin.Context) {
	err := vh.collector.ForceCollection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "collection_error",
			Message: "Failed to force collection",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Collection triggered successfully",
		"status":  "initiated",
	})
}

// GetCollectionStatus handles GET /api/collection/status
func (vh *VolumeHandler) GetCollectionStatus(c *gin.Context) {
	stats := vh.collector.GetStats()
	c.JSON(http.StatusOK, stats)
}

// HealthCheck handles GET /api/health
func (vh *VolumeHandler) HealthCheck(c *gin.Context) {
	health := models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Services:  make(map[string]models.ServiceHealth),
	}

	// Check database health
	if err := vh.db.HealthCheck(); err != nil {
		health.Services["database"] = models.ServiceHealth{
			Status:  "error",
			Message: err.Error(),
		}
		health.Status = "degraded"
	} else {
		health.Services["database"] = models.ServiceHealth{
			Status: "ok",
		}
	}

	// Check Polygon service health
	if err := vh.polygon.HealthCheck(); err != nil {
		health.Services["polygon"] = models.ServiceHealth{
			Status:  "error",
			Message: err.Error(),
		}
		health.Status = "degraded"
	} else {
		health.Services["polygon"] = models.ServiceHealth{
			Status: "ok",
		}
	}

	// Check collector health
	if err := vh.collector.HealthCheck(); err != nil {
		health.Services["collector"] = models.ServiceHealth{
			Status:  "error",
			Message: err.Error(),
		}
		health.Status = "degraded"
	} else {
		health.Services["collector"] = models.ServiceHealth{
			Status: "ok",
		}
	}

	// Set overall status
	if health.Status == "ok" {
		c.JSON(http.StatusOK, health)
	} else {
		c.JSON(http.StatusServiceUnavailable, health)
	}
}

// GetWatchedSymbols handles GET /api/symbols
func (vh *VolumeHandler) GetWatchedSymbols(c *gin.Context) {
	symbols, err := vh.db.GetWatchedSymbolsWithDetails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve watched symbols",
		})
		return
	}

	// Convert to slice of values for JSON response (symbols is already []models.WatchedSymbol)
	response := models.WatchedSymbolsResponse{
		Symbols: symbols,
		Count:   len(symbols),
	}

	c.JSON(http.StatusOK, response)
}

// AddWatchedSymbol handles POST /api/symbols
func (vh *VolumeHandler) AddWatchedSymbol(c *gin.Context) {
	var req models.WatchedSymbolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
		})
		return
	}

	// Validate symbol format (basic validation)
	if len(req.Symbol) < 1 || len(req.Symbol) > 10 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_symbol",
			Message: "Symbol must be between 1 and 10 characters",
		})
		return
	}

	// Convert to uppercase
	req.Symbol = strings.ToUpper(req.Symbol)

	err := vh.db.AddWatchedSymbol(req.Symbol, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to add watched symbol",
		})
		return
	}

	// Trigger immediate data collection for the new symbol
	go func() {
		if err := vh.collector.CollectSymbolDataNow(req.Symbol); err != nil {
			log.Printf("Failed to collect initial data for symbol %s: %v", req.Symbol, err)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Symbol added successfully",
		"symbol":  req.Symbol,
	})
}

// RemoveWatchedSymbol handles DELETE /api/symbols/:symbol
func (vh *VolumeHandler) RemoveWatchedSymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Symbol parameter is required",
		})
		return
	}

	symbol = strings.ToUpper(symbol)

	err := vh.db.RemoveWatchedSymbol(symbol)
	if err != nil {
		if err.Error() == fmt.Sprintf("symbol not found: %s", symbol) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "symbol_not_found",
				Message: "Symbol not found in watched list",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to remove watched symbol",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Symbol removed successfully",
		"symbol":  symbol,
	})
}

// CheckSymbolData handles GET /api/symbols/:symbol/check - checks if symbol has data
func (vh *VolumeHandler) CheckSymbolData(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Symbol parameter is required",
		})
		return
	}

	symbol = strings.ToUpper(symbol)

	// Check if we have any data for this symbol
	latestData, err := vh.db.GetLatestVolumeData(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to check symbol data",
		})
		return
	}

	hasData := latestData != nil
	var dataCount int64 = 0
	var lastUpdate *time.Time

	if hasData {
		// Get data count for this symbol
		counts, err := vh.db.GetDataCountBySymbol()
		if err == nil {
			dataCount = counts[symbol]
		}
		lastUpdate = &latestData.Timestamp
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":      symbol,
		"has_data":    hasData,
		"data_points": dataCount,
		"last_update": lastUpdate,
	})
}

// CollectSymbolData handles POST /api/symbols/:symbol/collect - manually trigger data collection
func (vh *VolumeHandler) CollectSymbolData(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Symbol parameter is required",
		})
		return
	}

	symbol = strings.ToUpper(symbol)

	// Check if symbol is being watched
	watchedSymbols, err := vh.db.GetWatchedSymbols()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to check watched symbols",
		})
		return
	}

	isWatched := false
	for _, ws := range watchedSymbols {
		if ws == symbol {
			isWatched = true
			break
		}
	}

	if !isWatched {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "symbol_not_watched",
			Message: "Symbol is not in the watchlist",
		})
		return
	}

	// Trigger data collection in background
	go func() {
		if err := vh.collector.CollectSymbolDataNow(symbol); err != nil {
			log.Printf("Failed to collect data for symbol %s: %v", symbol, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Data collection triggered for symbol",
		"symbol":  symbol,
		"status":  "initiated",
	})
}
