package database

import (
	"fmt"
	"time"

	"market-watch-go/internal/models"
)

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
		return nil, fmt.Errorf("failed to get latest volume data: %w", err)
	}

	return vd, nil
}

// GetVolumeStats calculates volume statistics for a symbol
func (db *DB) GetVolumeStats(symbol string, days int) (*models.VolumeStats, error) {
	// Get current volume (latest data point)
	latest, err := db.GetLatestVolumeData(symbol)
	if err != nil {
		return &models.VolumeStats{
			Symbol:          symbol,
			CurrentVolume:   0,
			AverageVolume:   0,
			VolumeRatio:     0,
			LastUpdate:      time.Time{},
			TotalDataPoints: 0,
		}, nil
	}

	// Calculate average volume over the period
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

// CleanupOldData removes old data (placeholder)
func (db *DB) CleanupOldData(days int) (int64, error) {
	// For now, just cleanup volume data
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

// HealthCheck checks database health
func (db *DB) HealthCheck() error {
	return db.Ping()
}

// GetDataCount returns total data count (placeholder)
func (db *DB) GetDataCount() (int64, error) {
	var count int64
	err := db.conn.QueryRow("SELECT COUNT(*) FROM volume_data").Scan(&count)
	return count, err
}

// GetDataCountBySymbol returns data count by symbol
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

// GetAllSymbols returns all symbols in the database
func (db *DB) GetAllSymbols() ([]string, error) {
	query := `SELECT DISTINCT symbol FROM volume_data ORDER BY symbol`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all symbols: %w", err)
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

// GetWatchedSymbolsWithDetails returns watched symbols with details
func (db *DB) GetWatchedSymbolsWithDetails() ([]models.WatchedSymbol, error) {
	query := `SELECT id, symbol, name, added_at, is_active FROM watched_symbols WHERE is_active = 1 ORDER BY symbol`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query watched symbols: %w", err)
	}
	defer rows.Close()

	var symbols []models.WatchedSymbol
	for rows.Next() {
		var ws models.WatchedSymbol
		if err := rows.Scan(&ws.ID, &ws.Symbol, &ws.Name, &ws.AddedAt, &ws.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan watched symbol: %w", err)
		}
		symbols = append(symbols, ws)
	}

	return symbols, nil
}
