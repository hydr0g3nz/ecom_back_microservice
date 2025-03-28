package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the service
type Config struct {
	// Server
	ServerPort     int
	GRPCServerPort int
	
	// MongoDB
	MongoURI      string
	MongoDB       string
	MongoUser     string
	MongoPassword string
	
	// Kafka
	KafkaBrokers []string
	KafkaGroupID string
	
	// JWT
	JWTSecret    string
	JWTExpiresIn time.Duration
	
	// CORS
	CORSAllowedOrigins []string
	
	// Timeouts
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	
	// Pagination
	DefaultPageSize int
	MaxPageSize     int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		// Server
		ServerPort:     getEnvAsInt("SERVER_PORT", 8080),
		GRPCServerPort: getEnvAsInt("GRPC_SERVER_PORT", 9090),
		
		// MongoDB
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:       getEnv("MONGO_DB", "ecom_orders"),
		MongoUser:     getEnv("MONGO_USER", ""),
		MongoPassword: getEnv("MONGO_PASSWORD", ""),
		
		// Kafka
		KafkaBrokers: getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "order-service"),
		
		// JWT
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiresIn: getEnvAsDuration("JWT_EXPIRES_IN", 24*time.Hour),
		
		// CORS
		CORSAllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
		
		// Timeouts
		ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 5*time.Second),
		WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", 120*time.Second),
		
		// Pagination
		DefaultPageSize: getEnvAsInt("DEFAULT_PAGE_SIZE", 10),
		MaxPageSize:     getEnvAsInt("MAX_PAGE_SIZE", 100),
	}
}

// Helper function to get an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Helper function to get an environment variable as an integer
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// Helper function to get an environment variable as a slice
func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}

// Helper function to get an environment variable as a duration
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
