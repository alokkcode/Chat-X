package models

import (
	"CHATX/config"
	"fmt"
)

type Room struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedBy int    `json:"created_by"`
}

// GetAllRooms retrieves all rooms with creator details
func GetAllRooms() ([]Room, error) {
	rows, err := config.DB.Query("SELECT id, name, created_by FROM rooms")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		err := rows.Scan(&room.ID, &room.Name, &room.CreatedBy)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// CreateRoom inserts a new room into the database
func CreateRoom(name string, createdBy int) error {
	_, err := config.DB.Exec("INSERT INTO rooms (name, created_by) VALUES (?, ?)", name, createdBy)
	return err
}

// GetRoomByID fetches details of a specific room
func GetRoomByID(roomID int) (*Room, error) {
	var room Room
	err := config.DB.QueryRow("SELECT id, name, created_by FROM rooms WHERE id = ?", roomID).Scan(&room.ID, &room.Name, &room.CreatedBy)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

// DeleteRoom removes a room (only if the requesting user created it)
func DeleteRoom(roomID, userID int) error {
	result, err := config.DB.Exec("DELETE FROM rooms WHERE id = ? AND created_by = ?", roomID, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("You are not allowed to delete this room")
	}

	_, _ = config.DB.Exec("DELETE FROM messages WHERE room_id = ?", roomID) // Cleanup related messages
	return nil
}

// GetLatestRoom fetches the most recently created room by a user
func GetLatestRoom(userID int) (*Room, error) {
	var room Room
	err := config.DB.QueryRow("SELECT id, name FROM rooms WHERE created_by = ? ORDER BY id DESC LIMIT 1", userID).Scan(&room.ID, &room.Name)
	if err != nil {
		return nil, err
	}
	return &room, nil
}