// backend/models/user.go
package models

import (
    "database/sql"
    "time"
)

type User struct {
    ID        int       `json:"id"`
    Nickname  string    `json:"nickname"`
    Age       int       `json:"age"`
    Gender    string    `json:"gender"`
    FirstName string    `json:"firstName"`
    LastName  string    `json:"lastName"`
    Email     string    `json:"email"`
    Password  string    `json:"-"` // Don't send password to client
    CreatedAt time.Time `json:"createdAt"`
}

type UserProfile struct {
    UserID      int       `json:"userId"`
    Bio         string    `json:"bio"`
    Avatar      string    `json:"avatar"`
    LastActive  time.Time `json:"lastActive"`
    PostCount   int       `json:"postCount"`
    CommentCount int      `json:"commentCount"`
}

// RegisterUser registers a new user in the database
func RegisterUser(db *sql.DB, user User, hashedPassword string) (int64, error) {
    query := `INSERT INTO users (nickname, age, gender, first_name, last_name, email, password) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
    
    result, err := db.Exec(query, user.Nickname, user.Age, user.Gender, user.FirstName, user.LastName, user.Email, hashedPassword)
    if err != nil {
        return 0, err
    }
    
    return result.LastInsertId()
}

// GetUserByNicknameOrEmail finds user by nickname or email for login
func GetUserByNicknameOrEmail(db *sql.DB, identifier string) (User, error) {
    var user User
    query := `SELECT id, nickname, age, gender, first_name, last_name, email, password, created_at
              FROM users 
              WHERE nickname = ? OR email = ?`
    
    row := db.QueryRow(query, identifier, identifier)
    err := row.Scan(&user.ID, &user.Nickname, &user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.CreatedAt)
    
    return user, err
}

// GetUserByID retrieves a user by their ID
func GetUserByID(db *sql.DB, id int) (User, error) {
    var user User
    query := `SELECT id, nickname, age, gender, first_name, last_name, email, created_at
              FROM users 
              WHERE id = ?`
    
    row := db.QueryRow(query, id)
    err := row.Scan(&user.ID, &user.Nickname, &user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt)
    
    return user, err
}

// GetUserProfile retrieves a user's profile
func GetUserProfile(db *sql.DB, userID int) (UserProfile, error) {
    var profile UserProfile
    profile.UserID = userID
    
    // Get bio and avatar
    query := `SELECT bio, avatar FROM user_profiles WHERE user_id = ?`
    row := db.QueryRow(query, userID)
    err := row.Scan(&profile.Bio, &profile.Avatar)
    
    // If profile doesn't exist yet, create an empty one
    if err == sql.ErrNoRows {
        profile.Bio = ""
        profile.Avatar = "default.png"
    } else if err != nil {
        return profile, err
    }
    
    // Get post count
    postQuery := `SELECT COUNT(*) FROM posts WHERE user_id = ?`
    row = db.QueryRow(postQuery, userID)
    err = row.Scan(&profile.PostCount)
    if err != nil {
        return profile, err
    }
    
    // Get comment count
    commentQuery := `SELECT COUNT(*) FROM comments WHERE user_id = ?`
    row = db.QueryRow(commentQuery, userID)
    err = row.Scan(&profile.CommentCount)
    if err != nil {
        return profile, err
    }
    
    return profile, nil
}

// UpdateUserProfile updates a user's profile
func UpdateUserProfile(db *sql.DB, profile UserProfile) error {
    // Check if profile exists
    var count int
    query := `SELECT COUNT(*) FROM user_profiles WHERE user_id = ?`
    row := db.QueryRow(query, profile.UserID)
    err := row.Scan(&count)
    if err != nil {
        return err
    }
    
    if count > 0 {
        // Update existing profile
        query = `UPDATE user_profiles SET bio = ?, avatar = ? WHERE user_id = ?`
        _, err = db.Exec(query, profile.Bio, profile.Avatar, profile.UserID)
    } else {
        // Create new profile
        query = `INSERT INTO user_profiles (user_id, bio, avatar) VALUES (?, ?, ?)`
        _, err = db.Exec(query, profile.UserID, profile.Bio, profile.Avatar)
    }
    
    return err
}