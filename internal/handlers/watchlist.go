package handlers

import (
	"net/http"
	"strconv"

	"market-watch-go/internal/database"
	"market-watch-go/internal/models"

	"github.com/gin-gonic/gin"
)

// WatchlistHandler handles watchlist-related requests
type WatchlistHandler struct {
	db *database.Database
}

// NewWatchlistHandler creates a new watchlist handler
func NewWatchlistHandler(db *database.Database) *WatchlistHandler {
	return &WatchlistHandler{db: db}
}

// Categories Endpoints

// GetCategories returns all watchlist categories
func (h *WatchlistHandler) GetCategories(c *gin.Context) {
	categories, err := h.db.GetWatchlistCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch categories",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"count":      len(categories),
	})
}

// CreateCategory creates a new watchlist category
func (h *WatchlistHandler) CreateCategory(c *gin.Context) {
	var category models.WatchlistCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if category.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Category name is required",
		})
		return
	}

	// Set default color if not provided
	if category.Color == "" {
		category.Color = "#007bff"
	}

	createdCategory, err := h.db.CreateWatchlistCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create category",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Category created successfully",
		"category": createdCategory,
	})
}

// UpdateCategory updates an existing category
func (h *WatchlistHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	var category models.WatchlistCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.db.UpdateWatchlistCategory(id, category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update category",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category updated successfully",
	})
}

// DeleteCategory deletes a category
func (h *WatchlistHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	if err := h.db.DeleteWatchlistCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete category",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}

// Stocks Endpoints

// GetStocks returns watchlist stocks, optionally filtered by category
func (h *WatchlistHandler) GetStocks(c *gin.Context) {
	var categoryID *int
	if categoryStr := c.Query("category_id"); categoryStr != "" {
		id, err := strconv.Atoi(categoryStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid category_id parameter",
			})
			return
		}
		categoryID = &id
	}

	stocks, err := h.db.GetWatchlistStocks(categoryID)
	if err != nil {
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
	var stock models.WatchlistStock
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

	addedStock, err := h.db.AddWatchlistStock(stock)
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
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid stock ID",
		})
		return
	}

	var stock models.WatchlistStock
	if err := c.ShouldBindJSON(&stock); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.db.UpdateWatchlistStock(id, stock); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update stock",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stock updated successfully",
	})
}

// DeleteStock removes a stock from the watchlist
func (h *WatchlistHandler) DeleteStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid stock ID",
		})
		return
	}

	if err := h.db.DeleteWatchlistStock(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete stock",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stock deleted successfully",
	})
}

// GetSummary returns watchlist summary statistics
func (h *WatchlistHandler) GetSummary(c *gin.Context) {
	summary, err := h.db.GetWatchlistSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch summary",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
	})
}

// Render watchlist page
func (h *WatchlistHandler) RenderWatchlistPage(c *gin.Context) {
	c.HTML(http.StatusOK, "watchlist.html", gin.H{
		"title": "Stock Watchlist",
	})
}
