// backend/models/session.go
package models

import (
    "database/sql"
    "github.com/gofrs/uuid"
    "time"
)

type Session struct {
    ID        string    `json:"id"`
    UserID    int       `json:"userId"`
    CreatedAt time.Time `json:"createdAt"`
}

// CreateSession creates a new session for a user
func CreateSession(db *sql.DB, userID int) (string, error) {
    uuid, err := uuid.NewV4()
    if err != nil {
        return "", err
    }
    
    sessionID := uuid.String()
    
    query := `INSERT INTO sessions (id, user_id) VALUES (?, ?)`
    _, err = db.Exec(query, sessionID, userID)
    if err != nil {
        return "", err
    }
    
    return sessionID, nil
}

// GetSessionByID retrieves a session by its ID
func GetSessionByID(db *sql.DB, sessionID string) (Session, error) {
    var session Session
    query := `SELECT id, user_id, created_at FROM sessions WHERE id = ?`
    
    row := db.QueryRow(query, sessionID)
    err := row.Scan(&session.ID, &session.UserID, &session.CreatedAt)
    
    return session, err
}

// DeleteSession removes a session from the database
func DeleteSession(db *sql.DB, sessionID string) error {
    query := `DELETE FROM sessions WHERE id = ?`
    _, err := db.Exec(query, sessionID)
    return err
}