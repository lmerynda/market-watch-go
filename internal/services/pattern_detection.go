package services

import (
	"fmt"
	"log"
	"time"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"
)

// PatternDetectionService handles automatic pattern detection for multiple pattern types
type PatternDetectionService struct {
	db           *database.DB
	taService    *TechnicalAnalysisService
	hsService    *HeadShouldersDetectionService
	emailService *EmailService
}

// NewPatternDetectionService creates a new pattern detection service
func NewPatternDetectionService(db *database.DB, taService *TechnicalAnalysisService, hsService *HeadShouldersDetectionService, emailService *EmailService) *PatternDetectionService {
	return &PatternDetectionService{
		db:           db,
		taService:    taService,
		hsService:    hsService,
		emailService: emailService,
	}
}

// AutoDetectPatternsForSymbol automatically detects all pattern types for a given symbol
func (pds *PatternDetectionService) AutoDetectPatternsForSymbol(symbol string) error {
	log.Printf("Starting automatic pattern detection for %s", symbol)

	// Detect Head & Shoulders patterns (both regular and inverse)
	if err := pds.detectHeadShouldersPatterns(symbol); err != nil {
		log.Printf("Failed to detect H&S patterns for %s: %v", symbol, err)
	}

	// TODO: Add other pattern detection methods
	// if err := pds.detectCupAndHandlePatterns(symbol); err != nil {
	//     log.Printf("Failed to detect Cup & Handle patterns for %s: %v", symbol, err)
	// }

	// if err := pds.detectTrianglePatterns(symbol); err != nil {
	//     log.Printf("Failed to detect Triangle patterns for %s: %v", symbol, err)
	// }

	log.Printf("Completed automatic pattern detection for %s", symbol)
	return nil
}

// AutoDetectPatternsForAllSymbols runs pattern detection for all watched symbols
func (pds *PatternDetectionService) AutoDetectPatternsForAllSymbols() error {
	symbols, err := pds.db.GetWatchedSymbols()
	if err != nil {
		return fmt.Errorf("failed to get watched symbols: %w", err)
	}

	log.Printf("Starting automatic pattern detection for %d symbols", len(symbols))

	for _, symbol := range symbols {
		if err := pds.AutoDetectPatternsForSymbol(symbol); err != nil {
			log.Printf("Failed pattern detection for %s: %v", symbol, err)
			continue
		}
	}

	log.Printf("Completed automatic pattern detection for all symbols")
	return nil
}

// detectHeadShouldersPatterns detects both regular and inverse head & shoulders patterns
func (pds *PatternDetectionService) detectHeadShouldersPatterns(symbol string) error {
	// Detect Inverse Head & Shoulders (bullish)
	_, err := pds.hsService.DetectInverseHeadShoulders(symbol)
	if err != nil {
		log.Printf("No inverse H&S pattern found for %s: %v", symbol, err)
	}

	// TODO: Add regular Head & Shoulders detection (bearish)
	// _, err = pds.hsService.DetectHeadShoulders(symbol)
	// if err != nil {
	//     log.Printf("No H&S pattern found for %s: %v", symbol, err)
	// }

	return nil
}

// StartPeriodicPatternDetection starts a background goroutine that periodically scans for patterns
func (pds *PatternDetectionService) StartPeriodicPatternDetection() {
	log.Printf("Starting periodic pattern detection service...")

	// Run pattern detection every 30 minutes
	ticker := time.NewTicker(30 * time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Printf("Running periodic pattern detection...")
				if err := pds.AutoDetectPatternsForAllSymbols(); err != nil {
					log.Printf("Periodic pattern detection failed: %v", err)
				}
			}
		}
	}()
}

// OnSymbolAdded is called when a new symbol is added to the watchlist
func (pds *PatternDetectionService) OnSymbolAdded(symbol string) {
	log.Printf("New symbol %s added to watchlist, triggering pattern detection", symbol)

	// Run pattern detection in a goroutine to avoid blocking
	go func() {
		if err := pds.AutoDetectPatternsForSymbol(symbol); err != nil {
			log.Printf("Failed to detect patterns for newly added symbol %s: %v", symbol, err)
		}
	}()
}

// OnDataUpdated is called when new price data is available for symbols
func (pds *PatternDetectionService) OnDataUpdated(symbols []string) {
	log.Printf("Price data updated for %d symbols, checking for pattern updates", len(symbols))

	go func() {
		for _, symbol := range symbols {
			// Only run pattern detection if we haven't run it recently for this symbol
			if pds.shouldRunPatternDetection(symbol) {
				if err := pds.AutoDetectPatternsForSymbol(symbol); err != nil {
					log.Printf("Failed to update patterns for %s: %v", symbol, err)
				}
			}
		}
	}()
}

// shouldRunPatternDetection checks if enough time has passed to run pattern detection again
func (pds *PatternDetectionService) shouldRunPatternDetection(symbol string) bool {
	// For now, always return true. In production, you might want to:
	// - Check when patterns were last detected for this symbol
	// - Only run if significant time has passed (e.g., 1 hour)
	// - Check if significant price movement has occurred
	return true
}

// GetPatternSummaryForSymbol returns a summary of all detected patterns for a symbol
func (pds *PatternDetectionService) GetPatternSummaryForSymbol(symbol string) (*PatternSummary, error) {
	// Get Head & Shoulders patterns
	hsPatterns, err := pds.db.GetHeadShouldersPatterns(&models.PatternFilter{
		Symbol: symbol,
		Limit:  100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get H&S patterns: %w", err)
	}

	summary := &PatternSummary{
		Symbol:                symbol,
		TotalPatterns:         len(hsPatterns),
		ActivePatterns:        0,
		CompletedPatterns:     0,
		HeadShouldersPatterns: hsPatterns,
		LastDetectionRun:      time.Now(),
	}

	for _, pattern := range hsPatterns {
		if pattern.IsComplete {
			summary.CompletedPatterns++
		} else {
			summary.ActivePatterns++
		}
	}

	return summary, nil
}

// PatternSummary represents a summary of all patterns for a symbol
type PatternSummary struct {
	Symbol                string                         `json:"symbol"`
	TotalPatterns         int                            `json:"total_patterns"`
	ActivePatterns        int                            `json:"active_patterns"`
	CompletedPatterns     int                            `json:"completed_patterns"`
	HeadShouldersPatterns []*models.HeadShouldersPattern `json:"head_shoulders_patterns"`
	// TODO: Add other pattern types
	// CupHandlePatterns     []*models.CupHandlePattern      `json:"cup_handle_patterns"`
	// TrianglePatterns      []*models.TrianglePattern       `json:"triangle_patterns"`
	LastDetectionRun time.Time `json:"last_detection_run"`
}
