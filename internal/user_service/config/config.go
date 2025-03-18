package config

import (
	"fmt"
	"os"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/pkg/jwt_service"
	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig       `yaml:"server"`
	Database DatabaseConfig     `yaml:"database"`
	JWT      jwt_service.Config `yaml:"jwt"`
	GRPC     GRPCConfig         `yaml:"grpc"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	IdleTimeout  time.Duration `yaml:"idleTimeout"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	User     string        `yaml:"user"`
	Password string        `yaml:"password"`
	Host     string        `yaml:"host"`
	Port     string        `yaml:"port"`
	Name     string        `yaml:"name"`
	MaxIdle  int           `yaml:"maxIdleConnections"`
	MaxOpen  int           `yaml:"maxOpenConnections"`
	MaxLife  time.Duration `yaml:"maxLifetime"`
}

// GRPCConfig contains gRPC server configuration
type GRPCConfig struct {
	Port string `yaml:"port"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Set default configuration
	config := &Config{
		Server: ServerConfig{
			Address:      "127.0.0.1:8080",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: DatabaseConfig{
			User:     "root",
			Password: "pass",
			Host:     "localhost",
			Port:     "3366",
			Name:     "ecom_user_service",
			MaxIdle:  25,
			MaxOpen:  25,
			MaxLife:  5 * time.Minute,
		},
		JWT: jwt_service.Config{
			SecretKey:            "secret_key",
			Issuer:               "user_service",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
		},
		GRPC: GRPCConfig{
			Port: "50051",
		},
	}

	// Read config file
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse YAML
	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Override with environment variables if they exist
	config = overrideWithEnv(config)

	return config, nil
}

// overrideWithEnv overrides config with environment variables
func overrideWithEnv(config *Config) *Config {
	// Server
	if value := os.Getenv("SERVER_ADDR"); value != "" {
		config.Server.Address = value
	}

	// Database
	if value := os.Getenv("DB_USER"); value != "" {
		config.Database.User = value
	}
	if value := os.Getenv("DB_PASSWORD"); value != "" {
		config.Database.Password = value
	}
	if value := os.Getenv("DB_HOST"); value != "" {
		config.Database.Host = value
	}
	if value := os.Getenv("DB_PORT"); value != "" {
		config.Database.Port = value
	}
	if value := os.Getenv("DB_NAME"); value != "" {
		config.Database.Name = value
	}

	// JWT
	if value := os.Getenv("JWT_SECRET_KEY"); value != "" {
		config.JWT.SecretKey = value
	}

	// GRPC
	if value := os.Getenv("GRPC_PORT"); value != "" {
		config.GRPC.Port = value
	}

	return config
}
