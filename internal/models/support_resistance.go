package models

import (
	"time"
)

// SupportResistanceLevel represents a detected support or resistance level
type SupportResistanceLevel struct {
	ID               int64     `json:"id" db:"id"`
	Symbol           string    `json:"symbol" db:"symbol"`
	Level            float64   `json:"level" db:"level"`
	LevelType        string    `json:"level_type" db:"level_type"` // 'support' or 'resistance'
	Strength         float64   `json:"strength" db:"strength"`
	Touches          int       `json:"touches" db:"touches"`
	FirstTouch       time.Time `json:"first_touch" db:"first_touch"`
	LastTouch        time.Time `json:"last_touch" db:"last_touch"`
	VolumeConfirmed  bool      `json:"volume_confirmed" db:"volume_confirmed"`
	AvgVolume        float64   `json:"avg_volume" db:"avg_volume"`
	MaxBouncePercent float64   `json:"max_bounce_percent" db:"max_bounce_percent"`
	AvgBouncePercent float64   `json:"avg_bounce_percent" db:"avg_bounce_percent"`
	TimeframeOrigin  string    `json:"timeframe_origin" db:"timeframe_origin"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	LastValidated    time.Time `json:"last_validated" db:"last_validated"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// GetAge returns the age of the level in days since first touch
func (sr *SupportResistanceLevel) GetAge() float64 {
	return time.Since(sr.FirstTouch).Hours() / 24
}

// IsRecent checks if the level was touched recently (within 24 hours)
func (sr *SupportResistanceLevel) IsRecent() bool {
	return time.Since(sr.LastTouch).Hours() <= 24
}

// PivotPoint represents a price pivot high or low
type PivotPoint struct {
	ID        int64     `json:"id" db:"id"`
	Symbol    string    `json:"symbol" db:"symbol"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Price     float64   `json:"price" db:"price"`
	PivotType string    `json:"pivot_type" db:"pivot_type"` // 'high' or 'low'
	Strength  int       `json:"strength" db:"strength"`
	Volume    int64     `json:"volume" db:"volume"`
	Confirmed bool      `json:"confirmed" db:"confirmed"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// SRLevelTouch represents a touch of a support/resistance level
type SRLevelTouch struct {
	ID              int64     `json:"id" db:"id"`
	LevelID         int64     `json:"level_id" db:"level_id"`
	Symbol          string    `json:"symbol" db:"symbol"`
	TouchTime       time.Time `json:"touch_time" db:"touch_time"`
	TouchPrice      float64   `json:"touch_price" db:"touch_price"`
	Level           float64   `json:"level" db:"level"`
	DistancePercent float64   `json:"distance_percent" db:"distance_percent"`
	BouncePercent   float64   `json:"bounce_percent" db:"bounce_percent"`
	VolumeAtTouch   int64     `json:"volume_at_touch" db:"volume_at_touch"`
	VolumeSpike     bool      `json:"volume_spike" db:"volume_spike"`
	BounceConfirmed bool      `json:"bounce_confirmed" db:"bounce_confirmed"`
	TimeAtLevel     int       `json:"time_at_level" db:"time_at_level"`
	TouchType       string    `json:"touch_type" db:"touch_type"` // 'test', 'break', 'bounce'
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// SRDetectionConfig holds configuration for S/R detection
type SRDetectionConfig struct {
	MinTouches                int     `json:"min_touches" yaml:"min_touches"`
	LookbackDays              int     `json:"lookback_days" yaml:"lookback_days"`
	StrengthCalculation       string  `json:"strength_calculation" yaml:"strength_calculation"`
	MinLevelDistancePercent   float64 `json:"min_level_distance_percent" yaml:"min_level_distance_percent"`
	LevelPenetrationTolerance float64 `json:"level_penetration_tolerance" yaml:"level_penetration_tolerance"`
	PivotStrength             int     `json:"pivot_strength" yaml:"pivot_strength"`
	VolumeConfirmationRatio   float64 `json:"volume_confirmation_ratio" yaml:"volume_confirmation_ratio"`
	MaxLevelAge               int     `json:"max_level_age" yaml:"max_level_age"`
	MinBouncePercent          float64 `json:"min_bounce_percent" yaml:"min_bounce_percent"`
}

// SRPriceRange represents a price range for filtering
type SRPriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// SRDetectionFilter represents filter parameters for S/R queries
type SRDetectionFilter struct {
	Symbol      string       `json:"symbol"`
	LevelType   string       `json:"level_type"`
	MinStrength float64      `json:"min_strength"`
	MaxStrength float64      `json:"max_strength"`
	MinTouches  int          `json:"min_touches"`
	MaxTouches  int          `json:"max_touches"`
	IsActive    *bool        `json:"is_active"`
	TimeRange   SRTimeRange  `json:"time_range"`
	PriceRange  SRPriceRange `json:"price_range"`
	Limit       int          `json:"limit"`
	Offset      int          `json:"offset"`
}

// SRTimeRange represents a time range for filtering
type SRTimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// SRAnalysisResult represents the result of S/R analysis
type SRAnalysisResult struct {
	Symbol            string                    `json:"symbol"`
	AnalysisTime      time.Time                 `json:"analysis_time"`
	SupportLevels     []*SupportResistanceLevel `json:"support_levels"`
	ResistanceLevels  []*SupportResistanceLevel `json:"resistance_levels"`
	CurrentPrice      float64                   `json:"current_price"`
	NearestSupport    *SupportResistanceLevel   `json:"nearest_support"`
	NearestResistance *SupportResistanceLevel   `json:"nearest_resistance"`
	KeyLevels         []*SupportResistanceLevel `json:"key_levels"`
	RecentTouches     []*SRLevelTouch           `json:"recent_touches"`
	LevelSummary      *SRLevelSummary           `json:"level_summary"`
}

// SRLevelSummary provides summary statistics about S/R levels
type SRLevelSummary struct {
	TotalLevels      int       `json:"total_levels"`
	SupportCount     int       `json:"support_count"`
	ResistanceCount  int       `json:"resistance_count"`
	AvgStrength      float64   `json:"avg_strength"`
	StrongestLevel   float64   `json:"strongest_level"`
	WeakestLevel     float64   `json:"weakest_level"`
	RecentTouchCount int       `json:"recent_touch_count"`
	LastTouchTime    time.Time `json:"last_touch_time"`
	LastAnalysis     time.Time `json:"last_analysis"`
}

// SRResponse represents API response for S/R queries
type SRResponse struct {
	Symbol  string                    `json:"symbol"`
	Levels  []*SupportResistanceLevel `json:"levels"`
	Summary *SRLevelSummary           `json:"summary"`
	Status  string                    `json:"status"`
	Message string                    `json:"message,omitempty"`
}
