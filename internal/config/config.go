package config

import (
	"fmt"
	"os"
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
	Email         EmailConfig         `yaml:"email"`
	WatchlistDefaults WatchlistDefaults `yaml:"watchlist_defaults"`
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
	Interval              time.Duration `yaml:"interval"`
	DefaultWatchedSymbols []string      `yaml:"default_watched_symbols"`
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

type EmailConfig struct {
	SMTPHost    string `yaml:"smtp_host"`
	SMTPPort    int    `yaml:"smtp_port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	FromName    string `yaml:"from_name"`
	FromAddress string `yaml:"from_address"`
	Enabled     bool   `yaml:"enabled"`
}

type WatchlistDefaults struct {
	Strategies []WatchlistStrategyConfig `yaml:"strategies"`
}

type WatchlistStrategyConfig struct {
	Name   string   `yaml:"name"`
	Color  string   `yaml:"color"`
	Stocks []string `yaml:"stocks"`
}

// Load reads configuration from file only (no environment variable overrides)
func Load(configPath string) (*Config, error) {
	cfg := &Config{}
	if configPath == "" {
		return nil, fmt.Errorf("config file path is required")
	}
	if err := loadFromYAML(cfg, configPath); err != nil {
		return nil, fmt.Errorf("failed to load config from YAML: %w", err)
	}
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

func validate(cfg *Config) error {
	if cfg.Polygon.APIKey == "" || cfg.Polygon.APIKey == "your_polygon_api_key_here" {
		return fmt.Errorf("polygon API key is required. Please set POLYGON_API_KEY environment variable or update the config file. Get your free API key at https://polygon.io/")
	}

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	if len(cfg.Collection.DefaultWatchedSymbols) == 0 {
		return fmt.Errorf("at least one default watched symbol must be configured for collection")
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
