package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
	Env  string `mapstructure:"env"`
}

type DatabaseConfig struct {
	TenantDBURL string `mapstructure:"tenant_db_url"`
	MaxConns    int32  `mapstructure:"max_conns"`
	MinConns    int32  `mapstructure:"min_conns"`
}

type AuthConfig struct {
	JWTSecret            string `mapstructure:"jwt_secret"`
	JWTExpirationHours   int    `mapstructure:"jwt_expiration_hours"`
	GoogleClientID       string `mapstructure:"google_client_id"`
	GoogleClientSecret   string `mapstructure:"google_client_secret"`
	MicrosoftClientID    string `mapstructure:"microsoft_client_id"`
	MicrosoftClientSecret string `mapstructure:"microsoft_client_secret"`
	RedirectURL          string `mapstructure:"redirect_url"`
	FrontendURL          string `mapstructure:"frontend_url"`
}

// LoadConfig reads configuration from file or environment variables
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/audity/")

	// Environment variables override
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUDITY_AUTH")

	// Set defaults
	viper.SetDefault("server.port", "8082")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.env", "development")
	viper.SetDefault("database.max_conns", 25)
	viper.SetDefault("database.min_conns", 5)
	viper.SetDefault("auth.jwt_expiration_hours", 24)
	viper.SetDefault("auth.frontend_url", "http://localhost:5173")

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found, using environment variables and defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &config, nil
}

// Validate checks if required configuration values are set
func (c *Config) Validate() error {
	if c.Database.TenantDBURL == "" {
		return fmt.Errorf("database.tenant_db_url is required")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required")
	}
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("auth.jwt_secret must be at least 32 characters")
	}
	return nil
}
