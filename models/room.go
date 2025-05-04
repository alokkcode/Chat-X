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



// GetLatestRoom fetches the most recently created room by a user
func GetLatestRoom(userID int) (*Room, error) {
	var room Room
	err := config.DB.QueryRow("SELECT id, name FROM rooms WHERE created_by = ? ORDER BY id DESC LIMIT 1", userID).Scan(&room.ID, &room.Name)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func GetRoomsByAdmin(adminID int) ([]Room, error) {
	rows, err := config.DB.Query("SELECT id, name FROM rooms WHERE created_by = ?", adminID)
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

func GetMessageCount(roomID int) (int, error) {
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM messages WHERE room_id = ?", roomID).Scan(&count)
	return count, err
}

func GetActiveUserCount(roomID int) (int, error) {
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM user_rooms WHERE room_id = ?", roomID).Scan(&count)
	return count, err
}



func IsRoomCreatedByAdmin(roomID, adminID int) (bool, error) {
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM rooms WHERE id = ? AND created_by = ?", roomID, adminID).Scan(&count)
	return count > 0, err
}




func DeleteRoom(roomID, adminID int) error {
	// Ensure the admin is the creator of the room before deleting
	isAdminRoom, err := IsRoomCreatedByAdmin(roomID, adminID)
	if err != nil || !isAdminRoom {
		return fmt.Errorf("You can only delete rooms you created")
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return err
	}

	// Delete all messages related to the room
	_, err = tx.Exec("DELETE FROM messages WHERE room_id = ?", roomID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete the room itself
	_, err = tx.Exec("DELETE FROM rooms WHERE id = ?", roomID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit() // Ensure everything is deleted before committing
}