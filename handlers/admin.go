package handlers

// TODO: Define admin-only handlers here (e.g., create/delete rooms)

import (
	"CHATX/models"
	"net/http"
	"encoding/json"
	"strconv"   
	"strings"
)

// HandleCreateRoom handles room creation by an admin
func HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Validate session
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized: No session", http.StatusUnauthorized)
		return
	}

	user, err := models.ValidateSessionToken(sessionCookie.Value)
	if err != nil || user.Role != "admin" {
		http.Error(w, "Unauthorized: Admin access required", http.StatusUnauthorized)
		return
	}

	// Parse JSON input instead of form values
	var requestData struct {
		RoomName string `json:"room_name"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.RoomName == "" {
		http.Error(w, "Room name required", http.StatusBadRequest)
		return
	}

	// Create room & associate it with the admin who created it
	err = models.CreateRoom(requestData.RoomName, user.ID)
	if err != nil {
		http.Error(w, "Could not create room", http.StatusInternalServerError)
		return
	}

	// Retrieve the newly created room
	newRoom, err := models.GetLatestRoom(user.ID)
	if err != nil {
		http.Error(w, "Could not retrieve new room", http.StatusInternalServerError)
		return
	}

	// Respond with JSON (DO NOT REDIRECT)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"room":    newRoom,
	})
}


// HandleDeleteRoom allows an admin to delete a room they created
func HandleDeleteRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete { // Change from http.MethodPost to http.MethodDelete
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	sessionCookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized: No session", http.StatusUnauthorized)
		return
	}

	user, err := models.ValidateSessionToken(sessionCookie.Value)
	if err != nil || user.Role != "admin" {
		http.Error(w, "Unauthorized: Admin access required", http.StatusUnauthorized)
		return
	}

	//Extract room ID from URL instead of reading JSON body
	roomIDStr := strings.TrimPrefix(r.URL.Path, "/api/delete-room/")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	//Check if admin owns the room before deleting
	isAdminRoom, err := models.IsRoomCreatedByAdmin(roomID, user.ID)
	if err != nil || !isAdminRoom {
		http.Error(w, "You can only delete rooms you created", http.StatusForbidden)
		return
	}

	// Delete the room and associated messages
	err = models.DeleteRoom(roomID, user.ID)
	if err != nil {
		http.Error(w, "Failed to delete room: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"success": "true", "message": "Room deleted successfully"})
}


