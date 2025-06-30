package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"market-watch-go/internal/database"
	"market-watch-go/internal/services"
)

// WatchlistRefreshHandler triggers stock refresh service
func WatchlistRefreshHandler(db *database.Database, stockService *services.StockService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[WATCHLIST] Triggering stock refresh...")
		
		err := stockService.RefreshAllStocks()
		if err != nil {
			log.Printf("[WATCHLIST] Stock refresh failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Stock refresh failed",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Stock refresh triggered successfully",
		})
	}
}
