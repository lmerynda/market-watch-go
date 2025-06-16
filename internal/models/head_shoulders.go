package models

import (
	"encoding/json"
	"time"
)

// Pattern type constants
const (
	SetupTypeInverseHeadShoulders = "inverse_head_shoulders"
	SetupTypeHeadShoulders        = "head_shoulders"
)

// Pattern phase constants
const (
	PhaseFormation     = "formation"
	PhaseBreakout      = "breakout"
	PhaseTargetPursuit = "target_pursuit"
	PhaseCompleted     = "completed"
)

// HeadShouldersPattern represents a detected head and shoulders pattern
type HeadShouldersPattern struct {
	ID          int64  `json:"id" db:"id"`
	SetupID     int64  `json:"setup_id" db:"setup_id"`
	Symbol      string `json:"symbol" db:"symbol"`
	PatternType string `json:"pattern_type" db:"pattern_type"` // "inverse_head_shoulders" or "head_shoulders"

	// Pattern Points
	LeftShoulderHigh  PatternPoint `json:"left_shoulder_high"`
	LeftShoulderLow   PatternPoint `json:"left_shoulder_low"`
	HeadHigh          PatternPoint `json:"head_high"`
	HeadLow           PatternPoint `json:"head_low"`
	RightShoulderHigh PatternPoint `json:"right_shoulder_high,omitempty"`
	RightShoulderLow  PatternPoint `json:"right_shoulder_low,omitempty"`

	// Neckline Analysis
	NecklineLevel  float64      `json:"neckline_level" db:"neckline_level"`
	NecklineSlope  float64      `json:"neckline_slope" db:"neckline_slope"`
	NecklineTouch1 PatternPoint `json:"neckline_touch1"`
	NecklineTouch2 PatternPoint `json:"neckline_touch2"`

	// Pattern Measurements
	PatternWidth  int64   `json:"pattern_width" db:"pattern_width"` // Duration in minutes
	PatternHeight float64 `json:"pattern_height" db:"pattern_height"`
	Symmetry      float64 `json:"symmetry" db:"symmetry"` // 0-100 score

	// Thesis Tracking
	ThesisComponents HeadShouldersThesis `json:"thesis_components"`

	// Status
	DetectedAt   time.Time `json:"detected_at" db:"detected_at"`
	LastUpdated  time.Time `json:"last_updated" db:"last_updated"`
	IsComplete   bool      `json:"is_complete" db:"is_complete"`
	CurrentPhase string    `json:"current_phase" db:"current_phase"`
}

// PatternPoint represents a specific point in the pattern
type PatternPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Price       float64   `json:"price"`
	Volume      int64     `json:"volume"`
	VolumeRatio float64   `json:"volume_ratio"` // vs average
}

// HeadShouldersThesis tracks all thesis components for the pattern
type HeadShouldersThesis struct {
	// Left Shoulder Components
	LeftShoulderFormed ThesisComponent `json:"left_shoulder_formed"`
	LeftShoulderVolume ThesisComponent `json:"left_shoulder_volume"`

	// Head Components
	HeadFormed      ThesisComponent `json:"head_formed"`
	HeadVolumeSpike ThesisComponent `json:"head_volume_spike"`
	HeadLowerLow    ThesisComponent `json:"head_lower_low"` // For inverse H&S

	// Right Shoulder Components
	RightShoulderFormed   ThesisComponent `json:"right_shoulder_formed"`
	RightShoulderSymmetry ThesisComponent `json:"right_shoulder_symmetry"`
	RightShoulderVolume   ThesisComponent `json:"right_shoulder_volume"`

	// Neckline Components
	NecklineEstablished ThesisComponent `json:"neckline_established"`
	NecklineRetest      ThesisComponent `json:"neckline_retest"`
	NecklineBreakout    ThesisComponent `json:"neckline_breakout"`
	BreakoutVolume      ThesisComponent `json:"breakout_volume"`

	// Target Components
	TargetProjected ThesisComponent `json:"target_projected"`
	PartialFillT1   ThesisComponent `json:"partial_fill_t1"`
	PartialFillT2   ThesisComponent `json:"partial_fill_t2"`
	FullTarget      ThesisComponent `json:"full_target"`

	// Overall Status
	CompletedComponents int     `json:"completed_components"`
	TotalComponents     int     `json:"total_components"`
	CompletionPercent   float64 `json:"completion_percent"`
	CurrentPhase        string  `json:"current_phase"`
}

// ThesisComponent represents a single component of the thesis
type ThesisComponent struct {
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	IsCompleted      bool       `json:"is_completed"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	IsRequired       bool       `json:"is_required"`
	Weight           float64    `json:"weight"`           // Importance weight
	ConfidenceLevel  float64    `json:"confidence_level"` // 0-100
	Evidence         []string   `json:"evidence"`         // Supporting data points
	LastChecked      time.Time  `json:"last_checked"`
	AutoDetected     bool       `json:"auto_detected"`
	NotificationSent bool       `json:"notification_sent"`
}

// HeadShouldersConfig holds configuration for pattern detection
type HeadShouldersConfig struct {
	MinPatternDuration   time.Duration `json:"min_pattern_duration" yaml:"min_pattern_duration"`
	MaxPatternDuration   time.Duration `json:"max_pattern_duration" yaml:"max_pattern_duration"`
	MinSymmetryScore     float64       `json:"min_symmetry_score" yaml:"min_symmetry_score"`
	MinVolumeIncrease    float64       `json:"min_volume_increase" yaml:"min_volume_increase"`
	NecklineDeviation    float64       `json:"neckline_deviation" yaml:"neckline_deviation"`
	TargetMultiplier     float64       `json:"target_multiplier" yaml:"target_multiplier"`
	MinHeadDepth         float64       `json:"min_head_depth" yaml:"min_head_depth"`
	MaxShoulderAsymmetry float64       `json:"max_shoulder_asymmetry" yaml:"max_shoulder_asymmetry"`
}

// PatternFilter represents filter parameters for pattern queries
type PatternFilter struct {
	Symbol      string      `json:"symbol"`
	PatternType string      `json:"pattern_type"`
	Phase       string      `json:"phase"`
	IsComplete  *bool       `json:"is_complete"`
	MinSymmetry float64     `json:"min_symmetry"`
	TimeRange   SRTimeRange `json:"time_range"`
	Limit       int         `json:"limit"`
	Offset      int         `json:"offset"`
}

// PatternAlert represents an alert for a pattern
type PatternAlert struct {
	ID               int64     `json:"id" db:"id"`
	PatternID        int64     `json:"pattern_id" db:"pattern_id"`
	Symbol           string    `json:"symbol" db:"symbol"`
	ComponentName    string    `json:"component_name" db:"component_name"`
	AlertType        string    `json:"alert_type" db:"alert_type"` // 'component_completed', 'breakout_confirmed', 'target_reached'
	Message          string    `json:"message" db:"message"`
	TriggeredAt      time.Time `json:"triggered_at" db:"triggered_at"`
	NotificationSent bool      `json:"notification_sent" db:"notification_sent"`
	EmailSent        bool      `json:"email_sent" db:"email_sent"`
}

// Methods for HeadShouldersPattern

// GetAge returns the age of the pattern in hours
func (hsp *HeadShouldersPattern) GetAge() float64 {
	return time.Since(hsp.DetectedAt).Hours()
}

// IsInFormation checks if the pattern is still forming
func (hsp *HeadShouldersPattern) IsInFormation() bool {
	return hsp.CurrentPhase == PhaseFormation && !hsp.IsComplete
}

// IsBreakoutPhase checks if the pattern is in breakout phase
func (hsp *HeadShouldersPattern) IsBreakoutPhase() bool {
	return hsp.CurrentPhase == PhaseBreakout
}

// IsTargetPursuit checks if the pattern is pursuing targets
func (hsp *HeadShouldersPattern) IsTargetPursuit() bool {
	return hsp.CurrentPhase == PhaseTargetPursuit
}

// CalculatePatternHeight calculates the height of the pattern
func (hsp *HeadShouldersPattern) CalculatePatternHeight() float64 {
	if hsp.PatternType == SetupTypeInverseHeadShoulders {
		// For inverse H&S, height is from head low to neckline
		return hsp.NecklineLevel - hsp.HeadLow.Price
	} else {
		// For regular H&S, height is from head high to neckline
		return hsp.HeadHigh.Price - hsp.NecklineLevel
	}
}

// CalculateTargetPrice calculates the projected target price
func (hsp *HeadShouldersPattern) CalculateTargetPrice() float64 {
	height := hsp.CalculatePatternHeight()
	if hsp.PatternType == SetupTypeInverseHeadShoulders {
		return hsp.NecklineLevel + height
	} else {
		return hsp.NecklineLevel - height
	}
}

// GetSymmetryScore calculates the symmetry score between shoulders
func (hsp *HeadShouldersPattern) GetSymmetryScore() float64 {
	if hsp.RightShoulderHigh.Price == 0 || hsp.RightShoulderLow.Price == 0 {
		return 0 // Right shoulder not formed yet
	}

	var leftHeight, rightHeight float64

	if hsp.PatternType == SetupTypeInverseHeadShoulders {
		leftHeight = hsp.LeftShoulderHigh.Price - hsp.LeftShoulderLow.Price
		rightHeight = hsp.RightShoulderHigh.Price - hsp.RightShoulderLow.Price
	} else {
		leftHeight = hsp.LeftShoulderHigh.Price - hsp.LeftShoulderLow.Price
		rightHeight = hsp.RightShoulderHigh.Price - hsp.RightShoulderLow.Price
	}

	if leftHeight == 0 || rightHeight == 0 {
		return 0
	}

	// Calculate symmetry as percentage
	diff := abs(leftHeight - rightHeight)
	avg := (leftHeight + rightHeight) / 2
	symmetry := (1 - (diff / avg)) * 100

	if symmetry < 0 {
		symmetry = 0
	}
	if symmetry > 100 {
		symmetry = 100
	}

	return symmetry
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Methods for HeadShouldersThesis

// InitializeThesis sets up the initial thesis components
func (hst *HeadShouldersThesis) InitializeThesis(patternType string) {
	now := time.Now()

	// Initialize all components with defaults
	hst.LeftShoulderFormed = ThesisComponent{
		Name:         "Left Shoulder Formed",
		Description:  "The left shoulder peak and trough have been identified",
		IsRequired:   true,
		Weight:       10.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.LeftShoulderVolume = ThesisComponent{
		Name:         "Left Shoulder Volume",
		Description:  "Volume characteristics at left shoulder are acceptable",
		IsRequired:   false,
		Weight:       5.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.HeadFormed = ThesisComponent{
		Name:         "Head Formed",
		Description:  "The head (main peak/trough) has been identified",
		IsRequired:   true,
		Weight:       15.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.HeadVolumeSpike = ThesisComponent{
		Name:         "Head Volume Spike",
		Description:  "Significant volume increase at head formation",
		IsRequired:   false,
		Weight:       8.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	if patternType == SetupTypeInverseHeadShoulders {
		hst.HeadLowerLow = ThesisComponent{
			Name:         "Head Lower Low",
			Description:  "Head forms a lower low than both shoulders",
			IsRequired:   true,
			Weight:       12.0,
			LastChecked:  now,
			AutoDetected: true,
		}
	}

	hst.RightShoulderFormed = ThesisComponent{
		Name:         "Right Shoulder Formed",
		Description:  "The right shoulder peak and trough have been identified",
		IsRequired:   true,
		Weight:       10.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.RightShoulderSymmetry = ThesisComponent{
		Name:         "Right Shoulder Symmetry",
		Description:  "Right shoulder shows good symmetry with left shoulder",
		IsRequired:   false,
		Weight:       8.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.RightShoulderVolume = ThesisComponent{
		Name:         "Right Shoulder Volume",
		Description:  "Volume characteristics at right shoulder are decreasing",
		IsRequired:   false,
		Weight:       5.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.NecklineEstablished = ThesisComponent{
		Name:         "Neckline Established",
		Description:  "Clear neckline drawn connecting shoulder highs/lows",
		IsRequired:   true,
		Weight:       12.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.NecklineRetest = ThesisComponent{
		Name:         "Neckline Retest",
		Description:  "Price retests the neckline after initial break",
		IsRequired:   false,
		Weight:       6.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.NecklineBreakout = ThesisComponent{
		Name:         "Neckline Breakout",
		Description:  "Price breaks above/below neckline with conviction",
		IsRequired:   true,
		Weight:       15.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.BreakoutVolume = ThesisComponent{
		Name:         "Breakout Volume",
		Description:  "Strong volume confirms the neckline breakout",
		IsRequired:   true,
		Weight:       10.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.TargetProjected = ThesisComponent{
		Name:         "Target Projected",
		Description:  "Price target calculated based on pattern height",
		IsRequired:   false,
		Weight:       3.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.PartialFillT1 = ThesisComponent{
		Name:         "Partial Target 1",
		Description:  "First target level (50% of projection) reached",
		IsRequired:   false,
		Weight:       4.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.PartialFillT2 = ThesisComponent{
		Name:         "Partial Target 2",
		Description:  "Second target level (75% of projection) reached",
		IsRequired:   false,
		Weight:       4.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.FullTarget = ThesisComponent{
		Name:         "Full Target",
		Description:  "Full price target (100% projection) reached",
		IsRequired:   false,
		Weight:       5.0,
		LastChecked:  now,
		AutoDetected: true,
	}

	hst.CurrentPhase = PhaseFormation
	hst.CalculateCompletion()
}

// CalculateCompletion updates the completion statistics
func (hst *HeadShouldersThesis) CalculateCompletion() {
	components := hst.GetAllComponents()
	completed := 0
	total := len(components)

	for _, component := range components {
		if component.IsCompleted {
			completed++
		}
	}

	hst.CompletedComponents = completed
	hst.TotalComponents = total
	if total > 0 {
		hst.CompletionPercent = (float64(completed) / float64(total)) * 100
	}
}

// GetAllComponents returns all thesis components
func (hst *HeadShouldersThesis) GetAllComponents() []*ThesisComponent {
	return []*ThesisComponent{
		&hst.LeftShoulderFormed,
		&hst.LeftShoulderVolume,
		&hst.HeadFormed,
		&hst.HeadVolumeSpike,
		&hst.HeadLowerLow,
		&hst.RightShoulderFormed,
		&hst.RightShoulderSymmetry,
		&hst.RightShoulderVolume,
		&hst.NecklineEstablished,
		&hst.NecklineRetest,
		&hst.NecklineBreakout,
		&hst.BreakoutVolume,
		&hst.TargetProjected,
		&hst.PartialFillT1,
		&hst.PartialFillT2,
		&hst.FullTarget,
	}
}

// GetFormationComponents returns formation phase components
func (hst *HeadShouldersThesis) GetFormationComponents() []*ThesisComponent {
	return []*ThesisComponent{
		&hst.LeftShoulderFormed,
		&hst.LeftShoulderVolume,
		&hst.HeadFormed,
		&hst.HeadVolumeSpike,
		&hst.HeadLowerLow,
		&hst.RightShoulderFormed,
		&hst.RightShoulderSymmetry,
		&hst.RightShoulderVolume,
		&hst.NecklineEstablished,
	}
}

// GetBreakoutComponents returns breakout phase components
func (hst *HeadShouldersThesis) GetBreakoutComponents() []*ThesisComponent {
	return []*ThesisComponent{
		&hst.NecklineRetest,
		&hst.NecklineBreakout,
		&hst.BreakoutVolume,
	}
}

// GetTargetComponents returns target phase components
func (hst *HeadShouldersThesis) GetTargetComponents() []*ThesisComponent {
	return []*ThesisComponent{
		&hst.TargetProjected,
		&hst.PartialFillT1,
		&hst.PartialFillT2,
		&hst.FullTarget,
	}
}

// UpdatePhase updates the current phase based on component completion
func (hst *HeadShouldersThesis) UpdatePhase() {
	formationComponents := hst.GetFormationComponents()
	breakoutComponents := hst.GetBreakoutComponents()

	formationComplete := true
	for _, comp := range formationComponents {
		if comp.IsRequired && !comp.IsCompleted {
			formationComplete = false
			break
		}
	}

	breakoutComplete := true
	for _, comp := range breakoutComponents {
		if comp.IsRequired && !comp.IsCompleted {
			breakoutComplete = false
			break
		}
	}

	if !formationComplete {
		hst.CurrentPhase = PhaseFormation
	} else if formationComplete && !breakoutComplete {
		hst.CurrentPhase = PhaseBreakout
	} else if breakoutComplete {
		hst.CurrentPhase = PhaseTargetPursuit
	}
}

// JSON marshaling methods for database storage

// MarshalPatternPointsJSON marshals pattern points to JSON for database storage
func MarshalPatternPointsJSON(pp PatternPoint) ([]byte, error) {
	return json.Marshal(pp)
}

// UnmarshalPatternPointsJSON unmarshals pattern points from JSON
func UnmarshalPatternPointsJSON(data []byte) (PatternPoint, error) {
	var pp PatternPoint
	err := json.Unmarshal(data, &pp)
	return pp, err
}

// MarshalThesisJSON marshals thesis to JSON for database storage
func MarshalThesisJSON(thesis HeadShouldersThesis) ([]byte, error) {
	return json.Marshal(thesis)
}

// UnmarshalThesisJSON unmarshals thesis from JSON
func UnmarshalThesisJSON(data []byte) (HeadShouldersThesis, error) {
	var thesis HeadShouldersThesis
	err := json.Unmarshal(data, &thesis)
	return thesis, err
}
