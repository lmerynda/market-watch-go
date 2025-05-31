package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	templatesPath string
	staticPath    string
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(templatesPath, staticPath string) *DashboardHandler {
	return &DashboardHandler{
		templatesPath: templatesPath,
		staticPath:    staticPath,
	}
}

// Index handles GET / - serves the main dashboard
func (dh *DashboardHandler) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":   "Market Watch - Volume Tracker",
		"symbols": []string{"PLTR", "TSLA", "BBAI", "MSFT", "NPWR"},
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
