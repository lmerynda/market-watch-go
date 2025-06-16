package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"market-watch-go/internal/config"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// New creates a new database connection
func New(cfg *config.Config) (*DB, error) {
	// Ensure data directory exists
	dataDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	conn.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	db := &DB{conn: conn}

	// Initialize database schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Run migrations
	if err := db.runMigrations(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize setup tables (includes trading setups schema)
	if err := db.CreateSetupTables(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize setup tables: %w", err)
	}

	// Initialize head and shoulders pattern tables
	if err := db.CreateHeadShouldersPatternTables(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize head and shoulders pattern tables: %w", err)
	}

	// Initialize falling wedge pattern tables
	if err := db.CreateFallingWedgePatternTables(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize falling wedge pattern tables: %w", err)
	}

	log.Printf("Database initialized at %s", cfg.Database.Path)
	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// initSchema creates the database tables and indexes
func (db *DB) initSchema() error {
	schema := `
	-- Price data table
	CREATE TABLE IF NOT EXISTS price_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		open_price REAL NOT NULL,
		high_price REAL NOT NULL,
		low_price REAL NOT NULL,
		close_price REAL NOT NULL,
		volume INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(symbol, timestamp)
	);

	-- Volume data table
	CREATE TABLE IF NOT EXISTS volume_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		volume INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(symbol, timestamp)
	);

	-- Technical indicators table (matching existing schema)
	CREATE TABLE IF NOT EXISTS technical_indicators (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		rsi_14 REAL,
		rsi_30 REAL,
		macd_line REAL,
		macd_signal REAL,
		macd_histogram REAL,
		sma_20 REAL,
		sma_50 REAL,
		sma_200 REAL,
		ema_20 REAL,
		ema_50 REAL,
		vwap REAL,
		volume_ratio REAL,
		bb_upper REAL,
		bb_middle REAL,
		bb_lower REAL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(symbol, timestamp)
	);

	-- Support/Resistance levels table
	CREATE TABLE IF NOT EXISTS support_resistance_levels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		level REAL NOT NULL,
		level_type TEXT NOT NULL CHECK (level_type IN ('support', 'resistance')),
		strength REAL NOT NULL DEFAULT 0,
		touches INTEGER DEFAULT 1,
		first_touch DATETIME NOT NULL,
		last_touch DATETIME NOT NULL,
		is_active BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Trading setups table
	CREATE TABLE IF NOT EXISTS trading_setups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		setup_type TEXT NOT NULL,
		direction TEXT NOT NULL CHECK (direction IN ('long', 'short')),
		current_price REAL,
		entry_price REAL,
		stop_loss REAL,
		target1 REAL,
		target2 REAL,
		target3 REAL,
		risk_reward_ratio REAL,
		confidence TEXT CHECK (confidence IN ('low', 'medium', 'high')),
		quality_score REAL,
		checklist_items TEXT,
		checklist_score INTEGER,
		status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'filled', 'stopped', 'completed', 'expired')),
		notes TEXT,
		detected_at DATETIME NOT NULL,
		expires_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Watched symbols table
	CREATE TABLE IF NOT EXISTS watched_symbols (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL UNIQUE,
		name TEXT,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN DEFAULT 1
	);

	-- Indicator alerts table
	CREATE TABLE IF NOT EXISTS indicator_alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		alert_type TEXT NOT NULL,
		indicator TEXT NOT NULL,
		value REAL NOT NULL,
		threshold REAL NOT NULL,
		message TEXT NOT NULL,
		triggered_at DATETIME NOT NULL,
		is_active BOOLEAN DEFAULT TRUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Create indexes for better performance
	CREATE INDEX IF NOT EXISTS idx_price_data_symbol_timestamp ON price_data(symbol, timestamp);
	CREATE INDEX IF NOT EXISTS idx_volume_data_symbol_timestamp ON volume_data(symbol, timestamp);
	CREATE INDEX IF NOT EXISTS idx_technical_indicators_symbol_timestamp ON technical_indicators(symbol, timestamp);
	CREATE INDEX IF NOT EXISTS idx_support_resistance_symbol_active ON support_resistance_levels(symbol, is_active);
	CREATE INDEX IF NOT EXISTS idx_trading_setups_symbol_status ON trading_setups(symbol, status, detected_at);
	CREATE INDEX IF NOT EXISTS idx_watched_symbols_active ON watched_symbols(is_active);
	CREATE INDEX IF NOT EXISTS idx_indicator_alerts_symbol ON indicator_alerts(symbol);
	CREATE INDEX IF NOT EXISTS idx_indicator_alerts_active ON indicator_alerts(is_active);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// EnsureConfigSymbolsWatched ensures that all symbols from config are in the watched_symbols table
func (db *DB) EnsureConfigSymbolsWatched(configSymbols []string) error {
	if len(configSymbols) == 0 {
		log.Printf("No symbols in config to watch")
		return nil
	}

	log.Printf("Ensuring config symbols are watched: %v", configSymbols)

	for _, symbol := range configSymbols {
		// Use AddWatchedSymbol which handles INSERT OR REPLACE
		err := db.AddWatchedSymbol(symbol, "") // Empty name, can be updated later
		if err != nil {
			log.Printf("Failed to ensure symbol %s is watched: %v", symbol, err)
			return err
		}
	}

	log.Printf("Successfully ensured %d config symbols are watched", len(configSymbols))
	return nil
}

// GetWatchedSymbols returns all active watched symbols
func (db *DB) GetWatchedSymbols() ([]string, error) {
	query := `SELECT symbol FROM watched_symbols WHERE is_active = 1 ORDER BY symbol`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query watched symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, fmt.Errorf("failed to scan symbol: %w", err)
		}
		symbols = append(symbols, symbol)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating symbol rows: %w", err)
	}

	return symbols, nil
}

// AddWatchedSymbol adds a new symbol to watch
func (db *DB) AddWatchedSymbol(symbol, name string) error {
	query := `INSERT OR REPLACE INTO watched_symbols (symbol, name, is_active) VALUES (?, ?, 1)`

	_, err := db.conn.Exec(query, symbol, name)
	if err != nil {
		return fmt.Errorf("failed to add watched symbol: %w", err)
	}

	return nil
}

// RemoveWatchedSymbol removes a symbol from being watched
func (db *DB) RemoveWatchedSymbol(symbol string) error {
	query := `UPDATE watched_symbols SET is_active = 0 WHERE symbol = ?`

	result, err := db.conn.Exec(query, symbol)
	if err != nil {
		return fmt.Errorf("failed to remove watched symbol: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("symbol %s not found", symbol)
	}

	return nil
}

// GetDatabaseStats returns statistics about the database
func (db *DB) GetDatabaseStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	tables := []string{
		"price_data",
		"volume_data",
		"technical_indicators",
		"support_resistance_levels",
		"trading_setups",
		"watched_symbols",
		"indicator_alerts",
	}

	for _, table := range tables {
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		err := db.conn.QueryRow(query).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to get count for table %s: %w", table, err)
		}
		stats[table] = count
	}

	return stats, nil
}

// runMigrations applies database migrations for schema updates
func (db *DB) runMigrations() error {
	// Add current_price column to trading_setups if it doesn't exist
	alterQuery := `ALTER TABLE trading_setups ADD COLUMN current_price REAL`

	// Check if column already exists by trying to add it
	// SQLite will return an error if column already exists, which we can ignore
	_, err := db.conn.Exec(alterQuery)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		log.Printf("Warning: Failed to add current_price column (might already exist): %v", err)
	}

	return nil
}

// Ping checks if the database connection is alive
func (db *DB) Ping() error {
	return db.conn.Ping()
}
