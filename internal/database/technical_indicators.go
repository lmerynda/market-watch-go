package database

import (
	"database/sql"
	"fmt"
	"time"

	"market-watch-go/internal/models"
)

// CreateTechnicalIndicatorsTable creates the technical_indicators table
func (db *DB) CreateTechnicalIndicatorsTable() error {
	query := `
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
	`

	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create technical_indicators table: %w", err)
	}

	// Create indexes for better performance
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_technical_indicators_symbol ON technical_indicators(symbol);`,
		`CREATE INDEX IF NOT EXISTS idx_technical_indicators_timestamp ON technical_indicators(timestamp);`,
		`CREATE INDEX IF NOT EXISTS idx_technical_indicators_symbol_timestamp ON technical_indicators(symbol, timestamp);`,
	}

	for _, indexQuery := range indexes {
		_, err := db.conn.Exec(indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// InsertTechnicalIndicators inserts technical indicators into the database
func (db *DB) InsertTechnicalIndicators(indicators *models.TechnicalIndicators) error {
	query := `
		INSERT OR REPLACE INTO technical_indicators 
		(symbol, timestamp, rsi_14, rsi_30, macd_line, macd_signal, macd_histogram,
		 sma_20, sma_50, sma_200, ema_20, ema_50, vwap, volume_ratio,
		 bb_upper, bb_middle, bb_lower, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		indicators.Symbol,
		indicators.Timestamp,
		indicators.RSI14,
		indicators.RSI30,
		indicators.MACD,
		indicators.MACDSignal,
		indicators.MACDHistogram,
		indicators.SMA20,
		indicators.SMA50,
		indicators.SMA200,
		indicators.EMA20,
		indicators.EMA50,
		indicators.VWAP,
		indicators.VolumeRatio,
		indicators.BBUpper,
		indicators.BBMiddle,
		indicators.BBLower,
		indicators.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert technical indicators: %w", err)
	}

	return nil
}

// GetLatestTechnicalIndicators retrieves the latest technical indicators for a symbol
func (db *DB) GetLatestTechnicalIndicators(symbol string) (*models.TechnicalIndicators, error) {
	query := `
		SELECT id, symbol, timestamp, rsi_14, rsi_30, macd_line, macd_signal, macd_histogram,
		       sma_20, sma_50, sma_200, ema_20, ema_50, vwap, volume_ratio,
		       bb_upper, bb_middle, bb_lower, created_at
		FROM technical_indicators 
		WHERE symbol = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`

	row := db.conn.QueryRow(query, symbol)

	indicators := &models.TechnicalIndicators{}
	err := row.Scan(
		&indicators.ID,
		&indicators.Symbol,
		&indicators.Timestamp,
		&indicators.RSI14,
		&indicators.RSI30,
		&indicators.MACD,
		&indicators.MACDSignal,
		&indicators.MACDHistogram,
		&indicators.SMA20,
		&indicators.SMA50,
		&indicators.SMA200,
		&indicators.EMA20,
		&indicators.EMA50,
		&indicators.VWAP,
		&indicators.VolumeRatio,
		&indicators.BBUpper,
		&indicators.BBMiddle,
		&indicators.BBLower,
		&indicators.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No data found
		}
		return nil, fmt.Errorf("failed to get latest technical indicators: %w", err)
	}

	return indicators, nil
}

// GetTechnicalIndicators retrieves technical indicators for a symbol within a time range
func (db *DB) GetTechnicalIndicators(filter *models.IndicatorFilter) ([]*models.TechnicalIndicators, error) {
	query := `
		SELECT id, symbol, timestamp, rsi_14, rsi_30, macd_line, macd_signal, macd_histogram,
		       sma_20, sma_50, sma_200, ema_20, ema_50, vwap, volume_ratio,
		       bb_upper, bb_middle, bb_lower, created_at
		FROM technical_indicators 
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
		return nil, fmt.Errorf("failed to query technical indicators: %w", err)
	}
	defer rows.Close()

	var indicators []*models.TechnicalIndicators
	for rows.Next() {
		ind := &models.TechnicalIndicators{}
		err := rows.Scan(
			&ind.ID,
			&ind.Symbol,
			&ind.Timestamp,
			&ind.RSI14,
			&ind.RSI30,
			&ind.MACD,
			&ind.MACDSignal,
			&ind.MACDHistogram,
			&ind.SMA20,
			&ind.SMA50,
			&ind.SMA200,
			&ind.EMA20,
			&ind.EMA50,
			&ind.VWAP,
			&ind.VolumeRatio,
			&ind.BBUpper,
			&ind.BBMiddle,
			&ind.BBLower,
			&ind.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan technical indicators: %w", err)
		}
		indicators = append(indicators, ind)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating technical indicators rows: %w", err)
	}

	return indicators, nil
}

// CreateIndicatorAlertsTable creates the indicator_alerts table
func (db *DB) CreateIndicatorAlertsTable() error {
	query := `
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
	`

	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create indicator_alerts table: %w", err)
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_indicator_alerts_symbol ON indicator_alerts(symbol);`,
		`CREATE INDEX IF NOT EXISTS idx_indicator_alerts_type ON indicator_alerts(alert_type);`,
		`CREATE INDEX IF NOT EXISTS idx_indicator_alerts_active ON indicator_alerts(is_active);`,
		`CREATE INDEX IF NOT EXISTS idx_indicator_alerts_triggered ON indicator_alerts(triggered_at);`,
	}

	for _, indexQuery := range indexes {
		_, err := db.conn.Exec(indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create indicator alerts index: %w", err)
		}
	}

	return nil
}

// InsertIndicatorAlert inserts an indicator alert into the database
func (db *DB) InsertIndicatorAlert(alert *models.IndicatorAlert) error {
	query := `
		INSERT INTO indicator_alerts 
		(symbol, alert_type, indicator, value, threshold, message, triggered_at, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		alert.Symbol,
		alert.AlertType,
		alert.Indicator,
		alert.Value,
		alert.Threshold,
		alert.Message,
		alert.TriggeredAt,
		alert.IsActive,
		alert.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert indicator alert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get alert ID: %w", err)
	}

	alert.ID = id
	return nil
}

// GetActiveIndicatorAlerts retrieves active alerts for a symbol
func (db *DB) GetActiveIndicatorAlerts(symbol string) ([]*models.IndicatorAlert, error) {
	query := `
		SELECT id, symbol, alert_type, indicator, value, threshold, message, triggered_at, is_active, created_at
		FROM indicator_alerts 
		WHERE symbol = ? AND is_active = TRUE
		ORDER BY triggered_at DESC
	`

	rows, err := db.conn.Query(query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to query active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.IndicatorAlert
	for rows.Next() {
		alert := &models.IndicatorAlert{}
		err := rows.Scan(
			&alert.ID,
			&alert.Symbol,
			&alert.AlertType,
			&alert.Indicator,
			&alert.Value,
			&alert.Threshold,
			&alert.Message,
			&alert.TriggeredAt,
			&alert.IsActive,
			&alert.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// DeactivateIndicatorAlert marks an alert as inactive
func (db *DB) DeactivateIndicatorAlert(alertID int64) error {
	query := `UPDATE indicator_alerts SET is_active = FALSE WHERE id = ?`

	_, err := db.conn.Exec(query, alertID)
	if err != nil {
		return fmt.Errorf("failed to deactivate alert: %w", err)
	}

	return nil
}

// CleanupOldIndicatorAlerts removes old alerts based on retention policy
func (db *DB) CleanupOldIndicatorAlerts(days int) (int64, error) {
	query := `DELETE FROM indicator_alerts WHERE created_at < datetime('now', '-' || ? || ' days')`

	result, err := db.conn.Exec(query, days)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old indicator alerts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetIndicatorAlertsStats returns statistics about indicator alerts
func (db *DB) GetIndicatorAlertsStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total alerts count
	var totalAlerts int64
	err := db.conn.QueryRow("SELECT COUNT(*) FROM indicator_alerts").Scan(&totalAlerts)
	if err != nil {
		return nil, fmt.Errorf("failed to get total alerts count: %w", err)
	}
	stats["total_alerts"] = totalAlerts

	// Active alerts count
	var activeAlerts int64
	err = db.conn.QueryRow("SELECT COUNT(*) FROM indicator_alerts WHERE is_active = TRUE").Scan(&activeAlerts)
	if err != nil {
		return nil, fmt.Errorf("failed to get active alerts count: %w", err)
	}
	stats["active_alerts"] = activeAlerts

	// Alerts by type
	alertsByType := make(map[string]int64)
	rows, err := db.conn.Query("SELECT alert_type, COUNT(*) FROM indicator_alerts WHERE is_active = TRUE GROUP BY alert_type")
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts by type: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alertType string
		var count int64
		if err := rows.Scan(&alertType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan alert type stats: %w", err)
		}
		alertsByType[alertType] = count
	}
	stats["alerts_by_type"] = alertsByType

	// Recent alerts (last 24 hours)
	var recentAlerts int64
	err = db.conn.QueryRow("SELECT COUNT(*) FROM indicator_alerts WHERE triggered_at > datetime('now', '-1 day')").Scan(&recentAlerts)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent alerts count: %w", err)
	}
	stats["recent_alerts_24h"] = recentAlerts

	return stats, nil
}

// CleanupOldTechnicalIndicators removes old technical indicator data
func (db *DB) CleanupOldTechnicalIndicators(days int) (int64, error) {
	query := `DELETE FROM technical_indicators WHERE created_at < datetime('now', '-' || ? || ' days')`

	result, err := db.conn.Exec(query, days)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old technical indicators: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetTechnicalIndicatorsStats returns statistics about technical indicators
func (db *DB) GetTechnicalIndicatorsStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total indicators count
	var totalIndicators int64
	err := db.conn.QueryRow("SELECT COUNT(*) FROM technical_indicators").Scan(&totalIndicators)
	if err != nil {
		return nil, fmt.Errorf("failed to get total indicators count: %w", err)
	}
	stats["total_indicators"] = totalIndicators

	// Indicators by symbol
	indicatorsBySymbol := make(map[string]int64)
	rows, err := db.conn.Query("SELECT symbol, COUNT(*) FROM technical_indicators GROUP BY symbol")
	if err != nil {
		return nil, fmt.Errorf("failed to get indicators by symbol: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var symbol string
		var count int64
		if err := rows.Scan(&symbol, &count); err != nil {
			return nil, fmt.Errorf("failed to scan symbol stats: %w", err)
		}
		indicatorsBySymbol[symbol] = count
	}
	stats["indicators_by_symbol"] = indicatorsBySymbol

	// Latest update time
	var latestUpdate time.Time
	err = db.conn.QueryRow("SELECT MAX(timestamp) FROM technical_indicators").Scan(&latestUpdate)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get latest update time: %w", err)
	}
	if err != sql.ErrNoRows {
		stats["latest_update"] = latestUpdate
	}

	return stats, nil
}
