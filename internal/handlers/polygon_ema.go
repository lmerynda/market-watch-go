package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"market-watch-go/internal/services"
)

// PolygonEMAHandler provides endpoints for Polygon EMA queries
func PolygonEMAHandler(emaService *services.PolygonEMAService) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Query("symbol")
		windowStr := c.DefaultQuery("window", "9")
		timespan := c.DefaultQuery("timespan", "day")
		limitStr := c.DefaultQuery("limit", "1")

		if symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
			return
		}

		window, err := strconv.Atoi(windowStr)
		if err != nil || window <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid window parameter"})
			return
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}

		resp, err := emaService.GetEMA(symbol, window, timespan, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch EMA", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// PolygonEMABatchHandler provides batch EMA queries for multiple windows
func PolygonEMABatchHandler(emaService *services.PolygonEMAService) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Query("symbol")
		windowsStr := c.DefaultQuery("windows", "9,50,200")
		timespan := c.DefaultQuery("timespan", "day")
		limitStr := c.DefaultQuery("limit", "1")

		if symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
			return
		}

		windowParts := strings.Split(windowsStr, ",")
		var windows []int
		for _, part := range windowParts {
			w, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil || w <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid windows parameter"})
				return
			}
			windows = append(windows, w)
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}

		resp, err := emaService.GetEMABatch(symbol, windows, timespan, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch EMA batch", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
