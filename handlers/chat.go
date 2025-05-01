package handlers

import (
	"CHATX/hub"
	"CHATX/models"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocket Upgrader for handling HTTP to WebSocket upgrade
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

// HandleHub shows the chat hub page
func HandleHub(w http.ResponseWriter, r *http.Request) {
	usernameCookie, err := r.Cookie("username")
	roleCookie, err2 := r.Cookie("role")

	if err != nil || err2 != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := usernameCookie.Value
	role := roleCookie.Value

	rooms, err := models.GetAllRooms()
	if err != nil {
		http.Error(w, "Could not fetch rooms", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/hub.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Username": username,
		"Role":     role,
		"Rooms":    rooms,
	})
}

// HandleJoinRoom renders the chatroom UI when joining a room
func HandleJoinRoom(w http.ResponseWriter, r *http.Request) {
	usernameCookie, err1 := r.Cookie("username")
	roomID := r.URL.Query().Get("room")

	if err1 != nil || roomID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Optional: fetch room name from DB for display
	room, err := models.GetRoomByID(roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/chatroom.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Username": usernameCookie.Value,
		"RoomID":   roomID,
		"RoomName": room.Name,
	})
}

// HandleWebSocket is responsible for upgrading HTTP requests to WebSocket connections
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	username := r.URL.Query().Get("user")

	if roomID == "" || username == "" {
		http.Error(w, "Missing room or user", http.StatusBadRequest)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Error upgrading to WebSocket", http.StatusInternalServerError)
		return
	}

	// Add user to the room
	hub.RoomsMutex.Lock()
	if hub.Rooms[roomID] == nil {
		hub.Rooms[roomID] = make(map[*websocket.Conn]bool)
	}
	hub.Rooms[roomID][conn] = true
	hub.RoomsMutex.Unlock()

	// Ensure cleanup on disconnect
	defer func() {
		hub.RoomsMutex.Lock()
		delete(hub.Rooms[roomID], conn)
		hub.RoomsMutex.Unlock()
		conn.Close()
	}()

	// Notify users of the new join
	notifyJoin := fmt.Sprintf("%s has joined the room", username)
	hub.BroadcastToRoom(roomID, notifyJoin)

	// Handle messaging within the room
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			// fmt.Println("Error reading message:", err)
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				fmt.Println("User disconnected normally")
			} else {
				fmt.Println("WebSocket error:", err)
			}
			break
		}

		// Broadcast message
		hub.BroadcastToRoom(roomID, fmt.Sprintf("%s: %s", username, string(p)))

		// Echo message back to sender
		if err := conn.WriteMessage(messageType, p); err != nil {
			fmt.Println("Error sending message:", err)
			break
		}
	}
}


