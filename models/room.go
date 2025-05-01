package models

// TODO: Define Room struct and DB logic herE

import (
	"CHATX/config"
	"database/sql"
	"fmt"
)

type Room struct {
	ID   int
	Name string
}

func GetAllRooms() ([]Room, error) {
	rows, err := config.DB.Query("SELECT id, name FROM rooms")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		err := rows.Scan(&room.ID, &room.Name)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func CreateRoom(name string) error {
	_, err := config.DB.Exec("INSERT INTO rooms (name) VALUES (?)", name)
	return err
}

func GetRoomByID(id string) (*Room, error) {
	var room Room
	err := config.DB.QueryRow("SELECT id, name FROM rooms WHERE id = ?", id).Scan(&room.ID, &room.Name)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

// AddUserToRoom adds a user to a specified room in the database
func AddUserToRoom(roomID, username string) error {
	// Get the user ID based on the username
	var userID int
	err := config.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return err
	}

	// Insert the user-room relationship into the user_rooms table
	_, err = config.DB.Exec("INSERT INTO user_rooms (user_id, room_id) VALUES (?, ?)", userID, roomID)
	if err != nil {
		return err
	}

	return nil
}