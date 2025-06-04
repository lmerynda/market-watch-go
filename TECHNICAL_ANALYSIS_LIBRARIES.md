# Technical Analysis Libraries for Market Watch Go

## 📚 Recommended Libraries

Based on research and your suggestion, here are the best Go libraries for technical analysis that we should integrate:

### 1. **github.com/sdcoffey/techan** ⭐ (Primary Choice)
- **Features**: Comprehensive technical analysis library
- **Indicators**: RSI, MACD, Moving Averages, Bollinger Bands, Williams %R, Stochastic, etc.
- **Patterns**: Support/Resistance detection, candlestick patterns
- **Data Structure**: Time series with OHLCV data
- **Community**: Well-maintained with good documentation

```go
import "github.com/sdcoffey/techan"

// Example usage:
series := techan.NewTimeSeries()
rsi := techan.NewRelativeStrengthIndexIndicator(techan.NewClosePriceIndicator(series), 14)
macd := techan.NewMACDIndicator(techan.NewClosePriceIndicator(series), 12, 26)
```

### 2. **github.com/cinar/indicator** ⭐ (Secondary/Backup)
- **Features**: Pure Go implementation, no dependencies
- **Indicators**: 40+ technical indicators
- **Performance**: Optimized for high-frequency calculations
- **API**: Simple and clean interface

```go
import "github.com/cinar/indicator"

rsi := indicator.Rsi(14, closes)
macd, signal, histogram := indicator.Macd(12, 26, 9, closes)
```

### 3. **github.com/markcheno/go-talib** (If needed)
- **Features**: Go wrapper for the famous TA-Lib C library
- **Pros**: Industry standard, extensive indicators
- **Cons**: CGO dependency, more complex setup
- **Use Case**: Only if we need very specific indicators not available elsewhere

## 🔄 Updated Implementation Strategy

### Library Integration Plan

1. **Primary**: Use `techan` for most indicators and pattern detection
2. **Fallback**: Use `cinar/indicator` for any missing functionality
3. **Custom**: Only implement custom logic for our specific S/R detection algorithm

### Updated Phase 1: Technical Analysis Foundation

#### Week 1: Library Integration & Core Indicators
- [ ] Add `github.com/sdcoffey/techan` to go.mod
- [ ] Add `github.com/cinar/indicator` as backup
- [ ] Create data conversion utilities (our PriceData → techan.TimeSeries)
- [ ] Implement RSI calculation service using techan
- [ ] Implement MACD calculation using techan
- [ ] Implement Moving Averages (SMA/EMA) using techan
- [ ] Create volume analysis service (ratios, VWAP)
- [ ] Add comprehensive unit tests for all indicators

#### Week 2: Support/Resistance Detection
- [ ] Research techan's support/resistance capabilities
- [ ] Implement custom S/R detection algorithm (if techan doesn't have exactly what we need)
- [ ] Create level strength scoring algorithm
- [ ] Add level validation and filtering
- [ ] Create S/R database schema and models
- [ ] Integrate S/R detection with data collection pipeline

### Code Structure with Libraries

```go
// internal/services/technical_analysis.go
type TechnicalAnalysisService struct {
    // TimeSeries cache for each symbol
    seriesCache map[string]*techan.TimeSeries
    mutex       sync.RWMutex
}

func (tas *TechnicalAnalysisService) CalculateRSI(symbol string, period int) (float64, error) {
    series := tas.getTimeSeries(symbol)
    rsi := techan.NewRelativeStrengthIndexIndicator(
        techan.NewClosePriceIndicator(series), 
        period,
    )
    return rsi.Calculate(series.LastIndex()).Float(), nil
}

func (tas *TechnicalAnalysisService) CalculateMACD(symbol string) (*MACDResult, error) {
    series := tas.getTimeSeries(symbol)
    macd := techan.NewMACDIndicator(
        techan.NewClosePriceIndicator(series), 
        12, 26,
    )
    signal := techan.NewEMAIndicator(macd, 9)
    
    lastIndex := series.LastIndex()
    return &MACDResult{
        MACD:      macd.Calculate(lastIndex).Float(),
        Signal:    signal.Calculate(lastIndex).Float(),
        Histogram: macd.Calculate(lastIndex).Sub(signal.Calculate(lastIndex)).Float(),
    }, nil
}
```

### Data Conversion Utilities

```go
// internal/services/data_converter.go
func ConvertToTechanSeries(priceData []*models.PriceData) *techan.TimeSeries {
    series := techan.NewTimeSeries()
    
    for _, data := range priceData {
        candle := techan.NewCandle(techan.NewTimePeriod(data.Timestamp, time.Minute))
        candle.OpenPrice = big.NewFromFloat(data.Open)
        candle.HighPrice = big.NewFromFloat(data.High)
        candle.LowPrice = big.NewFromFloat(data.Low)
        candle.ClosePrice = big.NewFromFloat(data.Close)
        candle.Volume = big.NewFromInt(data.Volume)
        
        series.AddCandle(candle)
    }
    
    return series
}
```

## 📦 Dependency Management

### Updated go.mod
```go
module market-watch-go

go 1.23.8

require (
    github.com/gin-gonic/gin v1.10.1
    github.com/mattn/go-sqlite3 v1.14.28
    github.com/robfig/cron/v3 v3.0.1
    github.com/sdcoffey/techan v0.12.1  // Primary TA library
    github.com/cinar/indicator v1.3.0   // Backup TA library
    gopkg.in/yaml.v3 v3.0.1
)
```

## 🎯 Benefits of Using Libraries

### Advantages
✅ **Battle-tested algorithms** - Industry standard implementations
✅ **Time savings** - No need to implement complex mathematical formulas
✅ **Accuracy** - Peer-reviewed and widely used calculations
✅ **Maintenance** - Libraries handle edge cases and optimizations
✅ **Documentation** - Well-documented APIs with examples
✅ **Performance** - Optimized implementations

### What We Still Need to Implement
🔧 **Custom S/R Detection** - Our specific bounce detection algorithm
🔧 **Setup Scoring System** - Our 100-point quality scoring
🔧 **Pattern Recognition** - Specific bounce patterns we want to detect
🔧 **Alert Logic** - Integration with our notification system
🔧 **Database Integration** - Storing and retrieving calculated indicators

## 🚀 Quick Start Integration

### Step 1: Add Dependencies
```bash
cd market-watch-go
go get github.com/sdcoffey/techan@latest
go get github.com/cinar/indicator@latest
go mod tidy
```

### Step 2: Create Technical Analysis Service
```go
// internal/services/technical_analysis.go
type TechnicalAnalysisService struct {
    db           *database.DB
    seriesCache  map[string]*techan.TimeSeries
    cacheExpiry  map[string]time.Time
    mutex        sync.RWMutex
}

func NewTechnicalAnalysisService(db *database.DB) *TechnicalAnalysisService {
    return &TechnicalAnalysisService{
        db:          db,
        seriesCache: make(map[string]*techan.TimeSeries),
        cacheExpiry: make(map[string]time.Time),
    }
}
```

### Step 3: Implement Core Indicators
```go
func (tas *TechnicalAnalysisService) GetIndicators(symbol string) (*TechnicalIndicators, error) {
    series, err := tas.getOrCreateTimeSeries(symbol)
    if err != nil {
        return nil, err
    }
    
    lastIndex := series.LastIndex()
    closePrices := techan.NewClosePriceIndicator(series)
    
    return &TechnicalIndicators{
        RSI14:      techan.NewRelativeStrengthIndexIndicator(closePrices, 14).Calculate(lastIndex).Float(),
        SMA20:      techan.NewSMAIndicator(closePrices, 20).Calculate(lastIndex).Float(),
        EMA20:      techan.NewEMAIndicator(closePrices, 20).Calculate(lastIndex).Float(),
        // ... more indicators
    }, nil
}
```

This approach will save us weeks of development time and ensure our technical analysis is accurate and reliable. We can focus on building the unique value-add features like intelligent setup detection, scoring algorithms, and alert systems.

## 🔄 Next Steps

1. **Add libraries to project** 
2. **Create technical analysis service layer**
3. **Implement data conversion utilities**
4. **Build core indicator calculations**
5. **Develop custom S/R detection on top of library foundation**

This library-based approach aligns perfectly with your suggestion and will dramatically accelerate our development timeline while ensuring accuracy.
