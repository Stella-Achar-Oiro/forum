package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Initialize initializes the database connection and sets up the schema
func Initialize() error {
	// Create database directory if it doesn't exist
	dbDir := filepath.Join(".", "database", "data")
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	dbPath := filepath.Join(dbDir, "forum.db")
	var err error

	// Open database connection
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Execute schema SQL
	schemaPath := filepath.Join(".", "database", "schema.sql")
	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}

	// Execute each statement separately
	statements := strings.Split(string(schemaSQL), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err = DB.Exec(stmt)
		if err != nil {
			// Ignore "table already exists" and "unique constraint" errors
			if !strings.Contains(err.Error(), "already exists") &&
				!strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return fmt.Errorf("failed to execute schema statement: %v", err)
			}
		}
	}

	log.Println("Database initialized successfully")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return DB
}
