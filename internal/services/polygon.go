package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"market-watch-go/internal/config"
	localmodels "market-watch-go/internal/models"
)

type PolygonService struct {
	client *http.Client
	cfg    *config.Config
}

// PolygonResponse represents the response from Polygon.io aggregates API
type PolygonResponse struct {
	Ticker       string          `json:"ticker"`
	QueryCount   int             `json:"queryCount"`
	ResultsCount int             `json:"resultsCount"`
	Adjusted     bool            `json:"adjusted"`
	Results      []PolygonResult `json:"results"`
	Status       string          `json:"status"`
	RequestID    string          `json:"request_id"`
	Count        int             `json:"count"`
}

// PolygonResult represents a single aggregate data point from Polygon.io
type PolygonResult struct {
	Open      float64 `json:"o"`  // Open price
	High      float64 `json:"h"`  // High price
	Low       float64 `json:"l"`  // Low price
	Close     float64 `json:"c"`  // Close price
	Volume    float64 `json:"v"`  // Volume
	VolumeWAP float64 `json:"vw"` // Volume weighted average price
	Timestamp int64   `json:"t"`  // Timestamp (Unix milliseconds)
	Count     int     `json:"n"`  // Number of transactions
}

// NewPolygonService creates a new Polygon.io service
func NewPolygonService(cfg *config.Config) *PolygonService {
	client := &http.Client{
		Timeout: cfg.Polygon.Timeout,
	}

	return &PolygonService{
		client: client,
		cfg:    cfg,
	}
}

// GetAggregates fetches aggregated data for a symbol
func (ps *PolygonService) GetAggregates(symbol string, from, to time.Time) ([]*localmodels.VolumeData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ps.cfg.Polygon.Timeout)
	defer cancel()

	// Format dates for Polygon API
	fromStr := from.Format("2006-01-02")
	toStr := to.Format("2006-01-02")

	// Build the URL for 5-minute aggregates
	url := fmt.Sprintf("%s/v2/aggs/ticker/%s/range/5/minute/%s/%s?adjusted=true&sort=asc&limit=50000&apikey=%s",
		ps.cfg.Polygon.BaseURL, symbol, fromStr, toStr, ps.cfg.Polygon.APIKey)

	// Make the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ps.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var polygonResp PolygonResponse
	if err := json.NewDecoder(resp.Body).Decode(&polygonResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if the response is successful (accept both OK and DELAYED)
	if polygonResp.Status != "OK" && polygonResp.Status != "DELAYED" {
		return nil, fmt.Errorf("API response status: %s", polygonResp.Status)
	}

	// Log if we got DELAYED response
	if polygonResp.Status == "DELAYED" {
		log.Printf("Received DELAYED response for %s (normal for free tier)", symbol)
	}

	var volumeData []*localmodels.VolumeData

	// Convert results to our format
	for _, result := range polygonResp.Results {
		// Convert timestamp from milliseconds to time.Time
		timestamp := time.Unix(result.Timestamp/1000, (result.Timestamp%1000)*1000000)

		vd := &localmodels.VolumeData{
			Symbol:    symbol,
			Timestamp: timestamp,
			Volume:    int64(result.Volume),
			CreatedAt: time.Now(),
		}

		volumeData = append(volumeData, vd)
	}

	log.Printf("Fetched %d data points for %s from %s to %s (Status: %s)",
		len(volumeData), symbol, fromStr, toStr, polygonResp.Status)

	// Debug: log first few data points
	if len(volumeData) > 0 {
		log.Printf("Sample data for %s: Volume=%d, Time=%s",
			symbol, volumeData[0].Volume, volumeData[0].Timestamp.Format("15:04:05"))
	}

	return volumeData, nil
}

// GetLatestAggregates fetches the most recent aggregated data
func (ps *PolygonService) GetLatestAggregates(symbol string, minutes int) ([]*localmodels.VolumeData, error) {
	to := time.Now()
	from := to.Add(-time.Duration(minutes) * time.Minute)

	return ps.GetAggregates(symbol, from, to)
}

// GetTodayAggregates fetches today's aggregated data
func (ps *PolygonService) GetTodayAggregates(symbol string) ([]*localmodels.VolumeData, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return ps.GetAggregates(symbol, today, now)
}

// GetMultipleSymbolAggregates fetches aggregates for multiple symbols
func (ps *PolygonService) GetMultipleSymbolAggregates(symbols []string, from, to time.Time) (map[string][]*localmodels.VolumeData, error) {
	result := make(map[string][]*localmodels.VolumeData)

	for _, symbol := range symbols {
		data, err := ps.GetAggregates(symbol, from, to)
		if err != nil {
			log.Printf("Failed to get aggregates for %s: %v", symbol, err)
			// Continue with other symbols even if one fails
			continue
		}
		result[symbol] = data

		// Add a small delay to avoid hitting rate limits
		time.Sleep(200 * time.Millisecond)
	}

	return result, nil
}

// GetCurrentPrice fetches the current price for a symbol using daily aggregates
func (ps *PolygonService) GetCurrentPrice(symbol string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ps.cfg.Polygon.Timeout)
	defer cancel()

	// Get the latest daily aggregate
	yesterday := time.Now().AddDate(0, 0, -1)
	today := time.Now()

	fromStr := yesterday.Format("2006-01-02")
	toStr := today.Format("2006-01-02")

	// Build the URL for daily aggregates
	url := fmt.Sprintf("%s/v2/aggs/ticker/%s/range/1/day/%s/%s?adjusted=true&sort=desc&limit=1&apikey=%s",
		ps.cfg.Polygon.BaseURL, symbol, fromStr, toStr, ps.cfg.Polygon.APIKey)

	// Make the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ps.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var polygonResp PolygonResponse
	if err := json.NewDecoder(resp.Body).Decode(&polygonResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if polygonResp.Status != "OK" || len(polygonResp.Results) == 0 {
		return 0, fmt.Errorf("no current price data available for %s", symbol)
	}

	return polygonResp.Results[0].Close, nil
}

// ValidateAPIKey checks if the API key is valid by making a test request
func (ps *PolygonService) ValidateAPIKey() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Make a simple test request for Apple stock
	yesterday := time.Now().AddDate(0, 0, -1)
	today := time.Now()

	fromStr := yesterday.Format("2006-01-02")
	toStr := today.Format("2006-01-02")

	url := fmt.Sprintf("%s/v2/aggs/ticker/AAPL/range/1/day/%s/%s?adjusted=true&sort=desc&limit=1&apikey=%s",
		ps.cfg.Polygon.BaseURL, fromStr, toStr, ps.cfg.Polygon.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ps.client.Do(req)
	if err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("API key validation failed: unauthorized (status 401)")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API key validation failed: status %d", resp.StatusCode)
	}

	var polygonResp PolygonResponse
	if err := json.NewDecoder(resp.Body).Decode(&polygonResp); err != nil {
		return fmt.Errorf("failed to decode validation response: %w", err)
	}

	if polygonResp.Status != "OK" && polygonResp.Status != "DELAYED" {
		return fmt.Errorf("API key validation failed: %s", polygonResp.Status)
	}

	// DELAYED status is acceptable for free tier API keys
	if polygonResp.Status == "DELAYED" {
		log.Printf("Polygon API validation successful (DELAYED response is normal for free tier)")
	}

	log.Printf("Polygon API key validation successful")
	return nil
}

// HealthCheck performs a health check on the Polygon service
func (ps *PolygonService) HealthCheck() error {
	return ps.ValidateAPIKey()
}

// GetLastTradingDay returns the last trading day
func (ps *PolygonService) GetLastTradingDay() time.Time {
	now := time.Now()

	// If it's weekend, go back to Friday
	for now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		now = now.AddDate(0, 0, -1)
	}

	// Since we're operating in 24/7 mode, we don't need to check market hours
	// Just ensure we're not on a weekend
	return now
}

// CollectCurrentData collects current volume data for all configured symbols
func (ps *PolygonService) CollectCurrentData() (map[string][]*localmodels.VolumeData, error) {
	// Get data for today
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return ps.GetMultipleSymbolAggregates(ps.cfg.Collection.Symbols, today, now)
}

// GetHistoricalData fetches historical data for a symbol over specified days
func (ps *PolygonService) GetHistoricalData(symbol string, days int) ([]*localmodels.VolumeData, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -days)

	return ps.GetAggregates(symbol, from, to)
}

// GetPriceAggregates fetches price aggregated data for a symbol
func (ps *PolygonService) GetPriceAggregates(symbol string, from, to time.Time) ([]*localmodels.PriceData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ps.cfg.Polygon.Timeout)
	defer cancel()

	// Format dates for Polygon API
	fromStr := from.Format("2006-01-02")
	toStr := to.Format("2006-01-02")

	// Build the URL for 5-minute aggregates
	url := fmt.Sprintf("%s/v2/aggs/ticker/%s/range/5/minute/%s/%s?adjusted=true&sort=asc&limit=50000&apikey=%s",
		ps.cfg.Polygon.BaseURL, symbol, fromStr, toStr, ps.cfg.Polygon.APIKey)

	// Make the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ps.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var polygonResp PolygonResponse
	if err := json.NewDecoder(resp.Body).Decode(&polygonResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if the response is successful (accept both OK and DELAYED)
	if polygonResp.Status != "OK" && polygonResp.Status != "DELAYED" {
		return nil, fmt.Errorf("API response status: %s", polygonResp.Status)
	}

	// Log if we got DELAYED response
	if polygonResp.Status == "DELAYED" {
		log.Printf("Received DELAYED response for %s (normal for free tier)", symbol)
	}

	var priceData []*localmodels.PriceData

	// Convert results to our format
	for _, result := range polygonResp.Results {
		// Convert timestamp from milliseconds to time.Time
		timestamp := time.Unix(result.Timestamp/1000, (result.Timestamp%1000)*1000000)

		pd := &localmodels.PriceData{
			Symbol:    symbol,
			Timestamp: timestamp,
			Open:      result.Open,
			High:      result.High,
			Low:       result.Low,
			Close:     result.Close,
			Volume:    int64(result.Volume),
			CreatedAt: time.Now(),
		}

		priceData = append(priceData, pd)
	}

	log.Printf("Fetched %d price data points for %s from %s to %s (Status: %s)",
		len(priceData), symbol, fromStr, toStr, polygonResp.Status)

	// Debug: log first few data points
	if len(priceData) > 0 {
		log.Printf("Sample price data for %s: OHLC=%.2f/%.2f/%.2f/%.2f, Volume=%d, Time=%s",
			symbol, priceData[0].Open, priceData[0].High, priceData[0].Low, priceData[0].Close,
			priceData[0].Volume, priceData[0].Timestamp.Format("15:04:05"))
	}

	return priceData, nil
}

// GetLatestPriceAggregates fetches the most recent price aggregated data
func (ps *PolygonService) GetLatestPriceAggregates(symbol string, minutes int) ([]*localmodels.PriceData, error) {
	to := time.Now()
	from := to.Add(-time.Duration(minutes) * time.Minute)

	return ps.GetPriceAggregates(symbol, from, to)
}

// GetTodayPriceAggregates fetches today's price aggregated data
func (ps *PolygonService) GetTodayPriceAggregates(symbol string) ([]*localmodels.PriceData, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return ps.GetPriceAggregates(symbol, today, now)
}

// GetHistoricalPriceData fetches historical price data for a symbol over specified days
func (ps *PolygonService) GetHistoricalPriceData(symbol string, days int) ([]*localmodels.PriceData, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -days)

	return ps.GetPriceAggregates(symbol, from, to)
}
