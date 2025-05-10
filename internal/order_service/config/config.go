// internal/order_service/config/config.go
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Kafka    KafkaConfig    `yaml:"kafka"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	IdleTimeout  time.Duration `yaml:"idleTimeout"`
}

// DatabaseConfig contains MongoDB database configuration
type DatabaseConfig struct {
	URI         string        `yaml:"uri"`
	Name        string        `yaml:"name"`
	PoolSize    uint64        `yaml:"poolSize"`
	ConnTimeout time.Duration `yaml:"connTimeout"`
}

// GRPCConfig contains gRPC server configuration
type GRPCConfig struct {
	Port string `yaml:"port"`
}

// KafkaConfig contains Kafka configuration
type KafkaConfig struct {
	Brokers string      `yaml:"brokers"`
	GroupID string      `yaml:"groupId"`
	Topics  KafkaTopics `yaml:"topics"`
}

// KafkaTopics contains Kafka topic configuration
type KafkaTopics struct {
	OrderEvents      string `yaml:"orderEvents"`
	InventoryEvents  string `yaml:"inventoryEvents"`
	PaymentEvents    string `yaml:"paymentEvents"`
	InventoryResults string `yaml:"inventoryResults"`
	PaymentResults   string `yaml:"paymentResults"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Set default configuration
	config := &Config{
		Server: ServerConfig{
			Address:      "127.0.0.1:8082", // Different port from other services
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: DatabaseConfig{
			URI:         "mongodb://localhost:27017",
			Name:        "ecom_order_service",
			PoolSize:    100,
			ConnTimeout: 30 * time.Second,
		},
		GRPC: GRPCConfig{
			Port: "50053", // Different port from other services
		},
		Kafka: KafkaConfig{
			Brokers: "localhost:9092",
			GroupID: "order-service",
			Topics: KafkaTopics{
				OrderEvents:      "order-events",
				InventoryEvents:  "inventory-events",
				PaymentEvents:    "payment-events",
				InventoryResults: "inventory-events-result",
				PaymentResults:   "payment-events-result",
			},
		},
	}

	// Read config file
	file, err := os.ReadFile(configPath)
	if err != nil {
		// If no file exists, use defaults
		if os.IsNotExist(err) {
			return config, nil
		}
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
	if value := os.Getenv("ORDER_SERVER_ADDR"); value != "" {
		config.Server.Address = value
	}

	// Database
	if value := os.Getenv("ORDER_DB_URI"); value != "" {
		config.Database.URI = value
	}
	if value := os.Getenv("ORDER_DB_NAME"); value != "" {
		config.Database.Name = value
	}

	// GRPC
	if value := os.Getenv("ORDER_GRPC_PORT"); value != "" {
		config.GRPC.Port = value
	}

	// Kafka
	if value := os.Getenv("KAFKA_BROKERS"); value != "" {
		config.Kafka.Brokers = value
	}
	if value := os.Getenv("KAFKA_GROUP_ID"); value != "" {
		config.Kafka.GroupID = value
	}

	return config
}
