package services

import (
	"log"
	"market-watch-go/internal/database"
)

// StockService handles stock data refresh operations
type StockService struct {
	db             *database.Database
	polygonService *PolygonService
	emaService     *PolygonEMAService
}

// NewStockService creates a new stock service
func NewStockService(db *database.Database, polygonService *PolygonService, emaService *PolygonEMAService) *StockService {
	return &StockService{
		db:             db,
		polygonService: polygonService,
		emaService:     emaService,
	}
}

// RefreshAllStocks refreshes price and EMA data for all stocks
func (ss *StockService) RefreshAllStocks() error {
	log.Printf("[STOCK_SERVICE] Starting stock refresh...")
	
	stocks, err := ss.db.GetStocks()
	if err != nil {
		return err
	}

	updated := 0
	for _, stock := range stocks {
		if err := ss.RefreshStock(stock.Symbol); err != nil {
			log.Printf("[STOCK_SERVICE] Failed to refresh %s: %v", stock.Symbol, err)
			continue
		}
		updated++
	}

	log.Printf("[STOCK_SERVICE] Refreshed %d stocks", updated)
	return nil
}

// RefreshStock refreshes data for a single stock
func (ss *StockService) RefreshStock(symbol string) error {
	// Fetch latest price
	price, err := ss.emaService.GetLastPrice(symbol)
	if err != nil {
		return err
	}

	// Fetch EMAs
	emas, err := ss.emaService.GetEMABatch(symbol, []int{9, 50, 200}, "day", 1)
	if err != nil {
		return err
	}

	var ema9, ema50, ema200 float64
	if ema, ok := emas[9]; ok && len(ema.Results.Values) > 0 {
		ema9 = ema.Results.Values[0].Value
	}
	if ema, ok := emas[50]; ok && len(ema.Results.Values) > 0 {
		ema50 = ema.Results.Values[0].Value
	}
	if ema, ok := emas[200]; ok && len(ema.Results.Values) > 0 {
		ema200 = ema.Results.Values[0].Value
	}

	return ss.db.RefreshStock(symbol, price, ema9, ema50, ema200)
}

// DeleteStock deletes a stock from the stocks table
func (ss *StockService) DeleteStock(stockID int) error {
	return ss.db.DeleteStock(stockID)
}