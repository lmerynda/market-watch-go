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

// SetupHandler handles setup detection API endpoints
type SetupHandler struct {
	db           *database.Database
	setupService *services.SetupDetectionService
}

// NewSetupHandler creates a new setup handler
func NewSetupHandler(db *database.Database, setupService *services.SetupDetectionService) *SetupHandler {
	return &SetupHandler{
		db:           db,
		setupService: setupService,
	}
}

// DetectSetups godoc
// @Summary Detect trading setups for a symbol
// @Description Analyze market data and detect potential trading setups with scoring
// @Tags setups
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {object} models.SetupDetectionResult
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/setups/{symbol}/detect [post]
func (h *SetupHandler) DetectSetups(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Run setup detection
	result, err := h.setupService.DetectSetups(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to detect setups: " + err.Error(),
		})
		return
	}

	// Store detected setups in database
	for _, setup := range result.SetupsFound {
		err := h.db.InsertTradingSetup(setup)
		if err != nil {
			result.Errors = append(result.Errors, "Failed to store setup: "+err.Error())
		}
	}

	c.JSON(http.StatusOK, result)
}

// GetSetups godoc
// @Summary Get trading setups for a symbol
// @Description Get all trading setups for a specific symbol with filtering options
// @Tags setups
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Param setup_type query string false "Setup type filter"
// @Param direction query string false "Direction filter: 'bullish' or 'bearish'"
// @Param status query string false "Status filter: 'active', 'triggered', 'expired', 'invalidated'"
// @Param min_quality query number false "Minimum quality score"
// @Param confidence query string false "Confidence filter: 'high', 'medium', 'low'"
// @Param is_active query boolean false "Filter for active setups only"
// @Param limit query int false "Maximum number of setups to return"
// @Success 200 {object} models.SetupResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/setups/{symbol} [get]
func (h *SetupHandler) GetSetups(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	// Parse query parameters
	setupType := c.Query("setup_type")
	direction := c.Query("direction")
	status := c.Query("status")
	confidence := c.Query("confidence")
	minQualityStr := c.DefaultQuery("min_quality", "0")
	limitStr := c.DefaultQuery("limit", "50")
	isActiveStr := c.Query("is_active")

	minQuality, err := strconv.ParseFloat(minQualityStr, 64)
	if err != nil {
		minQuality = 0
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	var isActive *bool
	if isActiveStr != "" {
		activeVal, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			isActive = &activeVal
		}
	}

	// Create filter
	filter := &models.SetupFilter{
		Symbol:          symbol,
		SetupType:       setupType,
		Direction:       direction,
		Status:          status,
		Confidence:      confidence,
		MinQualityScore: minQuality,
		IsActive:        isActive,
		Limit:           limit,
	}

	// Get setups from database
	setups, err := h.db.GetTradingSetups(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get setups: " + err.Error(),
		})
		return
	}

	// Get summary
	summary, err := h.db.GetSetupSummary(symbol)
	if err != nil {
		summary = &models.SetupSummary{} // Return empty summary on error
	}

	response := &models.SetupResponse{
		Symbol:  symbol,
		Setups:  setups,
		Summary: summary,
		Status:  "success",
	}

	c.JSON(http.StatusOK, response)
}

// GetSetupByID godoc
// @Summary Get a specific trading setup by ID
// @Description Get detailed information about a specific trading setup
// @Tags setups
// @Accept json
// @Produce json
// @Param id path int true "Setup ID"
// @Success 200 {object} models.TradingSetup
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/setups/id/{id} [get]
func (h *SetupHandler) GetSetupByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid setup ID",
		})
		return
	}

	// Get setup by ID
	setups, err := h.db.GetTradingSetups(&models.SetupFilter{Limit: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get setup: " + err.Error(),
		})
		return
	}

	// Find the specific setup (this is a simplified approach)
	var targetSetup *models.TradingSetup
	for _, setup := range setups {
		if setup.ID == id {
			targetSetup = setup
			break
		}
	}

	if targetSetup == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Setup not found",
		})
		return
	}

	c.JSON(http.StatusOK, targetSetup)
}

// UpdateSetupStatus godoc
// @Summary Update the status of a trading setup
// @Description Update the status of a specific trading setup
// @Tags setups
// @Accept json
// @Produce json
// @Param id path int true "Setup ID"
// @Param body body map[string]string true "Status update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/setups/id/{id}/status [put]
func (h *SetupHandler) UpdateSetupStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid setup ID",
		})
		return
	}

	var request struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"active":      true,
		"triggered":   true,
		"expired":     true,
		"invalidated": true,
	}

	if !validStatuses[request.Status] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid status. Must be one of: active, triggered, expired, invalidated",
		})
		return
	}

	// Get existing setup (simplified approach)
	setups, err := h.db.GetTradingSetups(&models.SetupFilter{Limit: 1000})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get setup: " + err.Error(),
		})
		return
	}

	var targetSetup *models.TradingSetup
	for _, setup := range setups {
		if setup.ID == id {
			targetSetup = setup
			break
		}
	}

	if targetSetup == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Setup not found",
		})
		return
	}

	// Update setup
	targetSetup.UpdateStatus(request.Status)
	if request.Notes != "" {
		targetSetup.Notes = request.Notes
	}

	err = h.db.UpdateTradingSetup(targetSetup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update setup: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"status":     "success",
		"message":    "Setup status updated successfully",
		"setup_id":   id,
		"new_status": request.Status,
		"updated_at": targetSetup.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// GetMultipleSetups godoc
// @Summary Get setups for multiple symbols
// @Description Get trading setups for all watched symbols or a specific list
// @Tags setups
// @Accept json
// @Produce json
// @Param symbols query string false "Comma-separated list of symbols (optional)"
// @Param setup_type query string false "Setup type filter"
// @Param direction query string false "Direction filter"
// @Param min_quality query number false "Minimum quality score" default(60)
// @Param is_active query boolean false "Filter for active setups only" default(true)
// @Success 200 {object} map[string]models.SetupResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/setups [get]
func (h *SetupHandler) GetMultipleSetups(c *gin.Context) {
	symbolsParam := c.Query("symbols")
	setupType := c.Query("setup_type")
	direction := c.Query("direction")
	minQualityStr := c.DefaultQuery("min_quality", "60")
	isActiveStr := c.DefaultQuery("is_active", "true")

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

	minQuality, err := strconv.ParseFloat(minQualityStr, 64)
	if err != nil {
		minQuality = 60
	}

	var isActive *bool
	if isActiveStr != "" {
		activeVal, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			isActive = &activeVal
		}
	}

	results := make(map[string]*models.SetupResponse)

	for _, symbol := range symbols {
		filter := &models.SetupFilter{
			Symbol:          symbol,
			SetupType:       setupType,
			Direction:       direction,
			MinQualityScore: minQuality,
			IsActive:        isActive,
			Limit:           20,
		}

		setups, err := h.db.GetTradingSetups(filter)
		if err != nil {
			results[symbol] = &models.SetupResponse{
				Symbol:  symbol,
				Status:  "error",
				Message: err.Error(),
			}
			continue
		}

		summary, _ := h.db.GetSetupSummary(symbol)
		if summary == nil {
			summary = &models.SetupSummary{}
		}

		results[symbol] = &models.SetupResponse{
			Symbol:  symbol,
			Setups:  setups,
			Summary: summary,
			Status:  "success",
		}
	}

	c.JSON(http.StatusOK, results)
}

// GetSetupSummary godoc
// @Summary Get setup summary statistics
// @Description Get comprehensive statistics about trading setups for a symbol
// @Tags setups
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {object} models.SetupSummary
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/setups/{symbol}/summary [get]
func (h *SetupHandler) GetSetupSummary(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	summary, err := h.db.GetSetupSummary(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get setup summary: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetHighQualitySetups godoc
// @Summary Get high quality setups
// @Description Get only high quality trading setups (score >= 80) across all symbols
// @Tags setups
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of setups to return" default(20)
// @Success 200 {array} models.TradingSetup
// @Failure 500 {object} models.ErrorResponse
// @Router /api/setups/high-quality [get]
func (h *SetupHandler) GetHighQualitySetups(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	filter := &models.SetupFilter{
		MinQualityScore: 80.0,
		IsActive:        utils.BoolPtr(true),
		Limit:           limit,
	}

	setups, err := h.db.GetTradingSetups(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get high quality setups: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, setups)
}

// ExpireOldSetups godoc
// @Summary Expire old setups
// @Description Mark old setups as expired based on their expiration time
// @Tags setups
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /api/setups/expire [post]
func (h *SetupHandler) ExpireOldSetups(c *gin.Context) {
	expiredCount, err := h.db.ExpireOldSetups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to expire old setups: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"status":        "success",
		"message":       "Old setups expired successfully",
		"expired_count": expiredCount,
		"expired_at":    time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// CleanupOldSetups godoc
// @Summary Cleanup old setup data
// @Description Remove old setup data based on retention policy
// @Tags setups
// @Accept json
// @Produce json
// @Param days query int false "Days to retain setup data" default(90)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/setups/cleanup [post]
func (h *SetupHandler) CleanupOldSetups(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "90")

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid days parameter",
		})
		return
	}

	deletedCount, err := h.db.CleanupOldSetupData(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to cleanup old setup data: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"status":         "success",
		"message":        "Old setup data cleaned up successfully",
		"deleted_count":  deletedCount,
		"retention_days": days,
		"cleanup_time":   time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetSetupChecklist godoc
// @Summary Get setup checklist
// @Description Get the detailed checklist for a specific setup
// @Tags setups
// @Accept json
// @Produce json
// @Param id path int true "Setup ID"
// @Success 200 {object} models.SetupChecklist
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/setups/id/{id}/checklist [get]
func (h *SetupHandler) GetSetupChecklist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid setup ID",
		})
		return
	}

	checklist, err := h.db.GetSetupChecklist(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get setup checklist: " + err.Error(),
		})
		return
	}

	if checklist == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Checklist not found for setup",
		})
		return
	}

	c.JSON(http.StatusOK, checklist)
}

// GetSetupsStats godoc
// @Summary Get setup statistics
// @Description Get comprehensive statistics about all trading setups
// @Tags setups
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /api/setups/stats [get]
func (h *SetupHandler) GetSetupsStats(c *gin.Context) {
	// Get all setups for stats calculation
	allSetups, err := h.db.GetTradingSetups(&models.SetupFilter{Limit: 10000})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get setups for statistics: " + err.Error(),
		})
		return
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"total_setups":      len(allSetups),
		"active_setups":     0,
		"high_quality":      0,
		"medium_quality":    0,
		"low_quality":       0,
		"bullish_setups":    0,
		"bearish_setups":    0,
		"avg_quality_score": 0.0,
		"avg_risk_reward":   0.0,
		"setup_types":       make(map[string]int),
		"by_symbol":         make(map[string]int),
		"last_updated":      time.Now(),
	}

	if len(allSetups) == 0 {
		c.JSON(http.StatusOK, stats)
		return
	}

	var totalQuality, totalRiskReward float64
	setupTypes := make(map[string]int)
	bySymbol := make(map[string]int)

	for _, setup := range allSetups {
		// Quality categories
		if setup.QualityScore >= 80 {
			stats["high_quality"] = stats["high_quality"].(int) + 1
		} else if setup.QualityScore >= 60 {
			stats["medium_quality"] = stats["medium_quality"].(int) + 1
		} else {
			stats["low_quality"] = stats["low_quality"].(int) + 1
		}

		// Direction
		if setup.Direction == "bullish" {
			stats["bullish_setups"] = stats["bullish_setups"].(int) + 1
		} else {
			stats["bearish_setups"] = stats["bearish_setups"].(int) + 1
		}

		// Active status
		if setup.IsActive() {
			stats["active_setups"] = stats["active_setups"].(int) + 1
		}

		// Averages
		totalQuality += setup.QualityScore
		totalRiskReward += setup.RiskRewardRatio

		// Groupings
		setupTypes[setup.SetupType]++
		bySymbol[setup.Symbol]++
	}

	stats["avg_quality_score"] = totalQuality / float64(len(allSetups))
	stats["avg_risk_reward"] = totalRiskReward / float64(len(allSetups))
	stats["setup_types"] = setupTypes
	stats["by_symbol"] = bySymbol

	c.JSON(http.StatusOK, stats)
}
