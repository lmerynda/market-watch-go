package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"market-watch-go/internal/config"
	"market-watch-go/internal/database"
	"market-watch-go/internal/handlers"
	"market-watch-go/internal/models"
	"market-watch-go/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from the existing config file
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration from configs/config.yaml: %v", err)
	}

	log.Printf("✅ Configuration loaded from configs/config.yaml")
	log.Printf("📊 Database: %s", cfg.Database.Path)
	log.Printf("🔗 Polygon API: %s", cfg.Polygon.BaseURL)
	log.Printf("📈 Symbols: %v", cfg.Collection.Symbols)

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Polygon service with API key from config
	polygonService := services.NewPolygonService(cfg)

	// Validate Polygon API key from config
	if err := polygonService.ValidateAPIKey(); err != nil {
		log.Printf("❌ Polygon API validation failed: %v", err)
		log.Printf("Note: Check API key in configs/config.yaml")
	} else {
		log.Printf("✅ Polygon API key validated successfully")
	}

	// Initialize collector service
	collectorService := services.NewCollectorService(db, polygonService, cfg)

	// Initialize technical analysis and other services
	taService := services.NewTechnicalAnalysisService(db, &services.TechnicalAnalysisConfig{})
	srService := services.NewSupportResistanceService(db, taService)
	setupService := services.NewSetupDetectionService(db, taService, srService)

	// Initialize handlers
	taHandler := handlers.NewTechnicalAnalysisHandler(db, taService)
	srHandler := handlers.NewSupportResistanceHandler(db, srService)
	setupHandler := handlers.NewSetupHandler(db, setupService)

	// Check if we have any price data, if not collect some
	count, err := db.GetPriceDataCount()
	if err != nil {
		log.Printf("Failed to check price data count: %v", err)
	} else if count == 0 {
		log.Printf("📊 No price data found. Collecting historical data for symbols: %v", cfg.Collection.Symbols)

		// Collect 7 days of historical data for all symbols
		go func() {
			if err := collectorService.CollectHistoricalData(7); err != nil {
				log.Printf("❌ Failed to collect historical data: %v", err)
			} else {
				log.Printf("✅ Historical data collection completed")
			}
		}()
	} else {
		log.Printf("✅ Found %d price data records in database", count)
	}

	// Start the collector service for real-time data collection
	if err := collectorService.Start(); err != nil {
		log.Printf("❌ Failed to start collector service: %v", err)
	} else {
		log.Printf("✅ Data collector started (interval: %v)", cfg.Collection.Interval)
	}

	// Setup Gin router
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("web/templates/*")

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "market-watch-go",
			"polygon": "connected",
			"config":  "configs/config.yaml",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Technical Analysis routes
		api.GET("/indicators/:symbol", taHandler.GetIndicators)
		api.GET("/indicators", taHandler.GetMultipleIndicators)

		// Support/Resistance routes
		api.GET("/support-resistance/:symbol/levels", srHandler.GetSupportResistanceLevels)
		api.POST("/support-resistance/:symbol/detect", srHandler.DetectSupportResistance)
		api.GET("/support-resistance/:symbol/nearest", srHandler.GetNearestLevels)
		api.GET("/support-resistance/:symbol/touches", srHandler.GetLevelTouches)
		api.GET("/support-resistance/:symbol/pivots", srHandler.GetPivotPoints)
		api.GET("/support-resistance/:symbol/summary", srHandler.GetLevelSummary)
		api.GET("/support-resistance/levels", srHandler.GetMultipleLevels)
		api.POST("/support-resistance/cleanup", srHandler.CleanupOldData)
		api.POST("/support-resistance/deactivate", srHandler.DeactivateOldLevels)

		// Setup Detection routes
		api.POST("/setups/:symbol/detect", setupHandler.DetectSetups)
		api.GET("/setups/:symbol", setupHandler.GetSetups)
		api.GET("/setups/id/:id", setupHandler.GetSetupByID)
		api.PUT("/setups/id/:id/status", setupHandler.UpdateSetupStatus)
		api.GET("/setups", setupHandler.GetMultipleSetups)
		api.GET("/setups/:symbol/summary", setupHandler.GetSetupSummary)
		api.GET("/setups/high-quality", setupHandler.GetHighQualitySetups)
		api.POST("/setups/expire", setupHandler.ExpireOldSetups)
		api.POST("/setups/cleanup", setupHandler.CleanupOldSetups)
		api.GET("/setups/id/:id/checklist", setupHandler.GetSetupChecklist)
		api.GET("/setups/stats", setupHandler.GetSetupsStats)

		// Data collection routes
		api.GET("/collection/status", func(c *gin.Context) {
			stats := collectorService.GetStats()
			c.JSON(http.StatusOK, gin.H{
				"status":              "running",
				"last_run":            stats.LastRun,
				"next_run":            stats.NextRun,
				"successful_runs":     stats.SuccessfulRuns,
				"failed_runs":         stats.FailedRuns,
				"collected_today":     stats.CollectedToday,
				"total_collected":     stats.TotalCollected,
				"is_running":          stats.IsRunning,
				"last_error":          stats.LastError,
				"active_symbols":      cfg.Collection.Symbols,
				"collection_interval": cfg.Collection.Interval.String(),
			})
		})

		api.POST("/collection/force", func(c *gin.Context) {
			if err := collectorService.ForceCollection(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to force collection",
					"details": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "Collection forced successfully",
			})
		})

		api.POST("/collection/historical/:days", func(c *gin.Context) {
			daysStr := c.Param("days")
			days, err := strconv.Atoi(daysStr)
			if err != nil || days <= 0 || days > 365 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid days parameter (must be 1-365)",
				})
				return
			}

			go func() {
				if err := collectorService.CollectHistoricalData(days); err != nil {
					log.Printf("Historical collection failed: %v", err)
				}
			}()

			c.JSON(http.StatusOK, gin.H{
				"message": fmt.Sprintf("Historical data collection started for %d days", days),
			})
		})

		// Price chart endpoint with real data from database
		api.GET("/price/:symbol/chart", func(c *gin.Context) {
			symbol := c.Param("symbol")
			rangeParam := c.DefaultQuery("range", "1D")

			// Get price data from database
			var from time.Time
			to := time.Now()

			switch rangeParam {
			case "1D":
				from = to.AddDate(0, 0, -1)
			case "1W":
				from = to.AddDate(0, 0, -7)
			case "1M":
				from = to.AddDate(0, -1, 0)
			default:
				from = to.AddDate(0, 0, -1)
			}

			priceData, err := db.GetPriceData(&models.PriceDataFilter{
				Symbol: symbol,
				From:   from,
				To:     to,
				Limit:  1000,
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to get price data",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"symbol": symbol,
				"range":  rangeParam,
				"data":   priceData,
				"count":  len(priceData),
				"from":   from,
				"to":     to,
			})
		})
	}

	// Dashboard route - serve the HTML template
	router.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "Market Watch Dashboard",
			"symbols": cfg.Collection.Symbols,
		})
	})

	// Static files
	router.Static("/static", "./web/static")

	// Frontend route with config information
	router.GET("/", func(c *gin.Context) {
		dataCount, _ := db.GetPriceDataCount()

		c.JSON(http.StatusOK, gin.H{
			"message":       "🚀 Market Watch API with Polygon.io",
			"version":       "1.0.0",
			"status":        "Operational",
			"config_source": "configs/config.yaml",
			"data_points":   dataCount,
			"features": []string{
				"📈 Real-time data from Polygon.io API",
				"🎯 Technical Analysis with custom indicators",
				"💡 Support/Resistance Detection (100-point scoring)",
				"✅ Trading Setup Detection (20-item checklist)",
				"⚡ Automated data collection & analysis",
				"🖥️ Interactive web dashboard",
			},
			"polygon_integration": gin.H{
				"status":       "active",
				"api_url":      cfg.Polygon.BaseURL,
				"symbols":      cfg.Collection.Symbols,
				"interval":     cfg.Collection.Interval.String(),
				"market_hours": cfg.Collection.MarketHours,
			},
			"database": gin.H{
				"path":            cfg.Database.Path,
				"max_connections": cfg.Database.MaxOpenConns,
			},
			"ui": gin.H{
				"dashboard": "/dashboard",
				"features": []string{
					"Real-time price charts with D3.js",
					"Technical indicators visualization",
					"Trading setups with quality scores",
					"Support/resistance level analysis",
					"Data collection monitoring",
				},
			},
			"endpoints": gin.H{
				"health":              "/health",
				"dashboard":           "/dashboard",
				"indicators":          "/api/indicators/:symbol",
				"support_resistance":  "/api/support-resistance/:symbol/levels",
				"setups":              "/api/setups/:symbol",
				"high_quality_setups": "/api/setups/high-quality",
				"setup_detection":     "/api/setups/:symbol/detect",
				"price_chart":         "/api/price/:symbol/chart",
				"collection_status":   "/api/collection/status",
				"force_collection":    "POST /api/collection/force",
			},
			"examples": gin.H{
				"view_dashboard":    "GET /dashboard",
				"get_indicators":    "GET /api/indicators/PLTR",
				"detect_setups":     "POST /api/setups/PLTR/detect",
				"get_high_quality":  "GET /api/setups/high-quality",
				"get_sr_levels":     "GET /api/support-resistance/PLTR/levels",
				"get_price_chart":   "GET /api/price/PLTR/chart?range=1W",
				"collection_status": "GET /api/collection/status",
				"force_collection":  "POST /api/collection/force",
				"historical_data":   "POST /api/collection/historical/7",
			},
		})
	})

	// Start server
	port := strconv.Itoa(cfg.Server.Port)
	log.Printf("🚀 Market Watch server starting on port %s", port)
	log.Printf("📊 Phase 3 Complete: Advanced Setup Detection & Scoring")
	log.Printf("📁 Config: configs/config.yaml")
	log.Printf("🔗 Polygon.io: %s", cfg.Polygon.BaseURL)
	log.Printf("📊 Database: %s", cfg.Database.Path)
	log.Printf("📈 Tracked symbols: %v", cfg.Collection.Symbols)
	log.Printf("⏰ Collection interval: %v", cfg.Collection.Interval)
	log.Printf("🎯 API available at: http://localhost:%s", port)
	log.Printf("🖥️ DASHBOARD available at: http://localhost:%s/dashboard", port)
	log.Printf("✨ Features: Real-time data + Technical Analysis + S/R Detection + Setup Intelligence + Web UI")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
