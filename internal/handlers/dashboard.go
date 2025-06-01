package handlers

import (
	"log"
	"net/http"
	"path/filepath"

	"market-watch-go/internal/database"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	templatesPath string
	staticPath    string
	db            *database.DB
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(templatesPath, staticPath string, db *database.DB) *DashboardHandler {
	return &DashboardHandler{
		templatesPath: templatesPath,
		staticPath:    staticPath,
		db:            db,
	}
}

// Index handles GET / - serves the main dashboard
func (dh *DashboardHandler) Index(c *gin.Context) {
	// Get current watched symbols from database
	symbols, err := dh.db.GetWatchedSymbols()
	if err != nil {
		log.Printf("Failed to load watched symbols for dashboard: %v", err)
		// Fallback to default symbols if database fails
		symbols = []string{"PLTR", "TSLA", "BBAI", "MSFT", "NPWR"}
	}

	// If no symbols are being watched, use defaults
	if len(symbols) == 0 {
		symbols = []string{"PLTR", "TSLA", "BBAI", "MSFT", "NPWR"}
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":   "Market Watch - Volume Tracker",
		"symbols": symbols,
	})
}

// ServeStatic serves static files (CSS, JS, etc.)
func (dh *DashboardHandler) ServeStatic(c *gin.Context) {
	// Get the requested file path
	requestedFile := c.Param("filepath")

	// Construct the full path
	fullPath := filepath.Join(dh.staticPath, requestedFile)

	// Serve the file
	c.File(fullPath)
}
