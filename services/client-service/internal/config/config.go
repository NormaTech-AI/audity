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
    Logging  LoggingConfig  `mapstructure:"logging"`
    Services ServicesConfig `mapstructure:"services"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	PostgresHost     string `mapstructure:"postgres_host"`
	PostgresPort     int    `mapstructure:"postgres_port"`
	PostgresUser     string `mapstructure:"postgres_user"`
	PostgresPassword string `mapstructure:"postgres_password"`
	PostgresDB       string `mapstructure:"postgres_db"`
	MaxConnections   int    `mapstructure:"max_connections"`
	MinConnections   int    `mapstructure:"min_connections"`
}

type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type AuthConfig struct {
	JWTSecret           string `mapstructure:"jwt_secret"`
	JWTExpirationHours  int    `mapstructure:"jwt_expiration_hours"`
}

type LoggingConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"`
}

type ServicesConfig struct {
    TenantBaseURL string `mapstructure:"tenant_base_url"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
	)
}
