package services

import (
	"testing"
	"time"

	"market-watch-go/internal/models"
)

// TestFallingWedgeDetection_LTBR tests the falling wedge detection with real LTBR data
func TestFallingWedgeDetection_LTBR(t *testing.T) {
	// Create test service with fixed configuration
	service := &FallingWedgeDetectionService{
		config: &models.FallingWedgeConfig{
			MinPatternDuration:  48 * time.Hour,  // 2 days minimum
			MaxPatternDuration:  480 * time.Hour, // 20 days maximum
			MinConvergence:      0.005,           // 0.5% minimum convergence (FIXED)
			MaxConvergence:      0.95,            // INCREASED to 95% to allow high convergence
			MinTouchPoints:      4,               // Minimum 4 touch points
			VolumeDecreaseRatio: 0.8,             // Volume should decrease to 80% or less
			BreakoutVolumeRatio: 1.5,             // Breakout volume should be 1.5x average
			MinWedgeHeight:      0.03,            // 3% minimum height
			MaxWedgeSlope:       -0.1,            // Maximum downward slope
		},
	}

	// LTBR test data from June 13-18, 2025 - EXPANDED with more points
	testData := []*models.PriceData{
		// June 13 - Initial decline
		{Timestamp: parseTime("2025-06-13T08:00:00Z"), High: 13.67, Low: 13.26, Close: 13.26, Volume: 7028},
		{Timestamp: parseTime("2025-06-13T08:30:00Z"), High: 13.29, Low: 13.23, Close: 13.23, Volume: 364},
		{Timestamp: parseTime("2025-06-13T09:00:00Z"), High: 13.27, Low: 13.25, Close: 13.25, Volume: 280},
		{Timestamp: parseTime("2025-06-13T09:30:00Z"), High: 13.25, Low: 13.20, Close: 13.20, Volume: 1963},
		{Timestamp: parseTime("2025-06-13T10:00:00Z"), High: 13.10, Low: 13.01, Close: 13.10, Volume: 2417},
		{Timestamp: parseTime("2025-06-13T10:30:00Z"), High: 12.81, Low: 12.68, Close: 12.69, Volume: 3830},
		{Timestamp: parseTime("2025-06-13T11:00:00Z"), High: 12.64, Low: 12.60, Close: 12.63, Volume: 21751}, // Low point
		{Timestamp: parseTime("2025-06-13T11:30:00Z"), High: 12.77, Low: 12.68, Close: 12.68, Volume: 2923},
		{Timestamp: parseTime("2025-06-13T12:00:00Z"), High: 12.88, Low: 12.62, Close: 12.81, Volume: 43116},
		{Timestamp: parseTime("2025-06-13T12:30:00Z"), High: 12.82, Low: 12.73, Close: 12.82, Volume: 5095},
		{Timestamp: parseTime("2025-06-13T13:00:00Z"), High: 12.89, Low: 12.69, Close: 12.69, Volume: 500},
		{Timestamp: parseTime("2025-06-13T13:30:00Z"), High: 13.15, Low: 12.65, Close: 12.93, Volume: 207232}, // Recovery
		{Timestamp: parseTime("2025-06-13T14:00:00Z"), High: 13.27, Low: 13.07, Close: 13.26, Volume: 13268},
		{Timestamp: parseTime("2025-06-13T14:30:00Z"), High: 13.10, Low: 12.93, Close: 12.97, Volume: 12805},
		{Timestamp: parseTime("2025-06-13T15:00:00Z"), High: 13.13, Low: 13.06, Close: 13.09, Volume: 6634},
		{Timestamp: parseTime("2025-06-13T15:30:00Z"), High: 13.41, Low: 13.36, Close: 13.38, Volume: 5051},
		{Timestamp: parseTime("2025-06-13T16:00:00Z"), High: 13.51, Low: 13.44, Close: 13.5, Volume: 15479}, // Higher high
		{Timestamp: parseTime("2025-06-13T16:30:00Z"), High: 13.55, Low: 13.47, Close: 13.50, Volume: 12204},
		{Timestamp: parseTime("2025-06-13T17:00:00Z"), High: 13.29, Low: 13.21, Close: 13.22, Volume: 6782},
		{Timestamp: parseTime("2025-06-13T17:30:00Z"), High: 13.04, Low: 13.00, Close: 13.02, Volume: 22541},
		{Timestamp: parseTime("2025-06-13T18:00:00Z"), High: 13.08, Low: 13.00, Close: 13.00, Volume: 4924},
		{Timestamp: parseTime("2025-06-13T18:30:00Z"), High: 12.87, Low: 12.76, Close: 12.78, Volume: 39524},
		{Timestamp: parseTime("2025-06-13T19:00:00Z"), High: 13.00, Low: 12.91, Close: 12.98, Volume: 2804},
		{Timestamp: parseTime("2025-06-13T19:30:00Z"), High: 12.97, Low: 12.89, Close: 12.89, Volume: 9183},
		{Timestamp: parseTime("2025-06-13T20:00:00Z"), High: 12.90, Low: 12.90, Close: 12.90, Volume: 7891},

		// June 16 - Breakout attempt
		{Timestamp: parseTime("2025-06-16T09:00:00Z"), High: 13.25, Low: 13.25, Close: 13.25, Volume: 281},
		{Timestamp: parseTime("2025-06-16T09:30:00Z"), High: 13.50, Low: 13.35, Close: 13.50, Volume: 1201}, // Higher low
		{Timestamp: parseTime("2025-06-16T10:00:00Z"), High: 13.40, Low: 13.32, Close: 13.33, Volume: 938},
		{Timestamp: parseTime("2025-06-16T10:30:00Z"), High: 13.35, Low: 13.25, Close: 13.30, Volume: 1642},
		{Timestamp: parseTime("2025-06-16T11:00:00Z"), High: 13.33, Low: 13.14, Close: 13.15, Volume: 689},
		{Timestamp: parseTime("2025-06-16T11:30:00Z"), High: 13.29, Low: 13.26, Close: 13.29, Volume: 1844},
		{Timestamp: parseTime("2025-06-16T12:00:00Z"), High: 13.50, Low: 13.15, Close: 13.50, Volume: 8474},
		{Timestamp: parseTime("2025-06-16T12:30:00Z"), High: 13.61, Low: 13.51, Close: 13.51, Volume: 2638},
		{Timestamp: parseTime("2025-06-16T13:00:00Z"), High: 13.63, Low: 13.53, Close: 13.63, Volume: 4922},
		{Timestamp: parseTime("2025-06-16T13:30:00Z"), High: 14.13, Low: 13.62, Close: 14.02, Volume: 181860}, // Spike high
		{Timestamp: parseTime("2025-06-16T14:00:00Z"), High: 14.18, Low: 13.86, Close: 14.08, Volume: 71330},
		{Timestamp: parseTime("2025-06-16T14:30:00Z"), High: 14.29, Low: 13.92, Close: 14.23, Volume: 41388}, // Highest point
		{Timestamp: parseTime("2025-06-16T15:00:00Z"), High: 14.11, Low: 13.95, Close: 13.95, Volume: 28087},
		{Timestamp: parseTime("2025-06-16T15:30:00Z"), High: 14.06, Low: 13.90, Close: 14.01, Volume: 28548},
		{Timestamp: parseTime("2025-06-16T16:00:00Z"), High: 14.17, Low: 13.99, Close: 14.16, Volume: 11166},
		{Timestamp: parseTime("2025-06-16T16:30:00Z"), High: 13.96, Low: 13.90, Close: 13.96, Volume: 17726},
		{Timestamp: parseTime("2025-06-16T17:00:00Z"), High: 13.97, Low: 13.87, Close: 13.97, Volume: 3387},
		{Timestamp: parseTime("2025-06-16T17:30:00Z"), High: 13.89, Low: 13.82, Close: 13.82, Volume: 6899},
		{Timestamp: parseTime("2025-06-16T18:00:00Z"), High: 13.73, Low: 13.63, Close: 13.73, Volume: 7448},
		{Timestamp: parseTime("2025-06-16T18:30:00Z"), High: 13.55, Low: 13.54, Close: 13.54, Volume: 3220},
		{Timestamp: parseTime("2025-06-16T19:00:00Z"), High: 13.55, Low: 13.50, Close: 13.55, Volume: 8618},
		{Timestamp: parseTime("2025-06-16T19:30:00Z"), High: 13.33, Low: 13.25, Close: 13.30, Volume: 12979},
		{Timestamp: parseTime("2025-06-16T20:00:00Z"), High: 13.60, Low: 13.60, Close: 13.60, Volume: 200},

		// June 17 - Continued decline
		{Timestamp: parseTime("2025-06-17T08:00:00Z"), High: 13.70, Low: 13.58, Close: 13.58, Volume: 504},
		{Timestamp: parseTime("2025-06-17T09:00:00Z"), High: 13.48, Low: 13.43, Close: 13.43, Volume: 265},
		{Timestamp: parseTime("2025-06-17T10:00:00Z"), High: 13.76, Low: 13.75, Close: 13.84, Volume: 1250},
		{Timestamp: parseTime("2025-06-17T11:00:00Z"), High: 13.65, Low: 13.61, Close: 13.61, Volume: 466},
		{Timestamp: parseTime("2025-06-17T12:00:00Z"), High: 13.90, Low: 13.85, Close: 13.90, Volume: 1160},
		{Timestamp: parseTime("2025-06-17T12:30:00Z"), High: 14.00, Low: 13.93, Close: 13.96, Volume: 1552}, // Local high
		{Timestamp: parseTime("2025-06-17T13:00:00Z"), High: 13.86, Low: 13.78, Close: 13.80, Volume: 631},
		{Timestamp: parseTime("2025-06-17T13:30:00Z"), High: 13.76, Low: 13.54, Close: 13.75, Volume: 65019}, // Break down
		{Timestamp: parseTime("2025-06-17T14:00:00Z"), High: 13.45, Low: 13.34, Close: 13.35, Volume: 13420},
		{Timestamp: parseTime("2025-06-17T14:30:00Z"), High: 13.65, Low: 13.50, Close: 13.53, Volume: 18973},
		{Timestamp: parseTime("2025-06-17T15:00:00Z"), High: 13.35, Low: 13.27, Close: 13.27, Volume: 5786},
		{Timestamp: parseTime("2025-06-17T15:30:00Z"), High: 13.25, Low: 13.22, Close: 13.25, Volume: 17221},
		{Timestamp: parseTime("2025-06-17T16:00:00Z"), High: 13.32, Low: 13.28, Close: 13.28, Volume: 9558},
		{Timestamp: parseTime("2025-06-17T16:30:00Z"), High: 13.19, Low: 13.16, Close: 13.18, Volume: 4387},
		{Timestamp: parseTime("2025-06-17T17:00:00Z"), High: 13.10, Low: 13.08, Close: 13.10, Volume: 5029},
		{Timestamp: parseTime("2025-06-17T17:30:00Z"), High: 13.04, Low: 12.95, Close: 13.04, Volume: 6493},
		{Timestamp: parseTime("2025-06-17T18:00:00Z"), High: 13.13, Low: 13.09, Close: 13.10, Volume: 3372},
		{Timestamp: parseTime("2025-06-17T18:30:00Z"), High: 13.02, Low: 12.96, Close: 13.00, Volume: 12112},
		{Timestamp: parseTime("2025-06-17T19:00:00Z"), High: 13.02, Low: 12.98, Close: 12.99, Volume: 5189},
		{Timestamp: parseTime("2025-06-17T19:30:00Z"), High: 13.06, Low: 13.04, Close: 13.06, Volume: 4367},
		{Timestamp: parseTime("2025-06-17T20:00:00Z"), High: 13.03, Low: 13.03, Close: 13.03, Volume: 937},

		// June 18 - Final decline
		{Timestamp: parseTime("2025-06-18T08:00:00Z"), High: 12.96, Low: 12.96, Close: 12.96, Volume: 7077}, // Lower low
		{Timestamp: parseTime("2025-06-18T09:00:00Z"), High: 13.15, Low: 13.15, Close: 13.15, Volume: 884},
		{Timestamp: parseTime("2025-06-18T10:00:00Z"), High: 13.19, Low: 13.15, Close: 13.15, Volume: 438},
		{Timestamp: parseTime("2025-06-18T11:00:00Z"), High: 13.07, Low: 13.00, Close: 13.00, Volume: 1307},
		{Timestamp: parseTime("2025-06-18T12:00:00Z"), High: 13.26, Low: 13.25, Close: 13.25, Volume: 100},
		{Timestamp: parseTime("2025-06-18T13:00:00Z"), High: 13.05, Low: 13.00, Close: 13.00, Volume: 2934},
		{Timestamp: parseTime("2025-06-18T13:30:00Z"), High: 13.16, Low: 12.75, Close: 12.85, Volume: 69481}, // Sharp drop
		{Timestamp: parseTime("2025-06-18T14:00:00Z"), High: 12.95, Low: 12.85, Close: 12.91, Volume: 21553},
		{Timestamp: parseTime("2025-06-18T14:30:00Z"), High: 13.25, Low: 13.19, Close: 13.23, Volume: 14661},
		{Timestamp: parseTime("2025-06-18T15:00:00Z"), High: 13.30, Low: 13.19, Close: 13.28, Volume: 13437}, // Recovery attempt
		{Timestamp: parseTime("2025-06-18T15:30:00Z"), High: 13.24, Low: 13.15, Close: 13.19, Volume: 11279},
		{Timestamp: parseTime("2025-06-18T16:00:00Z"), High: 13.04, Low: 12.97, Close: 13.04, Volume: 9707},
		{Timestamp: parseTime("2025-06-18T16:30:00Z"), High: 13.08, Low: 13.00, Close: 13.07, Volume: 8001},
		{Timestamp: parseTime("2025-06-18T17:00:00Z"), High: 13.17, Low: 13.13, Close: 13.14, Volume: 18702},
		{Timestamp: parseTime("2025-06-18T17:30:00Z"), High: 13.08, Low: 13.01, Close: 13.06, Volume: 11546},
		{Timestamp: parseTime("2025-06-18T18:00:00Z"), High: 13.13, Low: 13.00, Close: 13.00, Volume: 6805},
		{Timestamp: parseTime("2025-06-18T18:30:00Z"), High: 13.16, Low: 13.11, Close: 13.12, Volume: 4282},
		{Timestamp: parseTime("2025-06-18T19:00:00Z"), High: 12.88, Low: 12.80, Close: 12.88, Volume: 4563},
		{Timestamp: parseTime("2025-06-18T19:30:00Z"), High: 13.00, Low: 12.87, Close: 12.99, Volume: 7852},
		{Timestamp: parseTime("2025-06-18T20:00:00Z"), High: 13.09, Low: 13.01, Close: 13.02, Volume: 67144}, // End point
	}

	t.Run("AnalyzeFallingWedgePattern", func(t *testing.T) {
		pattern := service.analyzeFallingWedgePattern("LTBR", testData)

		if pattern == nil {
			t.Error("Expected falling wedge pattern to be detected for LTBR, but got nil")
			return
		}

		// Validate pattern properties
		if pattern.Symbol != "LTBR" {
			t.Errorf("Expected symbol LTBR, got %s", pattern.Symbol)
		}

		if pattern.PatternType != "falling_wedge" {
			t.Errorf("Expected pattern type 'falling_wedge', got %s", pattern.PatternType)
		}

		if pattern.UpperSlope >= 0 {
			t.Errorf("Upper slope should be negative, got %f", pattern.UpperSlope)
		}

		if pattern.LowerSlope <= 0 {
			t.Errorf("Lower slope should be positive, got %f", pattern.LowerSlope)
		}

		t.Logf("✅ LTBR Falling Wedge Pattern Detected!")
		t.Logf("   Upper slope: %f", pattern.UpperSlope)
		t.Logf("   Lower slope: %f", pattern.LowerSlope)
		t.Logf("   Convergence: %.2f%%", pattern.Convergence)
		t.Logf("   Height: %.2f%%", (pattern.PatternHeight/pattern.BreakoutLevel)*100)
		t.Logf("   Duration: %.1f hours", float64(pattern.PatternWidth)/60.0)
		t.Logf("   Breakout level: $%.4f", pattern.BreakoutLevel)
	})
}

// Helper function to parse time strings
func parseTime(timeStr string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05Z", timeStr)
	if err != nil {
		panic(err)
	}
	return t
}
