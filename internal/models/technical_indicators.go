package models

import (
	"time"
)

// TechnicalIndicators represents all calculated technical indicators for a symbol
type TechnicalIndicators struct {
	ID            int64     `json:"id" db:"id"`
	Symbol        string    `json:"symbol" db:"symbol"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	RSI14         float64   `json:"rsi_14" db:"rsi_14"`
	RSI30         float64   `json:"rsi_30" db:"rsi_30"`
	MACD          float64   `json:"macd" db:"macd_line"`
	MACDSignal    float64   `json:"macd_signal" db:"macd_signal"`
	MACDHistogram float64   `json:"macd_histogram" db:"macd_histogram"`
	
	EMA20         float64   `json:"ema_20" db:"ema_20"`
	EMA50         float64   `json:"ema_50" db:"ema_50"`
	VWAP          float64   `json:"vwap" db:"vwap"`
	VolumeRatio   float64   `json:"volume_ratio" db:"volume_ratio"`
	BBUpper       float64   `json:"bb_upper" db:"bb_upper"`
	BBMiddle      float64   `json:"bb_middle" db:"bb_middle"`
	BBLower       float64   `json:"bb_lower" db:"bb_lower"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// TechnicalIndicatorsResponse represents the API response for technical indicators
type TechnicalIndicatorsResponse struct {
	Symbol     string               `json:"symbol"`
	Indicators *TechnicalIndicators `json:"indicators"`
	Status     string               `json:"status"`
	Message    string               `json:"message,omitempty"`
}

// MACDData represents MACD indicator data
type MACDData struct {
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
}

// RSIData represents RSI indicator data
type RSIData struct {
	RSI14 float64 `json:"rsi_14"`
	RSI30 float64 `json:"rsi_30"`
}

// MovingAverageData represents moving average data
type MovingAverageData struct {
	
	EMA20  float64 `json:"ema_20"`
	EMA50  float64 `json:"ema_50"`
}

// BollingerBandsData represents Bollinger Bands data
type BollingerBandsData struct {
	Upper  float64 `json:"upper"`
	Middle float64 `json:"middle"`
	Lower  float64 `json:"lower"`
}

// VolumeAnalysisData represents volume analysis data (renamed to avoid conflict)
type VolumeAnalysisData struct {
	VWAP        float64 `json:"vwap"`
	VolumeRatio float64 `json:"volume_ratio"`
	AvgVolume   float64 `json:"avg_volume"`
}

// IndicatorFilter represents filter parameters for querying indicators
type IndicatorFilter struct {
	Symbol string    `json:"symbol"`
	From   time.Time `json:"from"`
	To     time.Time `json:"to"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

// IndicatorAlert represents an alert based on technical indicators
type IndicatorAlert struct {
	ID          int64     `json:"id" db:"id"`
	Symbol      string    `json:"symbol" db:"symbol"`
	AlertType   string    `json:"alert_type" db:"alert_type"` // 'rsi_oversold', 'rsi_overbought', 'macd_bullish', 'macd_bearish'
	Indicator   string    `json:"indicator" db:"indicator"`
	Value       float64   `json:"value" db:"value"`
	Threshold   float64   `json:"threshold" db:"threshold"`
	Message     string    `json:"message" db:"message"`
	TriggeredAt time.Time `json:"triggered_at" db:"triggered_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// IndicatorThresholds represents alert thresholds for indicators
type IndicatorThresholds struct {
	RSIOversold   float64 `json:"rsi_oversold" yaml:"rsi_oversold"`
	RSIOverbought float64 `json:"rsi_overbought" yaml:"rsi_overbought"`
	VolumeSpike   float64 `json:"volume_spike" yaml:"volume_spike"`
	MACDBullish   bool    `json:"macd_bullish" yaml:"macd_bullish"`
	MACDBearish   bool    `json:"macd_bearish" yaml:"macd_bearish"`
	BBOverbought  bool    `json:"bb_overbought" yaml:"bb_overbought"`
	BBOversold    bool    `json:"bb_oversold" yaml:"bb_oversold"`
}

// IndicatorSummary represents a summary of all indicators for a symbol
type IndicatorSummary struct {
	Symbol           string              `json:"symbol"`
	LastUpdate       time.Time           `json:"last_update"`
	CurrentPrice     float64             `json:"current_price"`
	RSI              *RSIData            `json:"rsi"`
	MACD             *MACDData           `json:"macd"`
	MovingAverages   *MovingAverageData  `json:"moving_averages"`
	BollingerBands   *BollingerBandsData `json:"bollinger_bands"`
	Volume           *VolumeAnalysisData `json:"volume"`
	ActiveAlerts     []*IndicatorAlert   `json:"active_alerts"`
	TrendDirection   string              `json:"trend_direction"`   // 'bullish', 'bearish', 'neutral'
	OverallSentiment string              `json:"overall_sentiment"` // 'strong_buy', 'buy', 'neutral', 'sell', 'strong_sell'
}

// GetTrendDirection analyzes indicators to determine trend direction
func (ti *TechnicalIndicators) GetTrendDirection() string {
	bullishSignals := 0
	bearishSignals := 0

	// RSI analysis
	if ti.RSI14 > 50 {
		bullishSignals++
	} else {
		bearishSignals++
	}

	// MACD analysis
	if ti.MACD > ti.MACDSignal {
		bullishSignals++
	} else {
		bearishSignals++
	}

	// Moving average analysis (price vs EMA20)
	// Note: This would need current price which we don't have in this struct
	// Will be implemented in the service layer

	if bullishSignals > bearishSignals {
		return "bullish"
	} else if bearishSignals > bullishSignals {
		return "bearish"
	}
	return "neutral"
}

// GetOverallSentiment provides an overall sentiment based on multiple indicators
func (ti *TechnicalIndicators) GetOverallSentiment(currentPrice float64) string {
	score := 0

	// RSI scoring
	if ti.RSI14 < 30 {
		score += 2 // Strong buy signal
	} else if ti.RSI14 < 50 {
		score += 1 // Buy signal
	} else if ti.RSI14 > 70 {
		score -= 2 // Strong sell signal
	} else if ti.RSI14 > 50 {
		score -= 1 // Sell signal
	}

	// MACD scoring
	if ti.MACD > ti.MACDSignal && ti.MACDHistogram > 0 {
		score += 2 // Strong bullish
	} else if ti.MACD > ti.MACDSignal {
		score += 1 // Bullish
	} else if ti.MACD < ti.MACDSignal && ti.MACDHistogram < 0 {
		score -= 2 // Strong bearish
	} else {
		score -= 1 // Bearish
	}

	// Moving average scoring
	if currentPrice > ti.EMA20 && ti.EMA20 > ti.EMA50 {
		score += 1 // Bullish trend
	} else if currentPrice < ti.EMA20 && ti.EMA20 < ti.EMA50 {
		score -= 1 // Bearish trend
	}

	// Bollinger Bands scoring
	if currentPrice < ti.BBLower {
		score += 1 // Oversold
	} else if currentPrice > ti.BBUpper {
		score -= 1 // Overbought
	}

	// Convert score to sentiment
	switch {
	case score >= 4:
		return "strong_buy"
	case score >= 2:
		return "buy"
	case score <= -4:
		return "strong_sell"
	case score <= -2:
		return "sell"
	default:
		return "neutral"
	}
}

// IsRSIOversold checks if RSI indicates oversold condition
func (ti *TechnicalIndicators) IsRSIOversold(threshold float64) bool {
	return ti.RSI14 < threshold
}

// IsRSIOverbought checks if RSI indicates overbought condition
func (ti *TechnicalIndicators) IsRSIOverbought(threshold float64) bool {
	return ti.RSI14 > threshold
}

// IsMACDBullish checks if MACD shows bullish signal
func (ti *TechnicalIndicators) IsMACDBullish() bool {
	return ti.MACD > ti.MACDSignal && ti.MACDHistogram > 0
}

// IsMACDBearish checks if MACD shows bearish signal
func (ti *TechnicalIndicators) IsMACDBearish() bool {
	return ti.MACD < ti.MACDSignal && ti.MACDHistogram < 0
}

// IsVolumeSpike checks if volume is significantly above average
func (ti *TechnicalIndicators) IsVolumeSpike(threshold float64) bool {
	return ti.VolumeRatio > threshold
}
