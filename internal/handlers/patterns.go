package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"
	"market-watch-go/internal/services"

	"github.com/gin-gonic/gin"
)

// PatternsHandler handles unified pattern detection for all pattern types
type PatternsHandler struct {
	db                  *database.DB
	patternService      *services.PatternDetectionService
	hsService           *services.HeadShouldersDetectionService
	fallingWedgeService *services.FallingWedgeDetectionService
}

// NewPatternsHandler creates a new unified patterns handler
func NewPatternsHandler(
	db *database.DB,
	patternService *services.PatternDetectionService,
	hsService *services.HeadShouldersDetectionService,
	fallingWedgeService *services.FallingWedgeDetectionService,
) *PatternsHandler {
	return &PatternsHandler{
		db:                  db,
		patternService:      patternService,
		hsService:           hsService,
		fallingWedgeService: fallingWedgeService,
	}
}

// ScanAllPatterns scans all watched symbols for all pattern types
func (h *PatternsHandler) ScanAllPatterns(c *gin.Context) {
	// Get all watched symbols
	symbols, err := h.db.GetWatchedSymbols()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get watched symbols",
			"details": err.Error(),
		})
		return
	}

	if len(symbols) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":              "No symbols to scan",
			"symbols_scanned":      0,
			"patterns_found":       0,
			"head_shoulders_found": 0,
			"falling_wedge_found":  0,
		})
		return
	}

	// Track scan results
	var headShouldersFound, fallingWedgeFound int
	var scanErrors []string
	scannedSymbols := 0

	// Scan each symbol for all pattern types
	for _, symbol := range symbols {
		scannedSymbols++

		// Scan for Head & Shoulders patterns
		_, err := h.hsService.DetectInverseHeadShoulders(symbol)
		if err == nil {
			headShouldersFound++
		}

		// Scan for Falling Wedge patterns
		_, err = h.fallingWedgeService.DetectFallingWedge(symbol)
		if err == nil {
			fallingWedgeFound++
		} else {
			// Only log non-pattern-not-found errors
			if err.Error() != "no valid falling wedge pattern found" &&
				err.Error() != "no valid inverse head and shoulders pattern found" &&
				err.Error() != "insufficient price data for pattern detection" {
				scanErrors = append(scanErrors, fmt.Sprintf("%s: %s", symbol, err.Error()))
			}
		}
	}

	totalPatternsFound := headShouldersFound + fallingWedgeFound

	// Prepare response
	response := gin.H{
		"message":              fmt.Sprintf("Pattern scan completed for %d symbols", scannedSymbols),
		"symbols_scanned":      scannedSymbols,
		"patterns_found":       totalPatternsFound,
		"head_shoulders_found": headShouldersFound,
		"falling_wedge_found":  fallingWedgeFound,
	}

	// Include errors if any
	if len(scanErrors) > 0 {
		response["errors"] = scanErrors
		response["message"] = fmt.Sprintf("Pattern scan completed for %d symbols with %d errors", scannedSymbols, len(scanErrors))
	}

	c.JSON(http.StatusOK, response)
}

// ScanSymbolPatterns scans a specific symbol for all pattern types
func (h *PatternsHandler) ScanSymbolPatterns(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Symbol parameter is required",
		})
		return
	}

	var headShouldersPattern *models.HeadShouldersPattern
	var fallingWedgePattern *models.FallingWedgePattern
	var scanErrors []string

	// Detect Head & Shoulders pattern
	hsPattern, err := h.hsService.DetectInverseHeadShoulders(symbol)
	if err == nil {
		headShouldersPattern = hsPattern
	} else if err.Error() != "no valid inverse head and shoulders pattern found" &&
		err.Error() != "insufficient price data for pattern detection" {
		scanErrors = append(scanErrors, fmt.Sprintf("Head & Shoulders: %s", err.Error()))
	}

	// Detect Falling Wedge pattern
	fwPattern, err := h.fallingWedgeService.DetectFallingWedge(symbol)
	if err == nil {
		fallingWedgePattern = fwPattern
	} else if err.Error() != "no valid falling wedge pattern found" &&
		err.Error() != "insufficient price data for pattern detection" {
		scanErrors = append(scanErrors, fmt.Sprintf("Falling Wedge: %s", err.Error()))
	}

	patternsFound := 0
	if headShouldersPattern != nil {
		patternsFound++
	}
	if fallingWedgePattern != nil {
		patternsFound++
	}

	response := gin.H{
		"symbol":                 symbol,
		"patterns_found":         patternsFound,
		"head_shoulders_pattern": headShouldersPattern,
		"falling_wedge_pattern":  fallingWedgePattern,
		"message":                fmt.Sprintf("Pattern detection completed for %s", symbol),
	}

	if len(scanErrors) > 0 {
		response["errors"] = scanErrors
	}

	c.JSON(http.StatusOK, response)
}

// GetAllPatterns returns all patterns of all types
func (h *PatternsHandler) GetAllPatterns(c *gin.Context) {
	// Parse query parameters
	symbolFilter := c.Query("symbol")
	patternType := c.Query("pattern_type") // "head_shoulders", "falling_wedge", or empty for all
	limitStr := c.DefaultQuery("limit", "100")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	response := gin.H{
		"patterns": gin.H{
			"head_shoulders": []interface{}{},
			"falling_wedge":  []interface{}{},
		},
		"total_count": 0,
	}

	// Get Head & Shoulders patterns
	if patternType == "" || patternType == "head_shoulders" {
		hsFilter := &models.PatternFilter{
			Symbol: symbolFilter,
			Limit:  limit,
		}
		hsPatterns, err := h.db.GetHeadShouldersPatterns(hsFilter)
		if err == nil && hsPatterns != nil {
			response["patterns"].(gin.H)["head_shoulders"] = hsPatterns
			response["total_count"] = response["total_count"].(int) + len(hsPatterns)
		}
	}

	// Get Falling Wedge patterns
	if patternType == "" || patternType == "falling_wedge" {
		fwFilter := &models.FallingWedgeFilter{
			Symbol: symbolFilter,
			Limit:  limit,
		}
		fwPatterns, err := h.db.GetFallingWedgePatterns(fwFilter)
		if err == nil && fwPatterns != nil {
			response["patterns"].(gin.H)["falling_wedge"] = fwPatterns
			response["total_count"] = response["total_count"].(int) + len(fwPatterns)
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetPatternsBySymbol returns all patterns for a specific symbol
func (h *PatternsHandler) GetPatternsBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Symbol parameter is required",
		})
		return
	}

	response := gin.H{
		"symbol": symbol,
		"patterns": gin.H{
			"head_shoulders": []interface{}{},
			"falling_wedge":  []interface{}{},
		},
		"total_count": 0,
	}

	// Get Head & Shoulders patterns for symbol
	hsFilter := &models.PatternFilter{
		Symbol: symbol,
		Limit:  100,
	}
	hsPatterns, err := h.db.GetHeadShouldersPatterns(hsFilter)
	if err == nil && hsPatterns != nil {
		response["patterns"].(gin.H)["head_shoulders"] = hsPatterns
		response["total_count"] = response["total_count"].(int) + len(hsPatterns)
	}

	// Get Falling Wedge patterns for symbol
	fwPatterns, err := h.db.GetFallingWedgePatternsBySymbol(symbol)
	if err == nil && fwPatterns != nil {
		response["patterns"].(gin.H)["falling_wedge"] = fwPatterns
		response["total_count"] = response["total_count"].(int) + len(fwPatterns)
	}

	c.JSON(http.StatusOK, response)
}

// GetPatternStatistics returns statistics for all pattern types
func (h *PatternsHandler) GetPatternStatistics(c *gin.Context) {
	// Get Head & Shoulders patterns
	hsPatterns, err := h.db.GetHeadShouldersPatterns(&models.PatternFilter{Limit: 10000})
	if err != nil {
		hsPatterns = []*models.HeadShouldersPattern{}
	}

	// Get Falling Wedge patterns
	fwPatterns, err := h.db.GetFallingWedgePatterns(&models.FallingWedgeFilter{Limit: 10000})
	if err != nil {
		fwPatterns = []*models.FallingWedgePattern{}
	}

	totalPatterns := len(hsPatterns) + len(fwPatterns)
	activePatterns := 0
	completedPatterns := 0

	// Count active/completed for Head & Shoulders
	for _, pattern := range hsPatterns {
		if pattern.IsComplete {
			completedPatterns++
		} else {
			activePatterns++
		}
	}

	// Count active/completed for Falling Wedge
	for _, pattern := range fwPatterns {
		if pattern.IsComplete {
			completedPatterns++
		} else {
			activePatterns++
		}
	}

	stats := gin.H{
		"total_patterns":     totalPatterns,
		"active_patterns":    activePatterns,
		"completed_patterns": completedPatterns,
		"pattern_types": gin.H{
			"head_shoulders": len(hsPatterns),
			"falling_wedge":  len(fwPatterns),
		},
		"last_updated": "now",
	}

	c.JSON(http.StatusOK, stats)
}
