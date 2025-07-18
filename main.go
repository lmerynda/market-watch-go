package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"market-watch-go/internal/config"
	"market-watch-go/internal/database"
	"market-watch-go/internal/handlers"
	"market-watch-go/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Improve log output: add timestamp and file info
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("[STARTUP] Logging to stderr (default for Go log package). If running in Docker or a dev container, check container logs or VS Code Output panel.")

	// Parse command line flags
	var (
		configPath     = flag.String("config", "configs/config.yaml", "Path to configuration file")
		historical     = flag.Int("historical", 0, "Collect historical data for N days (0 = disabled)")
		resetWatchlist = flag.Bool("reset-watchlist", false, "Reset watchlist to config defaults")
	)

	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Printf("Failed to load configuration: %v", err)
		log.Printf("\n" +
			"🔑 SETUP REQUIRED:\n" +
			"Please set your Polygon.io API key in configs/config.yaml under polygon.api_key\n" +
			"💡 Get a free API key at: https://polygon.io/\n" +
			"   1. Sign up for a free account\n" +
			"   2. Go to Dashboard -> API Keys\n" +
			"   3. Copy your API key\n")
		os.Exit(1)
	}

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Load configuration
	cfg, err = config.Load(*configPath)
	if err != nil {
		log.Printf("Failed to load configuration: %v", err)
		log.Printf("\n" +
			"🔑 SETUP REQUIRED:\n" +
			"Please set your Polygon.io API key in configs/config.yaml under polygon.api_key\n" +
			"💡 Get a free API key at: https://polygon.io/\n" +
			"   1. Sign up for a free account\n" +
			"   2. Go to Dashboard -> API Keys\n" +
			"   3. Copy your API key\n")
		os.Exit(1)
	}

	// Initialize database
	db, err = database.New(cfg)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Ensure default watched symbols are present if watchlist is empty
	watched, err := db.GetWatchedSymbols()
	if err != nil {
		log.Printf("Failed to check watched symbols: %v", err)
		os.Exit(1)
	}
	if len(watched) == 0 && len(cfg.Collection.DefaultWatchedSymbols) > 0 {
		log.Printf("No watched symbols found in DB. Adding default watched symbols from config: %v", cfg.Collection.DefaultWatchedSymbols)
		err := db.EnsureConfigSymbolsWatched(cfg.Collection.DefaultWatchedSymbols)
		if err != nil {
			log.Printf("Failed to add default watched symbols: %v", err)
			os.Exit(1)
		}
	}

	// Ensure strategies and stocks are properly associated in the database if resetWatchlist is set
	if *resetWatchlist {
		log.Printf("Resetting watchlist to config defaults...")
		if len(cfg.WatchlistDefaults.Strategies) > 0 {
			err := db.InitializeData(cfg.WatchlistDefaults.Strategies)
			if err != nil {
				log.Printf("Failed to reset watchlist strategies: %v", err)
				os.Exit(1)
			}
		}
	} else {
		// Only apply defaults if database is empty
		strategies, err := db.GetStrategies()
		if err != nil {
			log.Printf("Failed to check existing strategies: %v", err)
			os.Exit(1)
		}
		if len(strategies) == 0 && len(cfg.WatchlistDefaults.Strategies) > 0 {
			log.Printf("No strategies found in DB. Adding default strategies from config.")
			err := db.InitializeData(cfg.WatchlistDefaults.Strategies)
			if err != nil {
				log.Printf("Failed to ensure default strategies: %v", err)
				os.Exit(1)
			}
		}
	}

	// Initialize Polygon service
	polygonService := services.NewPolygonService(cfg)

	// Validate Polygon API key
	if err := polygonService.ValidateAPIKey(); err != nil {
		log.Printf("Failed to validate Polygon API key: %v", err)
		os.Exit(1)
	}

	// Initialize collector service
	collectorService := services.NewCollectorService(db, polygonService, cfg)

	// Collect historical data if requested, or default minimum for dashboard functionality
	historicalDays := *historical
	if historicalDays == 0 {
		// Default to 30 days to support all dashboard time ranges (1D, 1W, 2W, 1M)
		historicalDays = 30
		log.Printf("Auto-collecting 30 days of historical data to support all dashboard time ranges...")
	} else {
		log.Printf("Collecting historical data for %d days...", historicalDays)
	}

	if err := collectorService.CollectHistoricalData(historicalDays); err != nil {
		log.Printf("Failed to collect historical data: %v", err)
	} else {
		log.Printf("Historical data collection completed")
	}

	// Start the collector service
	if err := collectorService.Start(); err != nil {
		log.Printf("Failed to start collector service: %v", err)
		os.Exit(1)
	}
	defer collectorService.Stop()

	// Force initial collection to ensure we have some data
	log.Printf("Triggering initial data collection...")
	if err := collectorService.ForceCollection(); err != nil {
		log.Printf("Warning: Failed to trigger initial collection: %v", err)
	}

	// Initialize services
	taService := services.NewTechnicalAnalysisService(db, nil)
	srService := services.NewSupportResistanceService(db, taService)
	setupService := services.NewSetupDetectionService(db, taService, srService)

	// Initialize email service
	emailService := services.NewEmailService(cfg)

	// Initialize pattern detection services
	fallingWedgeService := services.NewFallingWedgeDetectionService(db, taService, emailService)
	hsService := services.NewHeadShouldersDetectionService(db, setupService, taService, emailService)
	patternService := services.NewPatternDetectionService(db, taService, hsService, emailService)

	// Initialize Polygon EMA service
	emaService := services.NewPolygonEMAService(cfg.Polygon.APIKey)

	// Initialize stock service
	stockService := services.NewStockService(db, polygonService, emaService)

	// Initialize handlers
	volumeHandler := handlers.NewVolumeHandler(db, collectorService, polygonService, patternService)
	priceHandler := handlers.NewPriceHandler(db, collectorService, polygonService)
	dashboardHandler := handlers.NewDashboardHandler("web/templates", "web/static", db)
	debugHandler := handlers.NewDebugHandler(db)
	taHandler := handlers.NewTechnicalAnalysisHandler(db, taService)
	setupHandler := handlers.NewSetupHandler(db, setupService)
	srHandler := handlers.NewSupportResistanceHandler(db, srService)
	fallingWedgeHandler := handlers.NewFallingWedgeHandler(db, fallingWedgeService)
	patternsHandler := handlers.NewPatternsHandler(db, patternService, hsService, fallingWedgeService)
	watchlistHandler := handlers.NewWatchlistHandler(db, stockService)

	// Polygon EMA handlers
	polygonEMAHandler := handlers.PolygonEMAHandler(emaService)
	polygonEMABatchHandler := handlers.PolygonEMABatchHandler(emaService)

	// Set up Gin router
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	// Global error logging middleware
	router.Use(func(c *gin.Context) {
		c.Next()
		for _, ginErr := range c.Errors {
			log.Printf("[GIN ERROR] %s %s | %v", c.Request.Method, c.Request.URL.Path, ginErr.Err)
		}
	})

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Load HTML templates
	router.LoadHTMLGlob("web/templates/*")

	// Add middleware to disable caching for static files in development
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})

	// Static files with no-cache headers for development
	router.Static("/static", "web/static")

	// Dashboard routes
	router.GET("/", dashboardHandler.Index)

	// Pattern Watcher page
	router.GET("/pattern-watcher", func(c *gin.Context) {
		c.HTML(200, "pattern-watcher.html", gin.H{
			"title": "Pattern Watcher",
		})
	})

	// Watchlist page
	router.GET("/watchlist", watchlistHandler.RenderWatchlistPage)

	// API routes
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", volumeHandler.HealthCheck)

		// Volume data endpoints
		volume := api.Group("/volume")
		{
			volume.GET("/:symbol", volumeHandler.GetVolumeData)
			volume.GET("/:symbol/latest", volumeHandler.GetLatestVolumeData)
			volume.GET("/:symbol/chart", volumeHandler.GetChartData)
		}

		// Price data endpoints for TradingView
		price := api.Group("/price")
		{
			price.GET("/:symbol", priceHandler.GetPriceData)
			price.GET("/:symbol/latest", priceHandler.GetLatestPriceData)
			price.GET("/:symbol/chart", priceHandler.GetPriceChartData)
			price.GET("/:symbol/stats", priceHandler.GetPriceStats)
		}

		// Dashboard endpoints
		dashboard := api.Group("/dashboard")
		{
			dashboard.GET("/summary", volumeHandler.GetDashboardSummary)
		}

		// Collection management endpoints
		collection := api.Group("/collection")
		{
			collection.GET("/status", volumeHandler.GetCollectionStatus)
			collection.POST("/force", volumeHandler.ForceCollection)
		}

		// Symbol management endpoints
		api.GET("/symbols", volumeHandler.GetWatchedSymbols)
		api.POST("/symbols", volumeHandler.AddWatchedSymbol)
		api.DELETE("/symbols/:symbol", volumeHandler.RemoveWatchedSymbol)
		api.GET("/symbols/:symbol/check", volumeHandler.CheckSymbolData)
		api.POST("/symbols/:symbol/collect", volumeHandler.CollectSymbolData)

		// Debug endpoints
		debug := api.Group("/debug")
		{
			debug.GET("/count", debugHandler.GetDataCount)
		}

		// Technical Analysis endpoints
		indicators := api.Group("/indicators")
		{
			indicators.GET("/:symbol", taHandler.GetIndicators)
			indicators.GET("/:symbol/summary", taHandler.GetIndicatorsSummary)
			indicators.GET("/:symbol/historical", taHandler.GetHistoricalIndicators)
			indicators.POST("/:symbol/update", taHandler.UpdateIndicators)
			indicators.GET("/:symbol/alerts", taHandler.CheckAlerts)
			indicators.GET("/:symbol/alerts/active", taHandler.GetActiveAlerts)
			indicators.POST("/:symbol/cache/invalidate", taHandler.InvalidateSymbolCache)
		}

		technicalAnalysis := api.Group("/technical-analysis")
		{
			technicalAnalysis.GET("/indicators", taHandler.GetMultipleIndicators)
			technicalAnalysis.GET("/stats", taHandler.GetStats)
			technicalAnalysis.GET("/cache/status", taHandler.GetCacheStatus)
			technicalAnalysis.POST("/cache/clear", taHandler.ClearCache)
		}

		// Setup Detection endpoints
		setups := api.Group("/setups")
		{
			setups.GET("/high-quality", setupHandler.GetHighQualitySetups)
			setups.GET("/", setupHandler.GetMultipleSetups)
			setups.POST("/expire", setupHandler.ExpireOldSetups)
			setups.POST("/cleanup", setupHandler.CleanupOldSetups)
			setups.GET("/stats", setupHandler.GetSetupsStats)
			setups.GET("/:symbol", setupHandler.GetSetups)
			setups.POST("/:symbol/detect", setupHandler.DetectSetups)
			setups.GET("/:symbol/summary", setupHandler.GetSetupSummary)
			setups.GET("/id/:id", setupHandler.GetSetupByID)
			setups.PUT("/id/:id/status", setupHandler.UpdateSetupStatus)
			setups.GET("/id/:id/checklist", setupHandler.GetSetupChecklist)
		}

		// Support/Resistance endpoints
		supportResistance := api.Group("/support-resistance")
		{
			supportResistance.GET("/levels", srHandler.GetMultipleLevels)
			supportResistance.POST("/cleanup", srHandler.CleanupOldData)
			supportResistance.POST("/deactivate", srHandler.DeactivateOldLevels)
			supportResistance.GET("/:symbol/levels", srHandler.GetSupportResistanceLevels)
			supportResistance.POST("/:symbol/detect", srHandler.DetectSupportResistance)
			supportResistance.GET("/:symbol/nearest", srHandler.GetNearestLevels)
			supportResistance.GET("/:symbol/touches", srHandler.GetLevelTouches)
			supportResistance.GET("/:symbol/pivots", srHandler.GetPivotPoints)
			supportResistance.GET("/:symbol/summary", srHandler.GetLevelSummary)
		}

		// Unified Patterns API - handles all pattern types
		patterns := api.Group("/patterns")
		{
			patterns.POST("/scan", patternsHandler.ScanAllPatterns)
			patterns.POST("/scan/:symbol", patternsHandler.ScanSymbolPatterns)
			patterns.GET("/", patternsHandler.GetAllPatterns)
			patterns.GET("/:symbol", patternsHandler.GetPatternsBySymbol)
			patterns.GET("/stats", patternsHandler.GetPatternStatistics)
		}

		// Head & Shoulders Pattern routes removed - use unified /api/patterns/ instead

		// Falling Wedge Pattern routes (legacy compatibility)
		fw := api.Group("/falling-wedge")
		{
			fw.GET("/patterns", fallingWedgeHandler.GetPatterns)
			fw.GET("/patterns/active", fallingWedgeHandler.GetActivePatterns)
			fw.GET("/patterns/stats", fallingWedgeHandler.GetPatternStatistics)
			fw.GET("/symbols/:symbol/patterns", fallingWedgeHandler.GetPatternsBySymbol)
			fw.POST("/symbols/:symbol/detect", fallingWedgeHandler.DetectPattern)
			fw.GET("/patterns/:id", fallingWedgeHandler.GetPatternDetails)
			fw.POST("/patterns/scan", fallingWedgeHandler.ScanPatterns)
		}

		// Watchlist routes
		watchlist := api.Group("/watchlist")
		{
			// Strategies
			watchlist.GET("/strategies", watchlistHandler.GetStrategies)
			watchlist.POST("/strategies", watchlistHandler.CreateStrategy)
			watchlist.PUT("/strategies/:id", watchlistHandler.UpdateStrategy)
			watchlist.DELETE("/strategies/:id", watchlistHandler.DeleteStrategy)

			// Backward compatibility routes (categories -> strategies)
			watchlist.GET("/categories", watchlistHandler.GetStrategies)
			watchlist.POST("/categories", watchlistHandler.CreateStrategy)
			watchlist.PUT("/categories/:id", watchlistHandler.UpdateStrategy)
			watchlist.DELETE("/categories/:id", watchlistHandler.DeleteStrategy)

			// Stocks
			watchlist.GET("/stocks", watchlistHandler.GetStocks)
			watchlist.POST("/stocks", watchlistHandler.AddStock)
			watchlist.PUT("/stocks/:id", watchlistHandler.UpdateStock)
			watchlist.DELETE("/stocks/:id", watchlistHandler.RemoveStock)

			// Refresh prices and EMAs for all stocks
			watchlist.POST("/refresh", handlers.WatchlistRefreshHandler(db, stockService))
		}

		// Polygon EMA endpoints
		api.GET("/polygon/ema", polygonEMAHandler)
		api.GET("/polygon/ema/batch", polygonEMABatchHandler)
	}

	// Start server
	log.Printf("Starting server on %s", cfg.GetAddress())
	log.Printf("Dashboard available at: http://%s", cfg.GetAddress())
	log.Printf("Pattern Watcher available at: http://%s/pattern-watcher", cfg.GetAddress())
	log.Printf("Unified Patterns API available at: http://%s/api/patterns/", cfg.GetAddress())
	log.Printf("API available at: http://%s/api", cfg.GetAddress())

	// Start server in a goroutine
	go func() {
		if err := router.Run(cfg.GetAddress()); err != nil {
			log.Printf("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down server...")

	// Stop collector service
	collectorService.Stop()

	log.Printf("Server shutdown complete")
}
