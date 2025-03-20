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
	Server    ServerConfig    `yaml:"server"`
	GRPC      GRPCConfig      `yaml:"grpc"`
	Cassandra CassandraConfig `yaml:"cassandra"`
	Kafka     KafkaConfig     `yaml:"kafka"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	IdleTimeout  time.Duration `yaml:"idleTimeout"`
}

// GRPCConfig contains gRPC server configuration
type GRPCConfig struct {
	Port string `yaml:"port"`
}

// CassandraConfig contains Cassandra configuration
type CassandraConfig struct {
	Hosts               []string      `yaml:"hosts"`
	Keyspace            string        `yaml:"keyspace"`
	Username            string        `yaml:"username"`
	Password            string        `yaml:"password"`
	Timeout             time.Duration `yaml:"timeout"`
	ConnectTimeout      time.Duration `yaml:"connectTimeout"`
	ReplicationStrategy string        `yaml:"replicationStrategy"`
	ReplicationFactor   int           `yaml:"replicationFactor"`
}

// KafkaConfig contains Kafka configuration
type KafkaConfig struct {
	Brokers       []string    `yaml:"brokers"`
	ConsumerGroup string      `yaml:"consumerGroup"`
	Topics        KafkaTopics `yaml:"topics"`
}

// KafkaTopics contains Kafka topic configuration
type KafkaTopics struct {
	Orders    string `yaml:"orders"`
	Payments  string `yaml:"payments"`
	Inventory string `yaml:"inventory"`
	Shipping  string `yaml:"shipping"`
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
		GRPC: GRPCConfig{
			Port: "50053", // Different port from other services
		},
		Cassandra: CassandraConfig{
			Hosts:               []string{"localhost"},
			Keyspace:            "order_service",
			Timeout:             5 * time.Second,
			ConnectTimeout:      10 * time.Second,
			ReplicationStrategy: "SimpleStrategy",
			ReplicationFactor:   1,
		},
		Kafka: KafkaConfig{
			Brokers:       []string{"localhost:9092"},
			ConsumerGroup: "order-service",
			Topics: KafkaTopics{
				Orders:    "orders",
				Payments:  "payments",
				Inventory: "inventory",
				Shipping:  "shipping",
			},
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
	if value := os.Getenv("ORDER_SERVER_ADDR"); value != "" {
		config.Server.Address = value
	}

	// GRPC
	if value := os.Getenv("ORDER_GRPC_PORT"); value != "" {
		config.GRPC.Port = value
	}

	// Cassandra
	if value := os.Getenv("ORDER_CASSANDRA_HOSTS"); value != "" {
		config.Cassandra.Hosts = []string{value} // Simple case for single host
	}
	if value := os.Getenv("ORDER_CASSANDRA_KEYSPACE"); value != "" {
		config.Cassandra.Keyspace = value
	}
	if value := os.Getenv("ORDER_CASSANDRA_USERNAME"); value != "" {
		config.Cassandra.Username = value
	}
	if value := os.Getenv("ORDER_CASSANDRA_PASSWORD"); value != "" {
		config.Cassandra.Password = value
	}

	// Kafka
	if value := os.Getenv("ORDER_KAFKA_BROKERS"); value != "" {
		config.Kafka.Brokers = []string{value} // Simple case for single broker
	}
	if value := os.Getenv("ORDER_KAFKA_CONSUMER_GROUP"); value != "" {
		config.Kafka.ConsumerGroup = value
	}

	return config
}
