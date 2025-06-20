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

// FallingWedgeHandler handles falling wedge pattern endpoints
type FallingWedgeHandler struct {
	db                  *database.DB
	fallingWedgeService *services.FallingWedgeDetectionService
}

// NewFallingWedgeHandler creates a new falling wedge handler
func NewFallingWedgeHandler(db *database.DB, fallingWedgeService *services.FallingWedgeDetectionService) *FallingWedgeHandler {
	return &FallingWedgeHandler{
		db:                  db,
		fallingWedgeService: fallingWedgeService,
	}
}

// DetectPattern detects falling wedge patterns for a symbol
func (h *FallingWedgeHandler) DetectPattern(c *gin.Context) {
	symbol := c.Param("symbol")

	pattern, err := h.fallingWedgeService.DetectFallingWedge(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to detect falling wedge pattern",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":  symbol,
		"pattern": pattern,
		"message": "Falling wedge pattern detected successfully",
	})
}

// GetPatterns retrieves falling wedge patterns with optional filtering
func (h *FallingWedgeHandler) GetPatterns(c *gin.Context) {
	filter := &models.FallingWedgeFilter{}

	// Parse query parameters
	if symbol := c.Query("symbol"); symbol != "" {
		filter.Symbol = symbol
	}

	if phase := c.Query("phase"); phase != "" {
		filter.Phase = phase
	}

	if isCompleteStr := c.Query("is_complete"); isCompleteStr != "" {
		if isComplete, err := strconv.ParseBool(isCompleteStr); err == nil {
			filter.IsComplete = &isComplete
		}
	}

	if minConvergenceStr := c.Query("min_convergence"); minConvergenceStr != "" {
		if minConvergence, err := strconv.ParseFloat(minConvergenceStr, 64); err == nil {
			filter.MinConvergence = minConvergence
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	} else {
		filter.Limit = 100 // Default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	patterns, err := h.db.GetFallingWedgePatterns(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get falling wedge patterns",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"patterns": patterns,
		"count":    len(patterns),
		"filter":   filter,
	})
}

// GetPatternsBySymbol retrieves falling wedge patterns for a specific symbol
func (h *FallingWedgeHandler) GetPatternsBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")

	patterns, err := h.db.GetFallingWedgePatternsBySymbol(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get falling wedge patterns for symbol",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":   symbol,
		"patterns": patterns,
		"count":    len(patterns),
	})
}

// GetPatternDetails retrieves detailed information about a specific falling wedge pattern
func (h *FallingWedgeHandler) GetPatternDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pattern ID",
		})
		return
	}

	pattern, err := h.db.GetFallingWedgePatternByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get falling wedge pattern",
			"details": err.Error(),
		})
		return
	}

	if pattern == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Pattern not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pattern": pattern,
	})
}

// GetActivePatterns retrieves all active falling wedge patterns
func (h *FallingWedgeHandler) GetActivePatterns(c *gin.Context) {
	patterns, err := h.db.GetActiveFallingWedgePatterns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get active falling wedge patterns",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"patterns": patterns,
		"count":    len(patterns),
	})
}

// GetPatternStatistics returns statistics about falling wedge patterns
func (h *FallingWedgeHandler) GetPatternStatistics(c *gin.Context) {
	// Get all patterns
	allPatterns, err := h.db.GetFallingWedgePatterns(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get patterns for statistics",
			"details": err.Error(),
		})
		return
	}

	// Calculate statistics
	stats := gin.H{
		"total_patterns":     len(allPatterns),
		"active_patterns":    0,
		"completed_patterns": 0,
		"avg_convergence":    0.0,
		"avg_duration_hours": 0.0,
		"phases": gin.H{
			"formation":      0,
			"breakout":       0,
			"target_pursuit": 0,
			"completed":      0,
		},
		"symbols": make(map[string]int),
	}

	if len(allPatterns) > 0 {
		var totalConvergence, totalDuration float64
		phases := map[string]int{
			"formation":      0,
			"breakout":       0,
			"target_pursuit": 0,
			"completed":      0,
		}
		symbols := make(map[string]int)

		for _, pattern := range allPatterns {
			if pattern.IsComplete {
				stats["completed_patterns"] = stats["completed_patterns"].(int) + 1
			} else {
				stats["active_patterns"] = stats["active_patterns"].(int) + 1
			}

			totalConvergence += pattern.Convergence
			totalDuration += float64(pattern.PatternWidth) / 60.0 // Convert minutes to hours

			phases[pattern.CurrentPhase]++
			symbols[pattern.Symbol]++
		}

		stats["avg_convergence"] = totalConvergence / float64(len(allPatterns))
		stats["avg_duration_hours"] = totalDuration / float64(len(allPatterns))
		stats["phases"] = phases
		stats["symbols"] = symbols
	}

	c.JSON(http.StatusOK, stats)
}

// ScanPatterns scans all watched symbols for falling wedge patterns
func (h *FallingWedgeHandler) ScanPatterns(c *gin.Context) {
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
			"message":         "No symbols to scan",
			"symbols_scanned": 0,
			"patterns_found":  0,
		})
		return
	}

	// Track scan results
	var patternsFound int
	var scanErrors []string
	scannedSymbols := 0

	// Scan each symbol for patterns
	for _, symbol := range symbols {
		scannedSymbols++
		_, err := h.fallingWedgeService.DetectFallingWedge(symbol)
		if err != nil {
			// Log error but continue scanning other symbols
			scanErrors = append(scanErrors, fmt.Sprintf("%s: %s", symbol, err.Error()))
			continue
		}
		patternsFound++
	}

	// Prepare response
	response := gin.H{
		"message":         fmt.Sprintf("Pattern scan completed for %d symbols", scannedSymbols),
		"symbols_scanned": scannedSymbols,
		"patterns_found":  patternsFound,
	}

	// Include errors if any
	if len(scanErrors) > 0 {
		response["errors"] = scanErrors
		response["message"] = fmt.Sprintf("Pattern scan completed for %d symbols with %d errors", scannedSymbols, len(scanErrors))
	}

	c.JSON(http.StatusOK, response)
}
