package hub

import (
	"sync"
	"fmt"
	"github.com/gorilla/websocket"
)

var (
	Rooms      = make(map[string]map[*websocket.Conn]bool) // roomID → set of connections
	RoomsMutex sync.Mutex
)


// BroadcastToRoom sends a message to all clients in a room
func BroadcastToRoom(roomID string, message string) {
	RoomsMutex.Lock()
	defer RoomsMutex.Unlock()

	clients, ok := Rooms[roomID]
	if !ok {
		fmt.Println("Room not found:", roomID)
		return
	}

	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			fmt.Println("Error broadcasting:", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}