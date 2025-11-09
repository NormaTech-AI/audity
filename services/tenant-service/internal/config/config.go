package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	MinIO    MinIOConfig    `mapstructure:"minio"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Crypto   CryptoConfig   `mapstructure:"crypto"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
	Env  string `mapstructure:"env"` // development, staging, production
}

type DatabaseConfig struct {
	TenantDBURL string `mapstructure:"tenant_db_url"`
	MaxConns    int32  `mapstructure:"max_conns"`
	MinConns    int32  `mapstructure:"min_conns"`
	// For provisioning new tenant databases
	PostgresHost     string `mapstructure:"postgres_host"`
	PostgresPort     int    `mapstructure:"postgres_port"`
	PostgresUser     string `mapstructure:"postgres_user"`
	PostgresPassword string `mapstructure:"postgres_password"`
}

type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
}

type AuthConfig struct {
	JWTSecret            string `mapstructure:"jwt_secret"`
	JWTExpirationHours   int    `mapstructure:"jwt_expiration_hours"`
	GoogleClientID       string `mapstructure:"google_client_id"`
	GoogleClientSecret   string `mapstructure:"google_client_secret"`
	MicrosoftClientID    string `mapstructure:"microsoft_client_id"`
	MicrosoftClientSecret string `mapstructure:"microsoft_client_secret"`
	RedirectURL          string `mapstructure:"redirect_url"`
}

type CryptoConfig struct {
	EncryptionKey string `mapstructure:"encryption_key"` // AES-256 key for encrypting DB passwords
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
	viper.SetEnvPrefix("AUDITY")

	// Set defaults
	viper.SetDefault("server.port", "8081")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.env", "development")
	viper.SetDefault("database.max_conns", 25)
	viper.SetDefault("database.min_conns", 5)
	viper.SetDefault("database.postgres_port", 5432)
	viper.SetDefault("minio.use_ssl", false)
	viper.SetDefault("auth.jwt_expiration_hours", 24)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; use defaults and env vars
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
	if c.MinIO.Endpoint == "" {
		return fmt.Errorf("minio.endpoint is required")
	}
	if c.MinIO.AccessKeyID == "" {
		return fmt.Errorf("minio.access_key_id is required")
	}
	if c.MinIO.SecretAccessKey == "" {
		return fmt.Errorf("minio.secret_access_key is required")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required")
	}
	if c.Crypto.EncryptionKey == "" {
		return fmt.Errorf("crypto.encryption_key is required")
	}
	if len(c.Crypto.EncryptionKey) != 32 {
		return fmt.Errorf("crypto.encryption_key must be 32 bytes for AES-256")
	}
	return nil
}
