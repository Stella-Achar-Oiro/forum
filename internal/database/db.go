package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbName = "forum.db"
)

var db *sql.DB

// InitDB initializes the database connection and creates tables
func InitDB() (*sql.DB, error) {
	if db != nil {
		return db, nil
	}

	// Create data directory if it doesn't exist
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	dbPath := filepath.Join(dataDir, dbName)
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Initialize schema
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %v", err)
	}

	log.Println("Database connection established successfully")
	return db, nil
}

// initSchema reads and executes the schema.sql file
func initSchema(db *sql.DB) error {
	schemaSQL, err := os.ReadFile("internal/database/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}

	_, err = db.Exec(string(schemaSQL))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %v", err)
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return db
}

// CloseDB closes the database connection
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
