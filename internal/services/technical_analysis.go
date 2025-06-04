package services

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"
)

// TechnicalAnalysisService provides technical analysis calculations
type TechnicalAnalysisService struct {
	db           *database.DB
	cache        map[string]*models.TechnicalIndicators
	cacheExpiry  map[string]time.Time
	mutex        sync.RWMutex
	cacheTimeout time.Duration
}

// TechnicalAnalysisConfig holds configuration for technical analysis
type TechnicalAnalysisConfig struct {
	CacheTimeout    time.Duration `yaml:"cache_timeout"`
	RSIPeriods      []int         `yaml:"rsi_periods"`
	RSIOversold     float64       `yaml:"rsi_oversold"`
	RSIOverbought   float64       `yaml:"rsi_overbought"`
	MACDFast        int           `yaml:"macd_fast"`
	MACDSlow        int           `yaml:"macd_slow"`
	MACDSignal      int           `yaml:"macd_signal"`
	MovingAverages  []int         `yaml:"moving_averages"`
	BollingerPeriod int           `yaml:"bollinger_period"`
	BollingerDev    float64       `yaml:"bollinger_deviation"`
	VolumePeriod    int           `yaml:"volume_average_period"`
}

// NewTechnicalAnalysisService creates a new technical analysis service
func NewTechnicalAnalysisService(db *database.DB, config *TechnicalAnalysisConfig) *TechnicalAnalysisService {
	if config == nil {
		config = &TechnicalAnalysisConfig{
			CacheTimeout:    5 * time.Minute,
			RSIPeriods:      []int{14, 30},
			RSIOversold:     30,
			RSIOverbought:   70,
			MACDFast:        12,
			MACDSlow:        26,
			MACDSignal:      9,
			MovingAverages:  []int{20, 50, 200},
			BollingerPeriod: 20,
			BollingerDev:    2.0,
			VolumePeriod:    20,
		}
	}

	return &TechnicalAnalysisService{
		db:           db,
		cache:        make(map[string]*models.TechnicalIndicators),
		cacheExpiry:  make(map[string]time.Time),
		cacheTimeout: config.CacheTimeout,
	}
}

// GetIndicators calculates all technical indicators for a symbol
func (tas *TechnicalAnalysisService) GetIndicators(symbol string) (*models.TechnicalIndicators, error) {
	// Check cache first
	tas.mutex.RLock()
	if cached, exists := tas.cache[symbol]; exists {
		if expiry, hasExpiry := tas.cacheExpiry[symbol]; hasExpiry && time.Now().Before(expiry) {
			tas.mutex.RUnlock()
			return cached, nil
		}
	}
	tas.mutex.RUnlock()

	// Get price data
	priceData, err := tas.db.GetPriceData(&models.PriceDataFilter{
		Symbol: symbol,
		From:   time.Now().AddDate(0, 0, -60), // Last 60 days for better indicators
		To:     time.Now(),
		Limit:  10000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get price data: %w", err)
	}

	if len(priceData) == 0 {
		return nil, fmt.Errorf("no price data available for symbol %s", symbol)
	}

	// Calculate indicators
	indicators := &models.TechnicalIndicators{
		Symbol:    symbol,
		Timestamp: time.Now(),
		CreatedAt: time.Now(),
	}

	// Extract price and volume arrays
	closes := make([]float64, len(priceData))
	highs := make([]float64, len(priceData))
	lows := make([]float64, len(priceData))
	volumes := make([]int64, len(priceData))

	for i, data := range priceData {
		closes[i] = data.Close
		highs[i] = data.High
		lows[i] = data.Low
		volumes[i] = data.Volume
	}

	// Calculate RSI
	if len(closes) >= 14 {
		indicators.RSI14 = tas.calculateRSI(closes, 14)
	}
	if len(closes) >= 30 {
		indicators.RSI30 = tas.calculateRSI(closes, 30)
	}

	// Calculate MACD
	if len(closes) >= 26 {
		macd, signal, histogram := tas.calculateMACD(closes, 12, 26, 9)
		indicators.MACD = macd
		indicators.MACDSignal = signal
		indicators.MACDHistogram = histogram
	}

	// Calculate Moving Averages
	if len(closes) >= 20 {
		indicators.SMA20 = tas.calculateSMA(closes, 20)
		indicators.EMA20 = tas.calculateEMA(closes, 20)
	}
	if len(closes) >= 50 {
		indicators.SMA50 = tas.calculateSMA(closes, 50)
		indicators.EMA50 = tas.calculateEMA(closes, 50)
	}
	if len(closes) >= 200 {
		indicators.SMA200 = tas.calculateSMA(closes, 200)
	}

	// Calculate VWAP
	if len(closes) >= 20 {
		indicators.VWAP = tas.calculateVWAP(highs, lows, closes, volumes, 20)
	}

	// Calculate Bollinger Bands
	if len(closes) >= 20 {
		upper, middle, lower := tas.calculateBollingerBands(closes, 20, 2.0)
		indicators.BBUpper = upper
		indicators.BBMiddle = middle
		indicators.BBLower = lower
	}

	// Calculate Volume Ratio
	indicators.VolumeRatio = tas.calculateVolumeRatio(volumes, 20)

	// Cache the result
	tas.mutex.Lock()
	tas.cache[symbol] = indicators
	tas.cacheExpiry[symbol] = time.Now().Add(tas.cacheTimeout)
	tas.mutex.Unlock()

	return indicators, nil
}

// calculateRSI calculates the Relative Strength Index
func (tas *TechnicalAnalysisService) calculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 0
	}

	gains := make([]float64, 0)
	losses := make([]float64, 0)

	// Calculate price changes
	for i := 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, -change)
		}
	}

	if len(gains) < period {
		return 0
	}

	// Calculate average gain and loss
	avgGain := tas.calculateSMA(gains[len(gains)-period:], period)
	avgLoss := tas.calculateSMA(losses[len(losses)-period:], period)

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// calculateMACD calculates MACD, Signal, and Histogram
func (tas *TechnicalAnalysisService) calculateMACD(prices []float64, fast, slow, signal int) (float64, float64, float64) {
	if len(prices) < slow {
		return 0, 0, 0
	}

	emaFast := tas.calculateEMA(prices, fast)
	emaSlow := tas.calculateEMA(prices, slow)
	macd := emaFast - emaSlow

	// For signal line, we'd need to calculate EMA of MACD values
	// Simplified: using the current MACD value as signal for now
	signalLine := macd * 0.9 // Simplified signal calculation
	histogram := macd - signalLine

	return macd, signalLine, histogram
}

// calculateSMA calculates Simple Moving Average
func (tas *TechnicalAnalysisService) calculateSMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	sum := 0.0
	start := len(prices) - period
	for i := start; i < len(prices); i++ {
		sum += prices[i]
	}

	return sum / float64(period)
}

// calculateEMA calculates Exponential Moving Average
func (tas *TechnicalAnalysisService) calculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	multiplier := 2.0 / float64(period+1)
	ema := tas.calculateSMA(prices[:period], period) // Start with SMA

	for i := period; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

// calculateVWAP calculates Volume Weighted Average Price
func (tas *TechnicalAnalysisService) calculateVWAP(highs, lows, closes []float64, volumes []int64, period int) float64 {
	if len(closes) < period {
		return 0
	}

	start := len(closes) - period
	totalPriceVolume := 0.0
	totalVolume := int64(0)

	for i := start; i < len(closes); i++ {
		typicalPrice := (highs[i] + lows[i] + closes[i]) / 3.0
		totalPriceVolume += typicalPrice * float64(volumes[i])
		totalVolume += volumes[i]
	}

	if totalVolume == 0 {
		return 0
	}

	return totalPriceVolume / float64(totalVolume)
}

// calculateBollingerBands calculates Bollinger Bands
func (tas *TechnicalAnalysisService) calculateBollingerBands(prices []float64, period int, stdDev float64) (float64, float64, float64) {
	if len(prices) < period {
		return 0, 0, 0
	}

	middle := tas.calculateSMA(prices, period)

	// Calculate standard deviation
	start := len(prices) - period
	variance := 0.0
	for i := start; i < len(prices); i++ {
		variance += math.Pow(prices[i]-middle, 2)
	}
	variance /= float64(period)
	standardDev := math.Sqrt(variance)

	upper := middle + (standardDev * stdDev)
	lower := middle - (standardDev * stdDev)

	return upper, middle, lower
}

// calculateVolumeRatio calculates current volume vs average volume
func (tas *TechnicalAnalysisService) calculateVolumeRatio(volumes []int64, period int) float64 {
	if len(volumes) < period {
		return 1.0
	}

	// Calculate average volume
	start := len(volumes) - period
	totalVolume := int64(0)
	for i := start; i < len(volumes)-1; i++ { // Exclude current volume
		totalVolume += volumes[i]
	}

	if totalVolume == 0 {
		return 1.0
	}

	avgVolume := float64(totalVolume) / float64(period-1)
	currentVolume := float64(volumes[len(volumes)-1])

	return currentVolume / avgVolume
}

// GetIndicatorsSummary returns a comprehensive summary of indicators for a symbol
func (tas *TechnicalAnalysisService) GetIndicatorsSummary(symbol string) (*models.IndicatorSummary, error) {
	indicators, err := tas.GetIndicators(symbol)
	if err != nil {
		return nil, err
	}

	// Get current price from latest price data
	latestPrice, err := tas.db.GetLatestPriceData(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest price for %s: %w", symbol, err)
	}

	currentPrice := 0.0
	if latestPrice != nil {
		currentPrice = latestPrice.Close
	}

	summary := &models.IndicatorSummary{
		Symbol:       symbol,
		LastUpdate:   indicators.Timestamp,
		CurrentPrice: currentPrice,
		RSI: &models.RSIData{
			RSI14: indicators.RSI14,
			RSI30: indicators.RSI30,
		},
		MACD: &models.MACDData{
			MACD:      indicators.MACD,
			Signal:    indicators.MACDSignal,
			Histogram: indicators.MACDHistogram,
		},
		MovingAverages: &models.MovingAverageData{
			SMA20:  indicators.SMA20,
			SMA50:  indicators.SMA50,
			SMA200: indicators.SMA200,
			EMA20:  indicators.EMA20,
			EMA50:  indicators.EMA50,
		},
		BollingerBands: &models.BollingerBandsData{
			Upper:  indicators.BBUpper,
			Middle: indicators.BBMiddle,
			Lower:  indicators.BBLower,
		},
		Volume: &models.VolumeAnalysisData{
			VWAP:        indicators.VWAP,
			VolumeRatio: indicators.VolumeRatio,
		},
		TrendDirection:   indicators.GetTrendDirection(),
		OverallSentiment: indicators.GetOverallSentiment(currentPrice),
	}

	return summary, nil
}

// InvalidateCache removes cached indicators for a symbol
func (tas *TechnicalAnalysisService) InvalidateCache(symbol string) {
	tas.mutex.Lock()
	defer tas.mutex.Unlock()

	delete(tas.cache, symbol)
	delete(tas.cacheExpiry, symbol)
	log.Printf("Invalidated cache for symbol: %s", symbol)
}

// ClearExpiredCache removes expired entries from cache
func (tas *TechnicalAnalysisService) ClearExpiredCache() {
	tas.mutex.Lock()
	defer tas.mutex.Unlock()

	now := time.Now()
	for symbol, expiry := range tas.cacheExpiry {
		if now.After(expiry) {
			delete(tas.cache, symbol)
			delete(tas.cacheExpiry, symbol)
			log.Printf("Cleared expired cache for symbol: %s", symbol)
		}
	}
}

// GetCacheStatus returns information about cached indicators
func (tas *TechnicalAnalysisService) GetCacheStatus() map[string]interface{} {
	tas.mutex.RLock()
	defer tas.mutex.RUnlock()

	status := map[string]interface{}{
		"cached_symbols": len(tas.cache),
		"cache_details":  make(map[string]interface{}),
	}

	details := status["cache_details"].(map[string]interface{})
	for symbol, _ := range tas.cache {
		expiry := tas.cacheExpiry[symbol]
		details[symbol] = map[string]interface{}{
			"expires_at":     expiry,
			"time_to_expiry": time.Until(expiry).String(),
		}
	}

	return status
}

// CheckIndicatorAlerts checks if any indicator-based alerts should be triggered
func (tas *TechnicalAnalysisService) CheckIndicatorAlerts(symbol string, thresholds *models.IndicatorThresholds) ([]*models.IndicatorAlert, error) {
	indicators, err := tas.GetIndicators(symbol)
	if err != nil {
		return nil, err
	}

	var alerts []*models.IndicatorAlert
	now := time.Now()

	// RSI Oversold Alert
	if indicators.IsRSIOversold(thresholds.RSIOversold) {
		alerts = append(alerts, &models.IndicatorAlert{
			Symbol:      symbol,
			AlertType:   "rsi_oversold",
			Indicator:   "RSI14",
			Value:       indicators.RSI14,
			Threshold:   thresholds.RSIOversold,
			Message:     fmt.Sprintf("RSI14 (%.2f) is below oversold threshold (%.2f)", indicators.RSI14, thresholds.RSIOversold),
			TriggeredAt: now,
			IsActive:    true,
			CreatedAt:   now,
		})
	}

	// RSI Overbought Alert
	if indicators.IsRSIOverbought(thresholds.RSIOverbought) {
		alerts = append(alerts, &models.IndicatorAlert{
			Symbol:      symbol,
			AlertType:   "rsi_overbought",
			Indicator:   "RSI14",
			Value:       indicators.RSI14,
			Threshold:   thresholds.RSIOverbought,
			Message:     fmt.Sprintf("RSI14 (%.2f) is above overbought threshold (%.2f)", indicators.RSI14, thresholds.RSIOverbought),
			TriggeredAt: now,
			IsActive:    true,
			CreatedAt:   now,
		})
	}

	// MACD Bullish Alert
	if thresholds.MACDBullish && indicators.IsMACDBullish() {
		alerts = append(alerts, &models.IndicatorAlert{
			Symbol:      symbol,
			AlertType:   "macd_bullish",
			Indicator:   "MACD",
			Value:       indicators.MACDHistogram,
			Threshold:   0,
			Message:     fmt.Sprintf("MACD shows bullish signal (MACD: %.4f > Signal: %.4f)", indicators.MACD, indicators.MACDSignal),
			TriggeredAt: now,
			IsActive:    true,
			CreatedAt:   now,
		})
	}

	// MACD Bearish Alert
	if thresholds.MACDBearish && indicators.IsMACDBearish() {
		alerts = append(alerts, &models.IndicatorAlert{
			Symbol:      symbol,
			AlertType:   "macd_bearish",
			Indicator:   "MACD",
			Value:       indicators.MACDHistogram,
			Threshold:   0,
			Message:     fmt.Sprintf("MACD shows bearish signal (MACD: %.4f < Signal: %.4f)", indicators.MACD, indicators.MACDSignal),
			TriggeredAt: now,
			IsActive:    true,
			CreatedAt:   now,
		})
	}

	// Volume Spike Alert
	if indicators.IsVolumeSpike(thresholds.VolumeSpike) {
		alerts = append(alerts, &models.IndicatorAlert{
			Symbol:      symbol,
			AlertType:   "volume_spike",
			Indicator:   "VolumeRatio",
			Value:       indicators.VolumeRatio,
			Threshold:   thresholds.VolumeSpike,
			Message:     fmt.Sprintf("Volume spike detected (%.1fx above average)", indicators.VolumeRatio),
			TriggeredAt: now,
			IsActive:    true,
			CreatedAt:   now,
		})
	}

	return alerts, nil
}

// UpdateIndicatorsForSymbol updates technical indicators for a specific symbol
func (tas *TechnicalAnalysisService) UpdateIndicatorsForSymbol(symbol string) error {
	// Invalidate cache to force fresh calculation
	tas.InvalidateCache(symbol)

	// Calculate fresh indicators
	indicators, err := tas.GetIndicators(symbol)
	if err != nil {
		return fmt.Errorf("failed to calculate indicators for %s: %w", symbol, err)
	}

	log.Printf("Updated technical indicators for %s: RSI14=%.2f, MACD=%.4f, SMA20=%.2f",
		symbol, indicators.RSI14, indicators.MACD, indicators.SMA20)

	return nil
}
