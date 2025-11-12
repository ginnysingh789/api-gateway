package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server         ServerConfig
	JWT            JWTConfig
	MongoDB        MongoDBConfig
	Redis          RedisConfig
	RateLimit      RateLimitConfig
	CircuitBreaker CircuitBreakerConfig
	Timeouts       TimeoutsConfig
	CORS           CORSConfig
	Logging        LoggingConfig
	Services       []ServiceConfig
}

type ServerConfig struct {
	Port        int
	Environment string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

type CircuitBreakerConfig struct {
	Threshold int
	Timeout   time.Duration
}

type TimeoutsConfig struct {
	Read  int
	Write int
	Idle  int
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type LoggingConfig struct {
	Level string
}

type ServiceConfig struct {
	Name      string   `yaml:"name"`
	URLs      []string `yaml:"urls"`
	HealthURL string   `yaml:"health_url"`
}

func LoadConfig() (*Config, error) {
	// Try to load from config file first
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath(".")

	// Read config file (optional)
	viper.ReadInConfig()

	// Environment variables take precedence
	viper.AutomaticEnv()

	config := &Config{
		Server: ServerConfig{
			Port:        getEnvAsInt("PORT", 8080),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			Expiry: parseDuration(getEnv("JWT_EXPIRY", "24h")),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGO_DATABASE", "api_gateway"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		RateLimit: RateLimitConfig{
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			Window:   time.Duration(getEnvAsInt("RATE_LIMIT_WINDOW", 60)) * time.Second,
		},
		CircuitBreaker: CircuitBreakerConfig{
			Threshold: getEnvAsInt("CIRCUIT_BREAKER_THRESHOLD", 5),
			Timeout:   time.Duration(getEnvAsInt("CIRCUIT_BREAKER_TIMEOUT", 30)) * time.Second,
		},
		Timeouts: TimeoutsConfig{
			Read:  getEnvAsInt("READ_TIMEOUT", 15),
			Write: getEnvAsInt("WRITE_TIMEOUT", 15),
			Idle:  getEnvAsInt("IDLE_TIMEOUT", 60),
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{getEnv("CORS_ALLOWED_ORIGINS", "*")},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	// Load services from config file if available
	if err := viper.UnmarshalKey("services", &config.Services); err == nil {
		fmt.Println("Loaded services from config file")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
