package database

import (
	"fmt"
	"time"

	"market-watch-go/internal/models"
)

// InsertPriceData inserts price data into the database
func (db *DB) InsertPriceData(data *models.PriceData) error {
	query := `
		INSERT OR REPLACE INTO price_data
		(symbol, timestamp, open_price, high_price, low_price, close_price, volume, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		data.Symbol,
		data.Timestamp,
		data.Open,
		data.High,
		data.Low,
		data.Close,
		data.Volume,
		data.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert price data: %w", err)
	}

	return nil
}

// InsertPriceDataBatch inserts multiple price data records in a transaction
func (db *DB) InsertPriceDataBatch(dataList []*models.PriceData) error {
	if len(dataList) == 0 {
		return nil
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO price_data
		(symbol, timestamp, open_price, high_price, low_price, close_price, volume, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, data := range dataList {
		_, err := stmt.Exec(
			data.Symbol,
			data.Timestamp,
			data.Open,
			data.High,
			data.Low,
			data.Close,
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

// GetPriceData retrieves price data for a symbol within a time range
func (db *DB) GetPriceData(filter *models.PriceDataFilter) ([]*models.PriceData, error) {
	query := `
		SELECT id, symbol, timestamp, open_price, high_price, low_price, close_price, volume, created_at
		FROM price_data
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
		return nil, fmt.Errorf("failed to query price data: %w", err)
	}
	defer rows.Close()

	var data []*models.PriceData
	for rows.Next() {
		pd := &models.PriceData{}
		err := rows.Scan(
			&pd.ID,
			&pd.Symbol,
			&pd.Timestamp,
			&pd.Open,
			&pd.High,
			&pd.Low,
			&pd.Close,
			&pd.Volume,
			&pd.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price data: %w", err)
		}
		data = append(data, pd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price data rows: %w", err)
	}

	return data, nil
}

// GetPriceDataRange retrieves price data for a symbol within a specific time range
func (db *DB) GetPriceDataRange(symbol string, startTime, endTime time.Time) ([]*models.PriceData, error) {
	filter := &models.PriceDataFilter{
		Symbol: symbol,
		From:   startTime,
		To:     endTime,
	}
	return db.GetPriceData(filter)
}

// GetLatestPriceData retrieves the latest price data for a symbol
func (db *DB) GetLatestPriceData(symbol string) (*models.PriceData, error) {
	query := `
		SELECT id, symbol, timestamp, open_price, high_price, low_price, close_price, volume, created_at
		FROM price_data
		WHERE symbol = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`

	row := db.conn.QueryRow(query, symbol)

	pd := &models.PriceData{}
	err := row.Scan(
		&pd.ID,
		&pd.Symbol,
		&pd.Timestamp,
		&pd.Open,
		&pd.High,
		&pd.Low,
		&pd.Close,
		&pd.Volume,
		&pd.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get latest price data: %w", err)
	}

	return pd, nil
}

// GetPriceDataCount returns the total number of price data records
func (db *DB) GetPriceDataCount() (int64, error) {
	query := `SELECT COUNT(*) FROM price_data`

	var count int64
	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get price data count: %w", err)
	}

	return count, nil
}

// GetPriceDataCountBySymbol returns the number of price data records for each symbol
func (db *DB) GetPriceDataCountBySymbol() (map[string]int64, error) {
	query := `SELECT symbol, COUNT(*) FROM price_data GROUP BY symbol`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get price data count by symbol: %w", err)
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

// CleanupOldPriceData removes price data older than the specified number of days
func (db *DB) CleanupOldPriceData(days int) (int64, error) {
	query := `DELETE FROM price_data WHERE created_at < datetime('now', '-' || ? || ' days')`

	result, err := db.conn.Exec(query, days)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old price data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetPriceStats calculates price statistics for a symbol
func (db *DB) GetPriceStats(symbol string, days int) (*models.PriceStats, error) {
	// Get current price (latest data point)
	latest, err := db.GetLatestPriceData(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest price data: %w", err)
	}
	if latest == nil {
		return &models.PriceStats{
			Symbol:             symbol,
			CurrentPrice:       0,
			OpenPrice:          0,
			HighPrice:          0,
			LowPrice:           0,
			PriceChange:        0,
			PriceChangePercent: 0,
			LastUpdate:         time.Time{},
		}, nil
	}

	// Calculate statistics over the last N days
	query := `
		SELECT 
			MIN(low_price) as min_low,
			MAX(high_price) as max_high,
			COUNT(*) as count
		FROM price_data 
		WHERE symbol = ? AND timestamp >= datetime('now', '-' || ? || ' days')
	`

	var minLow, maxHigh float64
	var count int
	err = db.conn.QueryRow(query, symbol, days).Scan(&minLow, &maxHigh, &count)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate price stats: %w", err)
	}

	// Get opening price (first record of the period)
	openQuery := `
		SELECT open_price 
		FROM price_data 
		WHERE symbol = ? AND timestamp >= datetime('now', '-' || ? || ' days')
		ORDER BY timestamp ASC 
		LIMIT 1
	`

	var openPrice float64
	err = db.conn.QueryRow(openQuery, symbol, days).Scan(&openPrice)
	if err != nil {
		openPrice = latest.Open // Fallback to current open
	}

	// Calculate price change
	priceChange := latest.Close - openPrice
	var priceChangePercent float64
	if openPrice > 0 {
		priceChangePercent = (priceChange / openPrice) * 100
	}

	return &models.PriceStats{
		Symbol:             symbol,
		CurrentPrice:       latest.Close,
		OpenPrice:          openPrice,
		HighPrice:          maxHigh,
		LowPrice:           minLow,
		PriceChange:        priceChange,
		PriceChangePercent: priceChangePercent,
		LastUpdate:         latest.Timestamp,
	}, nil
}

// GetSymbolsWithPriceData returns all symbols that have price data
func (db *DB) GetSymbolsWithPriceData() ([]string, error) {
	query := `SELECT DISTINCT symbol FROM price_data ORDER BY symbol`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols with price data: %w", err)
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

// GetPriceDataForAnalysis retrieves price data suitable for technical analysis
func (db *DB) GetPriceDataForAnalysis(symbol string, limit int) ([]*models.PriceData, error) {
	query := `
		SELECT id, symbol, timestamp, open_price, high_price, low_price, close_price, volume, created_at
		FROM price_data
		WHERE symbol = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := db.conn.Query(query, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query price data for analysis: %w", err)
	}
	defer rows.Close()

	var data []*models.PriceData
	for rows.Next() {
		pd := &models.PriceData{}
		err := rows.Scan(
			&pd.ID,
			&pd.Symbol,
			&pd.Timestamp,
			&pd.Open,
			&pd.High,
			&pd.Low,
			&pd.Close,
			&pd.Volume,
			&pd.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price data: %w", err)
		}
		data = append(data, pd)
	}

	// Reverse slice to get chronological order (oldest first)
	for i := len(data)/2 - 1; i >= 0; i-- {
		opp := len(data) - 1 - i
		data[i], data[opp] = data[opp], data[i]
	}

	return data, nil
}
