// backend/database/database.go
package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sql.DB {
	log.Println("Initializing database...")
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)                 // Maximum open connections
	db.SetMaxIdleConns(5)                  // Maximum idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatal("Error pinging database:", err)
	}
	log.Println("Database connection successful")

	log.Println("Creating tables...")
	createTables(db)
	log.Println("Database initialization completed")
	return db
}

func createTables(db *sql.DB) {
	// Users table
	createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        nickname TEXT UNIQUE NOT NULL,
        age INTEGER NOT NULL,
        gender TEXT NOT NULL,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	// Posts table
	createPostsTable := `
    CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        category TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users (id)
    );`

	// Comments table
	createCommentsTable := `
    CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        post_id INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        content TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (post_id) REFERENCES posts (id),
        FOREIGN KEY (user_id) REFERENCES users (id)
    );`

	// Messages table
	createMessagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id INTEGER NOT NULL,
		receiver_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		image_url TEXT DEFAULT '',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (sender_id) REFERENCES users (id),
		FOREIGN KEY (receiver_id) REFERENCES users (id)
	);`

	// Sessions table
	createSessionsTable := `
    CREATE TABLE IF NOT EXISTS sessions (
        id TEXT PRIMARY KEY,
        user_id INTEGER NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users (id)
    );`

	// User profiles table
	createUserProfilesTable := `
    CREATE TABLE IF NOT EXISTS user_profiles (
        user_id INTEGER PRIMARY KEY,
        bio TEXT DEFAULT '',
        avatar TEXT DEFAULT 'default.png',
        last_active TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users (id)
    );`

	// Execute all creation queries
	_, err := db.Exec(createUsersTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createPostsTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createCommentsTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createMessagesTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createSessionsTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createUserProfilesTable)
	if err != nil {
		log.Fatal(err)
	}

	// Create indexes for faster queries
	createMessagesIndex := `
	CREATE INDEX IF NOT EXISTS idx_messages_sender_receiver ON messages (sender_id, receiver_id);
	`
	_, err = db.Exec(createMessagesIndex)
	if err != nil {
		log.Fatal(err)
	}

	createPostsIndex := `
	CREATE INDEX IF NOT EXISTS idx_posts_user ON posts (user_id);
	`
	_, err = db.Exec(createPostsIndex)
	if err != nil {
		log.Fatal(err)
	}

	createCommentsIndex := `
	CREATE INDEX IF NOT EXISTS idx_comments_post ON comments (post_id);
	`
	_, err = db.Exec(createCommentsIndex)
	if err != nil {
		log.Fatal(err)
	}
}
