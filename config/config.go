package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	_"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var DB *sql.DB

func ConnectDB() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err.Error())
	}

	// Example: "username:password@tcp(localhost:3306)/chatx"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", // data source name 
		getEnv("DB_USER", "root"),
		getEnv("DB_PASS", "password"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "chatx"),
	)

	
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Database connection error: %s", err.Error())
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Database ping error: %s", err.Error())
	}

	fmt.Println("Connected to the database successfully.")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}


// GetDB returns the database instance
func GetDB() *sql.DB {
	return DB
}
