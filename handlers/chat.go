package handlers

import (
	"CHATX/hub"
	"CHATX/models"
	"fmt"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"
	"github.com/gorilla/websocket"
	"log"
	"CHATX/config"
)

// WebSocket Upgrader for handling HTTP to WebSocket upgrade
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

// HandleHub renders the chat hub page
func HandleHub(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := models.ValidateSessionToken(sessionCookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	rooms, err := models.GetAllRooms()
	if err != nil {
		http.Error(w, "Could not fetch rooms", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/hub.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Username": user.Username,
		"Role":     user.Role,
		"Rooms":    rooms,
	})
}

// HandleJoinRoom renders the chatroom UI when joining a room
func HandleJoinRoom(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := models.ValidateSessionToken(sessionCookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	roomIDInt, err := strconv.Atoi(roomID)
	if err != nil {
		http.Error(w, "Invalid room ID format", http.StatusBadRequest)
		return
	}

	room, err := models.GetRoomByID(roomIDInt)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	messages, err := models.GetMessagesByRoomID(roomIDInt, user.ID, user.Role) // Pass user ID & role
	if err != nil {
		messages = []models.Message{}
	}
	
	// Encode messages as a valid JSON string
	messagesJSON, err := json.Marshal(messages)
	if err != nil {
	http.Error(w, "Error encoding messages", http.StatusInternalServerError)
	return
}

	fmt.Println("Sending messages to template:", string(messagesJSON))
	tmpl := template.Must(template.ParseFiles("templates/chatroom.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Username": user.Username,
		"UserID":   user.ID,
		"Role":     user.Role,
		"RoomID":   roomID,
		"RoomName": room.Name,
		"MessagesJSON": template.JS(string(messagesJSON)),// Escapes JSON for safe embedding
	})
}

// HandleWebSocket upgrades HTTP requests to WebSocket connections
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized: No session", http.StatusUnauthorized)
		return
	}

	user, err := models.ValidateSessionToken(sessionCookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "Missing room ID", http.StatusBadRequest)
		return
	}

	roomIDInt, err := strconv.Atoi(roomID)
	if err != nil {
		log.Println("Error converting roomID to int:", err)
		http.Error(w, "Invalid room ID format", http.StatusBadRequest)
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

	// Notify users via WebSocket in JSON format
	joinMsg := map[string]interface{}{
		"type":     "system",
		"text":     fmt.Sprintf("%s has joined the room", user.Username),
		"username": "System",
	}
	joinJSON, _ := json.Marshal(joinMsg)
	hub.BroadcastToRoom(roomID, string(joinJSON))

	// Handle messaging within the room
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Save message & get message ID
		messageID, err := models.SaveMessageAndGetID(roomIDInt, user.ID, string(p))
		if err != nil {
			log.Println("Error saving message:", err)
			continue
		}

		// Broadcast message with metadata
		messageData := map[string]interface{}{
			"type":       "chat",
			"message_id": messageID,
			"username":   user.Username,
			"user_id":    user.ID,
			"role":       user.Role,
			"text":       string(p),
			"timestamp":  time.Now().Format(time.RFC3339),
		}
		messageJSON, _ := json.Marshal(messageData)
		hub.BroadcastToRoom(roomID, string(messageJSON))
	}
}

// HandleDeleteMessage handles form-based message deletion requests
func HandleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionCookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := models.ValidateSessionToken(sessionCookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	messageIDStr := r.FormValue("id")
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Delete message
	err = models.DeleteMessage(messageID, user.ID)
	if err != nil {
		http.Error(w, "Failed to delete message", http.StatusForbidden)
		return
	}

	// Get room ID to redirect back
	msg, _ := models.GetMessageByID(messageID)
	roomID := msg.RoomID

	// Notify clients about deletion
	deleteMsg := map[string]interface{}{
		"type":       "delete",
		"message_id": messageID,
	}
	deleteJSON, _ := json.Marshal(deleteMsg)
	hub.BroadcastToRoom(strconv.Itoa(roomID), string(deleteJSON))

	// Redirect back to the chat room
	http.Redirect(w, r, "/join?room="+strconv.Itoa(roomID), http.StatusSeeOther)
}

// HandleAPIDeleteMessage handles AJAX/JSON API requests for message deletion
func HandleAPIDeleteMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
		return
	}

	sessionCookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized: No session", http.StatusUnauthorized)
		return
	}

	user, err := models.ValidateSessionToken(sessionCookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var requestData struct {
		MessageID int `json:"message_id"`
		RoomID    int `json:"room_id"`
	}

	// Read request body and log it
	err = json.NewDecoder(r.Body).Decode(&requestData)
	fmt.Println("Received delete request:", requestData) //  Log the incoming request data

	if err != nil || requestData.MessageID == 0 || requestData.RoomID == 0 {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Delete message
	err = models.DeleteMessage(requestData.MessageID, user.ID)
	if err != nil {
		http.Error(w, "Failed to delete message", http.StatusForbidden)
		return
	}

	// Notify clients about deletion
	deleteMsg := map[string]interface{}{
		"type":       "delete",
		"message_id": requestData.MessageID,
	}
	deleteJSON, _ := json.Marshal(deleteMsg)
	hub.BroadcastToRoom(strconv.Itoa(requestData.RoomID), string(deleteJSON))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Message deleted successfully"})
}



// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	
	// Delete session from database if present
	sessionCookie, err := r.Cookie("session")
	if err == nil {
		_, _ = config.DB.Exec("DELETE FROM sessions WHERE token = ?", sessionCookie.Value)
	}
	
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}