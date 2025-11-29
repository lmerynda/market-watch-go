package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"market-watch-go/internal/services"
)

// OptionHandler provides a simple endpoint for a single hard-coded option contract
type OptionHandler struct {
	ms *services.MassiveService
}

func NewOptionHandler(ms *services.MassiveService) *OptionHandler {
	return &OptionHandler{ms: ms}
}

// GetCOINCallAggregates returns 1-minute aggregates for the COIN 11/14/2025 265C contract
// Route: GET /api/options/coin-2025-11-14-265C
func (h *OptionHandler) GetCOINCallAggregates(c *gin.Context) {
	// Hard-coded ticker for the requested contract
	ticker := "O:COIN251114C00265000"

	// Fetch aggregates for the expiration date (00:00 to 23:59 UTC)
	from := time.Date(2025, 11, 14, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)

	data, err := h.ms.GetOptionAggregates(context.Background(), ticker, from, to, 1, "minute")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map to simple response objects
	resp := make([]gin.H, 0, len(data))
	for _, d := range data {
		resp = append(resp, gin.H{
			"timestamp": d.Timestamp.UTC().Format(time.RFC3339),
			"open":      d.Open,
			"high":      d.High,
			"low":       d.Low,
			"close":     d.Close,
			"volume":    d.Volume,
		})
	}

	c.JSON(http.StatusOK, gin.H{"ticker": ticker, "data": resp})
}
