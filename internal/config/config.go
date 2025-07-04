package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	RedisURL       string
	RedisPassword  string
	RedisDB        int
	JWTSecret      string
	Port           string
	Environment    string
	LogLevel       string
	AllowedOrigins []string

	// Currency API configuration
	CurrencyAPIKey      string
	CurrencyAPIURL      string
	BaseCurrency        string
	SupportedCurrencies []string

	// SMTP configuration for email notifications
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string

	// File storage
	StoragePath string
	MaxFileSize int64
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading .env file, using system environment variables")
	}

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "10485760"), 10, 64) // 10MB default

	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost/gm58_hr?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        redisDB,
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080,http://localhost:5173"), ","),

		CurrencyAPIKey:      getEnv("CURRENCY_API_KEY", ""),
		CurrencyAPIURL:      getEnv("CURRENCY_API_URL", "https://api.exchangerate-api.com/v4/latest/"),
		BaseCurrency:        getEnv("BASE_CURRENCY", "USD"),
		SupportedCurrencies: strings.Split(getEnv("SUPPORTED_CURRENCIES", "USD,ZWL,ZAR,GBP,EUR"), ","),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),

		StoragePath: getEnv("STORAGE_PATH", "./storage"),
		MaxFileSize: maxFileSize,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
