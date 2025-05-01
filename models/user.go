package models

import (
	"CHATX/config"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Username string
	Password string
	Role     string
}

// Register a new user
func RegisterUser(username, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = config.DB.Exec("INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)",
		username, string(hashedPassword), role)
	return err
}

// Get user by username for login
func GetUserByUsername(username string) (*User, error) {
	row := config.DB.QueryRow("SELECT id, username, password_hash, role FROM users WHERE username = ?", username)
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("User not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserIDByUsername fetches the user ID from the database given a username
func GetUserIDByUsername(username string) (int, error) {
    db := config.GetDB() // Ensure we use the existing DB connection
    var userID int

    err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return 0, errors.New("User not found")
        }
        return 0, err
    }

    return userID, nil
}


// Compare hashed password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
