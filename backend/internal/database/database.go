package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	}

	var err error

	if dbType == "postgres" {
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
		)
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Printf("Failed to connect to PostgreSQL, falling back to SQLite: %v", err)
			DB, err = gorm.Open(sqlite.Open("script_kill.db"), &gorm.Config{})
			if err != nil {
				log.Fatalf("Failed to connect to SQLite: %v", err)
			}
			log.Println("Connected to SQLite database")
		} else {
			log.Println("Connected to PostgreSQL database")
		}
	} else {
		DB, err = gorm.Open(sqlite.Open("script_kill.db"), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to SQLite: %v", err)
		}
		log.Println("Connected to SQLite database")
	}
}
