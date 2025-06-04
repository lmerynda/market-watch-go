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

// SupportResistanceHandler handles S/R detection API endpoints
type SupportResistanceHandler struct {
	db        *database.DB
	srService *services.SupportResistanceService
}

// NewSupportResistanceHandler creates a new S/R handler
func NewSupportResistanceHandler(db *database.DB, srService *services.SupportResistanceService) *SupportResistanceHandler {
	return &SupportResistanceHandler{
		db:        db,
		srService: srService,
	}
}

// GetSupportResistanceLevels godoc
// @Summary Get support and resistance levels for a symbol
// @Description Get all active support and resistance levels for a specific symbol
// @Tags support-resistance
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Param level_type query string false "Level type: 'support', 'resistance', or 'both'" default(both)
// @Param min_strength query number false "Minimum strength score"
// @Param min_touches query int false "Minimum number of touches"
// @Param limit query int false "Maximum number of levels to return"
// @Success 200 {object} models.SRResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/support-resistance/{symbol}/levels [get]
func (h *SupportResistanceHandler) GetSupportResistanceLevels(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Parse query parameters
	levelType := c.DefaultQuery("level_type", "both")
	minStrengthStr := c.DefaultQuery("min_strength", "0")
	minTouchesStr := c.DefaultQuery("min_touches", "2")
	limitStr := c.DefaultQuery("limit", "50")

	minStrength, err := strconv.ParseFloat(minStrengthStr, 64)
	if err != nil {
		minStrength = 0
	}

	minTouches, err := strconv.Atoi(minTouchesStr)
	if err != nil {
		minTouches = 2
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	// Create filter
	filter := &models.SRDetectionFilter{
		Symbol:      symbol,
		LevelType:   levelType,
		MinStrength: minStrength,
		MinTouches:  minTouches,
		IsActive:    utils.BoolPtr(true),
		Limit:       limit,
	}

	// Get levels from database
	levels, err := h.db.GetSupportResistanceLevels(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get S/R levels: " + err.Error(),
		})
		return
	}

	// Get summary
	summary, err := h.db.GetSRLevelSummary(symbol)
	if err != nil {
		summary = &models.SRLevelSummary{} // Return empty summary on error
	}

	response := &models.SRResponse{
		Symbol:  symbol,
		Levels:  levels,
		Summary: summary,
		Status:  "success",
	}

	c.JSON(http.StatusOK, response)
}

// DetectSupportResistance godoc
// @Summary Detect support and resistance levels for a symbol
// @Description Analyze price data and detect new support and resistance levels
// @Tags support-resistance
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {object} models.SRAnalysisResult
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/support-resistance/{symbol}/detect [post]
func (h *SupportResistanceHandler) DetectSupportResistance(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Run S/R detection
	result, err := h.srService.DetectSupportResistanceLevels(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to detect S/R levels: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetNearestLevels godoc
// @Summary Get nearest support and resistance levels
// @Description Get the nearest support level below and resistance level above current price
// @Tags support-resistance
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/support-resistance/{symbol}/nearest [get]
func (h *SupportResistanceHandler) GetNearestLevels(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Get current price from latest price data
	latestPrice, err := h.db.GetLatestPriceData(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get current price: " + err.Error(),
		})
		return
	}

	if latestPrice == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "No price data available for symbol",
		})
		return
	}

	// Get nearest support and resistance levels
	nearestSupport, nearestResistance, err := h.db.GetNearestSupportResistance(symbol, latestPrice.Close)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get nearest levels: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"symbol":             symbol,
		"current_price":      latestPrice.Close,
		"nearest_support":    nearestSupport,
		"nearest_resistance": nearestResistance,
		"timestamp":          time.Now(),
	}

	// Calculate distances if levels exist
	if nearestSupport != nil {
		distance := ((latestPrice.Close - nearestSupport.Level) / latestPrice.Close) * 100
		response["support_distance_percent"] = distance
	}

	if nearestResistance != nil {
		distance := ((nearestResistance.Level - latestPrice.Close) / latestPrice.Close) * 100
		response["resistance_distance_percent"] = distance
	}

	c.JSON(http.StatusOK, response)
}

// GetLevelTouches godoc
// @Summary Get recent touches of support/resistance levels
// @Description Get recent instances where price touched support or resistance levels
// @Tags support-resistance
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Param hours query int false "Hours to look back for touches" default(24)
// @Param limit query int false "Maximum number of touches to return" default(20)
// @Success 200 {array} models.SRLevelTouch
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/support-resistance/{symbol}/touches [get]
func (h *SupportResistanceHandler) GetLevelTouches(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	hoursStr := c.DefaultQuery("hours", "24")
	limitStr := c.DefaultQuery("limit", "20")

	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 24
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	touches, err := h.db.GetRecentSRLevelTouches(symbol, hours, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get level touches: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, touches)
}

// GetPivotPoints godoc
// @Summary Get pivot points for a symbol
// @Description Get detected pivot highs and lows for analysis
// @Tags support-resistance
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Param from query string false "Start date (RFC3339 format)"
// @Param to query string false "End date (RFC3339 format)"
// @Param pivot_type query string false "Pivot type: 'high', 'low', or empty for both"
// @Success 200 {array} models.PivotPoint
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/support-resistance/{symbol}/pivots [get]
func (h *SupportResistanceHandler) GetPivotPoints(c *gin.Context) {
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
	pivotType := c.Query("pivot_type")

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

	pivots, err := h.db.GetPivotPoints(symbol, from, to, pivotType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get pivot points: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, pivots)
}

// GetLevelSummary godoc
// @Summary Get summary statistics for S/R levels
// @Description Get comprehensive statistics about support and resistance levels
// @Tags support-resistance
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {object} models.SRLevelSummary
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/support-resistance/{symbol}/summary [get]
func (h *SupportResistanceHandler) GetLevelSummary(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	summary, err := h.db.GetSRLevelSummary(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get level summary: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetMultipleLevels godoc
// @Summary Get S/R levels for multiple symbols
// @Description Get support and resistance levels for all watched symbols or a specific list
// @Tags support-resistance
// @Accept json
// @Produce json
// @Param symbols query string false "Comma-separated list of symbols (optional)"
// @Param level_type query string false "Level type filter" default(both)
// @Param min_strength query number false "Minimum strength score" default(20)
// @Success 200 {object} map[string]models.SRResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/support-resistance/levels [get]
func (h *SupportResistanceHandler) GetMultipleLevels(c *gin.Context) {
	symbolsParam := c.Query("symbols")
	levelType := c.DefaultQuery("level_type", "both")
	minStrengthStr := c.DefaultQuery("min_strength", "20")

	var symbols []string
	var err error

	if symbolsParam != "" {
		symbols = utils.ParseSymbols(symbolsParam)
	} else {
		symbols, err = h.db.GetWatchedSymbols()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to get watched symbols: " + err.Error(),
			})
			return
		}
	}

	minStrength, err := strconv.ParseFloat(minStrengthStr, 64)
	if err != nil {
		minStrength = 20
	}

	results := make(map[string]*models.SRResponse)

	for _, symbol := range symbols {
		filter := &models.SRDetectionFilter{
			Symbol:      symbol,
			LevelType:   levelType,
			MinStrength: minStrength,
			IsActive:    utils.BoolPtr(true),
			Limit:       20,
		}

		levels, err := h.db.GetSupportResistanceLevels(filter)
		if err != nil {
			results[symbol] = &models.SRResponse{
				Symbol:  symbol,
				Status:  "error",
				Message: err.Error(),
			}
			continue
		}

		summary, _ := h.db.GetSRLevelSummary(symbol)
		if summary == nil {
			summary = &models.SRLevelSummary{}
		}

		results[symbol] = &models.SRResponse{
			Symbol:  symbol,
			Levels:  levels,
			Summary: summary,
			Status:  "success",
		}
	}

	c.JSON(http.StatusOK, results)
}

// CleanupOldData godoc
// @Summary Cleanup old S/R data
// @Description Remove old support/resistance data based on retention policy
// @Tags support-resistance
// @Accept json
// @Produce json
// @Param days query int false "Days to retain data" default(90)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/support-resistance/cleanup [post]
func (h *SupportResistanceHandler) CleanupOldData(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "90")

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid days parameter",
		})
		return
	}

	deletedCount, err := h.db.CleanupOldSRData(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to cleanup old data: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"status":         "success",
		"message":        "Old S/R data cleaned up successfully",
		"deleted_count":  deletedCount,
		"retention_days": days,
		"cleanup_time":   time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// DeactivateOldLevels godoc
// @Summary Deactivate old S/R levels
// @Description Deactivate support/resistance levels that haven't been touched recently
// @Tags support-resistance
// @Accept json
// @Produce json
// @Param max_age_hours query int false "Maximum age in hours for active levels" default(720)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/support-resistance/deactivate [post]
func (h *SupportResistanceHandler) DeactivateOldLevels(c *gin.Context) {
	maxAgeStr := c.DefaultQuery("max_age_hours", "720") // 30 days default

	maxAgeHours, err := strconv.Atoi(maxAgeStr)
	if err != nil || maxAgeHours <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid max_age_hours parameter",
		})
		return
	}

	deactivatedCount, err := h.db.DeactivateOldSRLevels(maxAgeHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to deactivate old levels: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"status":            "success",
		"message":           "Old S/R levels deactivated successfully",
		"deactivated_count": deactivatedCount,
		"max_age_hours":     maxAgeHours,
		"deactivation_time": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}
