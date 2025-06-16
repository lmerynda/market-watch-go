package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"
	"market-watch-go/internal/services"

	"github.com/gin-gonic/gin"
)

// HeadShouldersHandler handles head and shoulders pattern API endpoints
type HeadShouldersHandler struct {
	db        *database.DB
	hsService *services.HeadShouldersDetectionService
}

// NewHeadShouldersHandler creates a new head and shoulders handler
func NewHeadShouldersHandler(db *database.DB, hsService *services.HeadShouldersDetectionService) *HeadShouldersHandler {
	return &HeadShouldersHandler{
		db:        db,
		hsService: hsService,
	}
}

// GetAllPatterns godoc
// @Summary Get all head and shoulders patterns
// @Description Get all head and shoulders patterns with filtering options
// @Tags head-shoulders
// @Accept json
// @Produce json
// @Param symbol query string false "Filter by symbol"
// @Param pattern_type query string false "Filter by pattern type: 'inverse_head_shoulders' or 'head_shoulders'"
// @Param phase query string false "Filter by phase: 'formation', 'breakout', 'target_pursuit', 'completed'"
// @Param is_complete query boolean false "Filter by completion status"
// @Param min_symmetry query number false "Minimum symmetry score"
// @Param limit query int false "Maximum number of patterns to return" default(50)
// @Success 200 {array} models.HeadShouldersPattern
// @Failure 500 {object} models.ErrorResponse
// @Router /api/head-shoulders/patterns [get]
func (h *HeadShouldersHandler) GetAllPatterns(c *gin.Context) {
	// Parse query parameters
	symbol := c.Query("symbol")
	patternType := c.Query("pattern_type")
	phase := c.Query("phase")
	isCompleteStr := c.Query("is_complete")
	minSymmetryStr := c.DefaultQuery("min_symmetry", "0")
	limitStr := c.DefaultQuery("limit", "50")

	var isComplete *bool
	if isCompleteStr != "" {
		completeVal, err := strconv.ParseBool(isCompleteStr)
		if err == nil {
			isComplete = &completeVal
		}
	}

	minSymmetry, err := strconv.ParseFloat(minSymmetryStr, 64)
	if err != nil {
		minSymmetry = 0
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	// Create filter
	filter := &models.PatternFilter{
		Symbol:      symbol,
		PatternType: patternType,
		Phase:       phase,
		IsComplete:  isComplete,
		MinSymmetry: minSymmetry,
		Limit:       limit,
	}

	// Get patterns from database
	patterns, err := h.db.GetHeadShouldersPatterns(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get patterns: " + err.Error(),
		})
		return
	}

	// Ensure we return an empty array instead of null
	if patterns == nil {
		patterns = []*models.HeadShouldersPattern{}
	}

	c.JSON(http.StatusOK, patterns)
}

// GetPatternsBySymbol godoc
// @Summary Get patterns for a specific symbol
// @Description Get all head and shoulders patterns for a specific symbol
// @Tags head-shoulders
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Param phase query string false "Filter by phase"
// @Param is_complete query boolean false "Filter by completion status"
// @Success 200 {array} models.HeadShouldersPattern
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/head-shoulders/patterns/{symbol} [get]
func (h *HeadShouldersHandler) GetPatternsBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	phase := c.Query("phase")
	isCompleteStr := c.Query("is_complete")

	var isComplete *bool
	if isCompleteStr != "" {
		completeVal, err := strconv.ParseBool(isCompleteStr)
		if err == nil {
			isComplete = &completeVal
		}
	}

	filter := &models.PatternFilter{
		Symbol:     symbol,
		Phase:      phase,
		IsComplete: isComplete,
		Limit:      100,
	}

	patterns, err := h.db.GetHeadShouldersPatterns(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get patterns: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, patterns)
}

// GetPatternDetails godoc
// @Summary Get detailed information about a specific pattern
// @Description Get comprehensive details about a head and shoulders pattern including thesis components
// @Tags head-shoulders
// @Accept json
// @Produce json
// @Param id path int true "Pattern ID"
// @Success 200 {object} models.HeadShouldersPattern
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/head-shoulders/patterns/{id}/details [get]
func (h *HeadShouldersHandler) GetPatternDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid pattern ID",
		})
		return
	}

	pattern, err := h.db.GetHeadShouldersPatternByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get pattern: " + err.Error(),
		})
		return
	}

	if pattern == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Pattern not found",
		})
		return
	}

	c.JSON(http.StatusOK, pattern)
}

// GetThesisComponents godoc
// @Summary Get thesis components for a pattern
// @Description Get all thesis components with their completion status for a specific pattern
// @Tags head-shoulders
// @Accept json
// @Produce json
// @Param id path int true "Pattern ID"
// @Success 200 {object} models.HeadShouldersThesis
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/head-shoulders/patterns/{id}/thesis [get]
func (h *HeadShouldersHandler) GetThesisComponents(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid pattern ID",
		})
		return
	}

	pattern, err := h.db.GetHeadShouldersPatternByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get pattern: " + err.Error(),
		})
		return
	}

	if pattern == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Pattern not found",
		})
		return
	}

	c.JSON(http.StatusOK, pattern.ThesisComponents)
}

// DetectPattern godoc
// @Summary Manually trigger pattern detection for a symbol
// @Description Trigger head and shoulders pattern detection for a specific symbol
// @Tags head-shoulders
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol"
// @Success 200 {object} models.HeadShouldersPattern
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/head-shoulders/patterns/{symbol}/detect [post]
func (h *HeadShouldersHandler) DetectPattern(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Symbol parameter is required",
		})
		return
	}

	pattern, err := h.hsService.DetectInverseHeadShoulders(symbol)
	if err != nil {
		// Log the full error for debugging
		log.Printf("Pattern detection failed for %s: %v", symbol, err)

		// Check if it's a "no pattern found" error vs a system error
		if err.Error() == "no valid inverse head and shoulders pattern found" ||
			err.Error() == "insufficient price data for pattern detection" {
			c.JSON(http.StatusOK, gin.H{
				"status":  "no_pattern",
				"message": err.Error(),
				"symbol":  symbol,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to detect pattern: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "pattern_detected",
		"message": "Pattern detected successfully",
		"pattern": pattern,
	})
}

// UpdateThesisComponent godoc
// @Summary Update a specific thesis component
// @Description Manually update the status of a thesis component
// @Tags head-shoulders
// @Accept json
// @Produce json
// @Param id path int true "Pattern ID"
// @Param component path string true "Component name"
// @Param body body map[string]interface{} true "Component update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/head-shoulders/patterns/{id}/thesis/{component} [put]
func (h *HeadShouldersHandler) UpdateThesisComponent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid pattern ID",
		})
		return
	}

	componentName := c.Param("component")
	if componentName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Component name is required",
		})
		return
	}

	var request struct {
		IsCompleted     bool     `json:"is_completed"`
		ConfidenceLevel float64  `json:"confidence_level"`
		Evidence        []string `json:"evidence"`
		Notes           string   `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get the pattern
	pattern, err := h.db.GetHeadShouldersPatternByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get pattern: " + err.Error(),
		})
		return
	}

	if pattern == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Pattern not found",
		})
		return
	}

	// Find and update the component
	components := pattern.ThesisComponents.GetAllComponents()
	var targetComponent *models.ThesisComponent

	for _, component := range components {
		if component.Name == componentName {
			targetComponent = component
			break
		}
	}

	if targetComponent == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Component not found",
		})
		return
	}

	// Update component
	now := time.Now()
	targetComponent.IsCompleted = request.IsCompleted
	targetComponent.ConfidenceLevel = request.ConfidenceLevel
	targetComponent.Evidence = request.Evidence
	targetComponent.LastChecked = now
	targetComponent.AutoDetected = false // Manual update

	if request.IsCompleted && targetComponent.CompletedAt == nil {
		targetComponent.CompletedAt = &now
	}

	// Recalculate completion statistics
	pattern.ThesisComponents.CalculateCompletion()
	pattern.ThesisComponents.UpdatePhase()
	pattern.LastUpdated = now

	// Save updated pattern
	err = h.db.UpdateHeadShouldersPattern(pattern)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update pattern: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"status":     "success",
		"message":    "Component updated successfully",
		"pattern_id": id,
		"component":  componentName,
		"updated_at": now,
		"completion": map[string]interface{}{
			"completed_components": pattern.ThesisComponents.CompletedComponents,
			"total_components":     pattern.ThesisComponents.TotalComponents,
			"completion_percent":   pattern.ThesisComponents.CompletionPercent,
			"current_phase":        pattern.ThesisComponents.CurrentPhase,
		},
	}

	c.JSON(http.StatusOK, response)
}

// MonitorAllPatterns godoc
// @Summary Trigger monitoring of all active patterns
// @Description Manually trigger the monitoring cycle for all active head and shoulders patterns
// @Tags head-shoulders
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /api/head-shoulders/patterns/monitor [post]
func (h *HeadShouldersHandler) MonitorAllPatterns(c *gin.Context) {
	err := h.hsService.MonitorActivePatterns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to monitor patterns: " + err.Error(),
		})
		return
	}

	// Get updated pattern count
	patterns, err := h.db.GetActiveHeadShouldersPatterns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get pattern count: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"status":           "success",
		"message":          "Pattern monitoring completed successfully",
		"patterns_updated": len(patterns),
		"monitored_at":     time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetPatternAlerts godoc
// @Summary Get alerts for a pattern
// @Description Get all alerts generated for a specific head and shoulders pattern
// @Tags head-shoulders
// @Accept json
// @Produce json
// @Param id path int true "Pattern ID"
// @Success 200 {array} models.PatternAlert
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/head-shoulders/patterns/{id}/alerts [get]
func (h *HeadShouldersHandler) GetPatternAlerts(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid pattern ID",
		})
		return
	}

	alerts, err := h.db.GetPatternAlerts(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get alerts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetPatternStatistics godoc
// @Summary Get pattern statistics
// @Description Get comprehensive statistics about all head and shoulders patterns
// @Tags head-shoulders
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /api/head-shoulders/patterns/stats [get]
func (h *HeadShouldersHandler) GetPatternStatistics(c *gin.Context) {
	// Get all patterns for statistics
	allPatterns, err := h.db.GetHeadShouldersPatterns(&models.PatternFilter{Limit: 10000})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get patterns for statistics: " + err.Error(),
		})
		return
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"total_patterns":       len(allPatterns),
		"active_patterns":      0,
		"completed_patterns":   0,
		"inverse_patterns":     0,
		"regular_patterns":     0,
		"formation_phase":      0,
		"breakout_phase":       0,
		"target_pursuit_phase": 0,
		"avg_symmetry":         0.0,
		"avg_completion":       0.0,
		"patterns_by_symbol":   make(map[string]int),
		"last_updated":         time.Now(),
	}

	if len(allPatterns) == 0 {
		c.JSON(http.StatusOK, stats)
		return
	}

	var totalSymmetry, totalCompletion float64
	patternsBySymbol := make(map[string]int)

	for _, pattern := range allPatterns {
		// Count by completion status
		if pattern.IsComplete {
			stats["completed_patterns"] = stats["completed_patterns"].(int) + 1
		} else {
			stats["active_patterns"] = stats["active_patterns"].(int) + 1
		}

		// Count by pattern type
		if pattern.PatternType == models.SetupTypeInverseHeadShoulders {
			stats["inverse_patterns"] = stats["inverse_patterns"].(int) + 1
		} else {
			stats["regular_patterns"] = stats["regular_patterns"].(int) + 1
		}

		// Count by phase
		switch pattern.CurrentPhase {
		case models.PhaseFormation:
			stats["formation_phase"] = stats["formation_phase"].(int) + 1
		case models.PhaseBreakout:
			stats["breakout_phase"] = stats["breakout_phase"].(int) + 1
		case models.PhaseTargetPursuit:
			stats["target_pursuit_phase"] = stats["target_pursuit_phase"].(int) + 1
		}

		// Calculate averages
		totalSymmetry += pattern.Symmetry
		totalCompletion += pattern.ThesisComponents.CompletionPercent

		// Count by symbol
		patternsBySymbol[pattern.Symbol]++
	}

	stats["avg_symmetry"] = totalSymmetry / float64(len(allPatterns))
	stats["avg_completion"] = totalCompletion / float64(len(allPatterns))
	stats["patterns_by_symbol"] = patternsBySymbol

	c.JSON(http.StatusOK, stats)
}

// GetPatternPerformance godoc
// @Summary Get pattern performance analysis
// @Description Get performance analysis for a specific head and shoulders pattern
// @Tags head-shoulders
// @Accept json
// @Produce json
// @Param id path int true "Pattern ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/head-shoulders/patterns/{id}/performance [get]
func (h *HeadShouldersHandler) GetPatternPerformance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid pattern ID",
		})
		return
	}

	pattern, err := h.db.GetHeadShouldersPatternByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get pattern: " + err.Error(),
		})
		return
	}

	if pattern == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Pattern not found",
		})
		return
	}

	// Get current price for performance calculation
	currentPrice := 0.0
	latestPrice, err := h.db.GetLatestPriceData(pattern.Symbol)
	if err == nil && latestPrice != nil {
		currentPrice = latestPrice.Close
	}

	targetPrice := pattern.CalculateTargetPrice()
	performance := map[string]interface{}{
		"pattern_id":    pattern.ID,
		"symbol":        pattern.Symbol,
		"pattern_type":  pattern.PatternType,
		"current_phase": pattern.CurrentPhase,
		"is_complete":   pattern.IsComplete,
		"detected_at":   pattern.DetectedAt,
		"last_updated":  pattern.LastUpdated,

		"price_levels": map[string]float64{
			"head_low":       pattern.HeadLow.Price,
			"neckline_level": pattern.NecklineLevel,
			"target_price":   targetPrice,
			"current_price":  currentPrice,
		},

		"pattern_metrics": map[string]interface{}{
			"pattern_height": pattern.PatternHeight,
			"pattern_width":  pattern.PatternWidth,
			"symmetry_score": pattern.Symmetry,
		},

		"thesis_progress": map[string]interface{}{
			"completed_components": pattern.ThesisComponents.CompletedComponents,
			"total_components":     pattern.ThesisComponents.TotalComponents,
			"completion_percent":   pattern.ThesisComponents.CompletionPercent,
			"current_phase":        pattern.ThesisComponents.CurrentPhase,
		},

		"performance_metrics": map[string]interface{}{
			"breakout_confirmed":  pattern.ThesisComponents.NecklineBreakout.IsCompleted,
			"target_1_reached":    pattern.ThesisComponents.PartialFillT1.IsCompleted,
			"target_2_reached":    pattern.ThesisComponents.PartialFillT2.IsCompleted,
			"full_target_reached": pattern.ThesisComponents.FullTarget.IsCompleted,
		},
	}

	// Calculate performance percentage if breakout occurred
	if pattern.ThesisComponents.NecklineBreakout.IsCompleted && currentPrice > 0 {
		performancePercent := ((currentPrice - pattern.NecklineLevel) / pattern.NecklineLevel) * 100
		targetPercent := ((targetPrice - pattern.NecklineLevel) / pattern.NecklineLevel) * 100

		performance["performance_analysis"] = map[string]interface{}{
			"breakout_performance_percent": performancePercent,
			"target_progress_percent":      (performancePercent / targetPercent) * 100,
			"distance_to_target":           targetPrice - currentPrice,
		}
	}

	c.JSON(http.StatusOK, performance)
}
