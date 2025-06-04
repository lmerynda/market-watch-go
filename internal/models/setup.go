package models

import (
	"time"
)

// TradingSetup represents a detected trading setup
type TradingSetup struct {
	ID           int64     `json:"id" db:"id"`
	Symbol       string    `json:"symbol" db:"symbol"`
	SetupType    string    `json:"setup_type" db:"setup_type"`       // 'support_bounce', 'resistance_bounce', 'breakout', etc.
	Direction    string    `json:"direction" db:"direction"`         // 'bullish', 'bearish'
	QualityScore float64   `json:"quality_score" db:"quality_score"` // 0-100 overall score
	Confidence   string    `json:"confidence" db:"confidence"`       // 'high', 'medium', 'low'
	Status       string    `json:"status" db:"status"`               // 'active', 'triggered', 'expired', 'invalidated'
	DetectedAt   time.Time `json:"detected_at" db:"detected_at"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	LastUpdated  time.Time `json:"last_updated" db:"last_updated"`

	// Price levels
	CurrentPrice float64 `json:"current_price" db:"current_price"`
	EntryPrice   float64 `json:"entry_price" db:"entry_price"`
	StopLoss     float64 `json:"stop_loss" db:"stop_loss"`
	Target1      float64 `json:"target1" db:"target1"`
	Target2      float64 `json:"target2" db:"target2"`
	Target3      float64 `json:"target3" db:"target3"`

	// Risk/Reward metrics
	RiskAmount      float64 `json:"risk_amount" db:"risk_amount"`
	RewardPotential float64 `json:"reward_potential" db:"reward_potential"`
	RiskRewardRatio float64 `json:"risk_reward_ratio" db:"risk_reward_ratio"`

	// Associated levels
	SupportLevel    *SupportResistanceLevel `json:"support_level,omitempty"`
	ResistanceLevel *SupportResistanceLevel `json:"resistance_level,omitempty"`
	KeyLevel        *SupportResistanceLevel `json:"key_level,omitempty"`

	// Scoring breakdown
	PriceActionScore float64 `json:"price_action_score" db:"price_action_score"`
	VolumeScore      float64 `json:"volume_score" db:"volume_score"`
	TechnicalScore   float64 `json:"technical_score" db:"technical_score"`
	RiskRewardScore  float64 `json:"risk_reward_score" db:"risk_reward_score"`

	// Checklist items
	Checklist *SetupChecklist `json:"checklist,omitempty"`

	// Metadata
	Notes     string    `json:"notes" db:"notes"`
	IsManual  bool      `json:"is_manual" db:"is_manual"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// SetupChecklist represents the criteria checklist for a setup
type SetupChecklist struct {
	SetupID int64 `json:"setup_id" db:"setup_id"`

	// Price Action Criteria (25 points max)
	MinLevelTouches ChecklistItem `json:"min_level_touches"` // 5 points
	BounceStrength  ChecklistItem `json:"bounce_strength"`   // 5 points
	TimeAtLevel     ChecklistItem `json:"time_at_level"`     // 5 points
	RejectionCandle ChecklistItem `json:"rejection_candle"`  // 5 points
	LevelDuration   ChecklistItem `json:"level_duration"`    // 5 points

	// Volume Criteria (25 points max)
	VolumeSpike        ChecklistItem `json:"volume_spike"`        // 5 points
	VolumeConfirmation ChecklistItem `json:"volume_confirmation"` // 5 points
	ApproachVolume     ChecklistItem `json:"approach_volume"`     // 5 points
	VWAPRelationship   ChecklistItem `json:"vwap_relationship"`   // 5 points
	RelativeVolume     ChecklistItem `json:"relative_volume"`     // 5 points

	// Technical Indicators (25 points max)
	RSICondition       ChecklistItem `json:"rsi_condition"`       // 5 points
	MovingAverage      ChecklistItem `json:"moving_average"`      // 5 points
	MACDSignal         ChecklistItem `json:"macd_signal"`         // 5 points
	MomentumDivergence ChecklistItem `json:"momentum_divergence"` // 5 points
	BollingerBands     ChecklistItem `json:"bollinger_bands"`     // 5 points

	// Risk Management (25 points max)
	StopLossDefined ChecklistItem `json:"stop_loss_defined"` // 5 points
	RiskRewardRatio ChecklistItem `json:"risk_reward_ratio"` // 5 points
	PositionSize    ChecklistItem `json:"position_size"`     // 5 points
	EntryPrecision  ChecklistItem `json:"entry_precision"`   // 5 points
	ExitStrategy    ChecklistItem `json:"exit_strategy"`     // 5 points

	// Calculated scores
	TotalScore        float64 `json:"total_score"`
	CompletedItems    int     `json:"completed_items"`
	TotalItems        int     `json:"total_items"`
	CompletionPercent float64 `json:"completion_percent"`

	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}

// ChecklistItem represents a single checklist criterion
type ChecklistItem struct {
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	IsCompleted    bool      `json:"is_completed"`
	IsRequired     bool      `json:"is_required"`
	Points         float64   `json:"points"`
	MaxPoints      float64   `json:"max_points"`
	AutoDetected   bool      `json:"auto_detected"`
	ManualOverride bool      `json:"manual_override"`
	Notes          string    `json:"notes"`
	LastChecked    time.Time `json:"last_checked"`
}

// SetupScoringConfig holds configuration for setup scoring
type SetupScoringConfig struct {
	// Quality thresholds
	HighQualityThreshold   float64 `json:"high_quality_threshold" yaml:"high_quality_threshold"`
	MediumQualityThreshold float64 `json:"medium_quality_threshold" yaml:"medium_quality_threshold"`
	LowQualityThreshold    float64 `json:"low_quality_threshold" yaml:"low_quality_threshold"`

	// Scoring weights
	PriceActionWeight float64 `json:"price_action_weight" yaml:"price_action_weight"`
	VolumeWeight      float64 `json:"volume_weight" yaml:"volume_weight"`
	TechnicalWeight   float64 `json:"technical_weight" yaml:"technical_weight"`
	RiskRewardWeight  float64 `json:"risk_reward_weight" yaml:"risk_reward_weight"`

	// Bounce criteria
	MinBouncePercent      float64 `json:"min_bounce_percent" yaml:"min_bounce_percent"`
	MinTimeAtLevelMinutes int     `json:"min_time_at_level_minutes" yaml:"min_time_at_level_minutes"`
	MaxLevelAgeDays       int     `json:"max_level_age_days" yaml:"max_level_age_days"`

	// Risk management
	MinRiskRewardRatio float64 `json:"min_risk_reward_ratio" yaml:"min_risk_reward_ratio"`
	MaxRiskPercent     float64 `json:"max_risk_percent" yaml:"max_risk_percent"`

	// Setup expiration
	SetupExpirationHours int `json:"setup_expiration_hours" yaml:"setup_expiration_hours"`
}

// SetupFilter represents filter parameters for setup queries
type SetupFilter struct {
	Symbol          string      `json:"symbol"`
	SetupType       string      `json:"setup_type"`
	Direction       string      `json:"direction"`
	Status          string      `json:"status"`
	MinQualityScore float64     `json:"min_quality_score"`
	MaxQualityScore float64     `json:"max_quality_score"`
	Confidence      string      `json:"confidence"`
	TimeRange       SRTimeRange `json:"time_range"`
	IsActive        *bool       `json:"is_active"`
	Limit           int         `json:"limit"`
	Offset          int         `json:"offset"`
}

// SetupDetectionResult represents the result of setup detection
type SetupDetectionResult struct {
	Symbol        string          `json:"symbol"`
	DetectionTime time.Time       `json:"detection_time"`
	SetupsFound   []*TradingSetup `json:"setups_found"`
	ActiveSetups  []*TradingSetup `json:"active_setups"`
	ExpiredSetups []*TradingSetup `json:"expired_setups"`
	Summary       *SetupSummary   `json:"summary"`
	Errors        []string        `json:"errors,omitempty"`
}

// SetupSummary provides summary statistics about setups
type SetupSummary struct {
	TotalSetups        int           `json:"total_setups"`
	ActiveCount        int           `json:"active_count"`
	HighQualityCount   int           `json:"high_quality_count"`
	MediumQualityCount int           `json:"medium_quality_count"`
	LowQualityCount    int           `json:"low_quality_count"`
	BullishCount       int           `json:"bullish_count"`
	BearishCount       int           `json:"bearish_count"`
	AvgQualityScore    float64       `json:"avg_quality_score"`
	AvgRiskReward      float64       `json:"avg_risk_reward"`
	BestSetup          *TradingSetup `json:"best_setup"`
	LastDetection      time.Time     `json:"last_detection"`
}

// SetupResponse represents API response for setup queries
type SetupResponse struct {
	Symbol  string          `json:"symbol"`
	Setups  []*TradingSetup `json:"setups"`
	Summary *SetupSummary   `json:"summary"`
	Status  string          `json:"status"`
	Message string          `json:"message,omitempty"`
}

// SetupAlert represents an alert for a trading setup
type SetupAlert struct {
	ID               int64     `json:"id" db:"id"`
	SetupID          int64     `json:"setup_id" db:"setup_id"`
	Symbol           string    `json:"symbol" db:"symbol"`
	AlertType        string    `json:"alert_type" db:"alert_type"` // 'new_setup', 'quality_change', 'triggered', 'invalidated'
	Message          string    `json:"message" db:"message"`
	Severity         string    `json:"severity" db:"severity"` // 'high', 'medium', 'low'
	IsActive         bool      `json:"is_active" db:"is_active"`
	TriggeredAt      time.Time `json:"triggered_at" db:"triggered_at"`
	NotificationSent bool      `json:"notification_sent" db:"notification_sent"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// Methods for TradingSetup

// GetAge returns the age of the setup in hours
func (ts *TradingSetup) GetAge() float64 {
	return time.Since(ts.DetectedAt).Hours()
}

// IsExpired checks if the setup has expired
func (ts *TradingSetup) IsExpired() bool {
	return time.Now().After(ts.ExpiresAt)
}

// IsActive checks if the setup is currently active
func (ts *TradingSetup) IsActive() bool {
	return ts.Status == "active" && !ts.IsExpired()
}

// GetConfidenceLevel returns confidence based on quality score
func (ts *TradingSetup) GetConfidenceLevel() string {
	switch {
	case ts.QualityScore >= 80:
		return "high"
	case ts.QualityScore >= 60:
		return "medium"
	default:
		return "low"
	}
}

// GetRiskAmount calculates risk amount from entry to stop loss
func (ts *TradingSetup) GetRiskAmount() float64 {
	if ts.Direction == "bullish" {
		return ts.EntryPrice - ts.StopLoss
	}
	return ts.StopLoss - ts.EntryPrice
}

// GetRewardPotential calculates potential reward to first target
func (ts *TradingSetup) GetRewardPotential() float64 {
	if ts.Direction == "bullish" {
		return ts.Target1 - ts.EntryPrice
	}
	return ts.EntryPrice - ts.Target1
}

// CalculateRiskRewardRatio calculates the risk:reward ratio
func (ts *TradingSetup) CalculateRiskRewardRatio() float64 {
	risk := ts.GetRiskAmount()
	if risk <= 0 {
		return 0
	}
	reward := ts.GetRewardPotential()
	return reward / risk
}

// UpdateStatus updates the setup status and timestamps
func (ts *TradingSetup) UpdateStatus(newStatus string) {
	ts.Status = newStatus
	ts.LastUpdated = time.Now()
}

// IsHighQuality checks if the setup meets high quality criteria
func (ts *TradingSetup) IsHighQuality() bool {
	return ts.QualityScore >= 80 && ts.RiskRewardRatio >= 2.0
}

// IsMediumQuality checks if the setup meets medium quality criteria
func (ts *TradingSetup) IsMediumQuality() bool {
	return ts.QualityScore >= 60 && ts.QualityScore < 80 && ts.RiskRewardRatio >= 1.5
}

// Methods for SetupChecklist

// CalculateScore calculates the total checklist score
func (sc *SetupChecklist) CalculateScore() {
	totalPoints := 0.0
	maxPoints := 0.0
	completedItems := 0
	totalItems := 0

	items := []*ChecklistItem{
		&sc.MinLevelTouches, &sc.BounceStrength, &sc.TimeAtLevel, &sc.RejectionCandle, &sc.LevelDuration,
		&sc.VolumeSpike, &sc.VolumeConfirmation, &sc.ApproachVolume, &sc.VWAPRelationship, &sc.RelativeVolume,
		&sc.RSICondition, &sc.MovingAverage, &sc.MACDSignal, &sc.MomentumDivergence, &sc.BollingerBands,
		&sc.StopLossDefined, &sc.RiskRewardRatio, &sc.PositionSize, &sc.EntryPrecision, &sc.ExitStrategy,
	}

	for _, item := range items {
		if item.MaxPoints > 0 {
			totalItems++
			maxPoints += item.MaxPoints
			if item.IsCompleted {
				completedItems++
				totalPoints += item.Points
			}
		}
	}

	sc.TotalScore = totalPoints
	sc.CompletedItems = completedItems
	sc.TotalItems = totalItems
	if totalItems > 0 {
		sc.CompletionPercent = (float64(completedItems) / float64(totalItems)) * 100
	}
	sc.LastUpdated = time.Now()
}

// GetPriceActionScore returns the price action component score
func (sc *SetupChecklist) GetPriceActionScore() float64 {
	score := 0.0
	items := []*ChecklistItem{&sc.MinLevelTouches, &sc.BounceStrength, &sc.TimeAtLevel, &sc.RejectionCandle, &sc.LevelDuration}
	for _, item := range items {
		if item.IsCompleted {
			score += item.Points
		}
	}
	return score
}

// GetVolumeScore returns the volume component score
func (sc *SetupChecklist) GetVolumeScore() float64 {
	score := 0.0
	items := []*ChecklistItem{&sc.VolumeSpike, &sc.VolumeConfirmation, &sc.ApproachVolume, &sc.VWAPRelationship, &sc.RelativeVolume}
	for _, item := range items {
		if item.IsCompleted {
			score += item.Points
		}
	}
	return score
}

// GetTechnicalScore returns the technical indicators component score
func (sc *SetupChecklist) GetTechnicalScore() float64 {
	score := 0.0
	items := []*ChecklistItem{&sc.RSICondition, &sc.MovingAverage, &sc.MACDSignal, &sc.MomentumDivergence, &sc.BollingerBands}
	for _, item := range items {
		if item.IsCompleted {
			score += item.Points
		}
	}
	return score
}

// GetRiskManagementScore returns the risk management component score
func (sc *SetupChecklist) GetRiskManagementScore() float64 {
	score := 0.0
	items := []*ChecklistItem{&sc.StopLossDefined, &sc.RiskRewardRatio, &sc.PositionSize, &sc.EntryPrecision, &sc.ExitStrategy}
	for _, item := range items {
		if item.IsCompleted {
			score += item.Points
		}
	}
	return score
}

// IsComplete checks if all required items are completed
func (sc *SetupChecklist) IsComplete() bool {
	items := []*ChecklistItem{
		&sc.MinLevelTouches, &sc.BounceStrength, &sc.TimeAtLevel, &sc.RejectionCandle, &sc.LevelDuration,
		&sc.VolumeSpike, &sc.VolumeConfirmation, &sc.ApproachVolume, &sc.VWAPRelationship, &sc.RelativeVolume,
		&sc.RSICondition, &sc.MovingAverage, &sc.MACDSignal, &sc.MomentumDivergence, &sc.BollingerBands,
		&sc.StopLossDefined, &sc.RiskRewardRatio, &sc.PositionSize, &sc.EntryPrecision, &sc.ExitStrategy,
	}

	for _, item := range items {
		if item.IsRequired && !item.IsCompleted {
			return false
		}
	}
	return true
}
