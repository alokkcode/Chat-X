package handlers

import (
	"CHATX/hub"
	"CHATX/models"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"github.com/gorilla/websocket"
	"log" 
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
	// Get username from cookie
	usernameCookie, err1 := r.Cookie("username")
	roleCookie, err2 := r.Cookie("role")
	roomID := r.URL.Query().Get("room")

	if err1 != nil || err2 != nil || roomID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user details
	username := usernameCookie.Value
	role := roleCookie.Value //Fetch user's role

	// Optional: fetch room name from DB for display
	room, err := models.GetRoomByID(roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Convert roomID from string to integer
	roomIDInt, err := strconv.Atoi(roomID)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	// Fetch messages for this room
	messages, err := models.GetMessagesByRoomID(roomIDInt)
	if err != nil {
		fmt.Println("Error fetching messages:", err)
		messages = []models.Message{} // Default to empty list if error occurs
	}

	// Pass the user's role and username to the template
	tmpl := template.Must(template.ParseFiles("templates/chatroom.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Username": username,
		"Role":     role,
		"RoomID":   roomID,
		"RoomName": room.Name,
		"Messages": messages, //passing messages to template as well
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

		// Get the user ID using the helper function
		userID, _ := models.GetUserIDByUsername(username) // Retrieves user ID from DB

		roomIDInt, err := strconv.Atoi(roomID) // Convert roomID string to int
		if err != nil {
    	log.Println("Error converting roomID to int:", err)
    	return
		}


		// Save the received message to the database
		err = models.SaveMessage(roomIDInt, userID, string(p))
		if err != nil {
			log.Println("Error saving message:", err)
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


func HandleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
		return
	}

	// Get current user
	usernameCookie, err1 := r.Cookie("username")
	roleCookie, err2 := r.Cookie("role")
	if err1 != nil || err2 != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	username := usernameCookie.Value
	role := roleCookie.Value

	// Get message ID from form
	msgIDStr := r.FormValue("id")
	msgID, err := strconv.Atoi(msgIDStr)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Get message owner ID
	message, err := models.GetMessageByID(msgID)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Fetch current user ID
	userID, err := models.GetUserIDByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	
	// Debugging log: Print details before attempting deletion
	fmt.Println("Deleting Message:", msgID, "User:", username, "Role:", role)

	// Only allow delete if admin or owner
	if role != "admin" && message.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Delete message
	err = models.DeleteMessage(msgID)
	if err != nil {
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	// Redirect back
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}



