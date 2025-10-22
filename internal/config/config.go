package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig     `yaml:"server"`
	Database  DatabaseConfig   `yaml:"database"`
	Auth      *AuthConfig      `yaml:"auth,omitempty"`
	PagerDuty *PagerDutyConfig `yaml:"pagerduty,omitempty"`
	OpsGenie  *OpsGenieConfig  `yaml:"opsgenie,omitempty"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// AuthConfig holds OIDC authentication configuration
type AuthConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Issuer       string `yaml:"issuer"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURL  string `yaml:"redirect_url"`
	SessionKey   string `yaml:"session_key,omitempty"`
}

// PagerDutyConfig holds PagerDuty API configuration
type PagerDutyConfig struct {
	APIKey string `yaml:"api_key"`
	APIURL string `yaml:"api_url,omitempty"`
}

// OpsGenieConfig holds OpsGenie API configuration
type OpsGenieConfig struct {
	APIKey string `yaml:"api_key"`
	APIURL string `yaml:"api_url,omitempty"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Server.Port)
	}

	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		fmt.Sscanf(dbPort, "%d", &cfg.Database.Port)
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		cfg.Database.User = dbUser
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		cfg.Database.Password = dbPass
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.Database.DBName = dbName
	}

	if pdKey := os.Getenv("PAGERDUTY_API_KEY"); pdKey != "" {
		if cfg.PagerDuty == nil {
			cfg.PagerDuty = &PagerDutyConfig{}
		}
		cfg.PagerDuty.APIKey = pdKey
	}

	if ogKey := os.Getenv("OPSGENIE_API_KEY"); ogKey != "" {
		if cfg.OpsGenie == nil {
			cfg.OpsGenie = &OpsGenieConfig{}
		}
		cfg.OpsGenie.APIKey = ogKey
	}

	// Auth environment variables
	if os.Getenv("AUTH_ENABLED") == "true" {
		if cfg.Auth == nil {
			cfg.Auth = &AuthConfig{}
		}
		cfg.Auth.Enabled = true
	}
	if issuer := os.Getenv("AUTH_ISSUER"); issuer != "" {
		if cfg.Auth == nil {
			cfg.Auth = &AuthConfig{}
		}
		cfg.Auth.Issuer = issuer
	}
	if clientID := os.Getenv("AUTH_CLIENT_ID"); clientID != "" {
		if cfg.Auth == nil {
			cfg.Auth = &AuthConfig{}
		}
		cfg.Auth.ClientID = clientID
	}
	if clientSecret := os.Getenv("AUTH_CLIENT_SECRET"); clientSecret != "" {
		if cfg.Auth == nil {
			cfg.Auth = &AuthConfig{}
		}
		cfg.Auth.ClientSecret = clientSecret
	}
	if redirectURL := os.Getenv("AUTH_REDIRECT_URL"); redirectURL != "" {
		if cfg.Auth == nil {
			cfg.Auth = &AuthConfig{}
		}
		cfg.Auth.RedirectURL = redirectURL
	}
	if sessionKey := os.Getenv("AUTH_SESSION_KEY"); sessionKey != "" {
		if cfg.Auth == nil {
			cfg.Auth = &AuthConfig{}
		}
		cfg.Auth.SessionKey = sessionKey
	}

	return &cfg, nil
}

// Default returns a default configuration
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Host:    "localhost",
			Port:    5432,
			User:    "outalator",
			Password: "outalator",
			DBName:  "outalator",
			SSLMode: "disable",
		},
	}
}
