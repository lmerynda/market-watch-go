package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"sort"
	"time"
)

type TestPriceData struct {
	Timestamp string  `json:"timestamp"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int64   `json:"volume"`
}

type TestData struct {
	Symbol    string          `json:"symbol"`
	PriceData []TestPriceData `json:"price_data"`
}

type PatternPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Price       float64   `json:"price"`
	Volume      int64     `json:"volume"`
	VolumeRatio float64   `json:"volume_ratio"`
}

type FallingWedgeConfig struct {
	MinPatternDuration  time.Duration
	MaxPatternDuration  time.Duration
	MinConvergence      float64
	MaxConvergence      float64
	MinTouchPoints      int
	VolumeDecreaseRatio float64
	BreakoutVolumeRatio float64
	MinWedgeHeight      float64
	MaxWedgeSlope       float64
}

type PatternResult struct {
	UpperLine     []PatternPoint
	LowerLine     []PatternPoint
	UpperSlope    float64
	LowerSlope    float64
	Duration      time.Duration
	Convergence   float64
	Height        float64
	VolumeProfile string
	Valid         bool
	FailReasons   []string
}

func main() {
	// Load test data
	data, err := ioutil.ReadFile("../test_ltbr_data.json")
	if err != nil {
		log.Fatal("Failed to read test data:", err)
	}

	var testData TestData
	if err := json.Unmarshal(data, &testData); err != nil {
		log.Fatal("Failed to parse test data:", err)
	}

	fmt.Printf("🔍 Analyzing %s Falling Wedge Pattern\n", testData.Symbol)
	fmt.Printf("📊 Data Points: %d\n", len(testData.PriceData))
	fmt.Printf("📅 Period: %s to %s\n\n", testData.PriceData[0].Timestamp, testData.PriceData[len(testData.PriceData)-1].Timestamp)

	// Initialize config (same as your detection service)
	config := &FallingWedgeConfig{
		MinPatternDuration:  48 * time.Hour,  // 2 days minimum
		MaxPatternDuration:  480 * time.Hour, // 20 days maximum
		MinConvergence:      0.02,            // 2% minimum convergence
		MaxConvergence:      0.15,            // 15% maximum convergence
		MinTouchPoints:      4,               // Minimum 4 touch points (2 per line)
		VolumeDecreaseRatio: 0.8,             // Volume should decrease to 80% or less
		BreakoutVolumeRatio: 1.5,             // Breakout volume should be 1.5x average
		MinWedgeHeight:      0.03,            // 3% minimum height
		MaxWedgeSlope:       -0.1,            // Maximum downward slope
	}

	// Run analysis
	analyzePattern(testData.PriceData, config)
}

func analyzePattern(priceData []TestPriceData, config *FallingWedgeConfig) {
	fmt.Println("🔍 FALLING WEDGE DETECTION ANALYSIS")
	fmt.Println("=====================================\n")

	// Step 1: Check minimum data requirement
	fmt.Printf("✅ Step 1: Data Requirements\n")
	fmt.Printf("   • Data points: %d (minimum 30 required)\n", len(priceData))
	if len(priceData) < 30 {
		fmt.Printf("   ❌ FAIL: Insufficient data points\n\n")
		return
	}
	fmt.Printf("   ✅ PASS: Sufficient data points\n\n")

	// Step 2: Find significant levels
	fmt.Printf("🎯 Step 2: Finding Significant Highs and Lows\n")
	highs, lows := findSignificantLevels(priceData)
	fmt.Printf("   • Significant highs found: %d\n", len(highs))
	fmt.Printf("   • Significant lows found: %d\n", len(lows))

	if len(highs) < 2 || len(lows) < 2 {
		fmt.Printf("   ❌ FAIL: Need at least 2 highs and 2 lows\n\n")
		return
	}
	fmt.Printf("   ✅ PASS: Sufficient pivot points\n\n")

	// Print the significant levels
	fmt.Printf("📈 Significant Highs:\n")
	for i, high := range highs {
		fmt.Printf("   %d. %s: $%.4f (Vol: %d)\n", i+1, high.Timestamp.Format("2006-01-02 15:04"), high.Price, high.Volume)
	}

	fmt.Printf("\n📉 Significant Lows:\n")
	for i, low := range lows {
		fmt.Printf("   %d. %s: $%.4f (Vol: %d)\n", i+1, low.Timestamp.Format("2006-01-02 15:04"), low.Price, low.Volume)
	}
	fmt.Println()

	// Step 3: Test trend line combinations
	fmt.Printf("📏 Step 3: Testing Trend Line Combinations\n")
	bestPattern := testBestPattern(highs, lows, priceData, config)

	if bestPattern != nil && bestPattern.Valid {
		fmt.Printf("   ✅ FOUND VALID FALLING WEDGE PATTERN!\n")
		printPatternDetails(bestPattern, config)
	} else {
		fmt.Printf("   ❌ NO VALID FALLING WEDGE PATTERN FOUND\n")

		// Debug: Show why combinations failed
		fmt.Printf("\n🔧 DEBUGGING: Why combinations failed...\n")
		debugCombinations(highs, lows, priceData, config)
	}
}

func findSignificantLevels(priceData []TestPriceData) (highs, lows []PatternPoint) {
	windowSize := 5 // Look for highs/lows over 5-period window

	for i := windowSize; i < len(priceData)-windowSize; i++ {
		current := priceData[i]
		isHigh := true
		isLow := true

		// Parse timestamp
		timestamp, _ := time.Parse("2006-01-02T15:04:05Z", current.Timestamp)

		// Check if current point is a significant high or low
		for j := i - windowSize; j <= i+windowSize; j++ {
			if j == i {
				continue
			}

			if priceData[j].High >= current.High {
				isHigh = false
			}
			if priceData[j].Low <= current.Low {
				isLow = false
			}
		}

		if isHigh {
			highs = append(highs, PatternPoint{
				Timestamp:   timestamp,
				Price:       current.High,
				Volume:      current.Volume,
				VolumeRatio: calculateVolumeRatio(priceData, i),
			})
		}

		if isLow {
			lows = append(lows, PatternPoint{
				Timestamp:   timestamp,
				Price:       current.Low,
				Volume:      current.Volume,
				VolumeRatio: calculateVolumeRatio(priceData, i),
			})
		}
	}

	// Sort by timestamp
	sort.Slice(highs, func(i, j int) bool {
		return highs[i].Timestamp.Before(highs[j].Timestamp)
	})
	sort.Slice(lows, func(i, j int) bool {
		return lows[i].Timestamp.Before(lows[j].Timestamp)
	})

	return highs, lows
}

func calculateVolumeRatio(priceData []TestPriceData, index int) float64 {
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

func testBestPattern(highs, lows []PatternPoint, priceData []TestPriceData, config *FallingWedgeConfig) *PatternResult {
	// Look for converging downward trending lines
	for i := 0; i < len(highs)-1; i++ {
		for j := i + 1; j < len(highs); j++ {
			upperLine := []PatternPoint{highs[i], highs[j]}

			for k := 0; k < len(lows)-1; k++ {
				for l := k + 1; l < len(lows); l++ {
					lowerLine := []PatternPoint{lows[k], lows[l]}

					if result := validateFallingWedge(upperLine, lowerLine, priceData, config); result.Valid {
						return result
					}
				}
			}
		}
	}
	return nil
}

func validateFallingWedge(upperLine, lowerLine []PatternPoint, priceData []TestPriceData, config *FallingWedgeConfig) *PatternResult {
	result := &PatternResult{
		UpperLine:   upperLine,
		LowerLine:   lowerLine,
		Valid:       true,
		FailReasons: []string{},
	}

	if len(upperLine) != 2 || len(lowerLine) != 2 {
		result.Valid = false
		result.FailReasons = append(result.FailReasons, "invalid line length")
		return result
	}

	// Calculate slopes
	result.UpperSlope = (upperLine[1].Price - upperLine[0].Price) / float64(upperLine[1].Timestamp.Sub(upperLine[0].Timestamp).Hours())
	result.LowerSlope = (lowerLine[1].Price - lowerLine[0].Price) / float64(lowerLine[1].Timestamp.Sub(lowerLine[0].Timestamp).Hours())

	// Both lines must trend downward
	if result.UpperSlope >= 0 {
		result.Valid = false
		result.FailReasons = append(result.FailReasons, "upper line not falling")
	}
	if result.LowerSlope >= 0 {
		result.Valid = false
		result.FailReasons = append(result.FailReasons, "lower line not falling")
	}

	// Upper line must fall faster than lower line (convergence)
	if result.UpperSlope >= result.LowerSlope {
		result.Valid = false
		result.FailReasons = append(result.FailReasons, "lines not converging (upper not falling faster)")
	}

	// Check pattern duration
	patternStart := upperLine[0].Timestamp
	if lowerLine[0].Timestamp.Before(patternStart) {
		patternStart = lowerLine[0].Timestamp
	}

	patternEnd := upperLine[1].Timestamp
	if lowerLine[1].Timestamp.After(patternEnd) {
		patternEnd = lowerLine[1].Timestamp
	}

	result.Duration = patternEnd.Sub(patternStart)
	if result.Duration < config.MinPatternDuration {
		result.Valid = false
		result.FailReasons = append(result.FailReasons, fmt.Sprintf("duration too short (%.1f hrs < %.1f hrs)", result.Duration.Hours(), config.MinPatternDuration.Hours()))
	}
	if result.Duration > config.MaxPatternDuration {
		result.Valid = false
		result.FailReasons = append(result.FailReasons, fmt.Sprintf("duration too long (%.1f hrs > %.1f hrs)", result.Duration.Hours(), config.MaxPatternDuration.Hours()))
	}

	// Check convergence
	startWidth := math.Abs(upperLine[0].Price - lowerLine[0].Price)
	endWidth := math.Abs(upperLine[1].Price - lowerLine[1].Price)
	result.Convergence = (startWidth - endWidth) / startWidth

	if result.Convergence < config.MinConvergence {
		result.Valid = false
		result.FailReasons = append(result.FailReasons, fmt.Sprintf("convergence too low (%.4f < %.4f)", result.Convergence, config.MinConvergence))
	}
	if result.Convergence > config.MaxConvergence {
		result.Valid = false
		result.FailReasons = append(result.FailReasons, fmt.Sprintf("convergence too high (%.4f > %.4f)", result.Convergence, config.MaxConvergence))
	}

	// Check minimum height
	maxHigh := math.Max(upperLine[0].Price, upperLine[1].Price)
	minLow := math.Min(lowerLine[0].Price, lowerLine[1].Price)
	result.Height = (maxHigh - minLow) / maxHigh

	if result.Height < config.MinWedgeHeight {
		result.Valid = false
		result.FailReasons = append(result.FailReasons, fmt.Sprintf("height too small (%.4f < %.4f)", result.Height, config.MinWedgeHeight))
	}

	return result
}

func debugCombinations(highs, lows []PatternPoint, priceData []TestPriceData, config *FallingWedgeConfig) {
	tested := 0

	for i := 0; i < len(highs)-1; i++ {
		for j := i + 1; j < len(highs); j++ {
			upperLine := []PatternPoint{highs[i], highs[j]}

			for k := 0; k < len(lows)-1; k++ {
				for l := k + 1; l < len(lows); l++ {
					lowerLine := []PatternPoint{lows[k], lows[l]}
					tested++

					result := validateFallingWedge(upperLine, lowerLine, priceData, config)
					if result.Valid {
						fmt.Printf("   ✅ Combination %d: VALID\n", tested)
						printPatternDetails(result, config)
						return
					} else {
						// Show why this combination failed
						debugFailedCombination(tested, result, config)
						if tested >= 10 { // Limit output
							fmt.Printf("   ... (showing first 10 failures)\n")
							break
						}
					}
				}
				if tested >= 10 {
					break
				}
			}
			if tested >= 10 {
				break
			}
		}
		if tested >= 10 {
			break
		}
	}

	fmt.Printf("   📊 Total combinations tested: %d\n", tested)
}

func debugFailedCombination(num int, result *PatternResult, config *FallingWedgeConfig) {
	fmt.Printf("   ❌ Combination %d: FAILED\n", num)
	fmt.Printf("      Upper: $%.4f->$%.4f (slope: %.6f)\n", result.UpperLine[0].Price, result.UpperLine[1].Price, result.UpperSlope)
	fmt.Printf("      Lower: $%.4f->$%.4f (slope: %.6f)\n", result.LowerLine[0].Price, result.LowerLine[1].Price, result.LowerSlope)
	fmt.Printf("      Duration: %.1f hours\n", result.Duration.Hours())
	fmt.Printf("      Convergence: %.4f\n", result.Convergence)
	fmt.Printf("      Height: %.4f\n", result.Height)
	fmt.Printf("      Reasons: %v\n\n", result.FailReasons)
}

func printPatternDetails(result *PatternResult, config *FallingWedgeConfig) {
	fmt.Printf("\n📊 PATTERN DETAILS\n")
	fmt.Printf("==================\n")
	fmt.Printf("Upper trend line: $%.4f -> $%.4f (slope: %.6f/hr)\n",
		result.UpperLine[0].Price, result.UpperLine[1].Price, result.UpperSlope)
	fmt.Printf("Lower trend line: $%.4f -> $%.4f (slope: %.6f/hr)\n",
		result.LowerLine[0].Price, result.LowerLine[1].Price, result.LowerSlope)
	fmt.Printf("Duration: %.1f hours (%.1f days)\n", result.Duration.Hours(), result.Duration.Hours()/24)
	fmt.Printf("Convergence: %.2f%%\n", result.Convergence*100)
	fmt.Printf("Height: %.2f%%\n", result.Height*100)
	fmt.Printf("Volume profile: %s\n", result.VolumeProfile)
}
