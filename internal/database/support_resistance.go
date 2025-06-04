package database

import (
	"database/sql"
	"fmt"
	"time"

	"market-watch-go/internal/models"
)

// CreateSupportResistanceTables creates all S/R related tables
func (db *DB) CreateSupportResistanceTables() error {
	tables := []string{
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
	}

	// Create tables
	for _, query := range tables {
		_, err := db.conn.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create S/R table: %w", err)
		}
	}

	// Create indexes for better performance
	indexes := []string{
		// Support/Resistance Levels indexes
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_symbol ON support_resistance_levels(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_type ON support_resistance_levels(level_type)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_strength ON support_resistance_levels(strength)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_active ON support_resistance_levels(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_last_touch ON support_resistance_levels(last_touch)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_symbol_type ON support_resistance_levels(symbol, level_type)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_levels_symbol_active ON support_resistance_levels(symbol, is_active)`,

		// Pivot Points indexes
		`CREATE INDEX IF NOT EXISTS idx_pivot_points_symbol ON pivot_points(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_pivot_points_timestamp ON pivot_points(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_pivot_points_type ON pivot_points(pivot_type)`,
		`CREATE INDEX IF NOT EXISTS idx_pivot_points_symbol_timestamp ON pivot_points(symbol, timestamp)`,

		// S/R Level Touches indexes
		`CREATE INDEX IF NOT EXISTS idx_sr_touches_level_id ON sr_level_touches(level_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_touches_symbol ON sr_level_touches(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_touches_time ON sr_level_touches(touch_time)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_touches_type ON sr_level_touches(touch_type)`,
		`CREATE INDEX IF NOT EXISTS idx_sr_touches_symbol_time ON sr_level_touches(symbol, touch_time)`,
	}

	for _, indexQuery := range indexes {
		_, err := db.conn.Exec(indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create S/R index: %w", err)
		}
	}

	return nil
}

// InsertSupportResistanceLevel inserts a new S/R level
func (db *DB) InsertSupportResistanceLevel(level *models.SupportResistanceLevel) error {
	query := `
		INSERT INTO support_resistance_levels 
		(symbol, level, level_type, strength, touches, first_touch, last_touch, 
		 volume_confirmed, avg_volume, max_bounce_percent, avg_bounce_percent, 
		 timeframe_origin, is_active, last_validated, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		level.Symbol, level.Level, level.LevelType, level.Strength, level.Touches,
		level.FirstTouch, level.LastTouch, level.VolumeConfirmed, level.AvgVolume,
		level.MaxBouncePercent, level.AvgBouncePercent, level.TimeframeOrigin,
		level.IsActive, level.LastValidated, level.CreatedAt, level.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert S/R level: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get S/R level ID: %w", err)
	}

	level.ID = id
	return nil
}

// UpdateSupportResistanceLevel updates an existing S/R level
func (db *DB) UpdateSupportResistanceLevel(level *models.SupportResistanceLevel) error {
	query := `
		UPDATE support_resistance_levels 
		SET strength = ?, touches = ?, last_touch = ?, volume_confirmed = ?, 
		    avg_volume = ?, max_bounce_percent = ?, avg_bounce_percent = ?, 
		    is_active = ?, last_validated = ?, updated_at = ?
		WHERE id = ?
	`

	level.UpdatedAt = time.Now()

	_, err := db.conn.Exec(query,
		level.Strength, level.Touches, level.LastTouch, level.VolumeConfirmed,
		level.AvgVolume, level.MaxBouncePercent, level.AvgBouncePercent,
		level.IsActive, level.LastValidated, level.UpdatedAt, level.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update S/R level: %w", err)
	}

	return nil
}

// GetSupportResistanceLevels retrieves S/R levels based on filter criteria
func (db *DB) GetSupportResistanceLevels(filter *models.SRDetectionFilter) ([]*models.SupportResistanceLevel, error) {
	query := `
		SELECT id, symbol, level, level_type, strength, touches, first_touch, last_touch,
		       volume_confirmed, avg_volume, max_bounce_percent, avg_bounce_percent,
		       timeframe_origin, is_active, last_validated, created_at, updated_at
		FROM support_resistance_levels 
		WHERE symbol = ?
	`
	args := []interface{}{filter.Symbol}

	// Add optional filters
	if filter.LevelType != "" && filter.LevelType != "both" {
		query += " AND level_type = ?"
		args = append(args, filter.LevelType)
	}

	if filter.MinStrength > 0 {
		query += " AND strength >= ?"
		args = append(args, filter.MinStrength)
	}

	if filter.MaxStrength > 0 {
		query += " AND strength <= ?"
		args = append(args, filter.MaxStrength)
	}

	if filter.IsActive != nil {
		query += " AND is_active = ?"
		args = append(args, *filter.IsActive)
	}

	if filter.MinTouches > 0 {
		query += " AND touches >= ?"
		args = append(args, filter.MinTouches)
	}

	if filter.PriceRange.Min > 0 {
		query += " AND level >= ?"
		args = append(args, filter.PriceRange.Min)
	}

	if filter.PriceRange.Max > 0 {
		query += " AND level <= ?"
		args = append(args, filter.PriceRange.Max)
	}

	// Order by strength descending
	query += " ORDER BY strength DESC, touches DESC"

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
		return nil, fmt.Errorf("failed to query S/R levels: %w", err)
	}
	defer rows.Close()

	var levels []*models.SupportResistanceLevel
	for rows.Next() {
		level := &models.SupportResistanceLevel{}
		err := rows.Scan(
			&level.ID, &level.Symbol, &level.Level, &level.LevelType,
			&level.Strength, &level.Touches, &level.FirstTouch, &level.LastTouch,
			&level.VolumeConfirmed, &level.AvgVolume, &level.MaxBouncePercent,
			&level.AvgBouncePercent, &level.TimeframeOrigin, &level.IsActive,
			&level.LastValidated, &level.CreatedAt, &level.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan S/R level: %w", err)
		}
		levels = append(levels, level)
	}

	return levels, nil
}

// GetNearestSupportResistance finds the nearest support and resistance levels to current price
func (db *DB) GetNearestSupportResistance(symbol string, currentPrice float64) (*models.SupportResistanceLevel, *models.SupportResistanceLevel, error) {
	// Find nearest support (below current price)
	supportQuery := `
		SELECT id, symbol, level, level_type, strength, touches, first_touch, last_touch,
		       volume_confirmed, avg_volume, max_bounce_percent, avg_bounce_percent,
		       timeframe_origin, is_active, last_validated, created_at, updated_at
		FROM support_resistance_levels 
		WHERE symbol = ? AND level_type = 'support' AND level < ? AND is_active = TRUE
		ORDER BY level DESC
		LIMIT 1
	`

	// Find nearest resistance (above current price)
	resistanceQuery := `
		SELECT id, symbol, level, level_type, strength, touches, first_touch, last_touch,
		       volume_confirmed, avg_volume, max_bounce_percent, avg_bounce_percent,
		       timeframe_origin, is_active, last_validated, created_at, updated_at
		FROM support_resistance_levels 
		WHERE symbol = ? AND level_type = 'resistance' AND level > ? AND is_active = TRUE
		ORDER BY level ASC
		LIMIT 1
	`

	var nearestSupport, nearestResistance *models.SupportResistanceLevel

	// Query nearest support
	row := db.conn.QueryRow(supportQuery, symbol, currentPrice)
	support := &models.SupportResistanceLevel{}
	err := row.Scan(
		&support.ID, &support.Symbol, &support.Level, &support.LevelType,
		&support.Strength, &support.Touches, &support.FirstTouch, &support.LastTouch,
		&support.VolumeConfirmed, &support.AvgVolume, &support.MaxBouncePercent,
		&support.AvgBouncePercent, &support.TimeframeOrigin, &support.IsActive,
		&support.LastValidated, &support.CreatedAt, &support.UpdatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, fmt.Errorf("failed to query nearest support: %w", err)
	}
	if err == nil {
		nearestSupport = support
	}

	// Query nearest resistance
	row = db.conn.QueryRow(resistanceQuery, symbol, currentPrice)
	resistance := &models.SupportResistanceLevel{}
	err = row.Scan(
		&resistance.ID, &resistance.Symbol, &resistance.Level, &resistance.LevelType,
		&resistance.Strength, &resistance.Touches, &resistance.FirstTouch, &resistance.LastTouch,
		&resistance.VolumeConfirmed, &resistance.AvgVolume, &resistance.MaxBouncePercent,
		&resistance.AvgBouncePercent, &resistance.TimeframeOrigin, &resistance.IsActive,
		&resistance.LastValidated, &resistance.CreatedAt, &resistance.UpdatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, fmt.Errorf("failed to query nearest resistance: %w", err)
	}
	if err == nil {
		nearestResistance = resistance
	}

	return nearestSupport, nearestResistance, nil
}

// InsertPivotPoint inserts a new pivot point
func (db *DB) InsertPivotPoint(pivot *models.PivotPoint) error {
	query := `
		INSERT INTO pivot_points 
		(symbol, timestamp, price, pivot_type, strength, volume, confirmed, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		pivot.Symbol, pivot.Timestamp, pivot.Price, pivot.PivotType,
		pivot.Strength, pivot.Volume, pivot.Confirmed, pivot.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert pivot point: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get pivot point ID: %w", err)
	}

	pivot.ID = id
	return nil
}

// GetPivotPoints retrieves pivot points for a symbol within a time range
func (db *DB) GetPivotPoints(symbol string, from, to time.Time, pivotType string) ([]*models.PivotPoint, error) {
	query := `
		SELECT id, symbol, timestamp, price, pivot_type, strength, volume, confirmed, created_at
		FROM pivot_points 
		WHERE symbol = ? AND timestamp BETWEEN ? AND ?
	`
	args := []interface{}{symbol, from, to}

	if pivotType != "" {
		query += " AND pivot_type = ?"
		args = append(args, pivotType)
	}

	query += " ORDER BY timestamp DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pivot points: %w", err)
	}
	defer rows.Close()

	var pivots []*models.PivotPoint
	for rows.Next() {
		pivot := &models.PivotPoint{}
		err := rows.Scan(
			&pivot.ID, &pivot.Symbol, &pivot.Timestamp, &pivot.Price,
			&pivot.PivotType, &pivot.Strength, &pivot.Volume,
			&pivot.Confirmed, &pivot.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pivot point: %w", err)
		}
		pivots = append(pivots, pivot)
	}

	return pivots, nil
}

// InsertSRLevelTouch records a touch of an S/R level
func (db *DB) InsertSRLevelTouch(touch *models.SRLevelTouch) error {
	query := `
		INSERT INTO sr_level_touches 
		(level_id, symbol, touch_time, touch_price, level, distance_percent, 
		 bounce_percent, volume_at_touch, volume_spike, bounce_confirmed, 
		 time_at_level, touch_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		touch.LevelID, touch.Symbol, touch.TouchTime, touch.TouchPrice,
		touch.Level, touch.DistancePercent, touch.BouncePercent,
		touch.VolumeAtTouch, touch.VolumeSpike, touch.BounceConfirmed,
		touch.TimeAtLevel, touch.TouchType, touch.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert S/R level touch: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get S/R level touch ID: %w", err)
	}

	touch.ID = id
	return nil
}

// GetRecentSRLevelTouches retrieves recent touches for S/R levels
func (db *DB) GetRecentSRLevelTouches(symbol string, hours int, limit int) ([]*models.SRLevelTouch, error) {
	query := `
		SELECT id, level_id, symbol, touch_time, touch_price, level, distance_percent,
		       bounce_percent, volume_at_touch, volume_spike, bounce_confirmed,
		       time_at_level, touch_type, created_at
		FROM sr_level_touches 
		WHERE symbol = ? AND touch_time > datetime('now', '-' || ? || ' hours')
		ORDER BY touch_time DESC
	`

	args := []interface{}{symbol, hours}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query S/R level touches: %w", err)
	}
	defer rows.Close()

	var touches []*models.SRLevelTouch
	for rows.Next() {
		touch := &models.SRLevelTouch{}
		err := rows.Scan(
			&touch.ID, &touch.LevelID, &touch.Symbol, &touch.TouchTime,
			&touch.TouchPrice, &touch.Level, &touch.DistancePercent,
			&touch.BouncePercent, &touch.VolumeAtTouch, &touch.VolumeSpike,
			&touch.BounceConfirmed, &touch.TimeAtLevel, &touch.TouchType,
			&touch.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan S/R level touch: %w", err)
		}
		touches = append(touches, touch)
	}

	return touches, nil
}

// DeactivateOldSRLevels deactivates S/R levels that haven't been touched recently
func (db *DB) DeactivateOldSRLevels(maxAgeHours int) (int64, error) {
	query := `
		UPDATE support_resistance_levels 
		SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP
		WHERE last_touch < datetime('now', '-' || ? || ' hours') AND is_active = TRUE
	`

	result, err := db.conn.Exec(query, maxAgeHours)
	if err != nil {
		return 0, fmt.Errorf("failed to deactivate old S/R levels: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetSRLevelSummary calculates summary statistics for S/R levels
func (db *DB) GetSRLevelSummary(symbol string) (*models.SRLevelSummary, error) {
	summary := &models.SRLevelSummary{}

	// Get basic counts and averages
	query := `
		SELECT 
			COUNT(*) as total_levels,
			COUNT(CASE WHEN level_type = 'support' AND is_active = TRUE THEN 1 END) as support_count,
			COUNT(CASE WHEN level_type = 'resistance' AND is_active = TRUE THEN 1 END) as resistance_count,
			AVG(CASE WHEN is_active = TRUE THEN strength END) as avg_strength,
			MAX(CASE WHEN is_active = TRUE THEN strength END) as strongest_level,
			MIN(CASE WHEN is_active = TRUE THEN strength END) as weakest_level
		FROM support_resistance_levels 
		WHERE symbol = ? AND is_active = TRUE
	`

	err := db.conn.QueryRow(query, symbol).Scan(
		&summary.TotalLevels, &summary.SupportCount, &summary.ResistanceCount,
		&summary.AvgStrength, &summary.StrongestLevel, &summary.WeakestLevel,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get S/R level summary: %w", err)
	}

	// Get recent touch information
	touchQuery := `
		SELECT COUNT(*), MAX(touch_time)
		FROM sr_level_touches 
		WHERE symbol = ? AND touch_time > datetime('now', '-24 hours')
	`

	err = db.conn.QueryRow(touchQuery, symbol).Scan(
		&summary.RecentTouchCount, &summary.LastTouchTime,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get recent touch info: %w", err)
	}

	return summary, nil
}

// CleanupOldSRData removes old S/R data based on retention policy
func (db *DB) CleanupOldSRData(days int) (int64, error) {
	var totalDeleted int64

	// Cleanup old touches
	touchResult, err := db.conn.Exec(
		"DELETE FROM sr_level_touches WHERE created_at < datetime('now', '-' || ? || ' days')",
		days,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old S/R touches: %w", err)
	}

	touchesDeleted, _ := touchResult.RowsAffected()
	totalDeleted += touchesDeleted

	// Cleanup old pivot points
	pivotResult, err := db.conn.Exec(
		"DELETE FROM pivot_points WHERE created_at < datetime('now', '-' || ? || ' days')",
		days,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old pivot points: %w", err)
	}

	pivotsDeleted, _ := pivotResult.RowsAffected()
	totalDeleted += pivotsDeleted

	// Cleanup inactive S/R levels older than retention period
	levelResult, err := db.conn.Exec(
		"DELETE FROM support_resistance_levels WHERE is_active = FALSE AND updated_at < datetime('now', '-' || ? || ' days')",
		days,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old S/R levels: %w", err)
	}

	levelsDeleted, _ := levelResult.RowsAffected()
	totalDeleted += levelsDeleted

	return totalDeleted, nil
}
