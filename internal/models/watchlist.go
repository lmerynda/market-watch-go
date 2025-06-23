package models

import "time"

// WatchlistCategory represents a category for organizing stocks
type WatchlistCategory struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"` // Hex color for UI
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// WatchlistStock represents a stock in the watchlist
type WatchlistStock struct {
	ID            int       `json:"id" db:"id"`
	Symbol        string    `json:"symbol" db:"symbol"`
	Name          string    `json:"name" db:"name"`
	CategoryID    *int      `json:"category_id" db:"category_id"`
	Notes         string    `json:"notes" db:"notes"`
	Tags          string    `json:"tags" db:"tags"` // Comma-separated tags
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

	// Joined data
	CategoryName  string `json:"category_name,omitempty" db:"category_name"`
	CategoryColor string `json:"category_color,omitempty" db:"category_color"`
}

// WatchlistSummary represents summary data for the watchlist
type WatchlistSummary struct {
	TotalStocks     int                 `json:"total_stocks"`
	TotalCategories int                 `json:"total_categories"`
	Categories      []WatchlistCategory `json:"categories"`
	RecentlyAdded   []WatchlistStock    `json:"recently_added"`
	TopGainers      []WatchlistStock    `json:"top_gainers"`
	TopLosers       []WatchlistStock    `json:"top_losers"`
}
