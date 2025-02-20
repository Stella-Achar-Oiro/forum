package models

import (
    "database/sql"
    "errors"
    "fmt"
    "time"
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    ID          int64     `json:"id"`
    Nickname    string    `json:"nickname"`
    Age         int       `json:"age"`
    Gender      string    `json:"gender"`
    FirstName   string    `json:"first_name"`
    LastName    string    `json:"last_name"`
    Email       string    `json:"email"`
    Password    string    `json:"-"` // Never sent to client
    CreatedAt   time.Time `json:"created_at"`
    LastSeen    time.Time `json:"last_seen"`
}

type UserRegistration struct {
    Nickname    string `json:"nickname"`
    Age         int    `json:"age"`
    Gender      string `json:"gender"`
    FirstName   string `json:"first_name"`
    LastName    string `json:"last_name"`
    Email       string `json:"email"`
    Password    string `json:"password"`
}

type UserLogin struct {
    Identity string `json:"identity"` // Can be either email or nickname
    Password string `json:"password"`
}

var (
    ErrUserNotFound      = errors.New("user not found")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUserExists        = errors.New("user already exists")
    ErrInvalidInput      = errors.New("invalid input")
)

// CreateUser creates a new user in the database
func CreateUser(db *sql.DB, reg *UserRegistration) (*User, error) {
    // Input validation
    if err := validateRegistration(reg); err != nil {
        return nil, fmt.Errorf("validation error: %w", err)
    }

    // Check if user exists
    exists, err := userExists(db, reg.Email, reg.Nickname)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, ErrUserExists
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("error hashing password: %w", err)
    }

    // Insert user
    query := `
        INSERT INTO users (nickname, age, gender, first_name, last_name, email, password_hash)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `
    result, err := db.Exec(query,
        reg.Nickname,
        reg.Age,
        reg.Gender,
        reg.FirstName,
        reg.LastName,
        reg.Email,
        string(hashedPassword),
    )
    if err != nil {
        return nil, fmt.Errorf("error creating user: %w", err)
    }

    id, err := result.LastInsertId()
    if err != nil {
        return nil, fmt.Errorf("error getting user id: %w", err)
    }

    // Return created user
    return GetUserByID(db, id)
}

// AuthenticateUser verifies user credentials and returns the user if valid
func AuthenticateUser(db *sql.DB, login *UserLogin) (*User, error) {
    query := `
        SELECT id, nickname, age, gender, first_name, last_name, email, password_hash, created_at, last_seen
        FROM users
        WHERE email = ? OR nickname = ?
    `
    
    var user User
    var passwordHash string
    
    err := db.QueryRow(query, login.Identity, login.Identity).Scan(
        &user.ID,
        &user.Nickname,
        &user.Age,
        &user.Gender,
        &user.FirstName,
        &user.LastName,
        &user.Email,
        &passwordHash,
        &user.CreatedAt,
        &user.LastSeen,
    )

    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("error querying user: %w", err)
    }

    // Verify password
    err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(login.Password))
    if err != nil {
        return nil, ErrInvalidCredentials
    }

    // Update last seen
    err = updateLastSeen(db, user.ID)
    if err != nil {
        // Log error but don't fail the authentication
        fmt.Printf("Error updating last seen: %v\n", err)
    }

    return &user, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(db *sql.DB, id int64) (*User, error) {
    var user User
    query := `
        SELECT id, nickname, age, gender, first_name, last_name, email, created_at, last_seen
        FROM users
        WHERE id = ?
    `
    err := db.QueryRow(query, id).Scan(
        &user.ID,
        &user.Nickname,
        &user.Age,
        &user.Gender,
        &user.FirstName,
        &user.LastName,
        &user.Email,
        &user.CreatedAt,
        &user.LastSeen,
    )

    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("error querying user: %w", err)
    }

    return &user, nil
}

// Helper functions

func validateRegistration(reg *UserRegistration) error {
    if reg.Nickname == "" || reg.Email == "" || reg.Password == "" ||
       reg.FirstName == "" || reg.LastName == "" || reg.Gender == "" {
        return ErrInvalidInput
    }
    if reg.Age < 13 { // Example age restriction
        return ErrInvalidInput
    }
    // Add more validation as needed
    return nil
}

func userExists(db *sql.DB, email, nickname string) (bool, error) {
    var exists bool
    query := `
        SELECT EXISTS(
            SELECT 1 FROM users 
            WHERE email = ? OR nickname = ?
        )
    `
    err := db.QueryRow(query, email, nickname).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("error checking user existence: %w", err)
    }
    return exists, nil
}

func updateLastSeen(db *sql.DB, userID int64) error {
    query := "UPDATE users SET last_seen = CURRENT_TIMESTAMP WHERE id = ?"
    _, err := db.Exec(query, userID)
    return err
}