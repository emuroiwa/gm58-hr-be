package main

import (
	"fmt"
	"gm58-hr-backend/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	fmt.Printf("Testing database connection to: %s\n", maskURL(cfg.DatabaseURL))

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("❌ Failed to get SQL DB instance: %v\n", err)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		fmt.Printf("❌ Failed to ping database: %v\n", err)
		return
	}

	fmt.Println("✅ Database connection successful!")
}

func maskURL(url string) string {
	if len(url) <= 20 {
		return "[DATABASE_URL]"
	}
	return url[:10] + "..." + url[len(url)-10:]
}
