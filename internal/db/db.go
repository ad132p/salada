package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB holds the database connection pool.
var DB *sql.DB

// ConnectDatabase initializes the database connection.
func ConnectDatabase() {
	// Build connection string from environment variables
	pass := os.Getenv("POSTGRES_PASSWORD")
	connStr := fmt.Sprintf("postgresql://salada:%s@salada-db:5432/salada?sslmode=disable", pass)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Ping the database to verify the connection
	if err = DB.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	// Set connection pool settings (optional, but good practice)
	DB.SetMaxOpenConns(25)                 // Max open connections
	DB.SetMaxIdleConns(25)                 // Max idle connections
	DB.SetConnMaxLifetime(5 * time.Minute) // Max lifetime for connections

	fmt.Println("Database connection established!")
}

// CloseDatabase closes the database connection.
func CloseDatabase() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			fmt.Println("Database connection closed.")
		}
	}
}
