package handlers

import (
	"net/http"
	"strconv"
	"time"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"
	"market-watch-go/internal/services"
	"market-watch-go/internal/utils"

	"github.com/gin-gonic/gin"
)

// TechnicalAnalysisHandler handles technical analysis API endpoints
type TechnicalAnalysisHandler struct {
	db        *database.DB
	taService *services.TechnicalAnalysisService
}

// NewTechnicalAnalysisHandler creates a new technical analysis handler
func NewTechnicalAnalysisHandler(db *database.DB, taService *services.TechnicalAnalysisService) *TechnicalAnalysisHandler {
	return &TechnicalAnalysisHandler{
		db:        db,
		taService: taService,
	}
}

// GetIndicators godoc
// @Summary Get technical indicators for a symbol
// @Description Get all technical indicators (RSI, MACD, MA, etc.) for a specific symbol
// @Tags technical-analysis
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {object} models.TechnicalIndicatorsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/technical-analysis/{symbol}/indicators [get]
func (h *TechnicalAnalysisHandler) GetIndicators(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	indicators, err := h.taService.GetIndicators(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to calculate technical indicators: " + err.Error(),
		})
		return
	}

	response := &models.TechnicalIndicatorsResponse{
		Symbol:     symbol,
		Indicators: indicators,
		Status:     "success",
	}

	c.JSON(http.StatusOK, response)
}

// GetIndicatorsSummary godoc
// @Summary Get technical indicators summary for a symbol
// @Description Get a comprehensive summary of technical indicators with sentiment analysis
// @Tags technical-analysis
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {object} models.IndicatorSummary
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/technical-analysis/{symbol}/summary [get]
func (h *TechnicalAnalysisHandler) GetIndicatorsSummary(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	summary, err := h.taService.GetIndicatorsSummary(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get indicators summary: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetMultipleIndicators godoc
// @Summary Get technical indicators for multiple symbols
// @Description Get technical indicators for all watched symbols or a specific list
// @Tags technical-analysis
// @Accept json
// @Produce json
// @Param symbols query string false "Comma-separated list of symbols (optional)"
// @Success 200 {object} map[string]models.TechnicalIndicatorsResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/technical-analysis/indicators [get]
func (h *TechnicalAnalysisHandler) GetMultipleIndicators(c *gin.Context) {
	symbolsParam := c.Query("symbols")
	var symbols []string
	var err error

	if symbolsParam != "" {
		// Parse comma-separated symbols from query parameter
		symbols = utils.ParseSymbols(symbolsParam)
	} else {
		// Get all watched symbols
		symbols, err = h.db.GetWatchedSymbols()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to get watched symbols: " + err.Error(),
			})
			return
		}
	}

	results := make(map[string]*models.TechnicalIndicatorsResponse)

	for _, symbol := range symbols {
		indicators, err := h.taService.GetIndicators(symbol)
		if err != nil {
			results[symbol] = &models.TechnicalIndicatorsResponse{
				Symbol:  symbol,
				Status:  "error",
				Message: err.Error(),
			}
			continue
		}

		results[symbol] = &models.TechnicalIndicatorsResponse{
			Symbol:     symbol,
			Indicators: indicators,
			Status:     "success",
		}
	}

	c.JSON(http.StatusOK, results)
}

// GetHistoricalIndicators godoc
// @Summary Get historical technical indicators
// @Description Get technical indicators for a symbol within a time range
// @Tags technical-analysis
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Param from query string false "Start date (RFC3339 format)"
// @Param to query string false "End date (RFC3339 format)"
// @Param limit query int false "Maximum number of records"
// @Param offset query int false "Number of records to skip"
// @Success 200 {array} models.TechnicalIndicators
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/technical-analysis/{symbol}/historical [get]
func (h *TechnicalAnalysisHandler) GetHistoricalIndicators(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Parse query parameters
	fromStr := c.DefaultQuery("from", time.Now().AddDate(0, 0, -7).Format(time.RFC3339))
	toStr := c.DefaultQuery("to", time.Now().Format(time.RFC3339))
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid 'from' date format. Use RFC3339 format",
		})
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid 'to' date format. Use RFC3339 format",
		})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	filter := &models.IndicatorFilter{
		Symbol: symbol,
		From:   from,
		To:     to,
		Limit:  limit,
		Offset: offset,
	}

	indicators, err := h.db.GetTechnicalIndicators(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get historical indicators: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, indicators)
}

// UpdateIndicators godoc
// @Summary Update technical indicators for a symbol
// @Description Force update of technical indicators for a specific symbol
// @Tags technical-analysis
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {object} models.TechnicalIndicatorsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/technical-analysis/{symbol}/update [post]
func (h *TechnicalAnalysisHandler) UpdateIndicators(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Force update by invalidating cache and recalculating
	err := h.taService.UpdateIndicatorsForSymbol(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update indicators: " + err.Error(),
		})
		return
	}

	// Get fresh indicators
	indicators, err := h.taService.GetIndicators(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get updated indicators: " + err.Error(),
		})
		return
	}

	// Store indicators in database
	err = h.db.InsertTechnicalIndicators(indicators)
	if err != nil {
		// Log error but don't fail the request
		c.Header("X-Warning", "Failed to store indicators in database: "+err.Error())
	}

	response := &models.TechnicalIndicatorsResponse{
		Symbol:     symbol,
		Indicators: indicators,
		Status:     "success",
		Message:    "Indicators updated successfully",
	}

	c.JSON(http.StatusOK, response)
}

// GetCacheStatus godoc
// @Summary Get technical analysis cache status
// @Description Get information about cached time series data
// @Tags technical-analysis
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/technical-analysis/cache/status [get]
func (h *TechnicalAnalysisHandler) GetCacheStatus(c *gin.Context) {
	status := h.taService.GetCacheStatus()
	c.JSON(http.StatusOK, status)
}

// ClearCache godoc
// @Summary Clear technical analysis cache
// @Description Clear expired cache entries for time series data
// @Tags technical-analysis
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/technical-analysis/cache/clear [post]
func (h *TechnicalAnalysisHandler) ClearCache(c *gin.Context) {
	h.taService.ClearExpiredCache()
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cache cleared successfully",
	})
}

// InvalidateSymbolCache godoc
// @Summary Invalidate cache for a specific symbol
// @Description Remove cached time series data for a specific symbol
// @Tags technical-analysis
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Router /api/technical-analysis/{symbol}/cache/invalidate [post]
func (h *TechnicalAnalysisHandler) InvalidateSymbolCache(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	h.taService.InvalidateCache(symbol)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cache invalidated for symbol: " + symbol,
	})
}

// CheckAlerts godoc
// @Summary Check indicator alerts for a symbol
// @Description Check if any indicator-based alerts should be triggered for a symbol
// @Tags technical-analysis
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {array} models.IndicatorAlert
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/technical-analysis/{symbol}/alerts [get]
func (h *TechnicalAnalysisHandler) CheckAlerts(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Default alert thresholds - these could be configurable
	thresholds := &models.IndicatorThresholds{
		RSIOversold:   30,
		RSIOverbought: 70,
		VolumeSpike:   200, // 200% of average
		MACDBullish:   true,
		MACDBearish:   true,
		BBOverbought:  true,
		BBOversold:    true,
	}

	alerts, err := h.taService.CheckIndicatorAlerts(symbol, thresholds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to check alerts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetActiveAlerts godoc
// @Summary Get active alerts for a symbol
// @Description Get all active indicator alerts from the database for a symbol
// @Tags technical-analysis
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {array} models.IndicatorAlert
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/technical-analysis/{symbol}/alerts/active [get]
func (h *TechnicalAnalysisHandler) GetActiveAlerts(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	alerts, err := h.db.GetActiveIndicatorAlerts(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get active alerts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetStats godoc
// @Summary Get technical indicators statistics
// @Description Get statistics about technical indicators data
// @Tags technical-analysis
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /api/technical-analysis/stats [get]
func (h *TechnicalAnalysisHandler) GetStats(c *gin.Context) {
	stats, err := h.db.GetTechnicalIndicatorsStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get statistics: " + err.Error(),
		})
		return
	}

	// Add alert statistics
	alertStats, err := h.db.GetIndicatorAlertsStats()
	if err != nil {
		c.Header("X-Warning", "Failed to get alert statistics: "+err.Error())
	} else {
		stats["alerts"] = alertStats
	}

	c.JSON(http.StatusOK, stats)
}
