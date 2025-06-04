package services

import (
	"fmt"
	"math"
	"time"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"
)

// SetupDetectionService handles trading setup detection and scoring
type SetupDetectionService struct {
	db        *database.DB
	taService *TechnicalAnalysisService
	srService *SupportResistanceService
	config    *models.SetupScoringConfig
}

// NewSetupDetectionService creates a new setup detection service
func NewSetupDetectionService(db *database.DB, taService *TechnicalAnalysisService, srService *SupportResistanceService) *SetupDetectionService {
	// Default configuration
	config := &models.SetupScoringConfig{
		HighQualityThreshold:   80.0,
		MediumQualityThreshold: 60.0,
		LowQualityThreshold:    40.0,
		PriceActionWeight:      25.0,
		VolumeWeight:           25.0,
		TechnicalWeight:        25.0,
		RiskRewardWeight:       25.0,
		MinBouncePercent:       2.0,
		MinTimeAtLevelMinutes:  30,
		MaxLevelAgeDays:        60,
		MinRiskRewardRatio:     1.5,
		MaxRiskPercent:         2.0,
		SetupExpirationHours:   24,
	}

	return &SetupDetectionService{
		db:        db,
		taService: taService,
		srService: srService,
		config:    config,
	}
}

// DetectSetups performs comprehensive setup detection for a symbol
func (sds *SetupDetectionService) DetectSetups(symbol string) (*models.SetupDetectionResult, error) {
	now := time.Now()

	result := &models.SetupDetectionResult{
		Symbol:        symbol,
		DetectionTime: now,
		SetupsFound:   []*models.TradingSetup{},
		ActiveSetups:  []*models.TradingSetup{},
		ExpiredSetups: []*models.TradingSetup{},
		Errors:        []string{},
	}

	// Get current market data
	currentPrice, err := sds.getCurrentPrice(symbol)
	if err != nil {
		result.Errors = append(result.Errors, "Failed to get current price: "+err.Error())
		return result, nil
	}

	// Get S/R analysis
	srAnalysis, err := sds.srService.DetectSupportResistanceLevels(symbol)
	if err != nil {
		result.Errors = append(result.Errors, "Failed to get S/R analysis: "+err.Error())
		return result, nil
	}

	// Get technical indicators
	indicators, err := sds.taService.GetIndicators(symbol)
	if err != nil {
		result.Errors = append(result.Errors, "Failed to get technical indicators: "+err.Error())
		return result, nil
	}

	// Detect different types of setups
	supportBounceSetups := sds.detectSupportBounceSetups(symbol, currentPrice, srAnalysis, indicators)
	resistanceBounceSetups := sds.detectResistanceBounceSetups(symbol, currentPrice, srAnalysis, indicators)
	breakoutSetups := sds.detectBreakoutSetups(symbol, currentPrice, srAnalysis, indicators)

	// Combine all detected setups
	allSetups := append(supportBounceSetups, resistanceBounceSetups...)
	allSetups = append(allSetups, breakoutSetups...)

	// Score and validate setups
	for _, setup := range allSetups {
		err := sds.scoreSetup(setup, srAnalysis, indicators)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to score setup %s: %v", setup.SetupType, err))
			continue
		}

		// Only include setups that meet minimum criteria
		if setup.QualityScore >= sds.config.LowQualityThreshold {
			result.SetupsFound = append(result.SetupsFound, setup)

			if setup.IsActive() {
				result.ActiveSetups = append(result.ActiveSetups, setup)
			}
		}
	}

	// Build summary
	result.Summary = sds.buildSetupSummary(result.SetupsFound)

	return result, nil
}

// detectSupportBounceSetups identifies potential support bounce setups
func (sds *SetupDetectionService) detectSupportBounceSetups(symbol string, currentPrice float64, srAnalysis *models.SRAnalysisResult, indicators *models.TechnicalIndicators) []*models.TradingSetup {
	var setups []*models.TradingSetup

	for _, supportLevel := range srAnalysis.SupportLevels {
		// Check if price is near support level (within 2%)
		distancePercent := math.Abs(currentPrice-supportLevel.Level) / supportLevel.Level * 100
		if distancePercent <= 2.0 && currentPrice >= supportLevel.Level*0.99 {

			// Create support bounce setup
			setup := &models.TradingSetup{
				Symbol:       symbol,
				SetupType:    "support_bounce",
				Direction:    "bullish",
				Status:       "active",
				DetectedAt:   time.Now(),
				ExpiresAt:    time.Now().Add(time.Duration(sds.config.SetupExpirationHours) * time.Hour),
				CurrentPrice: currentPrice,
				EntryPrice:   supportLevel.Level * 1.002, // Slight premium above support
				StopLoss:     supportLevel.Level * 0.995, // Just below support
				IsManual:     false,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			// Set targets based on resistance levels
			sds.setTargetLevels(setup, srAnalysis)

			// Associate the support level
			setup.SupportLevel = supportLevel
			setup.KeyLevel = supportLevel

			setups = append(setups, setup)
		}
	}

	return setups
}

// detectResistanceBounceSetups identifies potential resistance bounce setups
func (sds *SetupDetectionService) detectResistanceBounceSetups(symbol string, currentPrice float64, srAnalysis *models.SRAnalysisResult, indicators *models.TechnicalIndicators) []*models.TradingSetup {
	var setups []*models.TradingSetup

	for _, resistanceLevel := range srAnalysis.ResistanceLevels {
		// Check if price is near resistance level (within 2%)
		distancePercent := math.Abs(currentPrice-resistanceLevel.Level) / resistanceLevel.Level * 100
		if distancePercent <= 2.0 && currentPrice <= resistanceLevel.Level*1.01 {

			// Create resistance bounce setup
			setup := &models.TradingSetup{
				Symbol:       symbol,
				SetupType:    "resistance_bounce",
				Direction:    "bearish",
				Status:       "active",
				DetectedAt:   time.Now(),
				ExpiresAt:    time.Now().Add(time.Duration(sds.config.SetupExpirationHours) * time.Hour),
				CurrentPrice: currentPrice,
				EntryPrice:   resistanceLevel.Level * 0.998, // Slight discount below resistance
				StopLoss:     resistanceLevel.Level * 1.005, // Just above resistance
				IsManual:     false,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			// Set targets based on support levels
			sds.setTargetLevels(setup, srAnalysis)

			// Associate the resistance level
			setup.ResistanceLevel = resistanceLevel
			setup.KeyLevel = resistanceLevel

			setups = append(setups, setup)
		}
	}

	return setups
}

// detectBreakoutSetups identifies potential breakout setups
func (sds *SetupDetectionService) detectBreakoutSetups(symbol string, currentPrice float64, srAnalysis *models.SRAnalysisResult, indicators *models.TechnicalIndicators) []*models.TradingSetup {
	var setups []*models.TradingSetup

	// Check for resistance breakouts
	for _, resistanceLevel := range srAnalysis.ResistanceLevels {
		if currentPrice > resistanceLevel.Level*1.002 { // Price broke above resistance
			setup := &models.TradingSetup{
				Symbol:       symbol,
				SetupType:    "resistance_breakout",
				Direction:    "bullish",
				Status:       "active",
				DetectedAt:   time.Now(),
				ExpiresAt:    time.Now().Add(time.Duration(sds.config.SetupExpirationHours) * time.Hour),
				CurrentPrice: currentPrice,
				EntryPrice:   currentPrice,
				StopLoss:     resistanceLevel.Level * 0.998, // Below broken resistance
				IsManual:     false,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			sds.setTargetLevels(setup, srAnalysis)
			setup.ResistanceLevel = resistanceLevel
			setup.KeyLevel = resistanceLevel

			setups = append(setups, setup)
		}
	}

	// Check for support breakdowns
	for _, supportLevel := range srAnalysis.SupportLevels {
		if currentPrice < supportLevel.Level*0.998 { // Price broke below support
			setup := &models.TradingSetup{
				Symbol:       symbol,
				SetupType:    "support_breakdown",
				Direction:    "bearish",
				Status:       "active",
				DetectedAt:   time.Now(),
				ExpiresAt:    time.Now().Add(time.Duration(sds.config.SetupExpirationHours) * time.Hour),
				CurrentPrice: currentPrice,
				EntryPrice:   currentPrice,
				StopLoss:     supportLevel.Level * 1.002, // Above broken support
				IsManual:     false,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			sds.setTargetLevels(setup, srAnalysis)
			setup.SupportLevel = supportLevel
			setup.KeyLevel = supportLevel

			setups = append(setups, setup)
		}
	}

	return setups
}

// setTargetLevels sets target levels for a setup based on S/R analysis
func (sds *SetupDetectionService) setTargetLevels(setup *models.TradingSetup, srAnalysis *models.SRAnalysisResult) {
	if setup.Direction == "bullish" {
		// Find resistance levels above current price for targets
		var targets []float64
		for _, level := range srAnalysis.ResistanceLevels {
			if level.Level > setup.EntryPrice {
				targets = append(targets, level.Level)
			}
		}

		// Sort targets by distance from entry
		for i := 0; i < len(targets)-1; i++ {
			for j := i + 1; j < len(targets); j++ {
				if targets[i] > targets[j] {
					targets[i], targets[j] = targets[j], targets[i]
				}
			}
		}

		// Set up to 3 targets
		if len(targets) >= 1 {
			setup.Target1 = targets[0]
		}
		if len(targets) >= 2 {
			setup.Target2 = targets[1]
		}
		if len(targets) >= 3 {
			setup.Target3 = targets[2]
		}

		// If no resistance levels found, use percentage targets
		if len(targets) == 0 {
			setup.Target1 = setup.EntryPrice * 1.03 // 3% target
			setup.Target2 = setup.EntryPrice * 1.06 // 6% target
			setup.Target3 = setup.EntryPrice * 1.09 // 9% target
		}

	} else { // bearish
		// Find support levels below current price for targets
		var targets []float64
		for _, level := range srAnalysis.SupportLevels {
			if level.Level < setup.EntryPrice {
				targets = append(targets, level.Level)
			}
		}

		// Sort targets by distance from entry (closest first)
		for i := 0; i < len(targets)-1; i++ {
			for j := i + 1; j < len(targets); j++ {
				if targets[i] < targets[j] {
					targets[i], targets[j] = targets[j], targets[i]
				}
			}
		}

		// Set up to 3 targets
		if len(targets) >= 1 {
			setup.Target1 = targets[0]
		}
		if len(targets) >= 2 {
			setup.Target2 = targets[1]
		}
		if len(targets) >= 3 {
			setup.Target3 = targets[2]
		}

		// If no support levels found, use percentage targets
		if len(targets) == 0 {
			setup.Target1 = setup.EntryPrice * 0.97 // 3% target
			setup.Target2 = setup.EntryPrice * 0.94 // 6% target
			setup.Target3 = setup.EntryPrice * 0.91 // 9% target
		}
	}

	// Calculate risk/reward metrics
	setup.RiskAmount = setup.GetRiskAmount()
	setup.RewardPotential = setup.GetRewardPotential()
	setup.RiskRewardRatio = setup.CalculateRiskRewardRatio()
}

// scoreSetup calculates the quality score for a setup
func (sds *SetupDetectionService) scoreSetup(setup *models.TradingSetup, srAnalysis *models.SRAnalysisResult, indicators *models.TechnicalIndicators) error {
	// Create and populate checklist
	checklist := sds.createSetupChecklist(setup, srAnalysis, indicators)
	setup.Checklist = checklist

	// Calculate component scores
	setup.PriceActionScore = checklist.GetPriceActionScore()
	setup.VolumeScore = checklist.GetVolumeScore()
	setup.TechnicalScore = checklist.GetTechnicalScore()
	setup.RiskRewardScore = checklist.GetRiskManagementScore()

	// Calculate weighted overall score
	setup.QualityScore = (setup.PriceActionScore*sds.config.PriceActionWeight +
		setup.VolumeScore*sds.config.VolumeWeight +
		setup.TechnicalScore*sds.config.TechnicalWeight +
		setup.RiskRewardScore*sds.config.RiskRewardWeight) / 100.0

	// Set confidence level
	setup.Confidence = setup.GetConfidenceLevel()

	return nil
}

// createSetupChecklist creates and evaluates a checklist for a setup
func (sds *SetupDetectionService) createSetupChecklist(setup *models.TradingSetup, srAnalysis *models.SRAnalysisResult, indicators *models.TechnicalIndicators) *models.SetupChecklist {
	checklist := &models.SetupChecklist{
		SetupID: setup.ID,
	}

	// Initialize checklist items with default values
	sds.initializeChecklistItems(checklist)

	// Evaluate each criterion
	sds.evaluatePriceActionCriteria(checklist, setup, srAnalysis)
	sds.evaluateVolumeCriteria(checklist, setup, indicators)
	sds.evaluateTechnicalCriteria(checklist, setup, indicators)
	sds.evaluateRiskManagementCriteria(checklist, setup)

	// Calculate final scores
	checklist.CalculateScore()

	return checklist
}

// initializeChecklistItems sets up the checklist items with default configurations
func (sds *SetupDetectionService) initializeChecklistItems(checklist *models.SetupChecklist) {
	// Price Action Criteria
	checklist.MinLevelTouches = models.ChecklistItem{Name: "Minimum Level Touches", MaxPoints: 5, IsRequired: true}
	checklist.BounceStrength = models.ChecklistItem{Name: "Bounce Strength", MaxPoints: 5, IsRequired: true}
	checklist.TimeAtLevel = models.ChecklistItem{Name: "Time at Level", MaxPoints: 5, IsRequired: false}
	checklist.RejectionCandle = models.ChecklistItem{Name: "Rejection Candle", MaxPoints: 5, IsRequired: false}
	checklist.LevelDuration = models.ChecklistItem{Name: "Level Duration", MaxPoints: 5, IsRequired: false}

	// Volume Criteria
	checklist.VolumeSpike = models.ChecklistItem{Name: "Volume Spike", MaxPoints: 5, IsRequired: true}
	checklist.VolumeConfirmation = models.ChecklistItem{Name: "Volume Confirmation", MaxPoints: 5, IsRequired: false}
	checklist.ApproachVolume = models.ChecklistItem{Name: "Approach Volume", MaxPoints: 5, IsRequired: false}
	checklist.VWAPRelationship = models.ChecklistItem{Name: "VWAP Relationship", MaxPoints: 5, IsRequired: false}
	checklist.RelativeVolume = models.ChecklistItem{Name: "Relative Volume", MaxPoints: 5, IsRequired: false}

	// Technical Indicators
	checklist.RSICondition = models.ChecklistItem{Name: "RSI Condition", MaxPoints: 5, IsRequired: false}
	checklist.MovingAverage = models.ChecklistItem{Name: "Moving Average", MaxPoints: 5, IsRequired: false}
	checklist.MACDSignal = models.ChecklistItem{Name: "MACD Signal", MaxPoints: 5, IsRequired: false}
	checklist.MomentumDivergence = models.ChecklistItem{Name: "Momentum Divergence", MaxPoints: 5, IsRequired: false}
	checklist.BollingerBands = models.ChecklistItem{Name: "Bollinger Bands", MaxPoints: 5, IsRequired: false}

	// Risk Management
	checklist.StopLossDefined = models.ChecklistItem{Name: "Stop Loss Defined", MaxPoints: 5, IsRequired: true}
	checklist.RiskRewardRatio = models.ChecklistItem{Name: "Risk/Reward Ratio", MaxPoints: 5, IsRequired: true}
	checklist.PositionSize = models.ChecklistItem{Name: "Position Size", MaxPoints: 5, IsRequired: false}
	checklist.EntryPrecision = models.ChecklistItem{Name: "Entry Precision", MaxPoints: 5, IsRequired: false}
	checklist.ExitStrategy = models.ChecklistItem{Name: "Exit Strategy", MaxPoints: 5, IsRequired: false}
}

// evaluatePriceActionCriteria evaluates price action related criteria
func (sds *SetupDetectionService) evaluatePriceActionCriteria(checklist *models.SetupChecklist, setup *models.TradingSetup, srAnalysis *models.SRAnalysisResult) {
	keyLevel := setup.KeyLevel
	if keyLevel == nil {
		return
	}

	// Minimum Level Touches
	if keyLevel.Touches >= 3 {
		checklist.MinLevelTouches.IsCompleted = true
		checklist.MinLevelTouches.Points = 5
		checklist.MinLevelTouches.AutoDetected = true
	}

	// Bounce Strength
	if keyLevel.AvgBouncePercent >= sds.config.MinBouncePercent {
		checklist.BounceStrength.IsCompleted = true
		checklist.BounceStrength.Points = 5
		checklist.BounceStrength.AutoDetected = true
	}

	// Level Duration - fix type conversion
	levelAge := keyLevel.GetAge()
	if levelAge >= 5 && levelAge <= float64(sds.config.MaxLevelAgeDays) {
		checklist.LevelDuration.IsCompleted = true
		checklist.LevelDuration.Points = 5
		checklist.LevelDuration.AutoDetected = true
	}

	checklist.MinLevelTouches.LastChecked = time.Now()
	checklist.BounceStrength.LastChecked = time.Now()
	checklist.LevelDuration.LastChecked = time.Now()
}

// evaluateVolumeCriteria evaluates volume related criteria
func (sds *SetupDetectionService) evaluateVolumeCriteria(checklist *models.SetupChecklist, setup *models.TradingSetup, indicators *models.TechnicalIndicators) {
	// Volume Spike
	if indicators.VolumeRatio >= 1.5 {
		checklist.VolumeSpike.IsCompleted = true
		checklist.VolumeSpike.Points = 5
		checklist.VolumeSpike.AutoDetected = true
	}

	// VWAP Relationship
	if setup.Direction == "bullish" && setup.CurrentPrice > indicators.VWAP {
		checklist.VWAPRelationship.IsCompleted = true
		checklist.VWAPRelationship.Points = 5
		checklist.VWAPRelationship.AutoDetected = true
	} else if setup.Direction == "bearish" && setup.CurrentPrice < indicators.VWAP {
		checklist.VWAPRelationship.IsCompleted = true
		checklist.VWAPRelationship.Points = 5
		checklist.VWAPRelationship.AutoDetected = true
	}

	// Relative Volume
	if indicators.VolumeRatio >= 1.2 {
		checklist.RelativeVolume.IsCompleted = true
		checklist.RelativeVolume.Points = 5
		checklist.RelativeVolume.AutoDetected = true
	}

	checklist.VolumeSpike.LastChecked = time.Now()
	checklist.VWAPRelationship.LastChecked = time.Now()
	checklist.RelativeVolume.LastChecked = time.Now()
}

// evaluateTechnicalCriteria evaluates technical indicator criteria
func (sds *SetupDetectionService) evaluateTechnicalCriteria(checklist *models.SetupChecklist, setup *models.TradingSetup, indicators *models.TechnicalIndicators) {
	// RSI Condition
	if setup.Direction == "bullish" && indicators.RSI14 <= 40 {
		checklist.RSICondition.IsCompleted = true
		checklist.RSICondition.Points = 5
		checklist.RSICondition.AutoDetected = true
	} else if setup.Direction == "bearish" && indicators.RSI14 >= 60 {
		checklist.RSICondition.IsCompleted = true
		checklist.RSICondition.Points = 5
		checklist.RSICondition.AutoDetected = true
	}

	// Moving Average
	if setup.Direction == "bullish" && setup.CurrentPrice > indicators.SMA20 {
		checklist.MovingAverage.IsCompleted = true
		checklist.MovingAverage.Points = 5
		checklist.MovingAverage.AutoDetected = true
	} else if setup.Direction == "bearish" && setup.CurrentPrice < indicators.SMA20 {
		checklist.MovingAverage.IsCompleted = true
		checklist.MovingAverage.Points = 5
		checklist.MovingAverage.AutoDetected = true
	}

	// MACD Signal
	if setup.Direction == "bullish" && indicators.MACDHistogram > 0 {
		checklist.MACDSignal.IsCompleted = true
		checklist.MACDSignal.Points = 5
		checklist.MACDSignal.AutoDetected = true
	} else if setup.Direction == "bearish" && indicators.MACDHistogram < 0 {
		checklist.MACDSignal.IsCompleted = true
		checklist.MACDSignal.Points = 5
		checklist.MACDSignal.AutoDetected = true
	}

	checklist.RSICondition.LastChecked = time.Now()
	checklist.MovingAverage.LastChecked = time.Now()
	checklist.MACDSignal.LastChecked = time.Now()
}

// evaluateRiskManagementCriteria evaluates risk management criteria
func (sds *SetupDetectionService) evaluateRiskManagementCriteria(checklist *models.SetupChecklist, setup *models.TradingSetup) {
	// Stop Loss Defined
	if setup.StopLoss > 0 {
		checklist.StopLossDefined.IsCompleted = true
		checklist.StopLossDefined.Points = 5
		checklist.StopLossDefined.AutoDetected = true
	}

	// Risk/Reward Ratio
	if setup.RiskRewardRatio >= sds.config.MinRiskRewardRatio {
		checklist.RiskRewardRatio.IsCompleted = true
		checklist.RiskRewardRatio.Points = 5
		checklist.RiskRewardRatio.AutoDetected = true
	}

	// Exit Strategy
	if setup.Target1 > 0 {
		checklist.ExitStrategy.IsCompleted = true
		checklist.ExitStrategy.Points = 5
		checklist.ExitStrategy.AutoDetected = true
	}

	checklist.StopLossDefined.LastChecked = time.Now()
	checklist.RiskRewardRatio.LastChecked = time.Now()
	checklist.ExitStrategy.LastChecked = time.Now()
}

// buildSetupSummary creates a summary of detected setups
func (sds *SetupDetectionService) buildSetupSummary(setups []*models.TradingSetup) *models.SetupSummary {
	summary := &models.SetupSummary{
		TotalSetups:   len(setups),
		LastDetection: time.Now(),
	}

	if len(setups) == 0 {
		return summary
	}

	var totalScore, totalRiskReward float64
	var bestSetup *models.TradingSetup

	for _, setup := range setups {
		// Count by status
		if setup.IsActive() {
			summary.ActiveCount++
		}

		// Count by quality
		if setup.QualityScore >= sds.config.HighQualityThreshold {
			summary.HighQualityCount++
		} else if setup.QualityScore >= sds.config.MediumQualityThreshold {
			summary.MediumQualityCount++
		} else {
			summary.LowQualityCount++
		}

		// Count by direction
		if setup.Direction == "bullish" {
			summary.BullishCount++
		} else {
			summary.BearishCount++
		}

		// Track averages
		totalScore += setup.QualityScore
		totalRiskReward += setup.RiskRewardRatio

		// Find best setup
		if bestSetup == nil || setup.QualityScore > bestSetup.QualityScore {
			bestSetup = setup
		}
	}

	summary.AvgQualityScore = totalScore / float64(len(setups))
	summary.AvgRiskReward = totalRiskReward / float64(len(setups))
	summary.BestSetup = bestSetup

	return summary
}

// getCurrentPrice gets the current price for a symbol
func (sds *SetupDetectionService) getCurrentPrice(symbol string) (float64, error) {
	latestPrice, err := sds.db.GetLatestPriceData(symbol)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest price data: %w", err)
	}

	if latestPrice == nil {
		return 0, fmt.Errorf("no price data available for symbol %s", symbol)
	}

	return latestPrice.Close, nil
}

// UpdateSetupStatus updates the status of setups based on current market conditions
func (sds *SetupDetectionService) UpdateSetupStatus(symbol string) error {
	// This would be implemented to check existing setups and update their status
	// based on current price movements, trigger conditions, etc.
	return nil
}

// GetActiveSetups retrieves all active setups for a symbol
func (sds *SetupDetectionService) GetActiveSetups(symbol string) ([]*models.TradingSetup, error) {
	// This would query the database for active setups
	// For now, return empty slice
	return []*models.TradingSetup{}, nil
}
