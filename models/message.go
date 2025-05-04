package models

import (
	"CHATX/config"
	"database/sql"
	"time"
	"fmt"
)

type Message struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"room_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Deletable bool      `json:"deletable"`
}

// SaveMessageAndGetID stores a new message in the database and returns its ID
func SaveMessageAndGetID(roomID, userID int, content string) (int, error) {
	result, err := config.DB.Exec("INSERT INTO messages (room_id, user_id, content) VALUES (?, ?, ?)", 
		roomID, userID, content)
	if err != nil {
		return 0, err
	}

	// Get the last inserted message ID
	messageID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(messageID), nil
}


// GetMessagesByRoomID retrieves all messages for a specific room
func GetMessagesByRoomID(roomID int, currentUserID int, currentUserRole string) ([]Message, error) {
	rows, err := config.DB.Query(`
		SELECT m.id, m.room_id, u.username, u.role, m.content, m.timestamp, m.user_id
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.room_id = ?
		ORDER BY m.timestamp ASC
	`, roomID)
	if err != nil {
		fmt.Println("Error fetching messages from DB:", err)
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var senderID int
		var rawTimestamp []uint8

		err := rows.Scan(&msg.ID, &msg.RoomID, &msg.Username, &msg.Role, &msg.Content, &rawTimestamp, &senderID)
		if err != nil {
			fmt.Println("Error scanning message row:", err)
			return nil, err
		}

		parsedTime, err := time.Parse("2006-01-02 15:04:05", string(rawTimestamp))
		if err != nil {
			fmt.Println("Error parsing timestamp:", err)
			return nil, err
		}
		msg.Timestamp = parsedTime

		// FIX: Admins can delete ALL messages, users can delete ONLY their own
		msg.Deletable = (currentUserRole == "admin") || (currentUserID == senderID)

		messages = append(messages, msg)
	}

	fmt.Println("Fetched messages:", messages) //Log retrieved messages
	return messages, nil
}

// GetMessageByID fetches a specific message (used for deletion)
func GetMessageByID(msgID int) (*Message, error) {
	var msg Message
	err := config.DB.QueryRow("SELECT id, user_id FROM messages WHERE id = ?", msgID).Scan(&msg.ID, &msg.UserID)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// DeleteMessage deletes a message (only if the user is allowed)
func DeleteMessage(msgID, userID int) error {
	// Check if the user has permission to delete the message
	result, err := config.DB.Exec("DELETE FROM messages WHERE id = ? AND (user_id = ? OR (SELECT role FROM users WHERE id = ?) = 'admin')",
		msgID, userID, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}