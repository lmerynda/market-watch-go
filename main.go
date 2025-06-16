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
	log.Printf("📈 Config symbols: %v", cfg.Collection.Symbols)

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Ensure config symbols are in the watched symbols table
	if err := db.EnsureConfigSymbolsWatched(cfg.Collection.Symbols); err != nil {
		log.Fatalf("Failed to ensure config symbols are watched: %v", err)
	}

	// Get watched symbols from database (should now include config symbols)
	watchedSymbols, err := db.GetWatchedSymbols()
	if err != nil {
		log.Fatalf("Failed to get watched symbols: %v", err)
	}

	log.Printf("📈 Watched symbols from database: %v", watchedSymbols)

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

	// Initialize email service
	emailService := services.NewEmailService(cfg)

	// Initialize head and shoulders detection service
	hsService := services.NewHeadShouldersDetectionService(db, setupService, taService, emailService)

	// Initialize automatic pattern detection service
	patternService := services.NewPatternDetectionService(db, taService, hsService, emailService)

	// Initialize handlers
	taHandler := handlers.NewTechnicalAnalysisHandler(db, taService)
	srHandler := handlers.NewSupportResistanceHandler(db, srService)
	setupHandler := handlers.NewSetupHandler(db, setupService)
	emailHandler := handlers.NewEmailHandler(emailService)
	hsHandler := handlers.NewHeadShouldersHandler(db, hsService)

	// Check price data and collect recent historical data if needed
	count, err := db.GetPriceDataCount()
	if err != nil {
		log.Printf("Failed to check price data count: %v", err)
	} else {
		log.Printf("✅ Found %d price data records in database", count)

		if count == 0 {
			log.Printf("📊 No price data found. Collecting 7 days of historical data for symbols: %v", cfg.Collection.Symbols)
			// Collect 7 days of historical data for all symbols
			go func() {
				if err := collectorService.CollectHistoricalData(7); err != nil {
					log.Printf("❌ Failed to collect historical data: %v", err)
				} else {
					log.Printf("✅ Historical data collection completed")
				}
			}()
		} else {
			// Always collect 1 day of recent data on startup to ensure charts work
			log.Printf("📊 Collecting recent 1-day data to ensure chart availability")
			go func() {
				if err := collectorService.CollectHistoricalData(1); err != nil {
					log.Printf("❌ Failed to collect recent data: %v", err)
				} else {
					log.Printf("✅ Recent data collection completed")
				}
			}()
		}
	}

	// Start the collector service for real-time data collection
	if err := collectorService.Start(); err != nil {
		log.Printf("❌ Failed to start collector service: %v", err)
	} else {
		log.Printf("✅ Data collector started (interval: %v)", cfg.Collection.Interval)
	}

	// Start automatic pattern detection service
	patternService.StartPeriodicPatternDetection()
	log.Printf("✅ Automatic pattern detection started")

	// Run initial pattern detection for all watched symbols
	go func() {
		log.Printf("🔍 Running initial pattern detection for all symbols...")
		if err := patternService.AutoDetectPatternsForAllSymbols(); err != nil {
			log.Printf("❌ Initial pattern detection failed: %v", err)
		} else {
			log.Printf("✅ Initial pattern detection completed")
		}
	}()

	// Setup Gin router
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("web/templates/*")

	// Static files
	router.Static("/static", "./web/static")

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
		// API Documentation endpoint
		api.GET("/docs", func(c *gin.Context) {
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
					"🖥️ Interactive TradingView-style dashboard",
				},
				"polygon_integration": gin.H{
					"status":   "active",
					"api_url":  cfg.Polygon.BaseURL,
					"symbols":  cfg.Collection.Symbols,
					"interval": cfg.Collection.Interval.String(),
				},
				"database": gin.H{
					"path":            cfg.Database.Path,
					"max_connections": cfg.Database.MaxOpenConns,
				},
				"ui": gin.H{
					"dashboard": "/",
					"features": []string{
						"Real-time TradingView-style charts with D3.js",
						"Professional dark theme matching TradingView",
						"Technical indicators visualization",
						"Trading setups with quality scores",
						"Support/resistance level analysis",
						"Data collection monitoring",
						"Interactive candlestick charts",
						"Volume analysis with color coding",
					},
				},
				"endpoints": gin.H{
					"health":              "/health",
					"dashboard":           "/",
					"api_docs":            "/api/docs",
					"indicators":          "/api/indicators/:symbol",
					"support_resistance":  "/api/support-resistance/:symbol/levels",
					"setups":              "/api/setups/:symbol",
					"high_quality_setups": "/api/setups/high-quality",
					"setup_detection":     "/api/setups/:symbol/detect",
					"price_chart":         "/api/price/:symbol/chart",
					"collection_status":   "/api/collection/status",
					"force_collection":    "POST /api/collection/force",
					"email_status":        "/api/email/status",
					"email_test":          "POST /api/email/test",
					"email_alert":         "POST /api/email/alert",
				},
				"examples": gin.H{
					"view_dashboard":    "GET /",
					"api_documentation": "GET /api/docs",
					"get_indicators":    "GET /api/indicators/PLTR",
					"detect_setups":     "POST /api/setups/PLTR/detect",
					"get_high_quality":  "GET /api/setups/high-quality",
					"get_sr_levels":     "GET /api/support-resistance/PLTR/levels",
					"get_price_chart":   "GET /api/price/PLTR/chart?range=1W",
					"collection_status": "GET /api/collection/status",
					"force_collection":  "POST /api/collection/force",
					"historical_data":   "POST /api/collection/historical/7",
					"email_status":      "GET /api/email/status",
					"email_test":        "POST /api/email/test",
					"email_alert":       "POST /api/email/alert",
				},
			})
		})

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
			activeSymbols, _ := db.GetWatchedSymbols() // Get actual symbols from database
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
				"active_symbols":      activeSymbols,
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

		// Symbol management endpoints
		volumeHandler := handlers.NewVolumeHandler(db, collectorService, polygonService, patternService)
		api.GET("/symbols", volumeHandler.GetWatchedSymbols)
		api.POST("/symbols", volumeHandler.AddWatchedSymbol)
		api.DELETE("/symbols/:symbol", volumeHandler.RemoveWatchedSymbol)

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
			case "2W":
				from = to.AddDate(0, 0, -14)
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

		// Email notification routes
		api.GET("/email/status", emailHandler.GetEmailStatus)
		api.POST("/email/test", emailHandler.SendTestEmail)
		api.POST("/email/alert", emailHandler.SendTradingAlert)

		// Head and Shoulders Pattern routes
		hs := api.Group("/head-shoulders")
		{
			// General pattern routes
			hs.GET("/patterns", hsHandler.GetAllPatterns)
			hs.POST("/patterns/monitor", hsHandler.MonitorAllPatterns)
			hs.GET("/patterns/stats", hsHandler.GetPatternStatistics)

			// Symbol-based routes
			hs.GET("/symbols/:symbol/patterns", hsHandler.GetPatternsBySymbol)
			hs.POST("/symbols/:symbol/detect", hsHandler.DetectPattern)

			// ID-based routes (using different path structure)
			hs.GET("/pattern/:id/details", hsHandler.GetPatternDetails)
			hs.GET("/pattern/:id/thesis", hsHandler.GetThesisComponents)
			hs.PUT("/pattern/:id/thesis/:component", hsHandler.UpdateThesisComponent)
			hs.GET("/pattern/:id/alerts", hsHandler.GetPatternAlerts)
			hs.GET("/pattern/:id/performance", hsHandler.GetPatternPerformance)
		}
	}

	// Main route - serve the TradingView-style dashboard
	dashboardHandler := handlers.NewDashboardHandler("web/templates", "web/static", db)
	router.GET("/", dashboardHandler.Index)

	// Legacy dashboard route (redirect to main)
	router.GET("/dashboard", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	// Pattern Watcher page
	router.GET("/pattern-watcher", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pattern-watcher.html", gin.H{
			"title": "Pattern Watcher",
		})
	})

	// Legacy route for backward compatibility
	router.GET("/inverse-head-shoulders", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/pattern-watcher")
	})

	// Start server
	port := strconv.Itoa(cfg.Server.Port)
	log.Printf("🚀 Market Watch server starting on port %s", port)
	log.Printf("📊 Phase 3 Complete: Advanced Setup Detection & Scoring")
	log.Printf("📁 Config: configs/config.yaml")
	log.Printf("🔗 Polygon.io: %s", cfg.Polygon.BaseURL)
	log.Printf("📊 Database: %s", cfg.Database.Path)
	log.Printf("📈 Tracked symbols: %v", watchedSymbols)
	log.Printf("⏰ Collection interval: %v", cfg.Collection.Interval)
	log.Printf("🎯 DASHBOARD available at: http://localhost:%s", port)
	log.Printf("📚 API Documentation: http://localhost:%s/api/docs", port)
	log.Printf("✨ Features: TradingView-style UI + Real-time data + Technical Analysis + S/R Detection + Setup Intelligence")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
