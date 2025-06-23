package services

import (
	"context"
	"fmt"
	"log"
	"time"

	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
)

// PolygonEMAService wraps the official Polygon Go client for EMA queries
// You must set the API key in config.yaml

type PolygonEMAService struct {
	client *polygon.Client
}

func NewPolygonEMAService(apiKey string) *PolygonEMAService {
	if apiKey == "" {
		log.Fatal("Polygon API key is required for PolygonEMAService. Please set it in config.yaml.")
	}
	return &PolygonEMAService{
		client: polygon.New(apiKey),
	}
}

// GetEMA queries Polygon for the EMA for a symbol, window, and timespan (e.g. 9, 50, 200, "day")
func (s *PolygonEMAService) GetEMA(symbol string, window int, timespan string, limit int) (*models.GetEMAResponse, error) {
	params := models.GetEMAParams{
		Ticker: symbol,
	}.
		WithTimespan(models.Timespan(timespan)).
		WithAdjusted(true).
		WithWindow(window).
		WithSeriesType(models.SeriesType("close")).
		WithOrder(models.Order("desc")).
		WithLimit(limit)

	log.Printf("[PolygonEMAService] GetEMA request for %s (window %d): %+v", symbol, window, params)

	resp, err := s.client.GetEMA(context.Background(), params)
	if err != nil {
		log.Printf("[PolygonEMAService] GetEMA error for %s (window %d): %v", symbol, window, err)
		return nil, err
	}
	log.Printf("[PolygonEMAService] GetEMA response for %s (window %d): %+v", symbol, window, resp)
	return resp, nil
}

// GetEMABatch queries Polygon for multiple EMAs (e.g. 9, 50, 200) for a symbol
func (s *PolygonEMAService) GetEMABatch(symbol string, windows []int, timespan string, limit int) (map[int]*models.GetEMAResponse, error) {
	results := make(map[int]*models.GetEMAResponse)
	for _, w := range windows {
		resp, err := s.GetEMA(symbol, w, timespan, limit)
		if err != nil {
			return nil, err
		}
		results[w] = resp
	}
	return results, nil
}

// GetLastPrice fetches the last close price for a symbol from Polygon (free-tier compatible)
func (s *PolygonEMAService) GetLastPrice(symbol string) (float64, error) {
	from := models.Millis(time.Now().AddDate(0, 0, -7))
	to := models.Millis(time.Now())
	limit := 1
	params := &models.GetAggsParams{
		Ticker:     symbol,
		Multiplier: 1,
		Timespan:   "day",
		From:       from,
		To:         to,
		Limit:      &limit,
	}

	log.Printf("[PolygonEMAService] GetLastPrice (daily agg) request for %s: %+v", symbol, params)
	resp, err := s.client.GetAggs(context.Background(), params)
	if err != nil {
		log.Printf("[PolygonEMAService] GetLastPrice (daily agg) error for %s: %v", symbol, err)
		return 0, err
	}
	log.Printf("[PolygonEMAService] GetLastPrice (daily agg) response for %s: %+v", symbol, resp)
	if resp == nil || len(resp.Results) == 0 {
		return 0, fmt.Errorf("no current price data available for %s", symbol)
	}
	return resp.Results[0].Close, nil
}

// Helper to extract the latest EMA value from the response
func ExtractLatestEMA(resp *models.GetEMAResponse) (float64, error) {
	if resp == nil {
		return 0, nil
	}
	fmt.Printf("[DEBUG] EMA resp.Results: %+v\n", resp.Results)
	return 0, nil // placeholder until we know the correct field
}
