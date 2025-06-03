package models

import (
	"time"
)

// PriceData represents OHLC price data for a stock symbol
type PriceData struct {
	ID        int64     `json:"id" db:"id"`
	Symbol    string    `json:"symbol" db:"symbol"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Open      float64   `json:"open" db:"open_price"`
	High      float64   `json:"high" db:"high_price"`
	Low       float64   `json:"low" db:"low_price"`
	Close     float64   `json:"close" db:"close_price"`
	Volume    int64     `json:"volume" db:"volume"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PriceDataFilter represents filter parameters for querying price data
type PriceDataFilter struct {
	Symbol   string    `json:"symbol"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Interval string    `json:"interval"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// PriceDataResponse represents the API response for price data
type PriceDataResponse struct {
	Symbol       string              `json:"symbol"`
	Data         []TradingViewCandle `json:"data"`
	TotalRecords int                 `json:"total_records"`
	From         time.Time           `json:"from"`
	To           time.Time           `json:"to"`
	Interval     string              `json:"interval"`
}

// TradingViewCandle represents a candlestick data point for TradingView
type TradingViewCandle struct {
	Time   int64   `json:"time"`   // Unix timestamp
	Open   float64 `json:"open"`   // Open price
	High   float64 `json:"high"`   // High price
	Low    float64 `json:"low"`    // Low price
	Close  float64 `json:"close"`  // Close price
	Volume int64   `json:"volume"` // Volume
}

// PriceStats represents price statistics for a symbol
type PriceStats struct {
	Symbol             string    `json:"symbol"`
	CurrentPrice       float64   `json:"current_price"`
	OpenPrice          float64   `json:"open_price"`
	HighPrice          float64   `json:"high_price"`
	LowPrice           float64   `json:"low_price"`
	PriceChange        float64   `json:"price_change"`
	PriceChangePercent float64   `json:"price_change_percent"`
	LastUpdate         time.Time `json:"last_update"`
}

// ToTradingViewCandle converts PriceData to TradingViewCandle
func (pd *PriceData) ToTradingViewCandle() TradingViewCandle {
	return TradingViewCandle{
		Time:   pd.Timestamp.Unix(),
		Open:   pd.Open,
		High:   pd.High,
		Low:    pd.Low,
		Close:  pd.Close,
		Volume: pd.Volume,
	}
}

// ToPriceData converts PolygonAggregate to PriceData
func (pa *PolygonAggregate) ToPriceData(symbol string) *PriceData {
	return &PriceData{
		Symbol:    symbol,
		Timestamp: time.Unix(pa.Timestamp/1000, (pa.Timestamp%1000)*1000000), // Convert from milliseconds
		Open:      pa.Open,
		High:      pa.High,
		Low:       pa.Low,
		Close:     pa.Close,
		Volume:    pa.Volume,
		CreatedAt: time.Now(),
	}
}
