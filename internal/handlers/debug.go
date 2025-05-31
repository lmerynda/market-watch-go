package handlers

import (
	"net/http"

	"market-watch-go/internal/database"

	"github.com/gin-gonic/gin"
)

type DebugHandler struct {
	db *database.DB
}

// NewDebugHandler creates a new debug handler
func NewDebugHandler(db *database.DB) *DebugHandler {
	return &DebugHandler{
		db: db,
	}
}

// GetDataCount handles GET /api/debug/count
func (dh *DebugHandler) GetDataCount(c *gin.Context) {
	// Get total data count
	totalCount, err := dh.db.GetDataCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get data count",
			"details": err.Error(),
		})
		return
	}

	// Get data count by symbol
	countBySymbol, err := dh.db.GetDataCountBySymbol()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get data count by symbol",
			"details": err.Error(),
		})
		return
	}

	// Get all symbols
	symbols, err := dh.db.GetAllSymbols()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get symbols",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_records":   totalCount,
		"symbols":         symbols,
		"count_by_symbol": countBySymbol,
	})
}
