package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"market-watch-go/internal/models"
)

// CreateHeadShouldersPatternTables creates all head and shoulders pattern related tables
func (db *DB) CreateHeadShouldersPatternTables() error {
	tables := []string{
		// Head and Shoulders patterns table
		`CREATE TABLE IF NOT EXISTS head_shoulders_patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setup_id INTEGER NOT NULL,
			symbol TEXT NOT NULL,
			pattern_type TEXT NOT NULL CHECK (pattern_type IN ('inverse_head_shoulders', 'head_shoulders')),
			
			-- Pattern Points (JSON stored as TEXT)
			left_shoulder_high TEXT NOT NULL,
			left_shoulder_low TEXT NOT NULL,
			head_high TEXT NOT NULL,
			head_low TEXT NOT NULL,
			right_shoulder_high TEXT,
			right_shoulder_low TEXT,
			
			-- Neckline Analysis
			neckline_level REAL,
			neckline_slope REAL,
			neckline_touch1 TEXT,
			neckline_touch2 TEXT,
			
			-- Pattern Measurements
			pattern_width INTEGER, -- Duration in minutes
			pattern_height REAL,
			symmetry REAL,
			
			-- Thesis Components (JSON stored as TEXT)
			thesis_components TEXT NOT NULL DEFAULT '{}',
			
			-- Status
			detected_at DATETIME NOT NULL,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_complete BOOLEAN DEFAULT FALSE,
			current_phase TEXT DEFAULT 'formation' CHECK (current_phase IN ('formation', 'breakout', 'target_pursuit', 'completed')),
			
			FOREIGN KEY (setup_id) REFERENCES trading_setups(id)
		)`,

		// Thesis components tracking table
		`CREATE TABLE IF NOT EXISTS thesis_components (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pattern_id INTEGER NOT NULL,
			component_name TEXT NOT NULL,
			component_type TEXT NOT NULL, -- 'formation', 'volume', 'breakout', 'target'
			description TEXT,
			is_completed BOOLEAN DEFAULT FALSE,
			completed_at DATETIME,
			is_required BOOLEAN DEFAULT FALSE,
			weight REAL DEFAULT 1.0,
			confidence_level REAL DEFAULT 0.0,
			evidence TEXT, -- JSON array of evidence
			last_checked DATETIME NOT NULL,
			auto_detected BOOLEAN DEFAULT TRUE,
			notification_sent BOOLEAN DEFAULT FALSE,
			
			FOREIGN KEY (pattern_id) REFERENCES head_shoulders_patterns(id),
			UNIQUE(pattern_id, component_name)
		)`,

		// Pattern monitoring alerts
		`CREATE TABLE IF NOT EXISTS pattern_alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pattern_id INTEGER NOT NULL,
			symbol TEXT NOT NULL,
			component_name TEXT NOT NULL,
			alert_type TEXT NOT NULL, -- 'component_completed', 'breakout_confirmed', 'target_reached'
			message TEXT NOT NULL,
			triggered_at DATETIME NOT NULL,
			notification_sent BOOLEAN DEFAULT FALSE,
			email_sent BOOLEAN DEFAULT FALSE,
			
			FOREIGN KEY (pattern_id) REFERENCES head_shoulders_patterns(id)
		)`,
	}

	// Create tables
	for _, query := range tables {
		_, err := db.conn.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create head and shoulders pattern table: %w", err)
		}
	}

	// Create indexes for better performance
	indexes := []string{
		// Head and Shoulders patterns indexes
		`CREATE INDEX IF NOT EXISTS idx_hs_patterns_symbol ON head_shoulders_patterns(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_hs_patterns_type ON head_shoulders_patterns(pattern_type)`,
		`CREATE INDEX IF NOT EXISTS idx_hs_patterns_phase ON head_shoulders_patterns(current_phase)`,
		`CREATE INDEX IF NOT EXISTS idx_hs_patterns_complete ON head_shoulders_patterns(is_complete)`,
		`CREATE INDEX IF NOT EXISTS idx_hs_patterns_detected ON head_shoulders_patterns(detected_at)`,
		`CREATE INDEX IF NOT EXISTS idx_hs_patterns_symbol_phase ON head_shoulders_patterns(symbol, current_phase)`,
		`CREATE INDEX IF NOT EXISTS idx_hs_patterns_symbol_complete ON head_shoulders_patterns(symbol, is_complete)`,

		// Thesis components indexes
		`CREATE INDEX IF NOT EXISTS idx_thesis_pattern_id ON thesis_components(pattern_id)`,
		`CREATE INDEX IF NOT EXISTS idx_thesis_component_name ON thesis_components(component_name)`,
		`CREATE INDEX IF NOT EXISTS idx_thesis_completed ON thesis_components(is_completed)`,
		`CREATE INDEX IF NOT EXISTS idx_thesis_required ON thesis_components(is_required)`,
		`CREATE INDEX IF NOT EXISTS idx_thesis_notification ON thesis_components(notification_sent)`,

		// Pattern alerts indexes
		`CREATE INDEX IF NOT EXISTS idx_pattern_alerts_pattern_id ON pattern_alerts(pattern_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pattern_alerts_symbol ON pattern_alerts(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_pattern_alerts_type ON pattern_alerts(alert_type)`,
		`CREATE INDEX IF NOT EXISTS idx_pattern_alerts_triggered ON pattern_alerts(triggered_at)`,
		`CREATE INDEX IF NOT EXISTS idx_pattern_alerts_notification ON pattern_alerts(notification_sent)`,
	}

	for _, indexQuery := range indexes {
		_, err := db.conn.Exec(indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create head and shoulders pattern index: %w", err)
		}
	}

	return nil
}

// InsertHeadShouldersPattern inserts a new head and shoulders pattern
func (db *DB) InsertHeadShouldersPattern(pattern *models.HeadShouldersPattern) error {
	// Marshal pattern points to JSON
	leftShoulderHigh, _ := models.MarshalPatternPointsJSON(pattern.LeftShoulderHigh)
	leftShoulderLow, _ := models.MarshalPatternPointsJSON(pattern.LeftShoulderLow)
	headHigh, _ := models.MarshalPatternPointsJSON(pattern.HeadHigh)
	headLow, _ := models.MarshalPatternPointsJSON(pattern.HeadLow)
	rightShoulderHigh, _ := models.MarshalPatternPointsJSON(pattern.RightShoulderHigh)
	rightShoulderLow, _ := models.MarshalPatternPointsJSON(pattern.RightShoulderLow)
	necklineTouch1, _ := models.MarshalPatternPointsJSON(pattern.NecklineTouch1)
	necklineTouch2, _ := models.MarshalPatternPointsJSON(pattern.NecklineTouch2)

	// Marshal thesis components to JSON
	thesisData, err := models.MarshalThesisJSON(pattern.ThesisComponents)
	if err != nil {
		return fmt.Errorf("failed to marshal thesis components: %w", err)
	}

	query := `
		INSERT INTO head_shoulders_patterns 
		(setup_id, symbol, pattern_type, left_shoulder_high, left_shoulder_low,
		 head_high, head_low, right_shoulder_high, right_shoulder_low,
		 neckline_level, neckline_slope, neckline_touch1, neckline_touch2,
		 pattern_width, pattern_height, symmetry, thesis_components,
		 detected_at, last_updated, is_complete, current_phase)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		pattern.SetupID, pattern.Symbol, pattern.PatternType,
		string(leftShoulderHigh), string(leftShoulderLow),
		string(headHigh), string(headLow),
		string(rightShoulderHigh), string(rightShoulderLow),
		pattern.NecklineLevel, pattern.NecklineSlope,
		string(necklineTouch1), string(necklineTouch2),
		pattern.PatternWidth, pattern.PatternHeight, pattern.Symmetry,
		string(thesisData), pattern.DetectedAt, pattern.LastUpdated,
		pattern.IsComplete, pattern.CurrentPhase,
	)

	if err != nil {
		return fmt.Errorf("failed to insert head and shoulders pattern: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get pattern ID: %w", err)
	}

	pattern.ID = id

	// Insert thesis components
	err = db.insertThesisComponents(pattern.ID, &pattern.ThesisComponents)
	if err != nil {
		return fmt.Errorf("failed to insert thesis components: %w", err)
	}

	return nil
}

// UpdateHeadShouldersPattern updates an existing head and shoulders pattern
func (db *DB) UpdateHeadShouldersPattern(pattern *models.HeadShouldersPattern) error {
	// Marshal pattern points to JSON
	rightShoulderHigh, _ := models.MarshalPatternPointsJSON(pattern.RightShoulderHigh)
	rightShoulderLow, _ := models.MarshalPatternPointsJSON(pattern.RightShoulderLow)
	necklineTouch1, _ := models.MarshalPatternPointsJSON(pattern.NecklineTouch1)
	necklineTouch2, _ := models.MarshalPatternPointsJSON(pattern.NecklineTouch2)

	// Marshal thesis components to JSON
	thesisData, err := models.MarshalThesisJSON(pattern.ThesisComponents)
	if err != nil {
		return fmt.Errorf("failed to marshal thesis components: %w", err)
	}

	query := `
		UPDATE head_shoulders_patterns 
		SET right_shoulder_high = ?, right_shoulder_low = ?,
		    neckline_level = ?, neckline_slope = ?,
		    neckline_touch1 = ?, neckline_touch2 = ?,
		    pattern_width = ?, pattern_height = ?, symmetry = ?,
		    thesis_components = ?, last_updated = ?,
		    is_complete = ?, current_phase = ?
		WHERE id = ?
	`

	pattern.LastUpdated = time.Now()

	_, err = db.conn.Exec(query,
		string(rightShoulderHigh), string(rightShoulderLow),
		pattern.NecklineLevel, pattern.NecklineSlope,
		string(necklineTouch1), string(necklineTouch2),
		pattern.PatternWidth, pattern.PatternHeight, pattern.Symmetry,
		string(thesisData), pattern.LastUpdated,
		pattern.IsComplete, pattern.CurrentPhase, pattern.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update head and shoulders pattern: %w", err)
	}

	// Update thesis components
	err = db.updateThesisComponents(pattern.ID, &pattern.ThesisComponents)
	if err != nil {
		return fmt.Errorf("failed to update thesis components: %w", err)
	}

	return nil
}

// GetHeadShouldersPatterns retrieves patterns based on filter criteria
func (db *DB) GetHeadShouldersPatterns(filter *models.PatternFilter) ([]*models.HeadShouldersPattern, error) {
	query := `
		SELECT id, setup_id, symbol, pattern_type,
		       left_shoulder_high, left_shoulder_low, head_high, head_low,
		       right_shoulder_high, right_shoulder_low,
		       neckline_level, neckline_slope, neckline_touch1, neckline_touch2,
		       pattern_width, pattern_height, symmetry, thesis_components,
		       detected_at, last_updated, is_complete, current_phase
		FROM head_shoulders_patterns 
		WHERE 1=1
	`
	args := []interface{}{}

	// Add optional filters
	if filter.Symbol != "" {
		query += " AND symbol = ?"
		args = append(args, filter.Symbol)
	}

	if filter.PatternType != "" {
		query += " AND pattern_type = ?"
		args = append(args, filter.PatternType)
	}

	if filter.Phase != "" {
		query += " AND current_phase = ?"
		args = append(args, filter.Phase)
	}

	if filter.IsComplete != nil {
		query += " AND is_complete = ?"
		args = append(args, *filter.IsComplete)
	}

	if filter.MinSymmetry > 0 {
		query += " AND symmetry >= ?"
		args = append(args, filter.MinSymmetry)
	}

	// Order by detected date descending
	query += " ORDER BY detected_at DESC"

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
		return nil, fmt.Errorf("failed to query head and shoulders patterns: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice to ensure we never return nil
	patterns := make([]*models.HeadShouldersPattern, 0)

	for rows.Next() {
		pattern := &models.HeadShouldersPattern{}
		var leftShoulderHighJSON, leftShoulderLowJSON, headHighJSON, headLowJSON string
		var rightShoulderHighJSON, rightShoulderLowJSON string
		var necklineTouch1JSON, necklineTouch2JSON string
		var thesisJSON string

		err := rows.Scan(
			&pattern.ID, &pattern.SetupID, &pattern.Symbol, &pattern.PatternType,
			&leftShoulderHighJSON, &leftShoulderLowJSON, &headHighJSON, &headLowJSON,
			&rightShoulderHighJSON, &rightShoulderLowJSON,
			&pattern.NecklineLevel, &pattern.NecklineSlope,
			&necklineTouch1JSON, &necklineTouch2JSON,
			&pattern.PatternWidth, &pattern.PatternHeight, &pattern.Symmetry,
			&thesisJSON, &pattern.DetectedAt, &pattern.LastUpdated,
			&pattern.IsComplete, &pattern.CurrentPhase,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan head and shoulders pattern: %w", err)
		}

		// Unmarshal pattern points
		pattern.LeftShoulderHigh, _ = models.UnmarshalPatternPointsJSON([]byte(leftShoulderHighJSON))
		pattern.LeftShoulderLow, _ = models.UnmarshalPatternPointsJSON([]byte(leftShoulderLowJSON))
		pattern.HeadHigh, _ = models.UnmarshalPatternPointsJSON([]byte(headHighJSON))
		pattern.HeadLow, _ = models.UnmarshalPatternPointsJSON([]byte(headLowJSON))
		pattern.RightShoulderHigh, _ = models.UnmarshalPatternPointsJSON([]byte(rightShoulderHighJSON))
		pattern.RightShoulderLow, _ = models.UnmarshalPatternPointsJSON([]byte(rightShoulderLowJSON))
		pattern.NecklineTouch1, _ = models.UnmarshalPatternPointsJSON([]byte(necklineTouch1JSON))
		pattern.NecklineTouch2, _ = models.UnmarshalPatternPointsJSON([]byte(necklineTouch2JSON))

		// Unmarshal thesis components
		pattern.ThesisComponents, _ = models.UnmarshalThesisJSON([]byte(thesisJSON))

		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// GetHeadShouldersPatternByID retrieves a specific pattern by ID
func (db *DB) GetHeadShouldersPatternByID(id int64) (*models.HeadShouldersPattern, error) {
	patterns, err := db.GetHeadShouldersPatterns(&models.PatternFilter{Limit: 1})
	if err != nil {
		return nil, err
	}

	for _, pattern := range patterns {
		if pattern.ID == id {
			return pattern, nil
		}
	}

	return nil, sql.ErrNoRows
}

// GetActiveHeadShouldersPatterns retrieves all active (incomplete) patterns
func (db *DB) GetActiveHeadShouldersPatterns() ([]*models.HeadShouldersPattern, error) {
	incomplete := false
	return db.GetHeadShouldersPatterns(&models.PatternFilter{
		IsComplete: &incomplete,
	})
}

// insertThesisComponents inserts thesis components for a pattern
func (db *DB) insertThesisComponents(patternID int64, thesis *models.HeadShouldersThesis) error {
	components := thesis.GetAllComponents()

	for i, component := range components {
		if component == nil {
			continue
		}

		evidenceJSON, _ := json.Marshal(component.Evidence)
		componentType := db.getComponentType(i)

		query := `
			INSERT OR REPLACE INTO thesis_components 
			(pattern_id, component_name, component_type, description,
			 is_completed, completed_at, is_required, weight, confidence_level,
			 evidence, last_checked, auto_detected, notification_sent)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err := db.conn.Exec(query,
			patternID, component.Name, componentType, component.Description,
			component.IsCompleted, component.CompletedAt, component.IsRequired,
			component.Weight, component.ConfidenceLevel, string(evidenceJSON),
			component.LastChecked, component.AutoDetected, component.NotificationSent,
		)

		if err != nil {
			return fmt.Errorf("failed to insert thesis component %s: %w", component.Name, err)
		}
	}

	return nil
}

// updateThesisComponents updates thesis components for a pattern
func (db *DB) updateThesisComponents(patternID int64, thesis *models.HeadShouldersThesis) error {
	return db.insertThesisComponents(patternID, thesis) // Uses INSERT OR REPLACE
}

// getComponentType determines the component type based on index
func (db *DB) getComponentType(index int) string {
	switch {
	case index < 9: // Formation components (0-8)
		return "formation"
	case index < 12: // Breakout components (9-11)
		return "breakout"
	default: // Target components (12+)
		return "target"
	}
}

// InsertPatternAlert inserts a new pattern alert
func (db *DB) InsertPatternAlert(alert *models.PatternAlert) error {
	query := `
		INSERT INTO pattern_alerts 
		(pattern_id, symbol, component_name, alert_type, message, triggered_at, notification_sent, email_sent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		alert.PatternID, alert.Symbol, alert.ComponentName, alert.AlertType,
		alert.Message, alert.TriggeredAt, alert.NotificationSent, alert.EmailSent,
	)

	if err != nil {
		return fmt.Errorf("failed to insert pattern alert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get alert ID: %w", err)
	}

	alert.ID = id
	return nil
}

// GetPatternAlerts retrieves alerts for a pattern
func (db *DB) GetPatternAlerts(patternID int64) ([]*models.PatternAlert, error) {
	query := `
		SELECT id, pattern_id, symbol, component_name, alert_type, message,
		       triggered_at, notification_sent, email_sent
		FROM pattern_alerts 
		WHERE pattern_id = ?
		ORDER BY triggered_at DESC
	`

	rows, err := db.conn.Query(query, patternID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pattern alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.PatternAlert
	for rows.Next() {
		alert := &models.PatternAlert{}
		err := rows.Scan(
			&alert.ID, &alert.PatternID, &alert.Symbol, &alert.ComponentName,
			&alert.AlertType, &alert.Message, &alert.TriggeredAt,
			&alert.NotificationSent, &alert.EmailSent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pattern alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// UpdatePatternAlertNotificationStatus updates the notification status of an alert
func (db *DB) UpdatePatternAlertNotificationStatus(alertID int64, notificationSent, emailSent bool) error {
	query := `
		UPDATE pattern_alerts 
		SET notification_sent = ?, email_sent = ?
		WHERE id = ?
	`

	_, err := db.conn.Exec(query, notificationSent, emailSent, alertID)
	if err != nil {
		return fmt.Errorf("failed to update alert notification status: %w", err)
	}

	return nil
}

// GetPendingPatternAlerts retrieves alerts that haven't been processed
func (db *DB) GetPendingPatternAlerts() ([]*models.PatternAlert, error) {
	query := `
		SELECT id, pattern_id, symbol, component_name, alert_type, message,
		       triggered_at, notification_sent, email_sent
		FROM pattern_alerts 
		WHERE notification_sent = FALSE OR email_sent = FALSE
		ORDER BY triggered_at ASC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending pattern alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.PatternAlert
	for rows.Next() {
		alert := &models.PatternAlert{}
		err := rows.Scan(
			&alert.ID, &alert.PatternID, &alert.Symbol, &alert.ComponentName,
			&alert.AlertType, &alert.Message, &alert.TriggeredAt,
			&alert.NotificationSent, &alert.EmailSent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending pattern alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// CleanupOldPatternData removes old pattern data based on retention policy
func (db *DB) CleanupOldPatternData(days int) (int64, error) {
	var totalDeleted int64

	// Cleanup old thesis components
	thesisResult, err := db.conn.Exec(
		`DELETE FROM thesis_components WHERE pattern_id IN 
		 (SELECT id FROM head_shoulders_patterns WHERE detected_at < datetime('now', '-' || ? || ' days'))`,
		days,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old thesis components: %w", err)
	}

	thesisDeleted, _ := thesisResult.RowsAffected()
	totalDeleted += thesisDeleted

	// Cleanup old alerts
	alertResult, err := db.conn.Exec(
		`DELETE FROM pattern_alerts WHERE pattern_id IN 
		 (SELECT id FROM head_shoulders_patterns WHERE detected_at < datetime('now', '-' || ? || ' days'))`,
		days,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old pattern alerts: %w", err)
	}

	alertsDeleted, _ := alertResult.RowsAffected()
	totalDeleted += alertsDeleted

	// Cleanup old patterns
	patternResult, err := db.conn.Exec(
		"DELETE FROM head_shoulders_patterns WHERE detected_at < datetime('now', '-' || ? || ' days')",
		days,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old head and shoulders patterns: %w", err)
	}

	patternsDeleted, _ := patternResult.RowsAffected()
	totalDeleted += patternsDeleted

	return totalDeleted, nil
}
