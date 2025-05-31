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
			price DECIMAL(10,2),
			open_price DECIMAL(10,2),
			high_price DECIMAL(10,2),
			low_price DECIMAL(10,2),
			close_price DECIMAL(10,2),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(symbol, timestamp)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_symbol_timestamp ON volume_data(symbol, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_timestamp ON volume_data(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_symbol ON volume_data(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_created_at ON volume_data(created_at)`,
	}

	for _, query := range queries {
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
		(symbol, timestamp, volume, price, open_price, high_price, low_price, close_price, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		data.Symbol,
		data.Timestamp,
		data.Volume,
		data.Price,
		data.Open,
		data.High,
		data.Low,
		data.Close,
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
		(symbol, timestamp, volume, price, open_price, high_price, low_price, close_price, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
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
			data.Price,
			data.Open,
			data.High,
			data.Low,
			data.Close,
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
		SELECT id, symbol, timestamp, volume, price, open_price, high_price, low_price, close_price, created_at
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
			&vd.Price,
			&vd.Open,
			&vd.High,
			&vd.Low,
			&vd.Close,
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
		SELECT id, symbol, timestamp, volume, price, open_price, high_price, low_price, close_price, created_at
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
		&vd.Price,
		&vd.Open,
		&vd.High,
		&vd.Low,
		&vd.Close,
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
