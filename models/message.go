package models

// TODO: Define message struct and DB logic here

import (
    "CHATX/config"
    _"database/sql"
    "time"
)

type Message struct {
    ID        int
    RoomID    int
    UserID    int
    Content   string
    Timestamp time.Time
	Username  string
}

// SaveMessage inserts a new chat message
func SaveMessage(roomID int, userID int, content string) error {
    db := config.GetDB()
    _, err := db.Exec("INSERT INTO messages (room_id, user_id, content) VALUES (?, ?, ?)", roomID, userID, content)
    return err
}

// GetMessagesByRoom fetches messages for a room (optional for later)
func GetMessagesByRoom(roomID int) ([]Message, error) {
    db := config.GetDB()
    rows, err := db.Query("SELECT id, room_id, user_id, content, timestamp FROM messages WHERE room_id = ? ORDER BY timestamp ASC", roomID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var messages []Message
    for rows.Next() {
        var msg Message
        err := rows.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.Content, &msg.Timestamp)
        if err != nil {
            return nil, err
        }
        messages = append(messages, msg)
    }
    return messages, nil
}

// GetMessagesByRoomID returns the latest messages for a given room
func GetMessagesByRoomID(roomID int) ([]Message, error) {
    db := config.GetDB()
    rows, err := db.Query(`
        SELECT m.id, m.content, m.timestamp, u.username 
        FROM messages m
        JOIN users u ON m.user_id = u.id
        WHERE m.room_id = ?
        ORDER BY m.timestamp ASC`, roomID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var messages []Message
    for rows.Next() {
        var msg Message
        var username string
        var timestampStr string // Convert timestamp to a string first

        err := rows.Scan(&msg.ID, &msg.Content, &timestampStr, &username) // Scan timestamp as a string
        if err != nil {
            return nil, err
        }

        // Convert timestamp string to time.Time object
        msg.Timestamp, err = time.Parse("2006-01-02 15:04:05", timestampStr)
        if err != nil {
            return nil, err
        }

        msg.Username = username
        messages = append(messages, msg)
    }
    return messages, nil
}

// GetMessageByID fetches message details (used for deletion)
func GetMessageByID(msgID int) (Message, error) {
	db := config.GetDB()
	var msg Message
	err := db.QueryRow("SELECT id, user_id FROM messages WHERE id = ?", msgID).Scan(&msg.ID, &msg.UserID)
	if err != nil {
		return msg, err
	}
	return msg, nil
}

// DeleteMessage deletes a message from DB
func DeleteMessage(msgID int) error {
	db := config.GetDB()
	_, err := db.Exec("DELETE FROM messages WHERE id = ?", msgID)
	return err
}
