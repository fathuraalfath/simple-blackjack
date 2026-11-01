package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"simple-blackjack/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func LoadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
}

func ConnectDatabase() *gorm.DB {
	if DB != nil {
		return DB
	}
	LoadEnv()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("WARNING: DATABASE_URL not set in environment or .env. Using fallback local connection.")
		dsn = "host=localhost user=postgres password=postgres dbname=simple_blackjack port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// let gorm auto-sync tables
	if err := db.AutoMigrate(&models.User{}, &models.Game{}); err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	fmt.Println("Database connection established and schema auto-migrated.")
	DB = db
	return db
}
