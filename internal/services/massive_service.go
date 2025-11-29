package services

import (
	"context"
	"fmt"
	"time"

	massive "github.com/massive-com/client-go/v2/rest"
	mm "github.com/massive-com/client-go/v2/rest/models"

	"market-watch-go/internal/config"
	localmodels "market-watch-go/internal/models"
)

// MassiveService wraps the Massive.com REST client for market data.
type MassiveService struct {
	client *massive.Client
	cfg    *config.Config
}

// NewMassiveService creates a new MassiveService
func NewMassiveService(cfg *config.Config) *MassiveService {
	apiKey := cfg.Polygon.APIKey // kept in config for backward compatibility
	c := massive.New(apiKey)
	return &MassiveService{client: c, cfg: cfg}
}

// GetOptionAggregates fetches aggregates (bars) for an option contract ticker
// timespan is e.g. "minute"; multiplier is number of minutes per bar (1 for 1m)
func (ms *MassiveService) GetOptionAggregates(ctx context.Context, ticker string, from, to time.Time, multiplier int, timespan string) ([]*localmodels.PriceData, error) {
	// Build params
	base := mm.ListAggsParams{
		Ticker:     ticker,
		Multiplier: multiplier,
		Timespan:   mm.Timespan(timespan),
		From:       mm.Millis(from),
		To:         mm.Millis(to),
	}

	params := base.WithAdjusted(true).WithOrder(mm.Asc).WithLimit(50000)

	iter := ms.client.ListAggs(ctx, params)

	var out []*localmodels.PriceData

	for iter.Next() {
		a := iter.Item()

		// Map massive models.Agg -> local PriceData
		// Convert mm.Millis (alias to time.Time) to Unix milliseconds
		msUnix := time.Time(a.Timestamp).UnixMilli()
		pd := &localmodels.PriceData{
			Symbol:    ticker,
			Timestamp: time.Unix(msUnix/1000, (msUnix%1000)*1000000),
			Open:      a.Open,
			High:      a.High,
			Low:       a.Low,
			Close:     a.Close,
			Volume:    int64(a.Volume),
			CreatedAt: time.Now(),
		}

		out = append(out, pd)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list aggs: %w", err)
	}

	return out, nil
}
