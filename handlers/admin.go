package handlers

// TODO: Define admin-only handlers here (e.g., create/delete rooms)

import (
	"CHATX/models"
	"net/http"
	"encoding/json"
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
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get session token for authentication
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized: No session", http.StatusUnauthorized)
		return
	}

	// Validate session & get user details
	user, err := models.ValidateSessionToken(sessionCookie.Value)
	if err != nil || user.Role != "admin" {
		http.Error(w, "Unauthorized: Admin access required", http.StatusUnauthorized)
		return
	}

	// Parse room ID from request body
	var requestData struct {
		RoomID int `json:"room_id"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.RoomID == 0 {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	// Attempt to delete the room (only if user is the creator)
	err = models.DeleteRoom(requestData.RoomID, user.ID)
	if err != nil {
		http.Error(w, "Failed to delete room: "+err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Room deleted successfully"})
}


