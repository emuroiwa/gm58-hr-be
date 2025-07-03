package main

import (
	"fmt"
	"gm58-hr-backend/internal/config"
)

func main1() {
	fmt.Println("=== GM58-HR Environment Configuration ===")

	cfg := config.Load()

	fmt.Printf("Database URL: %s\n", maskSensitive(cfg.DatabaseURL))
	fmt.Printf("Redis URL: %s\n", cfg.RedisURL)
	fmt.Printf("JWT Secret: %s\n", maskSensitive(cfg.JWTSecret))
	fmt.Printf("Port: %s\n", cfg.Port)
	fmt.Printf("Environment: %s\n", cfg.Environment)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Base Currency: %s\n", cfg.BaseCurrency)
	fmt.Printf("Supported Currencies: %v\n", cfg.SupportedCurrencies)
	fmt.Printf("Currency API URL: %s\n", cfg.CurrencyAPIURL)
	fmt.Printf("Currency API Key: %s\n", maskSensitive(cfg.CurrencyAPIKey))
	fmt.Printf("Storage Path: %s\n", cfg.StoragePath)
	fmt.Printf("Max File Size: %d bytes\n", cfg.MaxFileSize)
	fmt.Printf("Allowed Origins: %v\n", cfg.AllowedOrigins)

	fmt.Println("\n=== SMTP Configuration ===")
	fmt.Printf("SMTP Host: %s\n", cfg.SMTPHost)
	fmt.Printf("SMTP Port: %s\n", cfg.SMTPPort)
	fmt.Printf("SMTP Username: %s\n", cfg.SMTPUsername)
	fmt.Printf("SMTP Password: %s\n", maskSensitive(cfg.SMTPPassword))
}

func maskSensitive(value string) string {
	if len(value) == 0 {
		return "[NOT SET]"
	}
	if len(value) <= 8 {
		return "[HIDDEN]"
	}
	return value[:4] + "..." + value[len(value)-4:]
}
