package models

import (
	"time"
)

// FallingWedgePattern represents a falling wedge chart pattern
type FallingWedgePattern struct {
	ID      int64  `json:"id" db:"id"`
	SetupID int64  `json:"setup_id" db:"setup_id"`
	Symbol  string `json:"symbol" db:"symbol"`

	// Pattern identification
	PatternType string `json:"pattern_type" db:"pattern_type"` // "falling_wedge"

	// Trend line points
	UpperTrendLine1 PatternPoint `json:"upper_trend_line_1" db:"upper_trend_line_1"`
	UpperTrendLine2 PatternPoint `json:"upper_trend_line_2" db:"upper_trend_line_2"`
	LowerTrendLine1 PatternPoint `json:"lower_trend_line_1" db:"lower_trend_line_1"`
	LowerTrendLine2 PatternPoint `json:"lower_trend_line_2" db:"lower_trend_line_2"`

	// Pattern metrics
	UpperSlope    float64 `json:"upper_slope" db:"upper_slope"`       // Slope of upper trend line
	LowerSlope    float64 `json:"lower_slope" db:"lower_slope"`       // Slope of lower trend line
	BreakoutLevel float64 `json:"breakout_level" db:"breakout_level"` // Current upper trend line level
	PatternWidth  int64   `json:"pattern_width" db:"pattern_width"`   // Duration in minutes
	PatternHeight float64 `json:"pattern_height" db:"pattern_height"` // Price range
	Convergence   float64 `json:"convergence" db:"convergence"`       // How much lines converge (%)

	// Volume analysis
	VolumeProfile string `json:"volume_profile" db:"volume_profile"` // "decreasing", "stable", "increasing"

	// Thesis tracking
	ThesisComponents FallingWedgeThesis `json:"thesis_components" db:"thesis_components"`

	// Status
	DetectedAt   time.Time `json:"detected_at" db:"detected_at"`
	LastUpdated  time.Time `json:"last_updated" db:"last_updated"`
	IsComplete   bool      `json:"is_complete" db:"is_complete"`
	CurrentPhase string    `json:"current_phase" db:"current_phase"` // "formation", "breakout", "target_pursuit", "completed"
}

// FallingWedgeThesis represents the thesis components for falling wedge pattern
type FallingWedgeThesis struct {
	// Formation components (40 points)
	DowntrendEstablished ThesisComponent `json:"downtrend_established"`  // 10 points
	ConvergingTrendLines ThesisComponent `json:"converging_trend_lines"` // 10 points
	MinimumTouchPoints   ThesisComponent `json:"minimum_touch_points"`   // 10 points
	VolumeDecline        ThesisComponent `json:"volume_decline"`         // 10 points

	// Breakout components (35 points)
	UpperTrendLineBreak ThesisComponent `json:"upper_trend_line_break"` // 15 points
	VolumeConfirmation  ThesisComponent `json:"volume_confirmation"`    // 10 points
	PriceCloseAboveLine ThesisComponent `json:"price_close_above_line"` // 10 points

	// Target components (25 points)
	PartialTarget ThesisComponent `json:"partial_target"` // 10 points
	FullTarget    ThesisComponent `json:"full_target"`    // 15 points

	// Calculated metrics
	CompletedComponents int     `json:"completed_components"`
	TotalComponents     int     `json:"total_components"`
	CompletionPercent   float64 `json:"completion_percent"`
	CurrentPhase        string  `json:"current_phase"`
}

// FallingWedgeConfig holds configuration for falling wedge detection
type FallingWedgeConfig struct {
	MinPatternDuration  time.Duration `json:"min_pattern_duration" yaml:"min_pattern_duration"`
	MaxPatternDuration  time.Duration `json:"max_pattern_duration" yaml:"max_pattern_duration"`
	MinConvergence      float64       `json:"min_convergence" yaml:"min_convergence"`             // Minimum convergence ratio
	MaxConvergence      float64       `json:"max_convergence" yaml:"max_convergence"`             // Maximum convergence ratio
	MinTouchPoints      int           `json:"min_touch_points" yaml:"min_touch_points"`           // Minimum trend line touches
	VolumeDecreaseRatio float64       `json:"volume_decrease_ratio" yaml:"volume_decrease_ratio"` // Volume should decrease
	BreakoutVolumeRatio float64       `json:"breakout_volume_ratio" yaml:"breakout_volume_ratio"` // Breakout volume increase
	MinWedgeHeight      float64       `json:"min_wedge_height" yaml:"min_wedge_height"`           // Minimum pattern height %
	MaxWedgeSlope       float64       `json:"max_wedge_slope" yaml:"max_wedge_slope"`             // Maximum downward slope
}

// FallingWedgeFilter represents filter parameters for falling wedge queries
type FallingWedgeFilter struct {
	Symbol         string      `json:"symbol"`
	Phase          string      `json:"phase"`
	IsComplete     *bool       `json:"is_complete"`
	MinConvergence float64     `json:"min_convergence"`
	TimeRange      SRTimeRange `json:"time_range"`
	Limit          int         `json:"limit"`
	Offset         int         `json:"offset"`
}

// Methods for FallingWedgePattern

// CalculateTargetPrice calculates the price target based on pattern height
func (fwp *FallingWedgePattern) CalculateTargetPrice() float64 {
	return fwp.BreakoutLevel + fwp.PatternHeight
}

// GetConvergenceScore returns a score based on how well lines converge
func (fwp *FallingWedgePattern) GetConvergenceScore() float64 {
	// Higher convergence = higher score (up to 100)
	return fwp.Convergence * 10 // 10% convergence = 100 score
}

// IsNearBreakout checks if price is near the breakout level
func (fwp *FallingWedgePattern) IsNearBreakout(currentPrice float64) bool {
	threshold := fwp.BreakoutLevel * 0.98 // Within 2% of breakout
	return currentPrice >= threshold
}

// GetPatternAge returns the age of the pattern in hours
func (fwp *FallingWedgePattern) GetPatternAge() float64 {
	return time.Since(fwp.DetectedAt).Hours()
}

// Methods for FallingWedgeThesis

// InitializeWedgeThesis initializes the thesis with default values
func (fwt *FallingWedgeThesis) InitializeWedgeThesis() {
	now := time.Now()

	// Formation components
	fwt.DowntrendEstablished = ThesisComponent{
		Name:            "Downtrend Established",
		Description:     "Clear downward price trend established",
		IsCompleted:     false,
		IsRequired:      true,
		Weight:          10.0,
		ConfidenceLevel: 0.0,
		Evidence:        []string{},
		LastChecked:     now,
		AutoDetected:    true,
	}

	fwt.ConvergingTrendLines = ThesisComponent{
		Name:            "Converging Trend Lines",
		Description:     "Upper and lower trend lines converge",
		IsCompleted:     false,
		IsRequired:      true,
		Weight:          10.0,
		ConfidenceLevel: 0.0,
		Evidence:        []string{},
		LastChecked:     now,
		AutoDetected:    true,
	}

	fwt.MinimumTouchPoints = ThesisComponent{
		Name:            "Minimum Touch Points",
		Description:     "At least 4 touch points on trend lines",
		IsCompleted:     false,
		IsRequired:      true,
		Weight:          10.0,
		ConfidenceLevel: 0.0,
		Evidence:        []string{},
		LastChecked:     now,
		AutoDetected:    true,
	}

	fwt.VolumeDecline = ThesisComponent{
		Name:            "Volume Decline",
		Description:     "Volume decreases as pattern forms",
		IsCompleted:     false,
		IsRequired:      false,
		Weight:          10.0,
		ConfidenceLevel: 0.0,
		Evidence:        []string{},
		LastChecked:     now,
		AutoDetected:    true,
	}

	// Breakout components
	fwt.UpperTrendLineBreak = ThesisComponent{
		Name:            "Upper Trend Line Break",
		Description:     "Price breaks above upper trend line",
		IsCompleted:     false,
		IsRequired:      true,
		Weight:          15.0,
		ConfidenceLevel: 0.0,
		Evidence:        []string{},
		LastChecked:     now,
		AutoDetected:    true,
	}

	fwt.VolumeConfirmation = ThesisComponent{
		Name:            "Volume Confirmation",
		Description:     "Breakout confirmed with volume spike",
		IsCompleted:     false,
		IsRequired:      false,
		Weight:          10.0,
		ConfidenceLevel: 0.0,
		Evidence:        []string{},
		LastChecked:     now,
		AutoDetected:    true,
	}

	fwt.PriceCloseAboveLine = ThesisComponent{
		Name:            "Price Close Above Line",
		Description:     "Price closes above trend line",
		IsCompleted:     false,
		IsRequired:      true,
		Weight:          10.0,
		ConfidenceLevel: 0.0,
		Evidence:        []string{},
		LastChecked:     now,
		AutoDetected:    true,
	}

	// Target components
	fwt.PartialTarget = ThesisComponent{
		Name:            "Partial Target",
		Description:     "50% of projected target reached",
		IsCompleted:     false,
		IsRequired:      false,
		Weight:          10.0,
		ConfidenceLevel: 0.0,
		Evidence:        []string{},
		LastChecked:     now,
		AutoDetected:    true,
	}

	fwt.FullTarget = ThesisComponent{
		Name:            "Full Target",
		Description:     "Full projected target reached",
		IsCompleted:     false,
		IsRequired:      false,
		Weight:          15.0,
		ConfidenceLevel: 0.0,
		Evidence:        []string{},
		LastChecked:     now,
		AutoDetected:    true,
	}

	fwt.TotalComponents = 9
	fwt.CalculateCompletion()
	fwt.UpdatePhase()
}

// GetAllComponents returns all thesis components
func (fwt *FallingWedgeThesis) GetAllComponents() []*ThesisComponent {
	return []*ThesisComponent{
		&fwt.DowntrendEstablished,
		&fwt.ConvergingTrendLines,
		&fwt.MinimumTouchPoints,
		&fwt.VolumeDecline,
		&fwt.UpperTrendLineBreak,
		&fwt.VolumeConfirmation,
		&fwt.PriceCloseAboveLine,
		&fwt.PartialTarget,
		&fwt.FullTarget,
	}
}

// CalculateCompletion calculates completion statistics
func (fwt *FallingWedgeThesis) CalculateCompletion() {
	components := fwt.GetAllComponents()
	completed := 0
	totalWeight := 0.0
	completedWeight := 0.0

	for _, component := range components {
		totalWeight += component.Weight
		if component.IsCompleted {
			completed++
			completedWeight += component.Weight
		}
	}

	fwt.CompletedComponents = completed
	fwt.TotalComponents = len(components)

	if totalWeight > 0 {
		fwt.CompletionPercent = (completedWeight / totalWeight) * 100
	}
}

// UpdatePhase updates the current phase based on completion
func (fwt *FallingWedgeThesis) UpdatePhase() {
	if fwt.FullTarget.IsCompleted {
		fwt.CurrentPhase = PhaseCompleted
	} else if fwt.UpperTrendLineBreak.IsCompleted {
		fwt.CurrentPhase = PhaseTargetPursuit
	} else if fwt.ConvergingTrendLines.IsCompleted && fwt.MinimumTouchPoints.IsCompleted {
		fwt.CurrentPhase = PhaseBreakout
	} else {
		fwt.CurrentPhase = PhaseFormation
	}
}

// GetFormationScore returns the formation phase score
func (fwt *FallingWedgeThesis) GetFormationScore() float64 {
	score := 0.0
	components := []*ThesisComponent{
		&fwt.DowntrendEstablished,
		&fwt.ConvergingTrendLines,
		&fwt.MinimumTouchPoints,
		&fwt.VolumeDecline,
	}

	for _, component := range components {
		if component.IsCompleted {
			score += component.Weight
		}
	}
	return score
}

// GetBreakoutScore returns the breakout phase score
func (fwt *FallingWedgeThesis) GetBreakoutScore() float64 {
	score := 0.0
	components := []*ThesisComponent{
		&fwt.UpperTrendLineBreak,
		&fwt.VolumeConfirmation,
		&fwt.PriceCloseAboveLine,
	}

	for _, component := range components {
		if component.IsCompleted {
			score += component.Weight
		}
	}
	return score
}

// GetTargetScore returns the target phase score
func (fwt *FallingWedgeThesis) GetTargetScore() float64 {
	score := 0.0
	components := []*ThesisComponent{
		&fwt.PartialTarget,
		&fwt.FullTarget,
	}

	for _, component := range components {
		if component.IsCompleted {
			score += component.Weight
		}
	}
	return score
}
