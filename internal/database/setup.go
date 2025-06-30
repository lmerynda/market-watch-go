package database

import (
	"fmt"

	"market-watch-go/internal/models"
)

// CreateSetupTables creates all setup related tables
func (db *Database) CreateSetupTables() error {
	tables := []string{
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

	// Create tables
	for _, query := range tables {
		_, err := db.conn.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create setup table: %w", err)
		}
	}

	// Create indexes for better performance
	indexes := []string{
		// Trading Setups indexes
		`CREATE INDEX IF NOT EXISTS idx_setups_symbol ON trading_setups(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_type ON trading_setups(setup_type)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_direction ON trading_setups(direction)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_status ON trading_setups(status)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_quality ON trading_setups(quality_score)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_detected ON trading_setups(detected_at)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_expires ON trading_setups(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_symbol_status ON trading_setups(symbol, status)`,
		`CREATE INDEX IF NOT EXISTS idx_setups_symbol_quality ON trading_setups(symbol, quality_score)`,

		// Setup Checklists indexes
		`CREATE INDEX IF NOT EXISTS idx_checklists_setup_id ON setup_checklists(setup_id)`,
		`CREATE INDEX IF NOT EXISTS idx_checklists_score ON setup_checklists(total_score)`,
		`CREATE INDEX IF NOT EXISTS idx_checklists_completion ON setup_checklists(completion_percent)`,

		// Setup Alerts indexes
		`CREATE INDEX IF NOT EXISTS idx_setup_alerts_setup_id ON setup_alerts(setup_id)`,
		`CREATE INDEX IF NOT EXISTS idx_setup_alerts_symbol ON setup_alerts(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_setup_alerts_type ON setup_alerts(alert_type)`,
		`CREATE INDEX IF NOT EXISTS idx_setup_alerts_active ON setup_alerts(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_setup_alerts_triggered ON setup_alerts(triggered_at)`,
	}

	for _, indexQuery := range indexes {
		_, err := db.conn.Exec(indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create setup index: %w", err)
		}
	}

	return nil
}

// GetSetupSummary calculates summary statistics for setups
func (db *Database) GetSetupSummary(symbol string) (*models.SetupSummary, error) {
	summary := &models.SetupSummary{}

	// Get basic counts and averages
	query := `
		SELECT 
			COUNT(*) as total_setups,
			COUNT(CASE WHEN status = 'active' AND expires_at > datetime('now') THEN 1 END) as active_count,
			COUNT(CASE WHEN quality_score >= 80 THEN 1 END) as high_quality_count,
			COUNT(CASE WHEN quality_score >= 60 AND quality_score < 80 THEN 1 END) as medium_quality_count,
			COUNT(CASE WHEN quality_score < 60 THEN 1 END) as low_quality_count,
			COUNT(CASE WHEN direction = 'bullish' THEN 1 END) as bullish_count,
			COUNT(CASE WHEN direction = 'bearish' THEN 1 END) as bearish_count,
			AVG(quality_score) as avg_quality_score,
			AVG(risk_reward_ratio) as avg_risk_reward,
			MAX(detected_at) as last_detection
		FROM trading_setups 
		WHERE symbol = ?
	`

	err := db.conn.QueryRow(query, symbol).Scan(
		&summary.TotalSetups, &summary.ActiveCount, &summary.HighQualityCount,
		&summary.MediumQualityCount, &summary.LowQualityCount, &summary.BullishCount,
		&summary.BearishCount, &summary.AvgQualityScore, &summary.AvgRiskReward,
		&summary.LastDetection,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get setup summary: %w", err)
	}

	return summary, nil
}

// ExpireOldSetups marks old setups as expired
func (db *Database) ExpireOldSetups() (int64, error) {
	query := `
		UPDATE trading_setups 
		SET status = 'expired', updated_at = CURRENT_TIMESTAMP
		WHERE status = 'active' AND expires_at <= datetime('now')
	`

	result, err := db.conn.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to expire old setups: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// CleanupOldSetupData removes old setup data based on retention policy
func (db *Database) CleanupOldSetupData(days int) (int64, error) {
	var totalDeleted int64

	// Cleanup old checklists
	checklistResult, err := db.conn.Exec(
		"DELETE FROM setup_checklists WHERE setup_id IN (SELECT id FROM trading_setups WHERE created_at < datetime('now', '-' || ? || ' days'))",
		days,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old setup checklists: %w", err)
	}

	checklistsDeleted, _ := checklistResult.RowsAffected()
	totalDeleted += checklistsDeleted

	// Cleanup old alerts
	alertResult, err := db.conn.Exec(
		"DELETE FROM setup_alerts WHERE setup_id IN (SELECT id FROM trading_setups WHERE created_at < datetime('now', '-' || ? || ' days'))",
		days,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old setup alerts: %w", err)
	}

	alertsDeleted, _ := alertResult.RowsAffected()
	totalDeleted += alertsDeleted

	// Cleanup old setups
	setupResult, err := db.conn.Exec(
		"DELETE FROM trading_setups WHERE created_at < datetime('now', '-' || ? || ' days')",
		days,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old trading setups: %w", err)
	}

	setupsDeleted, _ := setupResult.RowsAffected()
	totalDeleted += setupsDeleted

	return totalDeleted, nil
}

// InsertTradingSetup inserts a new trading setup into the database
func (db *Database) InsertTradingSetup(setup *models.TradingSetup) error {
	query := `
		INSERT INTO trading_setups (
			symbol, setup_type, direction, quality_score, confidence, status,
			detected_at, expires_at, current_price, entry_price, stop_loss,
			target1, target2, target3, risk_amount, reward_potential, risk_reward_ratio,
			price_action_score, volume_score, technical_score, risk_reward_score, notes, is_manual
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(
		query,
		setup.Symbol, setup.SetupType, setup.Direction, setup.QualityScore, setup.Confidence, setup.Status,
		setup.DetectedAt, setup.ExpiresAt, setup.CurrentPrice, setup.EntryPrice, setup.StopLoss,
		setup.Target1, setup.Target2, setup.Target3, setup.RiskAmount, setup.RewardPotential, setup.RiskRewardRatio,
		setup.PriceActionScore, setup.VolumeScore, setup.TechnicalScore, setup.RiskRewardScore, setup.Notes, setup.IsManual,
	)
	if err != nil {
		return fmt.Errorf("failed to insert trading setup: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	setup.ID = id // ID is int64 in the TradingSetup model
	return nil
}

// GetTradingSetups retrieves trading setups based on filter criteria
func (db *Database) GetTradingSetups(filter *models.SetupFilter) ([]*models.TradingSetup, error) {
	var setups []*models.TradingSetup
	var args []interface{}

	// Build the query based on filter
	query := `
		SELECT 
			id, symbol, setup_type, direction, quality_score, confidence, status,
			detected_at, expires_at, last_updated, current_price, entry_price, stop_loss,
			target1, target2, target3, risk_amount, reward_potential, risk_reward_ratio,
			price_action_score, volume_score, technical_score, risk_reward_score,
			notes, is_manual, created_at, updated_at
		FROM trading_setups
		WHERE 1=1
	`

	// Apply filters
	if filter.Symbol != "" {
		query += " AND symbol = ?"
		args = append(args, filter.Symbol)
	}

	if filter.SetupType != "" {
		query += " AND setup_type = ?"
		args = append(args, filter.SetupType)
	}

	if filter.Direction != "" {
		query += " AND direction = ?"
		args = append(args, filter.Direction)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.Confidence != "" {
		query += " AND confidence = ?"
		args = append(args, filter.Confidence)
	}

	if filter.MinQualityScore > 0 {
		query += " AND quality_score >= ?"
		args = append(args, filter.MinQualityScore)
	}

	if filter.IsActive != nil && *filter.IsActive {
		query += " AND status = 'active' AND expires_at > datetime('now')"
	}

	// Order by quality score descending
	query += " ORDER BY quality_score DESC"

	// Apply limit
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	// Execute query
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query trading setups: %w", err)
	}
	defer rows.Close()

	// Scan results
	for rows.Next() {
		setup := &models.TradingSetup{}
		err := rows.Scan(
			&setup.ID, &setup.Symbol, &setup.SetupType, &setup.Direction, &setup.QualityScore,
			&setup.Confidence, &setup.Status, &setup.DetectedAt, &setup.ExpiresAt, &setup.LastUpdated,
			&setup.CurrentPrice, &setup.EntryPrice, &setup.StopLoss, &setup.Target1, &setup.Target2,
			&setup.Target3, &setup.RiskAmount, &setup.RewardPotential, &setup.RiskRewardRatio,
			&setup.PriceActionScore, &setup.VolumeScore, &setup.TechnicalScore, &setup.RiskRewardScore,
			&setup.Notes, &setup.IsManual, &setup.CreatedAt, &setup.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trading setup: %w", err)
		}
		setups = append(setups, setup)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trading setups: %w", err)
	}

	return setups, nil
}

// UpdateTradingSetup updates an existing trading setup
func (db *Database) UpdateTradingSetup(setup *models.TradingSetup) error {
	query := `
		UPDATE trading_setups SET
			symbol = ?,
			setup_type = ?,
			direction = ?,
			quality_score = ?,
			confidence = ?,
			status = ?,
			detected_at = ?,
			expires_at = ?,
			current_price = ?,
			entry_price = ?,
			stop_loss = ?,
			target1 = ?,
			target2 = ?,
			target3 = ?,
			risk_amount = ?,
			reward_potential = ?,
			risk_reward_ratio = ?,
			price_action_score = ?,
			volume_score = ?,
			technical_score = ?,
			risk_reward_score = ?,
			notes = ?,
			is_manual = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.conn.Exec(
		query,
		setup.Symbol, setup.SetupType, setup.Direction, setup.QualityScore, setup.Confidence, setup.Status,
		setup.DetectedAt, setup.ExpiresAt, setup.CurrentPrice, setup.EntryPrice, setup.StopLoss,
		setup.Target1, setup.Target2, setup.Target3, setup.RiskAmount, setup.RewardPotential, setup.RiskRewardRatio,
		setup.PriceActionScore, setup.VolumeScore, setup.TechnicalScore, setup.RiskRewardScore, setup.Notes, setup.IsManual,
		setup.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update trading setup: %w", err)
	}

	return nil
}

// GetSetupChecklist retrieves the checklist for a trading setup
func (db *Database) GetSetupChecklist(setupID int64) (*models.SetupChecklist, error) {
	query := `
		SELECT 
			setup_id,
			
			-- Price Action Criteria
			min_level_touches_completed, min_level_touches_points,
			bounce_strength_completed, bounce_strength_points,
			time_at_level_completed, time_at_level_points,
			rejection_candle_completed, rejection_candle_points,
			level_duration_completed, level_duration_points,
			
			-- Volume Criteria
			volume_spike_completed, volume_spike_points,
			volume_confirmation_completed, volume_confirmation_points,
			approach_volume_completed, approach_volume_points,
			vwap_relationship_completed, vwap_relationship_points,
			relative_volume_completed, relative_volume_points,
			
			-- Technical Indicators
			rsi_condition_completed, rsi_condition_points,
			moving_average_completed, moving_average_points,
			macd_signal_completed, macd_signal_points,
			momentum_divergence_completed, momentum_divergence_points,
			bollinger_bands_completed, bollinger_bands_points,
			
			-- Risk Management
			stop_loss_defined_completed, stop_loss_defined_points,
			risk_reward_ratio_completed, risk_reward_ratio_points,
			position_size_completed, position_size_points,
			entry_precision_completed, entry_precision_points,
			exit_strategy_completed, exit_strategy_points,
			
			-- Summary
			total_score, completed_items, total_items, completion_percent,
			last_updated
		FROM setup_checklists
		WHERE setup_id = ?
	`

	checklist := &models.SetupChecklist{}

	var minLevelTouchesCompleted, bounceStrengthCompleted, timeAtLevelCompleted, rejectionCandleCompleted, levelDurationCompleted bool
	var minLevelTouchesPoints, bounceStrengthPoints, timeAtLevelPoints, rejectionCandlePoints, levelDurationPoints float64

	var volumeSpikeCompleted, volumeConfirmationCompleted, approachVolumeCompleted, vwapRelationshipCompleted, relativeVolumeCompleted bool
	var volumeSpikePoints, volumeConfirmationPoints, approachVolumePoints, vwapRelationshipPoints, relativeVolumePoints float64

	var rsiConditionCompleted, movingAverageCompleted, macdSignalCompleted, momentumDivergenceCompleted, bollingerBandsCompleted bool
	var rsiConditionPoints, movingAveragePoints, macdSignalPoints, momentumDivergencePoints, bollingerBandsPoints float64

	var stopLossDefinedCompleted, riskRewardRatioCompleted, positionSizeCompleted, entryPrecisionCompleted, exitStrategyCompleted bool
	var stopLossDefinedPoints, riskRewardRatioPoints, positionSizePoints, entryPrecisionPoints, exitStrategyPoints float64

	err := db.conn.QueryRow(query, setupID).Scan(
		&checklist.SetupID,

		&minLevelTouchesCompleted, &minLevelTouchesPoints,
		&bounceStrengthCompleted, &bounceStrengthPoints,
		&timeAtLevelCompleted, &timeAtLevelPoints,
		&rejectionCandleCompleted, &rejectionCandlePoints,
		&levelDurationCompleted, &levelDurationPoints,

		&volumeSpikeCompleted, &volumeSpikePoints,
		&volumeConfirmationCompleted, &volumeConfirmationPoints,
		&approachVolumeCompleted, &approachVolumePoints,
		&vwapRelationshipCompleted, &vwapRelationshipPoints,
		&relativeVolumeCompleted, &relativeVolumePoints,

		&rsiConditionCompleted, &rsiConditionPoints,
		&movingAverageCompleted, &movingAveragePoints,
		&macdSignalCompleted, &macdSignalPoints,
		&momentumDivergenceCompleted, &momentumDivergencePoints,
		&bollingerBandsCompleted, &bollingerBandsPoints,

		&stopLossDefinedCompleted, &stopLossDefinedPoints,
		&riskRewardRatioCompleted, &riskRewardRatioPoints,
		&positionSizeCompleted, &positionSizePoints,
		&entryPrecisionCompleted, &entryPrecisionPoints,
		&exitStrategyCompleted, &exitStrategyPoints,

		&checklist.TotalScore, &checklist.CompletedItems, &checklist.TotalItems, &checklist.CompletionPercent,
		&checklist.LastUpdated,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get setup checklist: %w", err)
	}

	// Set up the ChecklistItem structs
	checklist.MinLevelTouches = models.ChecklistItem{
		IsCompleted: minLevelTouchesCompleted,
		Points:      minLevelTouchesPoints,
		MaxPoints:   5.0,
	}

	checklist.BounceStrength = models.ChecklistItem{
		IsCompleted: bounceStrengthCompleted,
		Points:      bounceStrengthPoints,
		MaxPoints:   5.0,
	}

	checklist.TimeAtLevel = models.ChecklistItem{
		IsCompleted: timeAtLevelCompleted,
		Points:      timeAtLevelPoints,
		MaxPoints:   5.0,
	}

	checklist.RejectionCandle = models.ChecklistItem{
		IsCompleted: rejectionCandleCompleted,
		Points:      rejectionCandlePoints,
		MaxPoints:   5.0,
	}

	checklist.LevelDuration = models.ChecklistItem{
		IsCompleted: levelDurationCompleted,
		Points:      levelDurationPoints,
		MaxPoints:   5.0,
	}

	// Volume criteria
	checklist.VolumeSpike = models.ChecklistItem{
		IsCompleted: volumeSpikeCompleted,
		Points:      volumeSpikePoints,
		MaxPoints:   5.0,
	}

	checklist.VolumeConfirmation = models.ChecklistItem{
		IsCompleted: volumeConfirmationCompleted,
		Points:      volumeConfirmationPoints,
		MaxPoints:   5.0,
	}

	checklist.ApproachVolume = models.ChecklistItem{
		IsCompleted: approachVolumeCompleted,
		Points:      approachVolumePoints,
		MaxPoints:   5.0,
	}

	checklist.VWAPRelationship = models.ChecklistItem{
		IsCompleted: vwapRelationshipCompleted,
		Points:      vwapRelationshipPoints,
		MaxPoints:   5.0,
	}

	checklist.RelativeVolume = models.ChecklistItem{
		IsCompleted: relativeVolumeCompleted,
		Points:      relativeVolumePoints,
		MaxPoints:   5.0,
	}

	// Technical indicators
	checklist.RSICondition = models.ChecklistItem{
		IsCompleted: rsiConditionCompleted,
		Points:      rsiConditionPoints,
		MaxPoints:   5.0,
	}

	checklist.MovingAverage = models.ChecklistItem{
		IsCompleted: movingAverageCompleted,
		Points:      movingAveragePoints,
		MaxPoints:   5.0,
	}

	checklist.MACDSignal = models.ChecklistItem{
		IsCompleted: macdSignalCompleted,
		Points:      macdSignalPoints,
		MaxPoints:   5.0,
	}

	checklist.MomentumDivergence = models.ChecklistItem{
		IsCompleted: momentumDivergenceCompleted,
		Points:      momentumDivergencePoints,
		MaxPoints:   5.0,
	}

	checklist.BollingerBands = models.ChecklistItem{
		IsCompleted: bollingerBandsCompleted,
		Points:      bollingerBandsPoints,
		MaxPoints:   5.0,
	}

	// Risk management
	checklist.StopLossDefined = models.ChecklistItem{
		IsCompleted: stopLossDefinedCompleted,
		Points:      stopLossDefinedPoints,
		MaxPoints:   5.0,
	}

	checklist.RiskRewardRatio = models.ChecklistItem{
		IsCompleted: riskRewardRatioCompleted,
		Points:      riskRewardRatioPoints,
		MaxPoints:   5.0,
	}

	checklist.PositionSize = models.ChecklistItem{
		IsCompleted: positionSizeCompleted,
		Points:      positionSizePoints,
		MaxPoints:   5.0,
	}

	checklist.EntryPrecision = models.ChecklistItem{
		IsCompleted: entryPrecisionCompleted,
		Points:      entryPrecisionPoints,
		MaxPoints:   5.0,
	}

	checklist.ExitStrategy = models.ChecklistItem{
		IsCompleted: exitStrategyCompleted,
		Points:      exitStrategyPoints,
		MaxPoints:   5.0,
	}

	return checklist, nil
}
