package services

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"
)

// HeadShouldersDetectionService handles head and shoulders pattern detection and monitoring
type HeadShouldersDetectionService struct {
	db           *database.Database
	setupService *SetupDetectionService
	taService    *TechnicalAnalysisService
	emailService *EmailService
	config       *models.HeadShouldersConfig
}

// NewHeadShouldersDetectionService creates a new head and shoulders detection service
func NewHeadShouldersDetectionService(db *database.Database, setupService *SetupDetectionService, taService *TechnicalAnalysisService, emailService *EmailService) *HeadShouldersDetectionService {
	// Default configuration
	config := &models.HeadShouldersConfig{
		MinPatternDuration:   72 * time.Hour,  // 3 days minimum
		MaxPatternDuration:   720 * time.Hour, // 30 days maximum
		MinSymmetryScore:     60.0,            // 60% symmetry minimum
		MinVolumeIncrease:    1.2,             // 20% volume increase
		NecklineDeviation:    0.02,            // 2% deviation allowed
		TargetMultiplier:     1.0,             // 1:1 target projection
		MinHeadDepth:         0.05,            // 5% minimum head depth
		MaxShoulderAsymmetry: 0.3,             // 30% max asymmetry between shoulders
	}

	return &HeadShouldersDetectionService{
		db:           db,
		setupService: setupService,
		taService:    taService,
		emailService: emailService,
		config:       config,
	}
}

// DetectInverseHeadShoulders detects inverse head and shoulders patterns for a symbol
func (hsds *HeadShouldersDetectionService) DetectInverseHeadShoulders(symbol string) (*models.HeadShouldersPattern, error) {
	log.Printf("Detecting inverse head and shoulders pattern for %s", symbol)

	// Get price data for analysis (last 6 months)
	endTime := time.Now()
	startTime := endTime.Add(-180 * 24 * time.Hour) // 6 months

	priceData, err := hsds.db.GetPriceDataRange(symbol, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get price data: %w", err)
	}

	if len(priceData) < 50 {
		return nil, fmt.Errorf("insufficient price data for pattern detection")
	}

	// Find potential pattern points
	peaks, troughs := hsds.findPeaksAndTroughs(priceData)

	// Analyze for inverse head and shoulders pattern
	pattern := hsds.analyzeInverseHeadShouldersPattern(symbol, priceData, peaks, troughs)
	if pattern == nil {
		return nil, fmt.Errorf("no valid inverse head and shoulders pattern found")
	}

	// Create associated trading setup
	setup, err := hsds.createTradingSetup(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create trading setup: %w", err)
	}

	// Store the setup first to get ID
	err = hsds.db.InsertTradingSetup(setup)
	if err != nil {
		return nil, fmt.Errorf("failed to store trading setup: %w", err)
	}

	pattern.SetupID = setup.ID

	// Store the pattern
	err = hsds.db.InsertHeadShouldersPattern(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to store pattern: %w", err)
	}

	log.Printf("Successfully detected and stored inverse head and shoulders pattern for %s (ID: %d)", symbol, pattern.ID)
	return pattern, nil
}

// findPeaksAndTroughs identifies local peaks and troughs in price data
func (hsds *HeadShouldersDetectionService) findPeaksAndTroughs(priceData []*models.PriceData) (peaks, troughs []models.PatternPoint) {
	if len(priceData) < 5 {
		return
	}

	windowSize := 5 // Look for peaks/troughs over 5-period window

	for i := windowSize; i < len(priceData)-windowSize; i++ {
		current := priceData[i]
		isPeak := true
		isTrough := true

		// Check if current point is a local peak or trough
		for j := i - windowSize; j <= i+windowSize; j++ {
			if j == i {
				continue
			}

			if priceData[j].High >= current.High {
				isPeak = false
			}
			if priceData[j].Low <= current.Low {
				isTrough = false
			}
		}

		if isPeak {
			peaks = append(peaks, models.PatternPoint{
				Timestamp:   current.Timestamp,
				Price:       current.High,
				Volume:      current.Volume,
				VolumeRatio: hsds.calculateVolumeRatio(priceData, i),
			})
		}

		if isTrough {
			troughs = append(troughs, models.PatternPoint{
				Timestamp:   current.Timestamp,
				Price:       current.Low,
				Volume:      current.Volume,
				VolumeRatio: hsds.calculateVolumeRatio(priceData, i),
			})
		}
	}

	return peaks, troughs
}

// calculateVolumeRatio calculates volume ratio compared to recent average
func (hsds *HeadShouldersDetectionService) calculateVolumeRatio(priceData []*models.PriceData, index int) float64 {
	if index < 20 {
		return 1.0
	}

	// Calculate average volume over last 20 periods
	var totalVolume int64
	for i := index - 19; i < index; i++ {
		totalVolume += priceData[i].Volume
	}
	avgVolume := float64(totalVolume) / 20.0

	if avgVolume == 0 {
		return 1.0
	}

	return float64(priceData[index].Volume) / avgVolume
}

// analyzeInverseHeadShouldersPattern analyzes price data for inverse head and shoulders pattern
func (hsds *HeadShouldersDetectionService) analyzeInverseHeadShouldersPattern(symbol string, priceData []*models.PriceData, peaks, troughs []models.PatternPoint) *models.HeadShouldersPattern {
	if len(troughs) < 3 || len(peaks) < 2 {
		return nil
	}

	// Sort troughs by timestamp
	sort.Slice(troughs, func(i, j int) bool {
		return troughs[i].Timestamp.Before(troughs[j].Timestamp)
	})

	// Sort peaks by timestamp
	sort.Slice(peaks, func(i, j int) bool {
		return peaks[i].Timestamp.Before(peaks[j].Timestamp)
	})

	// Try to find valid inverse head and shoulders pattern
	for i := 0; i < len(troughs)-2; i++ {
		for j := i + 1; j < len(troughs)-1; j++ {
			for k := j + 1; k < len(troughs); k++ {
				leftShoulder := troughs[i]
				head := troughs[j]
				rightShoulder := troughs[k]

				// Check if this could be a valid inverse H&S pattern
				if hsds.isValidInverseHeadShouldersPattern(leftShoulder, head, rightShoulder, peaks) {
					pattern := hsds.buildInverseHeadShouldersPattern(symbol, leftShoulder, head, rightShoulder, peaks)
					if pattern != nil {
						return pattern
					}
				}
			}
		}
	}

	return nil
}

// isValidInverseHeadShouldersPattern checks if the given points form a valid inverse H&S pattern
func (hsds *HeadShouldersDetectionService) isValidInverseHeadShouldersPattern(leftShoulder, head, rightShoulder models.PatternPoint, peaks []models.PatternPoint) bool {
	// Head must be lower than both shoulders (inverse pattern)
	if head.Price >= leftShoulder.Price || head.Price >= rightShoulder.Price {
		return false
	}

	// Check minimum depth requirement
	avgShoulderPrice := (leftShoulder.Price + rightShoulder.Price) / 2
	headDepth := (avgShoulderPrice - head.Price) / avgShoulderPrice
	if headDepth < hsds.config.MinHeadDepth {
		return false
	}

	// Check shoulder symmetry
	shoulderAsymmetry := math.Abs(leftShoulder.Price-rightShoulder.Price) / math.Min(leftShoulder.Price, rightShoulder.Price)
	if shoulderAsymmetry > hsds.config.MaxShoulderAsymmetry {
		return false
	}

	// Check pattern duration
	patternDuration := rightShoulder.Timestamp.Sub(leftShoulder.Timestamp)
	if patternDuration < hsds.config.MinPatternDuration || patternDuration > hsds.config.MaxPatternDuration {
		return false
	}

	// Find peaks between shoulders for neckline calculation
	leftPeak := hsds.findPeakBetween(peaks, leftShoulder.Timestamp, head.Timestamp)
	rightPeak := hsds.findPeakBetween(peaks, head.Timestamp, rightShoulder.Timestamp)

	if leftPeak == nil || rightPeak == nil {
		return false
	}

	return true
}

// findPeakBetween finds a peak between two timestamps
func (hsds *HeadShouldersDetectionService) findPeakBetween(peaks []models.PatternPoint, start, end time.Time) *models.PatternPoint {
	var bestPeak *models.PatternPoint
	var highestPrice float64

	for i := range peaks {
		peak := &peaks[i]
		if peak.Timestamp.After(start) && peak.Timestamp.Before(end) {
			if bestPeak == nil || peak.Price > highestPrice {
				bestPeak = peak
				highestPrice = peak.Price
			}
		}
	}

	return bestPeak
}

// buildInverseHeadShouldersPattern constructs the complete pattern structure
func (hsds *HeadShouldersDetectionService) buildInverseHeadShouldersPattern(symbol string, leftShoulder, head, rightShoulder models.PatternPoint, peaks []models.PatternPoint) *models.HeadShouldersPattern {
	// Find shoulder peaks
	leftShoulderHigh := hsds.findPeakBetween(peaks, leftShoulder.Timestamp.Add(-24*time.Hour), leftShoulder.Timestamp.Add(24*time.Hour))
	rightShoulderHigh := hsds.findPeakBetween(peaks, rightShoulder.Timestamp.Add(-24*time.Hour), rightShoulder.Timestamp.Add(24*time.Hour))

	if leftShoulderHigh == nil || rightShoulderHigh == nil {
		return nil
	}

	// Find head peak (highest point around the head trough)
	headHigh := hsds.findPeakBetween(peaks, head.Timestamp.Add(-48*time.Hour), head.Timestamp.Add(48*time.Hour))
	if headHigh == nil {
		return nil
	}

	// Calculate neckline
	necklineLevel := (leftShoulderHigh.Price + rightShoulderHigh.Price) / 2
	necklineSlope := (rightShoulderHigh.Price - leftShoulderHigh.Price) / float64(rightShoulderHigh.Timestamp.Sub(leftShoulderHigh.Timestamp).Hours())

	// Create pattern
	pattern := &models.HeadShouldersPattern{
		Symbol:            symbol,
		PatternType:       models.SetupTypeInverseHeadShoulders,
		LeftShoulderHigh:  *leftShoulderHigh,
		LeftShoulderLow:   leftShoulder,
		HeadHigh:          *headHigh,
		HeadLow:           head,
		RightShoulderHigh: *rightShoulderHigh,
		RightShoulderLow:  rightShoulder,
		NecklineLevel:     necklineLevel,
		NecklineSlope:     necklineSlope,
		NecklineTouch1:    *leftShoulderHigh,
		NecklineTouch2:    *rightShoulderHigh,
		PatternWidth:      int64(rightShoulder.Timestamp.Sub(leftShoulder.Timestamp).Minutes()),
		PatternHeight:     necklineLevel - head.Price,
		DetectedAt:        time.Now(),
		LastUpdated:       time.Now(),
		CurrentPhase:      models.PhaseFormation,
		IsComplete:        false,
	}

	// Calculate symmetry score
	pattern.Symmetry = pattern.GetSymmetryScore()

	// Initialize thesis components
	pattern.ThesisComponents.InitializeThesis(models.SetupTypeInverseHeadShoulders)

	// Evaluate initial thesis components
	hsds.evaluateInitialThesis(pattern)

	return pattern
}

// evaluateInitialThesis evaluates the initial state of thesis components
func (hsds *HeadShouldersDetectionService) evaluateInitialThesis(pattern *models.HeadShouldersPattern) {
	now := time.Now()

	// Left shoulder formed - always completed for detected patterns
	pattern.ThesisComponents.LeftShoulderFormed.IsCompleted = true
	pattern.ThesisComponents.LeftShoulderFormed.CompletedAt = &pattern.LeftShoulderLow.Timestamp
	pattern.ThesisComponents.LeftShoulderFormed.ConfidenceLevel = 95.0
	pattern.ThesisComponents.LeftShoulderFormed.Evidence = []string{
		fmt.Sprintf("Left shoulder low at $%.2f on %s", pattern.LeftShoulderLow.Price, pattern.LeftShoulderLow.Timestamp.Format("2006-01-02")),
		fmt.Sprintf("Left shoulder high at $%.2f", pattern.LeftShoulderHigh.Price),
	}

	// Head formed - always completed for detected patterns
	pattern.ThesisComponents.HeadFormed.IsCompleted = true
	pattern.ThesisComponents.HeadFormed.CompletedAt = &pattern.HeadLow.Timestamp
	pattern.ThesisComponents.HeadFormed.ConfidenceLevel = 95.0
	pattern.ThesisComponents.HeadFormed.Evidence = []string{
		fmt.Sprintf("Head low at $%.2f on %s", pattern.HeadLow.Price, pattern.HeadLow.Timestamp.Format("2006-01-02")),
		fmt.Sprintf("Head forms lower low than shoulders"),
	}

	// Head lower low - check if head is significantly lower
	leftToHeadDepth := (pattern.LeftShoulderLow.Price - pattern.HeadLow.Price) / pattern.LeftShoulderLow.Price * 100
	if leftToHeadDepth >= 3.0 { // 3% minimum depth
		pattern.ThesisComponents.HeadLowerLow.IsCompleted = true
		pattern.ThesisComponents.HeadLowerLow.CompletedAt = &pattern.HeadLow.Timestamp
		pattern.ThesisComponents.HeadLowerLow.ConfidenceLevel = 90.0
		pattern.ThesisComponents.HeadLowerLow.Evidence = []string{
			fmt.Sprintf("Head is %.1f%% lower than left shoulder", leftToHeadDepth),
		}
	}

	// Right shoulder formed - completed if we have right shoulder data
	if pattern.RightShoulderLow.Price > 0 {
		pattern.ThesisComponents.RightShoulderFormed.IsCompleted = true
		pattern.ThesisComponents.RightShoulderFormed.CompletedAt = &pattern.RightShoulderLow.Timestamp
		pattern.ThesisComponents.RightShoulderFormed.ConfidenceLevel = 95.0
		pattern.ThesisComponents.RightShoulderFormed.Evidence = []string{
			fmt.Sprintf("Right shoulder low at $%.2f on %s", pattern.RightShoulderLow.Price, pattern.RightShoulderLow.Timestamp.Format("2006-01-02")),
			fmt.Sprintf("Right shoulder high at $%.2f", pattern.RightShoulderHigh.Price),
		}

		// Check symmetry
		if pattern.Symmetry >= hsds.config.MinSymmetryScore {
			pattern.ThesisComponents.RightShoulderSymmetry.IsCompleted = true
			pattern.ThesisComponents.RightShoulderSymmetry.CompletedAt = &pattern.RightShoulderLow.Timestamp
			pattern.ThesisComponents.RightShoulderSymmetry.ConfidenceLevel = pattern.Symmetry
			pattern.ThesisComponents.RightShoulderSymmetry.Evidence = []string{
				fmt.Sprintf("Shoulder symmetry score: %.1f%%", pattern.Symmetry),
			}
		}
	}

	// Neckline established
	pattern.ThesisComponents.NecklineEstablished.IsCompleted = true
	pattern.ThesisComponents.NecklineEstablished.CompletedAt = &now
	pattern.ThesisComponents.NecklineEstablished.ConfidenceLevel = 85.0
	pattern.ThesisComponents.NecklineEstablished.Evidence = []string{
		fmt.Sprintf("Neckline level at $%.2f", pattern.NecklineLevel),
		fmt.Sprintf("Neckline slope: %.4f", pattern.NecklineSlope),
	}

	// Target projected
	targetPrice := pattern.CalculateTargetPrice()
	pattern.ThesisComponents.TargetProjected.IsCompleted = true
	pattern.ThesisComponents.TargetProjected.CompletedAt = &now
	pattern.ThesisComponents.TargetProjected.ConfidenceLevel = 80.0
	pattern.ThesisComponents.TargetProjected.Evidence = []string{
		fmt.Sprintf("Target price projected at $%.2f", targetPrice),
		fmt.Sprintf("Pattern height: $%.2f", pattern.PatternHeight),
	}

	// Update completion statistics
	pattern.ThesisComponents.CalculateCompletion()
	pattern.ThesisComponents.UpdatePhase()
}

// createTradingSetup creates a trading setup from the pattern
func (hsds *HeadShouldersDetectionService) createTradingSetup(pattern *models.HeadShouldersPattern) (*models.TradingSetup, error) {
	targetPrice := pattern.CalculateTargetPrice()

	setup := &models.TradingSetup{
		Symbol:       pattern.Symbol,
		SetupType:    pattern.PatternType,
		Direction:    "long", // Inverse H&S is bullish (use 'long' to match schema)
		QualityScore: hsds.calculatePatternQuality(pattern),
		Status:       "active",
		DetectedAt:   pattern.DetectedAt,
		ExpiresAt:    pattern.DetectedAt.Add(24 * time.Hour), // 24 hours to enter
		CurrentPrice: pattern.RightShoulderLow.Price,
		EntryPrice:   pattern.NecklineLevel * 1.002, // Slight premium above neckline
		StopLoss:     pattern.HeadLow.Price * 0.98,  // Below head low
		Target1:      targetPrice * 0.5,             // 50% target
		Target2:      targetPrice * 0.75,            // 75% target
		Target3:      targetPrice,                   // Full target
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Calculate risk/reward ratio manually since methods don't exist
	if setup.EntryPrice > 0 && setup.StopLoss > 0 && setup.Target1 > 0 {
		risk := setup.EntryPrice - setup.StopLoss
		reward := setup.Target1 - setup.EntryPrice
		if risk > 0 {
			setup.RiskRewardRatio = reward / risk
		}
	}

	return setup, nil
}

// calculatePatternQuality calculates the overall quality score for the pattern
func (hsds *HeadShouldersDetectionService) calculatePatternQuality(pattern *models.HeadShouldersPattern) float64 {
	score := 0.0

	// Symmetry score (0-25 points)
	score += (pattern.Symmetry / 100.0) * 25.0

	// Pattern duration score (0-15 points)
	durationHours := float64(pattern.PatternWidth) / 60.0
	idealDuration := 168.0 // 1 week
	durationScore := math.Max(0, 15.0-(math.Abs(durationHours-idealDuration)/idealDuration)*15.0)
	score += durationScore

	// Head depth score (0-20 points)
	headDepth := pattern.PatternHeight / pattern.NecklineLevel * 100
	depthScore := math.Min(20.0, headDepth*4) // Max 20 points at 5% depth
	score += depthScore

	// Volume analysis score (0-15 points)
	volumeScore := 0.0
	if pattern.HeadLow.VolumeRatio > 1.2 {
		volumeScore += 8.0
	}
	if pattern.LeftShoulderLow.VolumeRatio > 1.0 {
		volumeScore += 3.5
	}
	if pattern.RightShoulderLow.VolumeRatio > 1.0 {
		volumeScore += 3.5
	}
	score += volumeScore

	// Completion score (0-25 points)
	completionScore := (pattern.ThesisComponents.CompletionPercent / 100.0) * 25.0
	score += completionScore

	return math.Min(100.0, score)
}

// MonitorActivePatterns monitors all active head and shoulders patterns
func (hsds *HeadShouldersDetectionService) MonitorActivePatterns() error {
	patterns, err := hsds.db.GetActiveHeadShouldersPatterns()
	if err != nil {
		return fmt.Errorf("failed to get active patterns: %w", err)
	}

	log.Printf("Monitoring %d active head and shoulders patterns", len(patterns))

	for _, pattern := range patterns {
		err := hsds.updatePatternThesis(pattern)
		if err != nil {
			log.Printf("Failed to update pattern thesis for %s: %v", pattern.Symbol, err)
			continue
		}

		err = hsds.db.UpdateHeadShouldersPattern(pattern)
		if err != nil {
			log.Printf("Failed to save updated pattern for %s: %v", pattern.Symbol, err)
		}
	}

	return nil
}

// updatePatternThesis updates the thesis components based on current market data
func (hsds *HeadShouldersDetectionService) updatePatternThesis(pattern *models.HeadShouldersPattern) error {
	// Get current price
	currentPrice, err := hsds.getCurrentPrice(pattern.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get current price: %w", err)
	}

	previousState := pattern.ThesisComponents

	// Check breakout conditions
	if pattern.CurrentPhase == models.PhaseFormation || pattern.CurrentPhase == models.PhaseBreakout {
		hsds.checkBreakoutConditions(pattern, currentPrice)
	}

	// Check target achievement
	if pattern.CurrentPhase == models.PhaseTargetPursuit {
		hsds.checkTargetAchievement(pattern, currentPrice)
	}

	// Update completion statistics
	pattern.ThesisComponents.CalculateCompletion()
	pattern.ThesisComponents.UpdatePhase()
	pattern.LastUpdated = time.Now()

	// Send notifications for newly completed components
	hsds.sendNotificationsForNewCompletions(pattern, &previousState)

	return nil
}

// checkBreakoutConditions checks if neckline breakout has occurred
func (hsds *HeadShouldersDetectionService) checkBreakoutConditions(pattern *models.HeadShouldersPattern, currentPrice float64) {
	now := time.Now()

	// Check neckline breakout (for inverse H&S, need price above neckline)
	if currentPrice > pattern.NecklineLevel && !pattern.ThesisComponents.NecklineBreakout.IsCompleted {
		pattern.ThesisComponents.NecklineBreakout.IsCompleted = true
		pattern.ThesisComponents.NecklineBreakout.CompletedAt = &now
		pattern.ThesisComponents.NecklineBreakout.ConfidenceLevel = 85.0
		pattern.ThesisComponents.NecklineBreakout.Evidence = []string{
			fmt.Sprintf("Price broke above neckline: $%.2f > $%.2f", currentPrice, pattern.NecklineLevel),
			fmt.Sprintf("Breakout confirmed at %s", now.Format("2006-01-02 15:04")),
		}

		// Update current phase
		pattern.CurrentPhase = models.PhaseTargetPursuit
	}

	// TODO: Add volume confirmation check for breakout
	// This would require getting current volume data
}

// checkTargetAchievement checks if price targets have been reached
func (hsds *HeadShouldersDetectionService) checkTargetAchievement(pattern *models.HeadShouldersPattern, currentPrice float64) {
	targetPrice := pattern.CalculateTargetPrice()
	now := time.Now()

	// Check partial target 1 (50% of projection)
	target1 := pattern.NecklineLevel + (targetPrice-pattern.NecklineLevel)*0.5
	if currentPrice >= target1 && !pattern.ThesisComponents.PartialFillT1.IsCompleted {
		pattern.ThesisComponents.PartialFillT1.IsCompleted = true
		pattern.ThesisComponents.PartialFillT1.CompletedAt = &now
		pattern.ThesisComponents.PartialFillT1.ConfidenceLevel = 90.0
		pattern.ThesisComponents.PartialFillT1.Evidence = []string{
			fmt.Sprintf("50%% target reached: $%.2f", target1),
		}
	}

	// Check partial target 2 (75% of projection)
	target2 := pattern.NecklineLevel + (targetPrice-pattern.NecklineLevel)*0.75
	if currentPrice >= target2 && !pattern.ThesisComponents.PartialFillT2.IsCompleted {
		pattern.ThesisComponents.PartialFillT2.IsCompleted = true
		pattern.ThesisComponents.PartialFillT2.CompletedAt = &now
		pattern.ThesisComponents.PartialFillT2.ConfidenceLevel = 95.0
		pattern.ThesisComponents.PartialFillT2.Evidence = []string{
			fmt.Sprintf("75%% target reached: $%.2f", target2),
		}
	}

	// Check full target
	if currentPrice >= targetPrice && !pattern.ThesisComponents.FullTarget.IsCompleted {
		pattern.ThesisComponents.FullTarget.IsCompleted = true
		pattern.ThesisComponents.FullTarget.CompletedAt = &now
		pattern.ThesisComponents.FullTarget.ConfidenceLevel = 100.0
		pattern.ThesisComponents.FullTarget.Evidence = []string{
			fmt.Sprintf("Full target reached: $%.2f", targetPrice),
		}

		// Mark pattern as complete
		pattern.IsComplete = true
		pattern.CurrentPhase = models.PhaseCompleted
	}
}

// getCurrentPrice gets the current price for a symbol
func (hsds *HeadShouldersDetectionService) getCurrentPrice(symbol string) (float64, error) {
	latestPrice, err := hsds.db.GetLatestPriceData(symbol)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest price data: %w", err)
	}

	if latestPrice == nil {
		return 0, fmt.Errorf("no price data available for symbol %s", symbol)
	}

	return latestPrice.Close, nil
}

// sendNotificationsForNewCompletions sends notifications for newly completed thesis components
func (hsds *HeadShouldersDetectionService) sendNotificationsForNewCompletions(pattern *models.HeadShouldersPattern, previousState *models.HeadShouldersThesis) {
	components := pattern.ThesisComponents.GetAllComponents()
	previousComponents := previousState.GetAllComponents()

	for i, component := range components {
		if component.IsCompleted && !component.NotificationSent {
			// Check if this is newly completed
			isNewlyCompleted := i < len(previousComponents) && !previousComponents[i].IsCompleted

			if isNewlyCompleted || !component.NotificationSent {
				// Create alert
				alert := &models.PatternAlert{
					PatternID:     pattern.ID,
					Symbol:        pattern.Symbol,
					ComponentName: component.Name,
					AlertType:     "component_completed",
					Message:       fmt.Sprintf("%s completed for %s inverse head and shoulders pattern", component.Name, pattern.Symbol),
					TriggeredAt:   time.Now(),
				}

				err := hsds.db.InsertPatternAlert(alert)
				if err != nil {
					log.Printf("Failed to insert pattern alert: %v", err)
				}

				// Send email notification if email service is available
				if hsds.emailService != nil {
					err = hsds.sendThesisComponentEmail(pattern, component)
					if err != nil {
						log.Printf("Failed to send email notification: %v", err)
					} else {
						alert.EmailSent = true
					}
				}

				component.NotificationSent = true
				alert.NotificationSent = true

				// Update alert status
				if alert.ID > 0 {
					hsds.db.UpdatePatternAlertNotificationStatus(alert.ID, true, alert.EmailSent)
				}
			}
		}
	}
}

// sendThesisComponentEmail sends an email notification for a completed thesis component
func (hsds *HeadShouldersDetectionService) sendThesisComponentEmail(pattern *models.HeadShouldersPattern, component *models.ThesisComponent) error {
	if hsds.emailService == nil {
		return fmt.Errorf("email service not available")
	}

	subject := fmt.Sprintf("📊 %s: %s Component Completed", pattern.Symbol, component.Name)

	// Create email content
	message := fmt.Sprintf(`
Pattern: %s - Inverse Head and Shoulders
Component: %s
Description: %s
Current Phase: %s
Completion: %d/%d components

Evidence:
%s

Target Price: $%.2f
Neckline Level: $%.2f
Completed At: %s
`,
		pattern.Symbol,
		component.Name,
		component.Description,
		pattern.CurrentPhase,
		pattern.ThesisComponents.CompletedComponents,
		pattern.ThesisComponents.TotalComponents,
		fmt.Sprintf("- %s", component.Evidence),
		pattern.CalculateTargetPrice(),
		pattern.NecklineLevel,
		component.CompletedAt.Format("2006-01-02 15:04:05"),
	)

	// Use the existing email service method
	emailMessage := &EmailMessage{
		To:      []string{"admin@example.com"}, // TODO: Get from config
		Subject: subject,
		Body:    message,
		IsHTML:  false,
	}

	return hsds.emailService.SendEmail(emailMessage)
}
