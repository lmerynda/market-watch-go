package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"market-watch-go/internal/models"
)

// CreateFallingWedgePatternTables creates all falling wedge pattern related tables
func (db *DB) CreateFallingWedgePatternTables() error {
	tables := []string{
		// Falling wedge patterns table
		`CREATE TABLE IF NOT EXISTS falling_wedge_patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setup_id INTEGER NOT NULL,
			symbol TEXT NOT NULL,
			pattern_type TEXT NOT NULL DEFAULT 'falling_wedge',
			
			-- Trend line points (JSON stored as TEXT)
			upper_trend_line_1 TEXT NOT NULL,
			upper_trend_line_2 TEXT NOT NULL,
			lower_trend_line_1 TEXT NOT NULL,
			lower_trend_line_2 TEXT NOT NULL,
			
			-- Pattern metrics
			upper_slope REAL NOT NULL,
			lower_slope REAL NOT NULL,
			breakout_level REAL NOT NULL,
			pattern_width INTEGER NOT NULL,
			pattern_height REAL NOT NULL,
			convergence REAL NOT NULL,
			
			-- Volume analysis
			volume_profile TEXT DEFAULT 'unknown',
			
			-- Thesis components (JSON stored as TEXT)
			thesis_components TEXT NOT NULL DEFAULT '{}',
			
			-- Status
			detected_at DATETIME NOT NULL,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_complete BOOLEAN DEFAULT FALSE,
			current_phase TEXT DEFAULT 'formation' CHECK (current_phase IN ('formation', 'breakout', 'target_pursuit', 'completed')),
			
			FOREIGN KEY (setup_id) REFERENCES trading_setups(id)
		)`,
	}

	// Create tables
	for _, query := range tables {
		_, err := db.conn.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create falling wedge pattern table: %w", err)
		}
	}

	// Create indexes for better performance
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_falling_wedge_symbol ON falling_wedge_patterns(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_falling_wedge_phase ON falling_wedge_patterns(current_phase)`,
		`CREATE INDEX IF NOT EXISTS idx_falling_wedge_complete ON falling_wedge_patterns(is_complete)`,
		`CREATE INDEX IF NOT EXISTS idx_falling_wedge_detected ON falling_wedge_patterns(detected_at)`,
		`CREATE INDEX IF NOT EXISTS idx_falling_wedge_symbol_phase ON falling_wedge_patterns(symbol, current_phase)`,
		`CREATE INDEX IF NOT EXISTS idx_falling_wedge_symbol_complete ON falling_wedge_patterns(symbol, is_complete)`,
	}

	for _, indexQuery := range indexes {
		_, err := db.conn.Exec(indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create falling wedge pattern index: %w", err)
		}
	}

	return nil
}

// InsertFallingWedgePattern inserts a new falling wedge pattern
func (db *DB) InsertFallingWedgePattern(pattern *models.FallingWedgePattern) error {
	// Marshal pattern points to JSON
	upperLine1JSON, err := json.Marshal(pattern.UpperTrendLine1)
	if err != nil {
		return fmt.Errorf("failed to marshal upper trend line 1: %w", err)
	}

	upperLine2JSON, err := json.Marshal(pattern.UpperTrendLine2)
	if err != nil {
		return fmt.Errorf("failed to marshal upper trend line 2: %w", err)
	}

	lowerLine1JSON, err := json.Marshal(pattern.LowerTrendLine1)
	if err != nil {
		return fmt.Errorf("failed to marshal lower trend line 1: %w", err)
	}

	lowerLine2JSON, err := json.Marshal(pattern.LowerTrendLine2)
	if err != nil {
		return fmt.Errorf("failed to marshal lower trend line 2: %w", err)
	}

	// Marshal thesis components to JSON
	thesisJSON, err := json.Marshal(pattern.ThesisComponents)
	if err != nil {
		return fmt.Errorf("failed to marshal thesis components: %w", err)
	}

	query := `
		INSERT INTO falling_wedge_patterns (
			setup_id, symbol, pattern_type,
			upper_trend_line_1, upper_trend_line_2, lower_trend_line_1, lower_trend_line_2,
			upper_slope, lower_slope, breakout_level, pattern_width, pattern_height, convergence,
			volume_profile, thesis_components,
			detected_at, last_updated, is_complete, current_phase
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		pattern.SetupID, pattern.Symbol, pattern.PatternType,
		string(upperLine1JSON), string(upperLine2JSON), string(lowerLine1JSON), string(lowerLine2JSON),
		pattern.UpperSlope, pattern.LowerSlope, pattern.BreakoutLevel,
		pattern.PatternWidth, pattern.PatternHeight, pattern.Convergence,
		pattern.VolumeProfile, string(thesisJSON),
		pattern.DetectedAt, pattern.LastUpdated, pattern.IsComplete, pattern.CurrentPhase,
	)

	if err != nil {
		return fmt.Errorf("failed to insert falling wedge pattern: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	pattern.ID = id
	return nil
}

// GetFallingWedgePatterns retrieves falling wedge patterns with optional filtering
func (db *DB) GetFallingWedgePatterns(filter *models.FallingWedgeFilter) ([]*models.FallingWedgePattern, error) {
	query := `
		SELECT id, setup_id, symbol, pattern_type,
		       upper_trend_line_1, upper_trend_line_2, lower_trend_line_1, lower_trend_line_2,
		       upper_slope, lower_slope, breakout_level, pattern_width, pattern_height, convergence,
		       volume_profile, thesis_components,
		       detected_at, last_updated, is_complete, current_phase
		FROM falling_wedge_patterns 
		WHERE 1=1
	`
	args := []interface{}{}

	// Add optional filters
	if filter != nil {
		if filter.Symbol != "" {
			query += " AND symbol = ?"
			args = append(args, filter.Symbol)
		}

		if filter.Phase != "" {
			query += " AND current_phase = ?"
			args = append(args, filter.Phase)
		}

		if filter.IsComplete != nil {
			query += " AND is_complete = ?"
			args = append(args, *filter.IsComplete)
		}

		if filter.MinConvergence > 0 {
			query += " AND convergence >= ?"
			args = append(args, filter.MinConvergence)
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
	} else {
		query += " ORDER BY detected_at DESC LIMIT 100"
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query falling wedge patterns: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice to ensure we never return nil
	patterns := make([]*models.FallingWedgePattern, 0)

	for rows.Next() {
		pattern := &models.FallingWedgePattern{}
		var upperLine1JSON, upperLine2JSON, lowerLine1JSON, lowerLine2JSON string
		var thesisJSON string

		err := rows.Scan(
			&pattern.ID, &pattern.SetupID, &pattern.Symbol, &pattern.PatternType,
			&upperLine1JSON, &upperLine2JSON, &lowerLine1JSON, &lowerLine2JSON,
			&pattern.UpperSlope, &pattern.LowerSlope, &pattern.BreakoutLevel,
			&pattern.PatternWidth, &pattern.PatternHeight, &pattern.Convergence,
			&pattern.VolumeProfile, &thesisJSON,
			&pattern.DetectedAt, &pattern.LastUpdated, &pattern.IsComplete, &pattern.CurrentPhase,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan falling wedge pattern: %w", err)
		}

		// Unmarshal pattern points
		if err := json.Unmarshal([]byte(upperLine1JSON), &pattern.UpperTrendLine1); err != nil {
			return nil, fmt.Errorf("failed to unmarshal upper trend line 1: %w", err)
		}
		if err := json.Unmarshal([]byte(upperLine2JSON), &pattern.UpperTrendLine2); err != nil {
			return nil, fmt.Errorf("failed to unmarshal upper trend line 2: %w", err)
		}
		if err := json.Unmarshal([]byte(lowerLine1JSON), &pattern.LowerTrendLine1); err != nil {
			return nil, fmt.Errorf("failed to unmarshal lower trend line 1: %w", err)
		}
		if err := json.Unmarshal([]byte(lowerLine2JSON), &pattern.LowerTrendLine2); err != nil {
			return nil, fmt.Errorf("failed to unmarshal lower trend line 2: %w", err)
		}

		// Parse thesis JSON
		if thesisJSON != "" {
			if err := json.Unmarshal([]byte(thesisJSON), &pattern.ThesisComponents); err != nil {
				// Log the error but continue - don't fail the entire query
				fmt.Printf("Failed to unmarshal thesis components for pattern %d: %v\n", pattern.ID, err)
			}
		}

		patterns = append(patterns, pattern)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating falling wedge patterns: %w", err)
	}

	return patterns, nil
}

// GetFallingWedgePatternByID retrieves a specific falling wedge pattern by ID
func (db *DB) GetFallingWedgePatternByID(id int64) (*models.FallingWedgePattern, error) {
	query := `
		SELECT id, setup_id, symbol, pattern_type,
		       upper_trend_line_1, upper_trend_line_2, lower_trend_line_1, lower_trend_line_2,
		       upper_slope, lower_slope, breakout_level, pattern_width, pattern_height, convergence,
		       volume_profile, thesis_components,
		       detected_at, last_updated, is_complete, current_phase
		FROM falling_wedge_patterns 
		WHERE id = ?
	`

	row := db.conn.QueryRow(query, id)

	pattern := &models.FallingWedgePattern{}
	var upperLine1JSON, upperLine2JSON, lowerLine1JSON, lowerLine2JSON string
	var thesisJSON string

	err := row.Scan(
		&pattern.ID, &pattern.SetupID, &pattern.Symbol, &pattern.PatternType,
		&upperLine1JSON, &upperLine2JSON, &lowerLine1JSON, &lowerLine2JSON,
		&pattern.UpperSlope, &pattern.LowerSlope, &pattern.BreakoutLevel,
		&pattern.PatternWidth, &pattern.PatternHeight, &pattern.Convergence,
		&pattern.VolumeProfile, &thesisJSON,
		&pattern.DetectedAt, &pattern.LastUpdated, &pattern.IsComplete, &pattern.CurrentPhase,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get falling wedge pattern: %w", err)
	}

	// Unmarshal pattern points
	if err := json.Unmarshal([]byte(upperLine1JSON), &pattern.UpperTrendLine1); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upper trend line 1: %w", err)
	}
	if err := json.Unmarshal([]byte(upperLine2JSON), &pattern.UpperTrendLine2); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upper trend line 2: %w", err)
	}
	if err := json.Unmarshal([]byte(lowerLine1JSON), &pattern.LowerTrendLine1); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lower trend line 1: %w", err)
	}
	if err := json.Unmarshal([]byte(lowerLine2JSON), &pattern.LowerTrendLine2); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lower trend line 2: %w", err)
	}

	// Parse thesis JSON
	if thesisJSON != "" {
		if err := json.Unmarshal([]byte(thesisJSON), &pattern.ThesisComponents); err != nil {
			fmt.Printf("Failed to unmarshal thesis components for pattern %d: %v\n", pattern.ID, err)
		}
	}

	return pattern, nil
}

// UpdateFallingWedgePattern updates an existing falling wedge pattern
func (db *DB) UpdateFallingWedgePattern(pattern *models.FallingWedgePattern) error {
	// Marshal thesis components to JSON
	thesisJSON, err := json.Marshal(pattern.ThesisComponents)
	if err != nil {
		return fmt.Errorf("failed to marshal thesis components: %w", err)
	}

	query := `
		UPDATE falling_wedge_patterns SET
			upper_slope = ?, lower_slope = ?, breakout_level = ?,
			pattern_width = ?, pattern_height = ?, convergence = ?,
			volume_profile = ?, thesis_components = ?,
			last_updated = ?, is_complete = ?, current_phase = ?
		WHERE id = ?
	`

	pattern.LastUpdated = time.Now()

	_, err = db.conn.Exec(query,
		pattern.UpperSlope, pattern.LowerSlope, pattern.BreakoutLevel,
		pattern.PatternWidth, pattern.PatternHeight, pattern.Convergence,
		pattern.VolumeProfile, string(thesisJSON),
		pattern.LastUpdated, pattern.IsComplete, pattern.CurrentPhase,
		pattern.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update falling wedge pattern: %w", err)
	}

	return nil
}

// GetActiveFallingWedgePatterns retrieves all active falling wedge patterns
func (db *DB) GetActiveFallingWedgePatterns() ([]*models.FallingWedgePattern, error) {
	isComplete := false
	filter := &models.FallingWedgeFilter{
		IsComplete: &isComplete,
		Limit:      1000,
	}
	return db.GetFallingWedgePatterns(filter)
}

// GetFallingWedgePatternsBySymbol retrieves all falling wedge patterns for a specific symbol
func (db *DB) GetFallingWedgePatternsBySymbol(symbol string) ([]*models.FallingWedgePattern, error) {
	filter := &models.FallingWedgeFilter{
		Symbol: symbol,
		Limit:  100,
	}
	return db.GetFallingWedgePatterns(filter)
}
