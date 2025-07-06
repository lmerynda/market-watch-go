package models

import "time"

// Strategy represents a trading strategy (formerly category)
type Strategy struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"` // Hex color for UI
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Stocks      []Stock   `json:"stocks,omitempty"`
}

// Stock represents a centralized stock in the watchlist
type Stock struct {
	ID            int       `json:"id" db:"id"`
	Symbol        string    `json:"symbol" db:"symbol"`
	Name          string    `json:"name" db:"name"`
	Notes         string    `json:"notes" db:"notes"`
	Price         float64   `json:"price" db:"price"`
	Change        float64   `json:"change" db:"change"`
	ChangePercent float64   `json:"change_percent" db:"change_percent"`
	Volume        int64     `json:"volume" db:"volume"`
	MarketCap     int64     `json:"market_cap" db:"market_cap"`
	AddedAt       time.Time `json:"added_at" db:"added_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	EMA9          float64   `json:"ema_9" db:"ema_9"`
	EMA50         float64   `json:"ema_50" db:"ema_50"`
	EMA200        float64   `json:"ema_200" db:"ema_200"`

	// Associated strategies (tags)
	Strategies []Strategy `json:"strategies,omitempty"`
}

// StockStrategy represents the many-to-many relationship between stocks and strategies
type StockStrategy struct {
	ID         int       `json:"id" db:"id"`
	StockID    int       `json:"stock_id" db:"stock_id"`
	StrategyID int       `json:"strategy_id" db:"strategy_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// WatchlistSummaryV2 represents summary data for the new strategy-based watchlist
type WatchlistSummaryV2 struct {
	TotalStocks     int        `json:"total_stocks"`
	TotalStrategies int        `json:"total_strategies"`
	Strategies      []Strategy `json:"strategies"`
	RecentlyAdded   []Stock    `json:"recently_added"`
	TopGainers      []Stock    `json:"top_gainers"`
	TopLosers       []Stock    `json:"top_losers"`
}
