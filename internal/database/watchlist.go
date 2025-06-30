package database

import (
	"fmt"
	"log"
	"market-watch-go/internal/models"
	"strings"
	"time"
)

// Database is an alias for DB to maintain consistency
type Database = DB

// CreateWatchlistTables creates the watchlist-related tables
func (db *Database) CreateWatchlistTables() error {
	// Categories table
	categoriesTable := `
	CREATE TABLE IF NOT EXISTS watchlist_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		color TEXT DEFAULT '#007bff',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Stocks table
	stocksTable := `
	CREATE TABLE IF NOT EXISTS watchlist_stocks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		name TEXT,
		category_id INTEGER,
		notes TEXT,
		tags TEXT,
		price REAL DEFAULT 0,
		change REAL DEFAULT 0,
		change_percent REAL DEFAULT 0,
		volume INTEGER DEFAULT 0,
		market_cap INTEGER DEFAULT 0,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category_id) REFERENCES watchlist_categories(id) ON DELETE SET NULL,
		UNIQUE(symbol)
	);`

	// Create tables
	if _, err := db.conn.Exec(categoriesTable); err != nil {
		return fmt.Errorf("failed to create watchlist_categories table: %v", err)
	}

	if _, err := db.conn.Exec(stocksTable); err != nil {
		return fmt.Errorf("failed to create watchlist_stocks table: %v", err)
	}

	log.Println("Watchlist tables created successfully")
	return nil
}

// Watchlist Categories Operations

func (db *Database) GetWatchlistCategories() ([]models.Strategy, error) {
	query := `
		SELECT id, name, description, color, created_at, updated_at
		FROM strategies
		ORDER BY name
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strategies []models.Strategy
	for rows.Next() {
		var strategy models.Strategy
		err := rows.Scan(
			&strategy.ID,
			&strategy.Name,
			&strategy.Description,
			&strategy.Color,
			&strategy.CreatedAt,
			&strategy.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		strategies = append(strategies, strategy)
	}

	return strategies, nil
}

func (db *Database) CreateWatchlistCategory(strategy models.Strategy) (*models.Strategy, error) {
	query := `
		INSERT INTO strategies (name, description, color)
		VALUES (?, ?, ?)
	`

	result, err := db.conn.Exec(query, strategy.Name, strategy.Description, strategy.Color)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	strategy.ID = int(id)
	strategy.CreatedAt = time.Now()
	strategy.UpdatedAt = time.Now()

	return &strategy, nil
}

func (db *Database) UpdateWatchlistCategory(id int, category models.WatchlistCategory) error {
	query := `
		UPDATE watchlist_categories 
		SET name = ?, description = ?, color = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.conn.Exec(query, category.Name, category.Description, category.Color, id)
	return err
}

func (db *Database) DeleteWatchlistCategory(id int) error {
	// First, update all stocks in this category to have no category
	updateQuery := `UPDATE watchlist_stocks SET category_id = NULL WHERE category_id = ?`
	_, err := db.conn.Exec(updateQuery, id)
	if err != nil {
		return err
	}

	// Then delete the category
	deleteQuery := `DELETE FROM watchlist_categories WHERE id = ?`
	_, err = db.conn.Exec(deleteQuery, id)
	return err
}

// Watchlist Stocks Operations

func (db *Database) GetWatchlistStocks(strategyID *int) ([]models.Stock, error) {
	var query string
	var args []interface{}

	if strategyID != nil {
		query = `
			SELECT 
				s.id, s.symbol, s.name, s.notes,
				s.price, s.change, s.change_percent, s.volume, s.market_cap,
				s.added_at, s.updated_at, s.ema_9, s.ema_50, s.ema_200
			FROM stocks s
			INNER JOIN stock_strategies ss ON s.id = ss.stock_id
			WHERE ss.strategy_id = ?
			ORDER BY s.symbol
		`
		args = append(args, *strategyID)
	} else {
		query = `
			SELECT 
				s.id, s.symbol, s.name, s.notes,
				s.price, s.change, s.change_percent, s.volume, s.market_cap,
				s.added_at, s.updated_at, s.ema_9, s.ema_50, s.ema_200
			FROM stocks s
			ORDER BY s.symbol
		`
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []models.Stock
	for rows.Next() {
		var stock models.Stock

		err := rows.Scan(
			&stock.ID,
			&stock.Symbol,
			&stock.Name,
			&stock.Notes,
			&stock.Price,
			&stock.Change,
			&stock.ChangePercent,
			&stock.Volume,
			&stock.MarketCap,
			&stock.AddedAt,
			&stock.UpdatedAt,
			&stock.EMA9,
			&stock.EMA50,
			&stock.EMA200,
		)
		if err != nil {
			return nil, err
		}

		// Load strategies for this stock
		strategies, err := db.GetStockStrategies(stock.ID)
		if err != nil {
			return nil, err
		}
		stock.Strategies = strategies

		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (db *Database) AddWatchlistStock(stock models.Stock) (*models.Stock, error) {
	return db.AddStock(stock)
}

func (db *Database) UpdateWatchlistStock(id int, stock models.WatchlistStock) error {
	query := `
		UPDATE watchlist_stocks 
		SET name = ?, category_id = ?, notes = ?, tags = ?, 
		    price = ?, change = ?, change_percent = ?, volume = ?, market_cap = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.conn.Exec(query,
		stock.Name,
		stock.CategoryID,
		stock.Notes,
		stock.Tags,
		stock.Price,
		stock.Change,
		stock.ChangePercent,
		stock.Volume,
		stock.MarketCap,
		id,
	)
	return err
}

func (db *Database) UpdateWatchlistStockWithEMA(id int, stock models.WatchlistStock) error {
	query := `
		UPDATE watchlist_stocks 
		SET name = ?, category_id = ?, notes = ?, tags = ?, 
		    price = ?, change = ?, change_percent = ?, volume = ?, market_cap = ?,
		    ema_9 = ?, ema_50 = ?, ema_200 = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.conn.Exec(query,
		stock.Name,
		stock.CategoryID,
		stock.Notes,
		stock.Tags,
		stock.Price,
		stock.Change,
		stock.ChangePercent,
		stock.Volume,
		stock.MarketCap,
		stock.EMA9,
		stock.EMA50,
		stock.EMA200,
		id,
	)
	return err
}

func (db *Database) DeleteWatchlistStock(id int) error {
	query := `DELETE FROM watchlist_stocks WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	return err
}

// WatchlistStockExists returns true if a stock with the given symbol exists in the watchlist
func (db *Database) WatchlistStockExists(symbol string) (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(1) FROM stocks WHERE symbol = ?", strings.ToUpper(symbol)).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
