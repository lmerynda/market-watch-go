package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"
)

// SupportResistanceService handles support and resistance level detection
type SupportResistanceService struct {
	db        *database.Database
	taService *TechnicalAnalysisService
	config    *models.SRDetectionConfig
}

// NewSupportResistanceService creates a new S/R detection service
func NewSupportResistanceService(db *database.Database, taService *TechnicalAnalysisService) *SupportResistanceService {
	// Default configuration
	config := &models.SRDetectionConfig{
		MinTouches:                3,
		LookbackDays:              30,
		StrengthCalculation:       "weighted",
		MinLevelDistancePercent:   1.0,
		LevelPenetrationTolerance: 0.5,
		PivotStrength:             5,
		VolumeConfirmationRatio:   1.5,
		MaxLevelAge:               60,
		MinBouncePercent:          2.0,
	}

	return &SupportResistanceService{
		db:        db,
		taService: taService,
		config:    config,
	}
}

// DetectSupportResistanceLevels performs comprehensive S/R detection for a symbol
func (srs *SupportResistanceService) DetectSupportResistanceLevels(symbol string) (*models.SRAnalysisResult, error) {
	now := time.Now()

	// Get price data for analysis
	priceData, err := srs.db.GetPriceData(&models.PriceDataFilter{
		Symbol: symbol,
		From:   now.AddDate(0, 0, -srs.config.LookbackDays),
		To:     now,
		Limit:  10000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get price data: %w", err)
	}

	if len(priceData) == 0 {
		return &models.SRAnalysisResult{
			Symbol:       symbol,
			AnalysisTime: now,
		}, nil
	}

	// Step 1: Detect pivot points
	pivots, err := srs.detectPivotPoints(symbol, priceData)
	if err != nil {
		return nil, fmt.Errorf("failed to detect pivot points: %w", err)
	}

	// Step 2: Cluster pivots into potential S/R levels
	potentialLevels := srs.clusterPivotsIntoLevels(pivots)

	// Step 3: Validate and score levels
	validatedLevels := srs.validateAndScoreLevels(symbol, potentialLevels, priceData)

	// Step 4: Update database with new/updated levels
	err = srs.updateSRLevelsInDatabase(symbol, validatedLevels)
	if err != nil {
		return nil, fmt.Errorf("failed to update S/R levels in database: %w", err)
	}

	// Step 5: Get comprehensive analysis result
	result, err := srs.buildAnalysisResult(symbol, validatedLevels, priceData)
	if err != nil {
		return nil, fmt.Errorf("failed to build analysis result: %w", err)
	}

	return result, nil
}

// detectPivotPoints identifies pivot highs and lows in price data
func (srs *SupportResistanceService) detectPivotPoints(symbol string, priceData []*models.PriceData) ([]*models.PivotPoint, error) {
	if len(priceData) < srs.config.PivotStrength*2+1 {
		return []*models.PivotPoint{}, nil
	}

	var pivots []*models.PivotPoint
	strength := srs.config.PivotStrength

	// Iterate through price data looking for pivot points
	for i := strength; i < len(priceData)-strength; i++ {
		current := priceData[i]

		// Check for pivot high
		isPivotHigh := true
		for j := i - strength; j <= i+strength; j++ {
			if j != i && priceData[j].High >= current.High {
				isPivotHigh = false
				break
			}
		}

		if isPivotHigh {
			pivot := &models.PivotPoint{
				Symbol:    symbol,
				Timestamp: current.Timestamp,
				Price:     current.High,
				PivotType: "high",
				Strength:  strength,
				Volume:    current.Volume,
				Confirmed: true,
				CreatedAt: time.Now(),
			}
			pivots = append(pivots, pivot)
		}

		// Check for pivot low
		isPivotLow := true
		for j := i - strength; j <= i+strength; j++ {
			if j != i && priceData[j].Low <= current.Low {
				isPivotLow = false
				break
			}
		}

		if isPivotLow {
			pivot := &models.PivotPoint{
				Symbol:    symbol,
				Timestamp: current.Timestamp,
				Price:     current.Low,
				PivotType: "low",
				Strength:  strength,
				Volume:    current.Volume,
				Confirmed: true,
				CreatedAt: time.Now(),
			}
			pivots = append(pivots, pivot)
		}
	}

	return pivots, nil
}

// clusterPivotsIntoLevels groups nearby pivots into potential S/R levels
func (srs *SupportResistanceService) clusterPivotsIntoLevels(pivots []*models.PivotPoint) []*models.SupportResistanceLevel {
	if len(pivots) == 0 {
		return []*models.SupportResistanceLevel{}
	}

	var levels []*models.SupportResistanceLevel
	minDistance := srs.config.MinLevelDistancePercent / 100.0

	// Sort pivots by price
	sort.Slice(pivots, func(i, j int) bool {
		return pivots[i].Price < pivots[j].Price
	})

	// Group nearby pivots
	i := 0
	for i < len(pivots) {
		cluster := []*models.PivotPoint{pivots[i]}
		j := i + 1

		// Find all pivots within distance threshold
		for j < len(pivots) {
			distance := math.Abs(pivots[j].Price-pivots[i].Price) / pivots[i].Price
			if distance <= minDistance {
				cluster = append(cluster, pivots[j])
				j++
			} else {
				break
			}
		}

		// Create level if cluster has enough touches
		if len(cluster) >= srs.config.MinTouches {
			level := srs.createLevelFromCluster(cluster)
			if level != nil {
				levels = append(levels, level)
			}
		}

		i = j
	}

	return levels
}

// createLevelFromCluster creates an S/R level from a cluster of pivots
func (srs *SupportResistanceService) createLevelFromCluster(cluster []*models.PivotPoint) *models.SupportResistanceLevel {
	if len(cluster) == 0 {
		return nil
	}

	// Calculate average price level
	var totalPrice, totalVolume float64
	var supportCount, resistanceCount int
	var firstTouch, lastTouch time.Time

	for i, pivot := range cluster {
		totalPrice += pivot.Price
		totalVolume += float64(pivot.Volume)

		if pivot.PivotType == "low" {
			supportCount++
		} else {
			resistanceCount++
		}

		if i == 0 || pivot.Timestamp.Before(firstTouch) {
			firstTouch = pivot.Timestamp
		}
		if i == 0 || pivot.Timestamp.After(lastTouch) {
			lastTouch = pivot.Timestamp
		}
	}

	avgPrice := totalPrice / float64(len(cluster))
	avgVolume := totalVolume / float64(len(cluster))

	// Determine level type based on majority
	levelType := "support"
	if resistanceCount > supportCount {
		levelType = "resistance"
	}

	now := time.Now()
	level := &models.SupportResistanceLevel{
		Symbol:          cluster[0].Symbol,
		Level:           avgPrice,
		LevelType:       levelType,
		Touches:         len(cluster),
		FirstTouch:      firstTouch,
		LastTouch:       lastTouch,
		AvgVolume:       avgVolume,
		TimeframeOrigin: "1m",
		IsActive:        true,
		LastValidated:   now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return level
}

// validateAndScoreLevels validates potential levels and calculates strength scores
func (srs *SupportResistanceService) validateAndScoreLevels(symbol string, potentialLevels []*models.SupportResistanceLevel, priceData []*models.PriceData) []*models.SupportResistanceLevel {
	var validatedLevels []*models.SupportResistanceLevel

	for _, level := range potentialLevels {
		// Calculate detailed metrics for the level
		err := srs.calculateLevelMetrics(level, priceData)
		if err != nil {
			continue // Skip levels with calculation errors
		}

		// Calculate strength score
		level.Strength = srs.calculateStrengthScore(level)

		// Only keep levels that meet minimum criteria
		if level.Strength >= 20 && level.Touches >= srs.config.MinTouches {
			validatedLevels = append(validatedLevels, level)
		}
	}

	return validatedLevels
}

// calculateLevelMetrics calculates detailed metrics for an S/R level
func (srs *SupportResistanceService) calculateLevelMetrics(level *models.SupportResistanceLevel, priceData []*models.PriceData) error {
	tolerance := srs.config.LevelPenetrationTolerance / 100.0
	var bounces []float64
	var touchVolumes []float64

	// Analyze price interactions with the level
	for _, candle := range priceData {
		distanceFromLevel := math.Abs(candle.Close-level.Level) / level.Level

		// Check if price touched the level
		if distanceFromLevel <= tolerance {
			touchVolumes = append(touchVolumes, float64(candle.Volume))

			// Look for bounce after touch (next few candles)
			bouncePercent := srs.calculateBounceFromLevel(level, candle, priceData)
			if bouncePercent > 0 {
				bounces = append(bounces, bouncePercent)
			}
		}
	}

	// Calculate bounce statistics
	if len(bounces) > 0 {
		level.AvgBouncePercent = average(bounces)
		level.MaxBouncePercent = maximum(bounces)
	}

	// Volume confirmation
	if len(touchVolumes) > 0 {
		avgTouchVolume := average(touchVolumes)

		// Get average volume for the period
		var allVolumes []float64
		for _, candle := range priceData {
			allVolumes = append(allVolumes, float64(candle.Volume))
		}
		avgVolume := average(allVolumes)

		level.VolumeConfirmed = avgTouchVolume >= avgVolume*srs.config.VolumeConfirmationRatio
		level.AvgVolume = avgTouchVolume
	}

	return nil
}

// calculateBounceFromLevel calculates bounce percentage after touching a level
func (srs *SupportResistanceService) calculateBounceFromLevel(level *models.SupportResistanceLevel, touchCandle *models.PriceData, priceData []*models.PriceData) float64 {
	// Find the touch candle index
	touchIndex := -1
	for i, candle := range priceData {
		if candle.Timestamp.Equal(touchCandle.Timestamp) {
			touchIndex = i
			break
		}
	}

	if touchIndex == -1 || touchIndex >= len(priceData)-5 {
		return 0
	}

	// Look at next 5 candles for bounce
	var maxBounce float64
	touchPrice := level.Level

	for i := touchIndex + 1; i < touchIndex+6 && i < len(priceData); i++ {
		candle := priceData[i]
		var bouncePercent float64

		if level.LevelType == "support" {
			// For support, bounce is upward movement
			bouncePercent = ((candle.High - touchPrice) / touchPrice) * 100
		} else {
			// For resistance, bounce is downward movement
			bouncePercent = ((touchPrice - candle.Low) / touchPrice) * 100
		}

		if bouncePercent > maxBounce {
			maxBounce = bouncePercent
		}
	}

	return maxBounce
}

// calculateStrengthScore calculates the strength score for an S/R level
func (srs *SupportResistanceService) calculateStrengthScore(level *models.SupportResistanceLevel) float64 {
	score := 0.0

	// Touch frequency (0-30 points)
	touchScore := float64(level.Touches) * 5.0
	if touchScore > 30 {
		touchScore = 30
	}
	score += touchScore

	// Bounce strength (0-25 points)
	if level.AvgBouncePercent > 0 {
		bounceScore := level.AvgBouncePercent * 2.5
		if bounceScore > 25 {
			bounceScore = 25
		}
		score += bounceScore
	}

	// Volume confirmation (0-20 points)
	if level.VolumeConfirmed {
		score += 20
	}

	// Age factor (0-15 points) - newer levels get higher scores
	ageInDays := time.Since(level.FirstTouch).Hours() / 24
	ageScore := 15 - (ageInDays * 0.25)
	if ageScore < 0 {
		ageScore = 0
	}
	score += ageScore

	// Recency factor (0-10 points) - recently touched levels get bonus
	hoursSinceLastTouch := time.Since(level.LastTouch).Hours()
	if hoursSinceLastTouch <= 24 {
		score += 10
	} else if hoursSinceLastTouch <= 168 { // 1 week
		score += 5
	}

	return score
}

// updateSRLevelsInDatabase updates or inserts S/R levels in the database
func (srs *SupportResistanceService) updateSRLevelsInDatabase(symbol string, levels []*models.SupportResistanceLevel) error {
	// Get existing levels from database
	existingLevels, err := srs.db.GetSupportResistanceLevels(&models.SRDetectionFilter{
		Symbol:   symbol,
		IsActive: boolPtr(true),
		Limit:    1000,
	})
	if err != nil {
		return fmt.Errorf("failed to get existing S/R levels: %w", err)
	}

	// Create maps for easier lookup
	existingMap := make(map[string]*models.SupportResistanceLevel)
	for _, existing := range existingLevels {
		key := fmt.Sprintf("%.4f_%s", existing.Level, existing.LevelType)
		existingMap[key] = existing
	}

	tolerance := srs.config.LevelPenetrationTolerance / 100.0

	// Process new levels
	for _, newLevel := range levels {
		// Check if similar level already exists
		var matchingLevel *models.SupportResistanceLevel
		for _, existing := range existingLevels {
			if existing.LevelType == newLevel.LevelType {
				distance := math.Abs(existing.Level-newLevel.Level) / existing.Level
				if distance <= tolerance {
					matchingLevel = existing
					break
				}
			}
		}

		if matchingLevel != nil {
			// Update existing level
			matchingLevel.Touches = newLevel.Touches
			matchingLevel.LastTouch = newLevel.LastTouch
			matchingLevel.Strength = newLevel.Strength
			matchingLevel.AvgBouncePercent = newLevel.AvgBouncePercent
			matchingLevel.MaxBouncePercent = newLevel.MaxBouncePercent
			matchingLevel.VolumeConfirmed = newLevel.VolumeConfirmed
			matchingLevel.AvgVolume = newLevel.AvgVolume
			matchingLevel.LastValidated = time.Now()

			err := srs.db.UpdateSupportResistanceLevel(matchingLevel)
			if err != nil {
				return fmt.Errorf("failed to update S/R level: %w", err)
			}
		} else {
			// Insert new level
			err := srs.db.InsertSupportResistanceLevel(newLevel)
			if err != nil {
				return fmt.Errorf("failed to insert S/R level: %w", err)
			}
		}
	}

	return nil
}

// buildAnalysisResult creates a comprehensive analysis result
func (srs *SupportResistanceService) buildAnalysisResult(symbol string, levels []*models.SupportResistanceLevel, priceData []*models.PriceData) (*models.SRAnalysisResult, error) {
	currentPrice := 0.0
	if len(priceData) > 0 {
		currentPrice = priceData[len(priceData)-1].Close
	}

	// Separate support and resistance levels
	var supportLevels, resistanceLevels []*models.SupportResistanceLevel
	for _, level := range levels {
		if level.LevelType == "support" {
			supportLevels = append(supportLevels, level)
		} else {
			resistanceLevels = append(resistanceLevels, level)
		}
	}

	// Find nearest levels
	nearestSupport, nearestResistance, err := srs.db.GetNearestSupportResistance(symbol, currentPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to get nearest S/R levels: %w", err)
	}

	// Get top 5 strongest levels
	keyLevels := srs.getKeyLevels(levels, 5)

	// Get recent touches
	recentTouches, err := srs.db.GetRecentSRLevelTouches(symbol, 24, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent touches: %w", err)
	}

	// Get level summary
	summary, err := srs.db.GetSRLevelSummary(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get S/R level summary: %w", err)
	}

	result := &models.SRAnalysisResult{
		Symbol:            symbol,
		AnalysisTime:      time.Now(),
		SupportLevels:     supportLevels,
		ResistanceLevels:  resistanceLevels,
		CurrentPrice:      currentPrice,
		NearestSupport:    nearestSupport,
		NearestResistance: nearestResistance,
		KeyLevels:         keyLevels,
		RecentTouches:     recentTouches,
		LevelSummary:      summary,
	}

	return result, nil
}

// getKeyLevels returns the strongest S/R levels
func (srs *SupportResistanceService) getKeyLevels(levels []*models.SupportResistanceLevel, count int) []*models.SupportResistanceLevel {
	// Sort by strength descending
	sort.Slice(levels, func(i, j int) bool {
		return levels[i].Strength > levels[j].Strength
	})

	if len(levels) <= count {
		return levels
	}

	return levels[:count]
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func maximum(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}
