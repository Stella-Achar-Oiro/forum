package models

import (
	"time"
)

// User represents a forum user
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // "-" means this field won't be included in JSON
	CreatedAt    time.Time `json:"created_at"`
}

// UserInput represents the data needed to create a new user
type UserInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Session represents a user session
type Session struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
