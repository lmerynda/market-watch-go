package handlers

import (
	"net/http"
	"strings"
	"time"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"
	"market-watch-go/internal/services"

	"github.com/gin-gonic/gin"
)

type PriceHandler struct {
	db        *database.DB
	collector *services.CollectorService
	polygon   *services.PolygonService
}

// NewPriceHandler creates a new price handler
func NewPriceHandler(db *database.DB, collector *services.CollectorService, polygon *services.PolygonService) *PriceHandler {
	return &PriceHandler{
		db:        db,
		collector: collector,
		polygon:   polygon,
	}
}

// GetPriceData handles GET /api/price/:symbol - returns OHLC price data for TradingView
func (ph *PriceHandler) GetPriceData(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Parse query parameters
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
	filter := &models.PriceDataFilter{
		Symbol: symbol,
		From:   from,
		To:     to,
		Limit:  1000,
	}

	data, err := ph.db.GetPriceData(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve price data",
		})
		return
	}

	// Convert to TradingView candlestick format
	candleData := make([]models.TradingViewCandle, 0)
	for _, pd := range data {
		candle := pd.ToTradingViewCandle()
		candleData = append(candleData, candle)
	}

	// Create response
	response := models.PriceDataResponse{
		Symbol:       symbol,
		Data:         candleData,
		TotalRecords: len(candleData),
		From:         from,
		To:           to,
		Interval:     rangeStr,
	}

	c.JSON(http.StatusOK, response)
}

// GetPriceChartData handles GET /api/price/:symbol/chart - returns TradingView compatible chart data
func (ph *PriceHandler) GetPriceChartData(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Parse query parameters
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
	filter := &models.PriceDataFilter{
		Symbol: symbol,
		From:   from,
		To:     to,
		Limit:  1000,
	}

	data, err := ph.db.GetPriceData(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve price data",
		})
		return
	}

	// Convert to simple format suitable for TradingView Lightweight Charts
	chartData := make([]models.TradingViewCandle, 0)
	for _, pd := range data {
		// Filter trading hours (9:30 AM to 4:00 PM ET)
		hour := pd.Timestamp.Hour()
		minute := pd.Timestamp.Minute()
		dayOfWeek := pd.Timestamp.Weekday()

		// Skip weekends
		if dayOfWeek == time.Saturday || dayOfWeek == time.Sunday {
			continue
		}

		// Skip non-trading hours
		if hour < 9 || (hour == 9 && minute < 30) || hour > 16 {
			continue
		}

		candle := pd.ToTradingViewCandle()
		chartData = append(chartData, candle)
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol": symbol,
		"data":   chartData,
		"range":  rangeStr,
		"from":   from,
		"to":     to,
	})
}

// GetLatestPriceData handles GET /api/price/:symbol/latest
func (ph *PriceHandler) GetLatestPriceData(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Symbol parameter is required",
		})
		return
	}

	symbol = strings.ToUpper(symbol)

	// Get latest data from database
	data, err := ph.db.GetLatestPriceData(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve latest price data",
		})
		return
	}

	if data == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "No price data found for symbol",
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetPriceStats handles GET /api/price/:symbol/stats
func (ph *PriceHandler) GetPriceStats(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Symbol parameter is required",
		})
		return
	}

	symbol = strings.ToUpper(symbol)

	// Get price statistics
	stats, err := ph.db.GetPriceStats(symbol, 1) // Get stats for today
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve price statistics",
		})
		return
	}

	if stats == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "No price statistics available for symbol",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
