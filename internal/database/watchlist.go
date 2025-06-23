package database

import (
	"database/sql"
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

func (db *Database) GetWatchlistCategories() ([]models.WatchlistCategory, error) {
	query := `
		SELECT id, name, description, color, created_at, updated_at
		FROM watchlist_categories
		ORDER BY name
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.WatchlistCategory
	for rows.Next() {
		var category models.WatchlistCategory
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Color,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (db *Database) CreateWatchlistCategory(category models.WatchlistCategory) (*models.WatchlistCategory, error) {
	query := `
		INSERT INTO watchlist_categories (name, description, color)
		VALUES (?, ?, ?)
	`

	result, err := db.conn.Exec(query, category.Name, category.Description, category.Color)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	category.ID = int(id)
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	return &category, nil
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

func (db *Database) GetWatchlistStocks(categoryID *int) ([]models.WatchlistStock, error) {
	var query string
	var args []interface{}

	if categoryID != nil {
		query = `
			SELECT 
				ws.id, ws.symbol, ws.name, ws.category_id, ws.notes, ws.tags,
				ws.price, ws.change, ws.change_percent, ws.volume, ws.market_cap,
				ws.added_at, ws.updated_at,
				wc.name as category_name, wc.color as category_color
			FROM watchlist_stocks ws
			LEFT JOIN watchlist_categories wc ON ws.category_id = wc.id
			WHERE ws.category_id = ?
			ORDER BY ws.symbol
		`
		args = append(args, *categoryID)
	} else {
		query = `
			SELECT 
				ws.id, ws.symbol, ws.name, ws.category_id, ws.notes, ws.tags,
				ws.price, ws.change, ws.change_percent, ws.volume, ws.market_cap,
				ws.added_at, ws.updated_at,
				wc.name as category_name, wc.color as category_color
			FROM watchlist_stocks ws
			LEFT JOIN watchlist_categories wc ON ws.category_id = wc.id
			ORDER BY ws.symbol
		`
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []models.WatchlistStock
	for rows.Next() {
		var stock models.WatchlistStock
		var categoryName, categoryColor sql.NullString

		err := rows.Scan(
			&stock.ID,
			&stock.Symbol,
			&stock.Name,
			&stock.CategoryID,
			&stock.Notes,
			&stock.Tags,
			&stock.Price,
			&stock.Change,
			&stock.ChangePercent,
			&stock.Volume,
			&stock.MarketCap,
			&stock.AddedAt,
			&stock.UpdatedAt,
			&categoryName,
			&categoryColor,
		)
		if err != nil {
			return nil, err
		}

		if categoryName.Valid {
			stock.CategoryName = categoryName.String
		}
		if categoryColor.Valid {
			stock.CategoryColor = categoryColor.String
		}

		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (db *Database) AddWatchlistStock(stock models.WatchlistStock) (*models.WatchlistStock, error) {
	query := `
		INSERT INTO watchlist_stocks (symbol, name, category_id, notes, tags, price, change, change_percent, volume, market_cap)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		strings.ToUpper(stock.Symbol),
		stock.Name,
		stock.CategoryID,
		stock.Notes,
		stock.Tags,
		stock.Price,
		stock.Change,
		stock.ChangePercent,
		stock.Volume,
		stock.MarketCap,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	stock.ID = int(id)
	stock.Symbol = strings.ToUpper(stock.Symbol)
	stock.AddedAt = time.Now()
	stock.UpdatedAt = time.Now()

	return &stock, nil
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

func (db *Database) DeleteWatchlistStock(id int) error {
	query := `DELETE FROM watchlist_stocks WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	return err
}

func (db *Database) GetWatchlistSummary() (*models.WatchlistSummary, error) {
	summary := &models.WatchlistSummary{}

	// Get total counts
	countQuery := `
		SELECT 
			(SELECT COUNT(*) FROM watchlist_stocks) as total_stocks,
			(SELECT COUNT(*) FROM watchlist_categories) as total_categories
	`
	err := db.conn.QueryRow(countQuery).Scan(&summary.TotalStocks, &summary.TotalCategories)
	if err != nil {
		return nil, err
	}

	// Get categories
	categories, err := db.GetWatchlistCategories()
	if err != nil {
		return nil, err
	}
	summary.Categories = categories

	// Get recently added stocks (last 10)
	recentQuery := `
		SELECT 
			ws.id, ws.symbol, ws.name, ws.category_id, ws.notes, ws.tags,
			ws.price, ws.change, ws.change_percent, ws.volume, ws.market_cap,
			ws.added_at, ws.updated_at,
			wc.name as category_name, wc.color as category_color
		FROM watchlist_stocks ws
		LEFT JOIN watchlist_categories wc ON ws.category_id = wc.id
		ORDER BY ws.added_at DESC
		LIMIT 10
	`
	recentStocks, err := db.queryWatchlistStocks(recentQuery)
	if err != nil {
		return nil, err
	}
	summary.RecentlyAdded = recentStocks

	// Get top gainers
	gainersQuery := `
		SELECT 
			ws.id, ws.symbol, ws.name, ws.category_id, ws.notes, ws.tags,
			ws.price, ws.change, ws.change_percent, ws.volume, ws.market_cap,
			ws.added_at, ws.updated_at,
			wc.name as category_name, wc.color as category_color
		FROM watchlist_stocks ws
		LEFT JOIN watchlist_categories wc ON ws.category_id = wc.id
		WHERE ws.change_percent > 0
		ORDER BY ws.change_percent DESC
		LIMIT 5
	`
	gainers, err := db.queryWatchlistStocks(gainersQuery)
	if err != nil {
		return nil, err
	}
	summary.TopGainers = gainers

	// Get top losers
	losersQuery := `
		SELECT 
			ws.id, ws.symbol, ws.name, ws.category_id, ws.notes, ws.tags,
			ws.price, ws.change, ws.change_percent, ws.volume, ws.market_cap,
			ws.added_at, ws.updated_at,
			wc.name as category_name, wc.color as category_color
		FROM watchlist_stocks ws
		LEFT JOIN watchlist_categories wc ON ws.category_id = wc.id
		WHERE ws.change_percent < 0
		ORDER BY ws.change_percent ASC
		LIMIT 5
	`
	losers, err := db.queryWatchlistStocks(losersQuery)
	if err != nil {
		return nil, err
	}
	summary.TopLosers = losers

	return summary, nil
}

// Helper function to query watchlist stocks
func (db *Database) queryWatchlistStocks(query string, args ...interface{}) ([]models.WatchlistStock, error) {
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []models.WatchlistStock
	for rows.Next() {
		var stock models.WatchlistStock
		var categoryName, categoryColor sql.NullString

		err := rows.Scan(
			&stock.ID,
			&stock.Symbol,
			&stock.Name,
			&stock.CategoryID,
			&stock.Notes,
			&stock.Tags,
			&stock.Price,
			&stock.Change,
			&stock.ChangePercent,
			&stock.Volume,
			&stock.MarketCap,
			&stock.AddedAt,
			&stock.UpdatedAt,
			&categoryName,
			&categoryColor,
		)
		if err != nil {
			return nil, err
		}

		if categoryName.Valid {
			stock.CategoryName = categoryName.String
		}
		if categoryColor.Valid {
			stock.CategoryColor = categoryColor.String
		}

		stocks = append(stocks, stock)
	}

	return stocks, nil
}

// WatchlistStockExists returns true if a stock with the given symbol exists in the watchlist
func (db *Database) WatchlistStockExists(symbol string) (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(1) FROM watchlist_stocks WHERE symbol = ?", strings.ToUpper(symbol)).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
