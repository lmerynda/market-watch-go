package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"market-watch-go/internal/database"
	"market-watch-go/internal/services"
)

// WatchlistRefreshHandler updates price and EMA values for all watchlist stocks
func WatchlistRefreshHandler(db *database.Database, emaService *services.PolygonEMAService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stocks, err := db.GetWatchlistStocks(nil)
		if err != nil {
			log.Printf("Failed to fetch watchlist stocks: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch watchlist stocks", "details": err.Error()})
			return
		}

		updated := 0
		errors := make(map[string]string)
		for _, stock := range stocks {
			log.Printf("[REFRESH] Processing symbol: %s", stock.Symbol)
			// Fetch latest price using GetLastPrice
			price, err := emaService.GetLastPrice(stock.Symbol)
			if err != nil {
				log.Printf("[REFRESH] Error fetching price for %s: %v", stock.Symbol, err)
				errors[stock.Symbol] = "price: " + err.Error()
				continue
			}
			log.Printf("[REFRESH] Price for %s: %v", stock.Symbol, price)

			// Fetch EMAs
			emas, err := emaService.GetEMABatch(stock.Symbol, []int{9, 50, 200}, "day", 1)
			if err != nil {
				log.Printf("[REFRESH] Error fetching EMAs for %s: %v", stock.Symbol, err)
				errors[stock.Symbol] = "ema: " + err.Error()
				continue
			}
			log.Printf("[REFRESH] EMA batch for %s: %+v", stock.Symbol, emas)

			// Update stock in DB
			stock.Price = price
			if ema9, ok := emas[9]; ok && len(ema9.Results.Values) > 0 {
				stock.EMA9 = ema9.Results.Values[0].Value
				log.Printf("[REFRESH] EMA9 for %s: %v", stock.Symbol, stock.EMA9)
			}
			if ema50, ok := emas[50]; ok && len(ema50.Results.Values) > 0 {
				stock.EMA50 = ema50.Results.Values[0].Value
				log.Printf("[REFRESH] EMA50 for %s: %v", stock.Symbol, stock.EMA50)
			}
			if ema200, ok := emas[200]; ok && len(ema200.Results.Values) > 0 {
				stock.EMA200 = ema200.Results.Values[0].Value
				log.Printf("[REFRESH] EMA200 for %s: %v", stock.Symbol, stock.EMA200)
			}
			log.Printf("[REFRESH] Updating DB for %s: Price=%v, EMA9=%v, EMA50=%v, EMA200=%v", stock.Symbol, stock.Price, stock.EMA9, stock.EMA50, stock.EMA200)
			err = db.UpdateWatchlistStockWithEMA(stock.ID, stock)
			if err != nil {
				log.Printf("[REFRESH] Error updating DB for %s: %v", stock.Symbol, err)
				errors[stock.Symbol] = "db: " + err.Error()
				continue
			}
			updated++
		}

		c.JSON(http.StatusOK, gin.H{
			"updated": updated,
			"errors": errors,
		})
	}
}
