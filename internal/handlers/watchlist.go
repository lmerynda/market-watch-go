package handlers

import (
	"net/http"
	"strconv"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"
	"market-watch-go/internal/services"

	"github.com/gin-gonic/gin"
)

// WatchlistHandler handles watchlist-related requests
type WatchlistHandler struct {
	db           *database.Database
	stockService *services.StockService
}

// NewWatchlistHandler creates a new watchlist handler
func NewWatchlistHandler(db *database.Database, stockService *services.StockService) *WatchlistHandler {
	return &WatchlistHandler{db: db, stockService: stockService}
}

// Strategy Endpoints

// GetStrategies returns all watchlist strategies
func (h *WatchlistHandler) GetStrategies(c *gin.Context) {
	strategies, err := h.db.GetStrategies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch strategies",
			"details": err.Error(),
		})
		return
	}

	for i, strategy := range strategies {
		stocks, err := h.db.GetStocksByStrategy(strategy.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch stocks for strategy",
				"details": err.Error(),
			})
			return
		}
		strategies[i].Stocks = stocks
	}

	c.JSON(http.StatusOK, gin.H{
		"strategies": strategies,
		"count":      len(strategies),
	})
}

// CreateStrategy creates a new watchlist strategy
func (h *WatchlistHandler) CreateStrategy(c *gin.Context) {
	var strategy models.Strategy
	if err := c.ShouldBindJSON(&strategy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if strategy.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Strategy name is required",
		})
		return
	}

	// Set default color if not provided
	if strategy.Color == "" {
		strategy.Color = "#007bff"
	}

	createdStrategy, err := h.db.CreateStrategy(strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create strategy",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Strategy created successfully",
		"strategy": createdStrategy,
	})
}

// UpdateStrategy updates an existing strategy
func (h *WatchlistHandler) UpdateStrategy(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid strategy ID",
		})
		return
	}

	var strategy models.Strategy
	if err := c.ShouldBindJSON(&strategy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if strategy.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Strategy name is required",
		})
		return
	}

	// Set default color if not provided
	if strategy.Color == "" {
		strategy.Color = "#007bff"
	}

	if err := h.db.UpdateStrategy(id, strategy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update strategy",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Strategy updated successfully",
	})
}

// DeleteStrategy deletes a strategy
func (h *WatchlistHandler) DeleteStrategy(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid strategy ID",
		})
		return
	}

	if err := h.db.DeleteStrategy(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete strategy",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Strategy deleted successfully",
	})
}

// Stocks Endpoints

// GetStocks returns watchlist stocks
func (h *WatchlistHandler) GetStocks(c *gin.Context) {
	stocks, err := h.db.GetStocks()
	if err != nil {
		c.Error(err) // Attach error for middleware logging
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch stocks",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stocks": stocks,
		"count":  len(stocks),
	})
}

// AddStock adds a new stock to the watchlist
func (h *WatchlistHandler) AddStock(c *gin.Context) {
	var stock models.Stock
	if err := c.ShouldBindJSON(&stock); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if stock.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Stock symbol is required",
		})
		return
	}

	addedStock, err := h.db.AddStock(stock)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add stock",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Stock added successfully",
		"stock":   addedStock,
	})
}

// UpdateStock updates an existing stock
func (h *WatchlistHandler) UpdateStock(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid stock ID",
		})
		return
	}

	var stock models.Stock
	if err := c.ShouldBindJSON(&stock); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Update only the notes field - other stock data comes from Polygon API
	if err := h.db.UpdateStockNotes(id, stock.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update stock notes",
			"details": err.Error(),
		})
		return
	}

	// Handle strategy assignments if provided
	if stock.Strategies != nil && len(stock.Strategies) > 0 {
		// First, remove all existing strategy associations
		if err := h.db.RemoveAllStockStrategies(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update stock strategies",
				"details": err.Error(),
			})
			return
		}

		// Then add the new ones
		for _, strategy := range stock.Strategies {
			if err := h.db.AddStockToStrategy(id, strategy.ID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to add stock to strategy",
					"details": err.Error(),
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stock updated successfully",
	})
}

// RemoveStock removes a stock from current strategy and deletes if no strategies remain
func (h *WatchlistHandler) RemoveStock(c *gin.Context) {
	stockID, _ := strconv.Atoi(c.Param("id"))
	strategyID, _ := strconv.Atoi(c.Query("strategy_id"))

	// Remove from specific strategy
	if err := h.db.RemoveStockFromStrategy(stockID, strategyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from strategy"})
		return
	}

	// Check if stock has any remaining strategies
	strategies, _ := h.db.GetStockStrategies(stockID)
	if len(strategies) == 0 {
		// Use stock service to delete stock
		h.stockService.DeleteStock(stockID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock removed successfully"})
}

// Render watchlist page
func (h *WatchlistHandler) RenderWatchlistPage(c *gin.Context) {
	c.HTML(http.StatusOK, "watchlist.html", gin.H{
		"title": "Stock Watchlist",
	})
}
