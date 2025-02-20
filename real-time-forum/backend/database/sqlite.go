package database

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"
    _ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the database connection and creates tables
func InitDB(dbPath string) error {
    var err error
    
    // Open database connection
    DB, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        return fmt.Errorf("error opening database: %v", err)
    }

    // Test the connection
    err = DB.Ping()
    if err != nil {
        return fmt.Errorf("error connecting to the database: %v", err)
    }

    // Read schema file
    schemaPath := filepath.Join("database", "schema.sql")
    schemaSQL, err := os.ReadFile(schemaPath)
    if err != nil {
        return fmt.Errorf("error reading schema file: %v", err)
    }

    // Execute schema
    _, err = DB.Exec(string(schemaSQL))
    if err != nil {
        return fmt.Errorf("error creating database schema: %v", err)
    }

    log.Println("Database initialized successfully")
    return nil
}

// CloseDB closes the database connection
func CloseDB() error {
    if DB != nil {
        return DB.Close()
    }
    return nil
}

// GetDB returns the database instance
func GetDB() *sql.DB {
    return DB
}