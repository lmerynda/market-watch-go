package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"market-watch-go/internal/config"
	"market-watch-go/internal/database"
	"market-watch-go/internal/handlers"
	"market-watch-go/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Parse command line flags
	var (
		configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
		envFile    = flag.String("env", "", "Path to environment file")
		historical = flag.Int("historical", 0, "Collect historical data for N days (0 = disabled)")
	)
	flag.Parse()

	// Load environment variables from .env file
	if *envFile != "" {
		if err := config.LoadEnvFile(*envFile); err != nil {
			log.Printf("Warning: Failed to load environment file %s: %v", *envFile, err)
		}
	} else {
		// Try to load .env file from current directory
		if err := config.LoadEnvFile(".env"); err != nil {
			log.Printf("Warning: Failed to load .env file: %v", err)
		}
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Printf("Failed to load configuration: %v", err)
		log.Printf("\n" +
			"🔑 SETUP REQUIRED:\n" +
			"Please set your Polygon.io API key in one of these ways:\n" +
			"1. Edit the .env file and replace 'your_polygon_api_key_here' with your actual API key\n" +
			"2. Set environment variable: export POLYGON_API_KEY=your_actual_api_key\n" +
			"3. Edit configs/config.yaml and update the polygon.api_key field\n\n" +
			"💡 Get a free API key at: https://polygon.io/\n" +
			"   1. Sign up for a free account\n" +
			"   2. Go to Dashboard -> API Keys\n" +
			"   3. Copy your API key\n")
		os.Exit(1)
	}

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Polygon service
	polygonService := services.NewPolygonService(cfg)

	// Validate Polygon API key
	if err := polygonService.ValidateAPIKey(); err != nil {
		log.Fatalf("Failed to validate Polygon API key: %v", err)
	}

	// Initialize collector service
	collectorService := services.NewCollectorService(db, polygonService, cfg)

	// Collect historical data if requested
	if *historical > 0 {
		log.Printf("Collecting historical data for %d days...", *historical)
		if err := collectorService.CollectHistoricalData(*historical); err != nil {
			log.Printf("Failed to collect historical data: %v", err)
		} else {
			log.Printf("Historical data collection completed")
		}
	}

	// Start the collector service
	if err := collectorService.Start(); err != nil {
		log.Fatalf("Failed to start collector service: %v", err)
	}
	defer collectorService.Stop()

	// Force initial collection to ensure we have some data
	log.Printf("Triggering initial data collection...")
	if err := collectorService.ForceCollection(); err != nil {
		log.Printf("Warning: Failed to trigger initial collection: %v", err)
	}

	// Initialize handlers
	volumeHandler := handlers.NewVolumeHandler(db, collectorService, polygonService)
	dashboardHandler := handlers.NewDashboardHandler("web/templates", "web/static")
	debugHandler := handlers.NewDebugHandler(db)

	// Set up Gin router
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

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

	// Static files
	router.Static("/static", "web/static")

	// Dashboard routes
	router.GET("/", dashboardHandler.Index)

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

		// Debug endpoints
		debug := api.Group("/debug")
		{
			debug.GET("/count", debugHandler.GetDataCount)
		}
	}

	// Start server
	log.Printf("Starting server on %s", cfg.GetAddress())
	log.Printf("Dashboard available at: http://%s", cfg.GetAddress())
	log.Printf("API available at: http://%s/api", cfg.GetAddress())

	// Start server in a goroutine
	go func() {
		if err := router.Run(cfg.GetAddress()); err != nil {
			log.Fatalf("Failed to start server: %v", err)
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
