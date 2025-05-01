package models

// TODO: Define Room struct and DB logic herE

import "CHATX/config"

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
