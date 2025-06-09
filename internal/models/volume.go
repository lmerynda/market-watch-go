package models

import (
	"time"
)

// VolumeData represents trading volume data for a stock symbol
type VolumeData struct {
	ID        int64     `json:"id" db:"id"`
	Symbol    string    `json:"symbol" db:"symbol"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Volume    int64     `json:"volume" db:"volume"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// VolumeDataResponse represents the API response for volume data
type VolumeDataResponse struct {
	Symbol       string       `json:"symbol"`
	Data         []VolumeData `json:"data"`
	TotalRecords int          `json:"total_records"`
	From         time.Time    `json:"from"`
	To           time.Time    `json:"to"`
	Interval     string       `json:"interval"`
}

// VolumeStats represents volume statistics for a symbol
type VolumeStats struct {
	Symbol          string    `json:"symbol"`
	CurrentVolume   int64     `json:"current_volume"`
	AverageVolume   float64   `json:"average_volume"`
	VolumeRatio     float64   `json:"volume_ratio"`
	LastUpdate      time.Time `json:"last_update"`
	TotalDataPoints int       `json:"total_data_points"`
}

// DashboardSummary represents the summary data for the dashboard
type DashboardSummary struct {
	Symbols        []VolumeStats `json:"symbols"`
	LastUpdate     time.Time     `json:"last_update"`
	CollectionMode string        `json:"collection_mode"` // "24/7" to indicate continuous collection
}

// PolygonAggregateResponse represents the response from Polygon.io aggregates API
type PolygonAggregateResponse struct {
	Ticker       string             `json:"ticker"`
	QueryCount   int                `json:"queryCount"`
	ResultsCount int                `json:"resultsCount"`
	Adjusted     bool               `json:"adjusted"`
	Results      []PolygonAggregate `json:"results"`
	Status       string             `json:"status"`
	RequestID    string             `json:"request_id"`
	Count        int                `json:"count"`
}

// PolygonAggregate represents a single aggregate data point from Polygon.io
type PolygonAggregate struct {
	Open      float64 `json:"o"`  // Open price
	High      float64 `json:"h"`  // High price
	Low       float64 `json:"l"`  // Low price
	Close     float64 `json:"c"`  // Close price
	Volume    int64   `json:"v"`  // Volume
	VolumeWAP float64 `json:"vw"` // Volume weighted average price
	Timestamp int64   `json:"t"`  // Timestamp (Unix milliseconds)
	Count     int     `json:"n"`  // Number of transactions
}

// ToVolumeData converts PolygonAggregate to VolumeData
func (pa *PolygonAggregate) ToVolumeData(symbol string) *VolumeData {
	return &VolumeData{
		Symbol:    symbol,
		Timestamp: time.Unix(pa.Timestamp/1000, (pa.Timestamp%1000)*1000000), // Convert from milliseconds
		Volume:    pa.Volume,
		CreatedAt: time.Now(),
	}
}

// VolumeDataFilter represents filter parameters for querying volume data
type VolumeDataFilter struct {
	Symbol   string    `json:"symbol"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Interval string    `json:"interval"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Services  map[string]ServiceHealth `json:"services"`
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// CollectionStatus represents the status of data collection
type CollectionStatus struct {
	LastRun        time.Time `json:"last_run"`
	NextRun        time.Time `json:"next_run"`
	SuccessfulRuns int       `json:"successful_runs"`
	FailedRuns     int       `json:"failed_runs"`
	LastError      string    `json:"last_error,omitempty"`
	CollectedToday int       `json:"collected_today"`
	IsRunning      bool      `json:"is_running"`
}

// ChartDataPoint represents a data point for Chart.js
type ChartDataPoint struct {
	X string `json:"x"` // Timestamp as ISO string
	Y int64  `json:"y"` // Volume value
}

// ChartData represents the structure for Chart.js data
type ChartData struct {
	Labels   []string       `json:"labels"`
	Datasets []ChartDataset `json:"datasets"`
}

// ChartDataset represents a dataset for Chart.js
type ChartDataset struct {
	Label           string           `json:"label"`
	Data            []ChartDataPoint `json:"data"`
	BorderColor     string           `json:"borderColor"`
	BackgroundColor string           `json:"backgroundColor"`
	Fill            bool             `json:"fill"`
	Tension         float64          `json:"tension"`
}

// GetChartColors returns predefined colors for charts
func GetChartColors() map[string]ChartColors {
	return map[string]ChartColors{
		"PLTR": {Border: "#1f77b4", Background: "rgba(31, 119, 180, 0.1)"},
		"TSLA": {Border: "#ff7f0e", Background: "rgba(255, 127, 14, 0.1)"},
		"BBAI": {Border: "#2ca02c", Background: "rgba(44, 160, 44, 0.1)"},
		"MSFT": {Border: "#d62728", Background: "rgba(214, 39, 40, 0.1)"},
		"NPWR": {Border: "#9467bd", Background: "rgba(148, 103, 189, 0.1)"},
	}
}

// ChartColors represents the color scheme for a chart
type ChartColors struct {
	Border     string `json:"border"`
	Background string `json:"background"`
}

// TimeRange represents a time range for data queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Label string    `json:"label"`
}

// GetCommonTimeRanges returns commonly used time ranges
func GetCommonTimeRanges() []TimeRange {
	now := time.Now()
	return []TimeRange{
		{
			Start: now.AddDate(0, 0, -1),
			End:   now,
			Label: "1D",
		},
		{
			Start: now.AddDate(0, 0, -7),
			End:   now,
			Label: "1W",
		},
		{
			Start: now.AddDate(0, 0, -14),
			End:   now,
			Label: "2W",
		},
		{
			Start: now.AddDate(0, -1, 0),
			End:   now,
			Label: "1M",
		},
	}
}

// WatchedSymbol represents a symbol being watched for data collection
type WatchedSymbol struct {
	ID       int64     `json:"id" db:"id"`
	Symbol   string    `json:"symbol" db:"symbol"`
	Name     string    `json:"name" db:"name"`
	AddedAt  time.Time `json:"added_at" db:"added_at"`
	IsActive bool      `json:"is_active" db:"is_active"`
}

// WatchedSymbolRequest represents a request to add a watched symbol
type WatchedSymbolRequest struct {
	Symbol string `json:"symbol" binding:"required"`
	Name   string `json:"name"`
}

// WatchedSymbolsResponse represents the response for watched symbols
type WatchedSymbolsResponse struct {
	Symbols []WatchedSymbol `json:"symbols"`
	Count   int             `json:"count"`
}
