package database

import (
	"database/sql"
	"fmt"

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
		if err == sql.ErrNoRows {
			return nil, nil // No data found
		}
		return nil, fmt.Errorf("failed to get latest price data: %w", err)
	}

	return pd, nil
}

// GetPriceStats calculates price statistics for a symbol
func (db *DB) GetPriceStats(symbol string, days int) (*models.PriceStats, error) {
	// Get current price (latest data point)
	latest, err := db.GetLatestPriceData(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest price data: %w", err)
	}
	if latest == nil {
		return nil, nil
	}

	// Get opening price for the day
	query := `
		SELECT open_price, high_price, low_price, close_price
		FROM price_data 
		WHERE symbol = ? AND DATE(timestamp) = DATE(?)
		ORDER BY timestamp ASC
		LIMIT 1
	`

	var openPrice, highPrice, lowPrice, closePrice float64
	err = db.conn.QueryRow(query, symbol, latest.Timestamp).Scan(&openPrice, &highPrice, &lowPrice, &closePrice)
	if err != nil {
		if err == sql.ErrNoRows {
			// Use latest data if no data for today
			openPrice = latest.Open
			highPrice = latest.High
			lowPrice = latest.Low
		} else {
			return nil, fmt.Errorf("failed to get daily price data: %w", err)
		}
	}

	// Calculate price change and percentage
	priceChange := latest.Close - openPrice
	var priceChangePercent float64
	if openPrice > 0 {
		priceChangePercent = (priceChange / openPrice) * 100
	}

	return &models.PriceStats{
		Symbol:             symbol,
		CurrentPrice:       latest.Close,
		OpenPrice:          openPrice,
		HighPrice:          highPrice,
		LowPrice:           lowPrice,
		PriceChange:        priceChange,
		PriceChangePercent: priceChangePercent,
		LastUpdate:         latest.Timestamp,
	}, nil
}

// CleanupOldPriceData removes old price data based on retention policy
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
