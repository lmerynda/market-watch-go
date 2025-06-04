package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"market-watch-go/internal/config"
	"market-watch-go/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
	cfg  *config.Config
}

// New creates a new database connection
func New(cfg *config.Config) (*DB, error) {
	conn, err := sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	conn.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		conn: conn,
		cfg:  cfg,
	}

	// Run migrations
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// migrate runs database migrations
func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS volume_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			volume INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(symbol, timestamp)
		)`,
		`CREATE TABLE IF NOT EXISTS price_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			open_price DECIMAL(10,2),
			high_price DECIMAL(10,2),
			low_price DECIMAL(10,2),
			close_price DECIMAL(10,2),
			volume INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(symbol, timestamp)
		)`,
		`CREATE TABLE IF NOT EXISTS watched_symbols (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol TEXT NOT NULL UNIQUE,
			name TEXT,
			added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_active BOOLEAN DEFAULT TRUE
		)`,
		// Technical Indicators Table
		`CREATE TABLE IF NOT EXISTS technical_indicators (
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
		)`,
		// Indicator Alerts Table
		`CREATE TABLE IF NOT EXISTS indicator_alerts (
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
		)`,
		// Support/Resistance Levels table
		`CREATE TABLE IF NOT EXISTS support_resistance_levels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol TEXT NOT NULL,
			level REAL NOT NULL,
			level_type TEXT NOT NULL CHECK (level_type IN ('support', 'resistance')),
			strength REAL NOT NULL DEFAULT 0,
			touches INTEGER NOT NULL DEFAULT 0,
			first_touch DATETIME NOT NULL,
			last_touch DATETIME NOT NULL,
			volume_confirmed BOOLEAN DEFAULT FALSE,
			avg_volume REAL DEFAULT 0,
			max_bounce_percent REAL DEFAULT 0,
			avg_bounce_percent REAL DEFAULT 0,
			timeframe_origin TEXT DEFAULT '1m',
			is_active BOOLEAN DEFAULT TRUE,
			last_validated DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Pivot Points table
		`CREATE TABLE IF NOT EXISTS pivot_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			price REAL NOT NULL,
			pivot_type TEXT NOT NULL CHECK (pivot_type IN ('high', 'low')),
			strength INTEGER NOT NULL DEFAULT 1,
			volume INTEGER NOT NULL DEFAULT 0,
			confirmed BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// S/R Level Touches table
		`CREATE TABLE IF NOT EXISTS sr_level_touches (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			level_id INTEGER NOT NULL,
			symbol TEXT NOT NULL,
			touch_time DATETIME NOT NULL,
			touch_price REAL NOT NULL,
			level REAL NOT NULL,
			distance_percent REAL NOT NULL,
			bounce_percent REAL DEFAULT 0,
			volume_at_touch INTEGER DEFAULT 0,
			volume_spike BOOLEAN DEFAULT FALSE,
			bounce_confirmed BOOLEAN DEFAULT FALSE,
			time_at_level INTEGER DEFAULT 0,
			touch_type TEXT DEFAULT 'test' CHECK (touch_type IN ('test', 'break', 'bounce')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (level_id) REFERENCES support_resistance_levels(id)
		)`,
		// Trading Setups table
		`CREATE TABLE IF NOT EXISTS trading_setups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol TEXT NOT NULL,
			setup_type TEXT NOT NULL,
			direction TEXT NOT NULL CHECK (direction IN ('bullish', 'bearish')),
			quality_score REAL NOT NULL DEFAULT 0,
			confidence TEXT DEFAULT 'low' CHECK (confidence IN ('high', 'medium', 'low')),
			status TEXT DEFAULT 'active' CHECK (status IN ('active', 'triggered', 'expired', 'invalidated')),
			detected_at DATETIME NOT NULL,
			expires_at DATETIME NOT NULL,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			
			current_price REAL NOT NULL,
			entry_price REAL NOT NULL,
			stop_loss REAL NOT NULL,
			target1 REAL DEFAULT 0,
			target2 REAL DEFAULT 0,
			target3 REAL DEFAULT 0,
			
			risk_amount REAL DEFAULT 0,
			reward_potential REAL DEFAULT 0,
			risk_reward_ratio REAL DEFAULT 0,
			
			price_action_score REAL DEFAULT 0,
			volume_score REAL DEFAULT 0,
			technical_score REAL DEFAULT 0,
			risk_reward_score REAL DEFAULT 0,
			
			notes TEXT DEFAULT '',
			is_manual BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Setup Checklists table
		`CREATE TABLE IF NOT EXISTS setup_checklists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setup_id INTEGER NOT NULL,
			
			-- Price Action Criteria
			min_level_touches_completed BOOLEAN DEFAULT FALSE,
			min_level_touches_points REAL DEFAULT 0,
			bounce_strength_completed BOOLEAN DEFAULT FALSE,
			bounce_strength_points REAL DEFAULT 0,
			time_at_level_completed BOOLEAN DEFAULT FALSE,
			time_at_level_points REAL DEFAULT 0,
			rejection_candle_completed BOOLEAN DEFAULT FALSE,
			rejection_candle_points REAL DEFAULT 0,
			level_duration_completed BOOLEAN DEFAULT FALSE,
			level_duration_points REAL DEFAULT 0,
			
			-- Volume Criteria
			volume_spike_completed BOOLEAN DEFAULT FALSE,
			volume_spike_points REAL DEFAULT 0,
			volume_confirmation_completed BOOLEAN DEFAULT FALSE,
			volume_confirmation_points REAL DEFAULT 0,
			approach_volume_completed BOOLEAN DEFAULT FALSE,
			approach_volume_points REAL DEFAULT 0,
			vwap_relationship_completed BOOLEAN DEFAULT FALSE,
			vwap_relationship_points REAL DEFAULT 0,
			relative_volume_completed BOOLEAN DEFAULT FALSE,
			relative_volume_points REAL DEFAULT 0,
			
			-- Technical Indicators
			rsi_condition_completed BOOLEAN DEFAULT FALSE,
			rsi_condition_points REAL DEFAULT 0,
			moving_average_completed BOOLEAN DEFAULT FALSE,
			moving_average_points REAL DEFAULT 0,
			macd_signal_completed BOOLEAN DEFAULT FALSE,
			macd_signal_points REAL DEFAULT 0,
			momentum_divergence_completed BOOLEAN DEFAULT FALSE,
			momentum_divergence_points REAL DEFAULT 0,
			bollinger_bands_completed BOOLEAN DEFAULT FALSE,
			bollinger_bands_points REAL DEFAULT 0,
			
			-- Risk Management
			stop_loss_defined_completed BOOLEAN DEFAULT FALSE,
			stop_loss_defined_points REAL DEFAULT 0,
			risk_reward_ratio_completed BOOLEAN DEFAULT FALSE,
			risk_reward_ratio_points REAL DEFAULT 0,
			position_size_completed BOOLEAN DEFAULT FALSE,
			position_size_points REAL DEFAULT 0,
			entry_precision_completed BOOLEAN DEFAULT FALSE,
			entry_precision_points REAL DEFAULT 0,
			exit_strategy_completed BOOLEAN DEFAULT FALSE,
			exit_strategy_points REAL DEFAULT 0,
			
			-- Summary
			total_score REAL DEFAULT 0,
			completed_items INTEGER DEFAULT 0,
			total_items INTEGER DEFAULT 0,
			completion_percent REAL DEFAULT 0,
			
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (setup_id) REFERENCES trading_setups(id)
		)`,
		// Setup Alerts table
		`CREATE TABLE IF NOT EXISTS setup_alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setup_id INTEGER NOT NULL,
			symbol TEXT NOT NULL,
			alert_type TEXT NOT NULL,
			message TEXT NOT NULL,
			severity TEXT DEFAULT 'medium' CHECK (severity IN ('high', 'medium', 'low')),
			is_active BOOLEAN DEFAULT TRUE,
			triggered_at DATETIME NOT NULL,
			notification_sent BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (setup_id) REFERENCES trading_setups(id)
		)`,
	}

	// Create all indexes
	indexes := []string{
		// Volume data indexes
		`CREATE INDEX IF NOT EXISTS idx_volume_symbol_timestamp ON volume_data(symbol, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_volume_timestamp ON volume_data(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_volume_symbol ON volume_data(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_volume_created_at ON volume_data(created_at)`,

		// Price data indexes
		`CREATE INDEX IF NOT EXISTS idx_price_symbol_timestamp ON price_data(symbol, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_price_timestamp ON price_data(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_price_symbol ON price_data(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_price_created_at ON price_data(created_at)`,

		// Watched symbols indexes
		`CREATE INDEX IF NOT EXISTS idx_watched_symbols_active ON watched_symbols(is_active)`,

		// Technical indicators indexes
		`CREATE INDEX IF NOT EXISTS idx_technical_indicators_symbol ON technical_indicators(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_technical_indicators_timestamp ON technical_indicators(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_technical_indicators_symbol_timestamp ON technical_indicators(symbol, timestamp)`,

		// Indicator alerts indexes
		`CREATE INDEX IF NOT EXISTS idx_indicator_alerts_symbol ON indicator_alerts(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_indicator_alerts_type ON indicator_alerts(alert_type)`,
		`CREATE INDEX IF NOT EXISTS idx_indicator_alerts_active ON indicator_alerts(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_indicator_alerts_triggered ON indicator_alerts(triggered_at)`,

		// Support/resistance levels indexes
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_symbol ON support_resistance_levels(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_type ON support_resistance_levels(level_type)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_strength ON support_resistance_levels(strength)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_active ON support_resistance_levels(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_last_touch ON support_resistance_levels(last_touch)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_symbol_type ON support_resistance_levels(symbol, level_type)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_symbol_active ON support_resistance_levels(symbol, is_active)`,

		// Pivot points indexes
		`CREATE INDEX IF NOT EXISTS idx_pivot_points_symbol ON pivot_points(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_pivot_points_timestamp ON pivot_points(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_pivot_points_type ON pivot_points(pivot_type)`,
		`CREATE INDEX IF NOT EXISTS idx_pivot_points_symbol_timestamp ON pivot_points(symbol, timestamp)`,

		// S/R level touches indexes
		`CREATE INDEX IF NOT EXISTS idx_sr_touches_level_id ON sr_level_touches(level_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_touches_symbol ON sr_level_touches(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_touches_time ON sr_level_touches(touch_time)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_touches_type ON sr_level_touches(touch_type)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_touches_symbol_time ON sr_level_touches(symbol, touch_time)`,

		// Trading setups indexes
		`CREATE INDEX IF NOT EXISTS idx_setups_symbol ON trading_setups(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_type ON trading_setups(setup_type)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_direction ON trading_setups(direction)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_status ON trading_setups(status)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_quality ON trading_setups(quality_score)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_detected ON trading_setups(detected_at)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_expires ON trading_setups(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_symbol_status ON trading_setups(symbol, status)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_symbol_quality ON trading_setups(symbol, quality_score)`,

		// Setup checklists indexes
		`CREATE INDEX IF NOT EXISTS idx_checklists_setup_id ON setup_checklists(setup_id)`,
		`CREATE INDEX IF NOT EXISTS idx_checklists_score ON setup_checklists(total_score)`,
		`CREATE INDEX IF NOT EXISTS idx_checklists_completion ON setup_checklists(completion_percent)`,

		// Setup alerts indexes
		`CREATE INDEX IF NOT EXISTS idx_setup_alerts_setup_id ON setup_alerts(setup_id)`,
		`CREATE INDEX IF NOT EXISTS idx_setup_alerts_symbol ON setup_alerts(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_setup_alerts_type ON setup_alerts(alert_type)`,
		`CREATE INDEX IF NOT EXISTS idx_setup_alerts_active ON setup_alerts(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_setup_alerts_triggered ON setup_alerts(triggered_at)`,
	}

	// Insert default symbols if they don't exist
	defaultSymbols := []string{
		"INSERT OR IGNORE INTO watched_symbols (symbol, name) VALUES ('PLTR', 'Palantir Technologies Inc.')",
		"INSERT OR IGNORE INTO watched_symbols (symbol, name) VALUES ('TSLA', 'Tesla Inc.')",
		"INSERT OR IGNORE INTO watched_symbols (symbol, name) VALUES ('BBAI', 'BigBear.ai Holdings Inc.')",
		"INSERT OR IGNORE INTO watched_symbols (symbol, name) VALUES ('MSFT', 'Microsoft Corporation')",
		"INSERT OR IGNORE INTO watched_symbols (symbol, name) VALUES ('NPWR', 'NET Power Inc.')",
	}

	// Combine all queries
	allQueries := append(queries, indexes...)
	allQueries = append(allQueries, defaultSymbols...)

	for _, query := range allQueries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// InsertVolumeData inserts volume data into the database
func (db *DB) InsertVolumeData(data *models.VolumeData) error {
	query := `
		INSERT OR REPLACE INTO volume_data
		(symbol, timestamp, volume, created_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		data.Symbol,
		data.Timestamp,
		data.Volume,
		data.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert volume data: %w", err)
	}

	return nil
}

// InsertVolumeDataBatch inserts multiple volume data records in a transaction
func (db *DB) InsertVolumeDataBatch(dataList []*models.VolumeData) error {
	if len(dataList) == 0 {
		return nil
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO volume_data
		(symbol, timestamp, volume, created_at)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, data := range dataList {
		_, err := stmt.Exec(
			data.Symbol,
			data.Timestamp,
			data.Volume,
			data.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to execute batch insert: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetVolumeData retrieves volume data for a symbol within a time range
func (db *DB) GetVolumeData(filter *models.VolumeDataFilter) ([]*models.VolumeData, error) {
	query := `
		SELECT id, symbol, timestamp, volume, created_at
		FROM volume_data
		WHERE symbol = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC
	`

	args := []interface{}{filter.Symbol, filter.From, filter.To}

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query volume data: %w", err)
	}
	defer rows.Close()

	var data []*models.VolumeData
	for rows.Next() {
		vd := &models.VolumeData{}
		err := rows.Scan(
			&vd.ID,
			&vd.Symbol,
			&vd.Timestamp,
			&vd.Volume,
			&vd.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan volume data: %w", err)
		}
		data = append(data, vd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating volume data rows: %w", err)
	}

	return data, nil
}

// GetLatestVolumeData retrieves the latest volume data for a symbol
func (db *DB) GetLatestVolumeData(symbol string) (*models.VolumeData, error) {
	query := `
		SELECT id, symbol, timestamp, volume, created_at
		FROM volume_data
		WHERE symbol = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`

	row := db.conn.QueryRow(query, symbol)

	vd := &models.VolumeData{}
	err := row.Scan(
		&vd.ID,
		&vd.Symbol,
		&vd.Timestamp,
		&vd.Volume,
		&vd.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No data found
		}
		return nil, fmt.Errorf("failed to get latest volume data: %w", err)
	}

	return vd, nil
}

// GetVolumeStats calculates volume statistics for a symbol
func (db *DB) GetVolumeStats(symbol string, days int) (*models.VolumeStats, error) {
	// Get current volume (latest data point)
	latest, err := db.GetLatestVolumeData(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest volume data: %w", err)
	}
	if latest == nil {
		return &models.VolumeStats{
			Symbol:          symbol,
			CurrentVolume:   0,
			AverageVolume:   0,
			VolumeRatio:     0,
			LastUpdate:      time.Time{},
			TotalDataPoints: 0,
		}, nil
	}

	// Calculate average volume over the last N days
	query := `
		SELECT AVG(volume), COUNT(*)
		FROM volume_data 
		WHERE symbol = ? AND timestamp >= datetime('now', '-' || ? || ' days')
	`

	var avgVolume float64
	var count int
	err = db.conn.QueryRow(query, symbol, days).Scan(&avgVolume, &count)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate volume stats: %w", err)
	}

	// Calculate volume ratio
	var volumeRatio float64
	if avgVolume > 0 {
		volumeRatio = float64(latest.Volume) / avgVolume
	}

	return &models.VolumeStats{
		Symbol:          symbol,
		CurrentVolume:   latest.Volume,
		AverageVolume:   avgVolume,
		VolumeRatio:     volumeRatio,
		LastUpdate:      latest.Timestamp,
		TotalDataPoints: count,
	}, nil
}

// GetAllSymbols returns all symbols that have volume data
func (db *DB) GetAllSymbols() ([]string, error) {
	query := `SELECT DISTINCT symbol FROM volume_data ORDER BY symbol`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
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

	return symbols, nil
}

// CleanupOldData removes volume data older than the specified number of days
func (db *DB) CleanupOldData(days int) (int64, error) {
	query := `DELETE FROM volume_data WHERE created_at < datetime('now', '-' || ? || ' days')`

	result, err := db.conn.Exec(query, days)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetDataCount returns the total number of volume data records
func (db *DB) GetDataCount() (int64, error) {
	query := `SELECT COUNT(*) FROM volume_data`

	var count int64
	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get data count: %w", err)
	}

	return count, nil
}

// GetDataCountBySymbol returns the number of volume data records for each symbol
func (db *DB) GetDataCountBySymbol() (map[string]int64, error) {
	query := `SELECT symbol, COUNT(*) FROM volume_data GROUP BY symbol`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get data count by symbol: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var symbol string
		var count int64
		if err := rows.Scan(&symbol, &count); err != nil {
			return nil, fmt.Errorf("failed to scan symbol count: %w", err)
		}
		counts[symbol] = count
	}

	return counts, nil
}

// HealthCheck performs a basic health check on the database
func (db *DB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.conn.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test a simple query
	var count int
	err := db.conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM volume_data LIMIT 1").Scan(&count)
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	return nil
}

// GetWatchedSymbols returns all active watched symbols
func (db *DB) GetWatchedSymbols() ([]string, error) {
	query := `SELECT symbol FROM watched_symbols WHERE is_active = TRUE ORDER BY symbol`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get watched symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, fmt.Errorf("failed to scan watched symbol: %w", err)
		}
		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

// AddWatchedSymbol adds a new symbol to watch
func (db *DB) AddWatchedSymbol(symbol, name string) error {
	query := `INSERT OR REPLACE INTO watched_symbols (symbol, name, is_active) VALUES (?, ?, TRUE)`

	_, err := db.conn.Exec(query, symbol, name)
	if err != nil {
		return fmt.Errorf("failed to add watched symbol: %w", err)
	}

	return nil
}

// RemoveWatchedSymbol marks a symbol as inactive
func (db *DB) RemoveWatchedSymbol(symbol string) error {
	query := `UPDATE watched_symbols SET is_active = FALSE WHERE symbol = ?`

	result, err := db.conn.Exec(query, symbol)
	if err != nil {
		return fmt.Errorf("failed to remove watched symbol: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("symbol not found: %s", symbol)
	}

	return nil
}

// GetWatchedSymbolsWithDetails returns all active watched symbols with details
func (db *DB) GetWatchedSymbolsWithDetails() ([]*models.WatchedSymbol, error) {
	query := `SELECT id, symbol, name, added_at, is_active FROM watched_symbols WHERE is_active = TRUE ORDER BY symbol`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get watched symbols with details: %w", err)
	}
	defer rows.Close()

	var symbols []*models.WatchedSymbol
	for rows.Next() {
		ws := &models.WatchedSymbol{}
		err := rows.Scan(&ws.ID, &ws.Symbol, &ws.Name, &ws.AddedAt, &ws.IsActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan watched symbol details: %w", err)
		}
		symbols = append(symbols, ws)
	}

	return symbols, nil
}
