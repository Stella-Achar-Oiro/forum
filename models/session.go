package models

import (
	"errors"
	"log"
	"time"

	"github.com/Stella-Achar-Oiro/forum/database"
	"github.com/gofrs/uuid"
)

// Session represents a user session
type Session struct {
	ID         string    `json:"id"`
	UserID     int       `json:"userId"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	LastActive time.Time `json:"lastActive"`
}

// CreateSession creates a new session for a user
func CreateSession(userID int, expirationHours int) (Session, error) {
	log.Printf("Creating session for user %d", userID)

	// Generate a UUID for the session
	id, err := uuid.NewV4()
	if err != nil {
		log.Printf("Failed to generate UUID: %v", err)
		return Session{}, err
	}

	// Calculate expiration time
	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(expirationHours) * time.Hour)

	// Format timestamps in ISO 8601 format
	nowStr := now.Format(time.RFC3339)
	expiresStr := expiresAt.Format(time.RFC3339)

	// Create the session
	session := Session{
		ID:         id.String(),
		UserID:     userID,
		CreatedAt:  now,
		ExpiresAt:  expiresAt,
		LastActive: now,
	}

	// Insert into database
	_, err = database.DB.Exec(
		"INSERT INTO sessions (id, user_id, created_at, expires_at, last_active) VALUES (?, ?, ?, ?, ?)",
		session.ID, session.UserID, nowStr, expiresStr, nowStr,
	)
	if err != nil {
		log.Printf("Failed to insert session: %v", err)
		return Session{}, err
	}

	log.Printf("Successfully created session %s for user %d", session.ID, userID)
	return session, nil
}

// GetSessionByID retrieves a session by its ID
func GetSessionByID(sessionID string) (Session, error) {
	log.Printf("Getting session by ID: %s", sessionID)

	var session Session
	var expiresStr string
	var createdStr string
	var lastActiveStr string

	err := database.DB.QueryRow(
		"SELECT id, user_id, created_at, expires_at, last_active FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&session.ID, &session.UserID, &createdStr, &expiresStr, &lastActiveStr)
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		return Session{}, err
	}

	// Try parsing with different time formats
	formats := []string{
		time.RFC3339,                // ISO 8601 format (try this first since we store in this format)
		"2006-01-02 15:04:05",       // SQLite default format
		"2006-01-02T15:04:05Z07:00", // Another common format
	}

	var parseErr error
	for _, format := range formats {
		session.CreatedAt, parseErr = time.Parse(format, createdStr)
		if parseErr == nil {
			break
		}
	}
	if parseErr != nil {
		log.Printf("Failed to parse created_at timestamp: %v", parseErr)
		return Session{}, parseErr
	}

	for _, format := range formats {
		session.ExpiresAt, parseErr = time.Parse(format, expiresStr)
		if parseErr == nil {
			break
		}
	}
	if parseErr != nil {
		log.Printf("Failed to parse expires_at timestamp: %v", parseErr)
		return Session{}, parseErr
	}

	// Parse the last_active timestamp
	for _, format := range formats {
		session.LastActive, parseErr = time.Parse(format, lastActiveStr)
		if parseErr == nil {
			break
		}
	}
	if parseErr != nil {
		log.Printf("Failed to parse last_active timestamp: %v", parseErr)
		// If we can't parse it, just set it to creation time
		session.LastActive = session.CreatedAt
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		log.Printf("Session %s has expired", sessionID)
		// Delete the expired session
		_, _ = database.DB.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
		return Session{}, errors.New("session has expired")
	}

	log.Printf("Successfully retrieved session for user %d", session.UserID)
	return session, nil
}

// UpdateSessionActivity updates the last active timestamp for a session
func UpdateSessionActivity(sessionID string) error {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	// Update the last_active timestamp
	_, err := database.DB.Exec(
		"UPDATE sessions SET last_active = ? WHERE id = ?",
		nowStr, sessionID,
	)

	if err != nil {
		log.Printf("Failed to update session activity: %v", err)
		return err
	}

	log.Printf("Successfully updated activity for session %s", sessionID)
	return nil
}

// ExtendSession extends the expiration time of a session
func ExtendSession(sessionID string, expirationHours int) error {
	newExpiresAt := time.Now().UTC().Add(time.Duration(expirationHours) * time.Hour)
	expiresStr := newExpiresAt.Format(time.RFC3339)

	// Update the expires_at timestamp
	_, err := database.DB.Exec(
		"UPDATE sessions SET expires_at = ? WHERE id = ?",
		expiresStr, sessionID,
	)

	if err != nil {
		log.Printf("Failed to extend session: %v", err)
		return err
	}

	log.Printf("Successfully extended session %s", sessionID)
	return nil
}

// DeleteSession deletes a session
func DeleteSession(sessionID string) error {
	_, err := database.DB.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

// DeleteUserSessions deletes all sessions for a user
func DeleteUserSessions(userID int) error {
	_, err := database.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}

// DeleteExpiredSessions deletes all expired sessions
func DeleteExpiredSessions() error {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := database.DB.Exec("DELETE FROM sessions WHERE expires_at < ?", now)
	if err != nil {
		log.Printf("Failed to delete expired sessions: %v", err)
		return err
	}

	return nil
}

// DeleteInactiveSessions deletes sessions that haven't been active for a certain amount of time
func DeleteInactiveSessions(inactivityMinutes int) error {
	cutoffTime := time.Now().UTC().Add(-time.Duration(inactivityMinutes) * time.Minute).Format(time.RFC3339)

	_, err := database.DB.Exec("DELETE FROM sessions WHERE last_active < ?", cutoffTime)
	if err != nil {
		log.Printf("Failed to delete inactive sessions: %v", err)
		return err
	}

	return nil
}
