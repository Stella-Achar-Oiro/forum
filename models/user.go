package models

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/Stella-Achar-Oiro/forum/database"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Not exported in JSON
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Age       int       `json:"age"`
	Gender    string    `json:"gender"`
	CreatedAt time.Time `json:"createdAt"`
	LastLogin time.Time `json:"lastLogin"`
	IsOnline  bool      `json:"isOnline"`
}

// UserForPublic represents a user with limited information for public display
type UserForPublic struct {
	ID        int    `json:"id"`
	Nickname  string `json:"nickname"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	IsOnline  bool   `json:"isOnline"`
}

// CreateUser creates a new user in the database
func CreateUser(user User, plaintextPassword string) (int, error) {
	log.Printf("Creating user with nickname: %s", user.Nickname)

	// Check if user already exists
	var id int
	err := database.DB.QueryRow("SELECT id FROM users WHERE nickname = ? OR email = ?", user.Nickname, user.Email).Scan(&id)
	if err == nil {
		log.Printf("User already exists with nickname or email")
		return 0, errors.New("user with this nickname or email already exists")
	} else if err != sql.ErrNoRows {
		log.Printf("Database error checking existing user: %v", err)
		return 0, err
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return 0, err
	}

	// Format current time in ISO 8601 format
	now := time.Now().UTC().Format(time.RFC3339)

	// Insert new user
	result, err := database.DB.Exec(
		"INSERT INTO users (nickname, email, password, first_name, last_name, age, gender, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		user.Nickname, user.Email, string(hashedPassword), user.FirstName, user.LastName, user.Age, user.Gender, now,
	)
	if err != nil {
		log.Printf("Failed to insert user: %v", err)
		return 0, err
	}

	// Get the newly created user ID
	userID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v", err)
		return 0, err
	}

	log.Printf("Successfully created user with ID: %d", userID)
	return int(userID), nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(id int) (User, error) {
	log.Printf("Getting user by ID: %d", id)

	var user User
	var createdAtStr, lastLoginStr sql.NullString

	err := database.DB.QueryRow(
		"SELECT id, nickname, email, password, first_name, last_name, age, gender, created_at, last_login, is_online FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Nickname, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Age, &user.Gender, &createdAtStr, &lastLoginStr, &user.IsOnline)

	if err != nil {
		log.Printf("Failed to get user by ID: %v", err)
		return User{}, err
	}

	// Set default timestamps
	user.CreatedAt = time.Now()  // Default to current time if not set
	user.LastLogin = time.Time{} // Default to zero time if not set

	// Parse created_at timestamp if present
	if createdAtStr.Valid && createdAtStr.String != "" {
		// Try parsing with different time formats
		formats := []string{
			time.RFC3339,                // ISO 8601 format
			"2006-01-02 15:04:05",       // SQLite default format
			"2006-01-02T15:04:05Z07:00", // Another common format
		}

		var parseErr error
		for _, format := range formats {
			user.CreatedAt, parseErr = time.Parse(format, createdAtStr.String)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			log.Printf("Warning: Failed to parse created_at timestamp: %v", parseErr)
		}
	}

	// Parse last_login timestamp if present
	if lastLoginStr.Valid && lastLoginStr.String != "" {
		// Try parsing with different time formats
		formats := []string{
			time.RFC3339,                // ISO 8601 format
			"2006-01-02 15:04:05",       // SQLite default format
			"2006-01-02T15:04:05Z07:00", // Another common format
		}

		var parseErr error
		for _, format := range formats {
			user.LastLogin, parseErr = time.Parse(format, lastLoginStr.String)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			log.Printf("Warning: Failed to parse last_login timestamp: %v", parseErr)
		}
	}

	log.Printf("Successfully retrieved user: %s", user.Nickname)
	return user, nil
}

// GetUserByCredentials retrieves a user by their nickname or email and verifies the password
func GetUserByCredentials(identifier, password string) (User, error) {
	var user User
	var createdAtStr, lastLoginStr sql.NullString

	err := database.DB.QueryRow(
		"SELECT id, nickname, email, password, first_name, last_name, age, gender, created_at, last_login, is_online FROM users WHERE nickname = ? OR email = ?",
		identifier, identifier,
	).Scan(&user.ID, &user.Nickname, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Age, &user.Gender, &createdAtStr, &lastLoginStr, &user.IsOnline)
	if err != nil {
		return User{}, err
	}

	// Set default timestamps
	user.CreatedAt = time.Now()  // Default to current time if not set
	user.LastLogin = time.Time{} // Default to zero time if not set

	// Parse created_at timestamp if present
	if createdAtStr.Valid && createdAtStr.String != "" {
		// Try parsing with different time formats
		formats := []string{
			time.RFC3339,                // ISO 8601 format
			"2006-01-02 15:04:05",       // SQLite default format
			"2006-01-02T15:04:05Z07:00", // Another common format
		}

		var parseErr error
		for _, format := range formats {
			user.CreatedAt, parseErr = time.Parse(format, createdAtStr.String)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			log.Printf("Warning: Failed to parse created_at timestamp: %v", parseErr)
		}
	}

	// Parse last_login timestamp if present
	if lastLoginStr.Valid && lastLoginStr.String != "" {
		// Try parsing with different time formats
		formats := []string{
			time.RFC3339,                // ISO 8601 format
			"2006-01-02 15:04:05",       // SQLite default format
			"2006-01-02T15:04:05Z07:00", // Another common format
		}

		var parseErr error
		for _, format := range formats {
			user.LastLogin, parseErr = time.Parse(format, lastLoginStr.String)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			log.Printf("Warning: Failed to parse last_login timestamp: %v", parseErr)
		}
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return User{}, errors.New("invalid password")
	}

	return user, nil
}

// UpdateUserOnlineStatus updates the user's online status
func UpdateUserOnlineStatus(userID int, isOnline bool) error {
	_, err := database.DB.Exec("UPDATE users SET is_online = ?, last_login = CURRENT_TIMESTAMP WHERE id = ?", isOnline, userID)
	return err
}

// GetAllUsers retrieves all users
func GetAllUsers() ([]UserForPublic, error) {
	rows, err := database.DB.Query("SELECT id, nickname, first_name, last_name, is_online FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserForPublic
	for rows.Next() {
		var user UserForPublic
		if err := rows.Scan(&user.ID, &user.Nickname, &user.FirstName, &user.LastName, &user.IsOnline); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// ToPublic converts a User to UserForPublic
func (u User) ToPublic() UserForPublic {
	return UserForPublic{
		ID:        u.ID,
		Nickname:  u.Nickname,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		IsOnline:  u.IsOnline,
	}
}

// GetUserByEmail retrieves a user by their email address
func GetUserByEmail(email string) (User, error) {
	var user User
	var createdAtStr, lastLoginStr sql.NullString

	err := database.DB.QueryRow(
		"SELECT id, nickname, email, password, first_name, last_name, age, gender, created_at, last_login, is_online FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Nickname, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Age, &user.Gender, &createdAtStr, &lastLoginStr, &user.IsOnline)

	if err != nil {
		return User{}, err
	}

	// Set default timestamps
	user.CreatedAt = time.Now()  // Default to current time if not set
	user.LastLogin = time.Time{} // Default to zero time if not set

	// Parse created_at timestamp if present
	if createdAtStr.Valid && createdAtStr.String != "" {
		// Try parsing with different time formats
		formats := []string{
			time.RFC3339,                // ISO 8601 format
			"2006-01-02 15:04:05",       // SQLite default format
			"2006-01-02T15:04:05Z07:00", // Another common format
		}

		var parseErr error
		for _, format := range formats {
			user.CreatedAt, parseErr = time.Parse(format, createdAtStr.String)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			log.Printf("Warning: Failed to parse created_at timestamp: %v", parseErr)
		}
	}

	// Parse last_login timestamp if present
	if lastLoginStr.Valid && lastLoginStr.String != "" {
		// Try parsing with different time formats
		formats := []string{
			time.RFC3339,                // ISO 8601 format
			"2006-01-02 15:04:05",       // SQLite default format
			"2006-01-02T15:04:05Z07:00", // Another common format
		}

		var parseErr error
		for _, format := range formats {
			user.LastLogin, parseErr = time.Parse(format, lastLoginStr.String)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			log.Printf("Warning: Failed to parse last_login timestamp: %v", parseErr)
		}
	}

	return user, nil
}
