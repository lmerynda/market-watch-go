package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Database      DatabaseConfig      `yaml:"database"`
	Polygon       PolygonConfig       `yaml:"polygon"`
	Collection    CollectionConfig    `yaml:"collection"`
	Logging       LoggingConfig       `yaml:"logging"`
	DataRetention DataRetentionConfig `yaml:"data_retention"`
}

type ServerConfig struct {
	Port         int           `yaml:"port"`
	Host         string        `yaml:"host"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	Path            string        `yaml:"path"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type PolygonConfig struct {
	APIKey        string        `yaml:"api_key"`
	BaseURL       string        `yaml:"base_url"`
	Timeout       time.Duration `yaml:"timeout"`
	RetryAttempts int           `yaml:"retry_attempts"`
}

type CollectionConfig struct {
	Interval    time.Duration `yaml:"interval"`
	Symbols     []string      `yaml:"symbols"`
	MarketHours MarketHours   `yaml:"market_hours"`
}

type MarketHours struct {
	Start    string `yaml:"start"`
	End      string `yaml:"end"`
	Timezone string `yaml:"timezone"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type DataRetentionConfig struct {
	Days            int           `yaml:"days"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

// Load reads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	// Default configuration
	cfg := &Config{
		Server: ServerConfig{
			Port:         8080,
			Host:         "localhost",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Path:            "./data/market-watch.db",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Polygon: PolygonConfig{
			BaseURL:       "https://api.polygon.io",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
		},
		Collection: CollectionConfig{
			Interval: 5 * time.Minute,
			Symbols:  []string{"PLTR", "TSLA", "BBAI", "MSFT", "NPWR"},
			MarketHours: MarketHours{
				Start:    "09:30",
				End:      "16:00",
				Timezone: "America/New_York",
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		DataRetention: DataRetentionConfig{
			Days:            30,
			CleanupInterval: 24 * time.Hour,
		},
	}

	// Load from YAML file if it exists
	if configPath != "" {
		if err := loadFromYAML(cfg, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from YAML: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnv(cfg)

	// Validate configuration
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func loadFromYAML(cfg *Config, configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Expand environment variables in YAML
	expanded := os.ExpandEnv(string(data))

	return yaml.Unmarshal([]byte(expanded), cfg)
}

func loadFromEnv(cfg *Config) {
	// Server configuration
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}

	// Database configuration
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}

	// Polygon configuration
	if apiKey := os.Getenv("POLYGON_API_KEY"); apiKey != "" {
		fmt.Printf("Using Polygon API Key from environment variable: %s\n", apiKey)
		cfg.Polygon.APIKey = apiKey
	}

	// Collection configuration
	if interval := os.Getenv("COLLECTION_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			cfg.Collection.Interval = d
		}
	}
	if symbols := os.Getenv("TRACKED_SYMBOLS"); symbols != "" {
		cfg.Collection.Symbols = strings.Split(symbols, ",")
		// Trim whitespace from symbols
		for i, symbol := range cfg.Collection.Symbols {
			cfg.Collection.Symbols[i] = strings.TrimSpace(symbol)
		}
	}

	// Logging configuration
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = logLevel
	}
}

func validate(cfg *Config) error {
	if cfg.Polygon.APIKey == "" || cfg.Polygon.APIKey == "your_polygon_api_key_here" {
		return fmt.Errorf("polygon API key is required. Please set POLYGON_API_KEY environment variable or update the config file. Get your free API key at https://polygon.io/")
	}

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	if len(cfg.Collection.Symbols) == 0 {
		return fmt.Errorf("at least one symbol must be configured for collection")
	}

	if cfg.Collection.Interval < time.Minute {
		return fmt.Errorf("collection interval must be at least 1 minute")
	}

	return nil
}

// GetAddress returns the server address in host:port format
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsMarketHours checks if the current time is within market hours
func (c *Config) IsMarketHours() bool {
	loc, err := time.LoadLocation(c.Collection.MarketHours.Timezone)
	if err != nil {
		// Default to Eastern Time
		loc, _ = time.LoadLocation("America/New_York")
	}

	now := time.Now().In(loc)

	// Skip weekends
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return false
	}

	// Parse market hours
	start, err := time.Parse("15:04", c.Collection.MarketHours.Start)
	if err != nil {
		return false
	}

	end, err := time.Parse("15:04", c.Collection.MarketHours.End)
	if err != nil {
		return false
	}

	// Create times for today
	todayStart := time.Date(now.Year(), now.Month(), now.Day(),
		start.Hour(), start.Minute(), 0, 0, loc)
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(),
		end.Hour(), end.Minute(), 0, 0, loc)

	return now.After(todayStart) && now.Before(todayEnd)
}
