package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main11() {
	var (
		dbURL = flag.String("db", "", "Database URL")
		dir   = flag.String("dir", "migrations", "Migrations directory")
		cmd   = flag.String("cmd", "up", "Command: up, down, or status")
	)
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
		if *dbURL == "" {
			log.Fatal("Database URL is required")
		}
	}

	db, err := gorm.Open(postgres.Open(*dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create migrations table if it doesn't exist
	sqlDB, _ := db.DB()
	_, err = sqlDB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal("Failed to create migrations table:", err)
	}

	switch *cmd {
	case "up":
		runMigrations(db, *dir, true)
	case "down":
		runMigrations(db, *dir, false)
	case "status":
		showStatus(db, *dir)
	default:
		log.Fatal("Unknown command:", *cmd)
	}
}

func runMigrations(db *gorm.DB, dir string, up bool) {
	files, err := getMigrationFiles(dir, up)
	if err != nil {
		log.Fatal("Failed to read migrations:", err)
	}

	appliedMigrations := getAppliedMigrations(db)

	for _, file := range files {
		version := extractVersion(file)

		if up {
			if _, applied := appliedMigrations[version]; applied {
				fmt.Printf("Migration %s already applied, skipping\n", version)
				continue
			}
		} else {
			if _, applied := appliedMigrations[version]; !applied {
				fmt.Printf("Migration %s not applied, skipping\n", version)
				continue
			}
		}

		content, err := ioutil.ReadFile(filepath.Join(dir, file))
		if err != nil {
			log.Fatal("Failed to read migration file:", err)
		}

		fmt.Printf("Running migration %s\n", file)

		sqlDB, _ := db.DB()
		_, err = sqlDB.Exec(string(content))
		if err != nil {
			log.Fatal("Failed to execute migration:", err)
		}

		// Update migrations table
		if up {
			db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		} else {
			db.Exec("DELETE FROM schema_migrations WHERE version = ?", version)
		}

		fmt.Printf("Migration %s completed\n", file)
	}
}

func getMigrationFiles(dir string, up bool) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var migrations []string
	suffix := ".up.sql"
	if !up {
		suffix = ".down.sql"
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), suffix) {
			migrations = append(migrations, file.Name())
		}
	}

	sort.Strings(migrations)
	if !up {
		// Reverse order for down migrations
		for i := len(migrations)/2 - 1; i >= 0; i-- {
			opp := len(migrations) - 1 - i
			migrations[i], migrations[opp] = migrations[opp], migrations[i]
		}
	}

	return migrations, nil
}

func getAppliedMigrations(db *gorm.DB) map[string]bool {
	var versions []string
	db.Raw("SELECT version FROM schema_migrations").Scan(&versions)

	applied := make(map[string]bool)
	for _, version := range versions {
		applied[version] = true
	}

	return applied
}

func extractVersion(filename string) string {
	parts := strings.Split(filename, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return filename
}

func showStatus(db *gorm.DB, dir string) {
	appliedMigrations := getAppliedMigrations(db)

	upFiles, _ := getMigrationFiles(dir, true)

	fmt.Println("Migration Status:")
	fmt.Println("================")

	for _, file := range upFiles {
		version := extractVersion(file)
		status := "PENDING"
		if _, applied := appliedMigrations[version]; applied {
			status = "APPLIED"
		}

		fmt.Printf("%s: %s\n", version, status)
	}
}
