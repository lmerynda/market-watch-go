package database

import (
	"database/sql"
	"fmt"
	"log"
	"market-watch-go/internal/config"
	"market-watch-go/internal/models"
	"strings"
	"time"
)

// CreateStrategyTables creates the new strategy-based tables
func (db *Database) CreateStrategyTables() error {
	// Strategies table (replaces categories)
	strategiesTable := `
	CREATE TABLE IF NOT EXISTS strategies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		color TEXT DEFAULT '#007bff',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Centralized stocks table
	stocksTable := `
	CREATE TABLE IF NOT EXISTS stocks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL UNIQUE,
		name TEXT,
		notes TEXT,
		price REAL DEFAULT 0,
		change REAL DEFAULT 0,
		change_percent REAL DEFAULT 0,
		volume INTEGER DEFAULT 0,
		market_cap INTEGER DEFAULT 0,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		ema_9 REAL DEFAULT 0,
		ema_50 REAL DEFAULT 0,
		ema_200 REAL DEFAULT 0
	);`

	// Many-to-many relationship table
	stockStrategiesTable := `
	CREATE TABLE IF NOT EXISTS stock_strategies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		stock_id INTEGER NOT NULL,
		strategy_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (stock_id) REFERENCES stocks(id) ON DELETE CASCADE,
		FOREIGN KEY (strategy_id) REFERENCES strategies(id) ON DELETE CASCADE,
		UNIQUE(stock_id, strategy_id)
	);`

	// Create tables
	if _, err := db.conn.Exec(strategiesTable); err != nil {
		return fmt.Errorf("failed to create strategies table: %v", err)
	}

	if _, err := db.conn.Exec(stocksTable); err != nil {
		return fmt.Errorf("failed to create stocks table: %v", err)
	}

	if _, err := db.conn.Exec(stockStrategiesTable); err != nil {
		return fmt.Errorf("failed to create stock_strategies table: %v", err)
	}

	log.Println("Strategy tables created successfully")
	return nil
}

// Strategy Operations

func (db *Database) GetStrategies() ([]models.Strategy, error) {
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

func (db *Database) CreateStrategy(strategy models.Strategy) (*models.Strategy, error) {
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

// UpdateStrategy updates an existing strategy
func (db *Database) UpdateStrategy(id int, strategy models.Strategy) error {
	query := `
		UPDATE strategies 
		SET name = ?, description = ?, color = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.conn.Exec(query, strategy.Name, strategy.Description, strategy.Color, id)
	return err
}

// DeleteStrategy deletes a strategy and removes all associations
func (db *Database) DeleteStrategy(id int) error {
	// First delete all associations in the stock_strategies table
	deleteAssociationsQuery := `DELETE FROM stock_strategies WHERE strategy_id = ?`
	_, err := db.conn.Exec(deleteAssociationsQuery, id)
	if err != nil {
		return err
	}

	// Then delete the strategy
	deleteStrategyQuery := `DELETE FROM strategies WHERE id = ?`
	_, err = db.conn.Exec(deleteStrategyQuery, id)
	return err
}

// Stock Operations

func (db *Database) GetStocks() ([]models.Stock, error) {
	query := `
		SELECT id, symbol, name, notes, price, change, change_percent, 
		       volume, market_cap, added_at, updated_at, ema_9, ema_50, ema_200
		FROM stocks
		ORDER BY symbol
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []models.Stock
	for rows.Next() {
		var stock models.Stock
		err := rows.Scan(
			&stock.ID, &stock.Symbol, &stock.Name, &stock.Notes,
			&stock.Price, &stock.Change, &stock.ChangePercent,
			&stock.Volume, &stock.MarketCap, &stock.AddedAt, &stock.UpdatedAt,
			&stock.EMA9, &stock.EMA50, &stock.EMA200,
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

func (db *Database) GetStocksByStrategy(strategyID int) ([]models.Stock, error) {
	query := `
		SELECT s.id, s.symbol, s.name, s.notes, s.price, s.change, s.change_percent,
		       s.volume, s.market_cap, s.added_at, s.updated_at, s.ema_9, s.ema_50, s.ema_200
		FROM stocks s
		INNER JOIN stock_strategies ss ON s.id = ss.stock_id
		WHERE ss.strategy_id = ?
		ORDER BY s.symbol
	`

	rows, err := db.conn.Query(query, strategyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []models.Stock
	for rows.Next() {
		var stock models.Stock
		err := rows.Scan(
			&stock.ID, &stock.Symbol, &stock.Name, &stock.Notes,
			&stock.Price, &stock.Change, &stock.ChangePercent,
			&stock.Volume, &stock.MarketCap, &stock.AddedAt, &stock.UpdatedAt,
			&stock.EMA9, &stock.EMA50, &stock.EMA200,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (db *Database) AddStock(stock models.Stock) (*models.Stock, error) {
	query := `
		INSERT INTO stocks (symbol, name, notes, price, change, change_percent, volume, market_cap)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		strings.ToUpper(stock.Symbol), stock.Name, stock.Notes,
		stock.Price, stock.Change, stock.ChangePercent,
		stock.Volume, stock.MarketCap,
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

// UpdateStockNotes updates only the notes field for a stock
// This is appropriate for the strategy module as notes are user-editable metadata
func (db *Database) UpdateStockNotes(id int, notes string) error {
	query := `
		UPDATE stocks 
		SET notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.conn.Exec(query, notes, id)
	return err
}

func (db *Database) StockExists(symbol string) (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(1) FROM stocks WHERE symbol = ?", strings.ToUpper(symbol)).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *Database) GetStockBySymbol(symbol string) (*models.Stock, error) {
	query := `
		SELECT id, symbol, name, notes, price, change, change_percent,
		       volume, market_cap, added_at, updated_at, ema_9, ema_50, ema_200
		FROM stocks
		WHERE symbol = ?
	`

	var stock models.Stock
	err := db.conn.QueryRow(query, strings.ToUpper(symbol)).Scan(
		&stock.ID, &stock.Symbol, &stock.Name, &stock.Notes,
		&stock.Price, &stock.Change, &stock.ChangePercent,
		&stock.Volume, &stock.MarketCap, &stock.AddedAt, &stock.UpdatedAt,
		&stock.EMA9, &stock.EMA50, &stock.EMA200,
	)
	if err != nil {
		return nil, err
	}

	// Load strategies
	strategies, err := db.GetStockStrategies(stock.ID)
	if err != nil {
		return nil, err
	}
	stock.Strategies = strategies

	return &stock, nil
}

// Stock-Strategy Relationship Operations

func (db *Database) AddStockToStrategy(stockID, strategyID int) error {
	query := `INSERT INTO stock_strategies (stock_id, strategy_id) VALUES (?, ?)`
	_, err := db.conn.Exec(query, stockID, strategyID)
	return err
}

func (db *Database) RemoveStockFromStrategy(stockID, strategyID int) error {
	query := `DELETE FROM stock_strategies WHERE stock_id = ? AND strategy_id = ?`
	_, err := db.conn.Exec(query, stockID, strategyID)
	return err
}

// RemoveAllStockStrategies removes all strategy associations for a stock
func (db *Database) RemoveAllStockStrategies(stockID int) error {
	query := `DELETE FROM stock_strategies WHERE stock_id = ?`
	_, err := db.conn.Exec(query, stockID)
	return err
}

func (db *Database) GetStockStrategies(stockID int) ([]models.Strategy, error) {
	query := `
		SELECT s.id, s.name, s.description, s.color, s.created_at, s.updated_at
		FROM strategies s
		INNER JOIN stock_strategies ss ON s.id = ss.strategy_id
		WHERE ss.stock_id = ?
		ORDER BY s.name
	`

	rows, err := db.conn.Query(query, stockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strategies []models.Strategy
	for rows.Next() {
		var strategy models.Strategy
		err := rows.Scan(
			&strategy.ID, &strategy.Name, &strategy.Description,
			&strategy.Color, &strategy.CreatedAt, &strategy.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		strategies = append(strategies, strategy)
	}

	return strategies, nil
}

// Migration helper to populate strategies from config
func (db *Database) EnsureConfigStrategies(strategies []config.WatchlistStrategyConfig) error {
	for _, strategyConfig := range strategies {
		// Check if strategy exists
		var strategyID int
		err := db.conn.QueryRow("SELECT id FROM strategies WHERE name = ?", strategyConfig.Name).Scan(&strategyID)
		if err == sql.ErrNoRows {
			// Create strategy
			strategy := models.Strategy{
				Name:  strategyConfig.Name,
				Color: strategyConfig.Color,
			}
			createdStrategy, err := db.CreateStrategy(strategy)
			if err != nil {
				return fmt.Errorf("failed to create strategy %s: %v", strategyConfig.Name, err)
			}
			strategyID = createdStrategy.ID
		} else if err != nil {
			return err
		}

		// Add stocks to strategy
		for _, symbol := range strategyConfig.Stocks {
			// Ensure stock exists
			var stockID int
			err := db.conn.QueryRow("SELECT id FROM stocks WHERE symbol = ?", strings.ToUpper(symbol)).Scan(&stockID)
			if err == sql.ErrNoRows {
				// Create stock
				stock := models.Stock{Symbol: strings.ToUpper(symbol)}
				createdStock, err := db.AddStock(stock)
				if err != nil {
					return fmt.Errorf("failed to create stock %s: %v", symbol, err)
				}
				stockID = createdStock.ID
			} else if err != nil {
				return err
			}

			// Add to strategy (ignore if already exists)
			db.AddStockToStrategy(stockID, strategyID)
		}
	}
	return nil
}
