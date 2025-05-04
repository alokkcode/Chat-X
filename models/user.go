package models

import (
	"CHATX/config"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	Role         string
}

// Register a new user
func RegisterUser(username, email, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = config.DB.Exec("INSERT INTO users (username, email, password_hash, role) VALUES (?, ?, ?, ?)",
		username, email, string(hashedPassword), role)
	return err
}

// Get user by email for login
func GetUserByEmail(email string) (*User, error) {
	row := config.DB.QueryRow("SELECT id, username, email, password_hash, role FROM users WHERE email = ?", email)
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Fetch user ID from database given an email
func GetUserIDByEmail(email string) (int, error) {
	db := config.GetDB()
	var userID int

	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("user not found")
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

// GenerateSessionToken creates a random session token
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32) // 256-bit token
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// StoreSessionToken saves the session token in the database
func StoreSessionToken(userID int, token string) error {
	_, err := config.DB.Exec("INSERT INTO sessions (user_id, token) VALUES (?, ?)", userID, token)
	return err
}

// ValidateSessionToken checks if the session token is valid
func ValidateSessionToken(token string) (*User, error) {
	row := config.DB.QueryRow(`
		SELECT users.id, users.username, users.email, users.role 
		FROM users 
		INNER JOIN sessions ON users.id = sessions.user_id 
		WHERE sessions.token = ?`, token)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid session")
		}
		return nil, err
	}
	return &user, nil
}
