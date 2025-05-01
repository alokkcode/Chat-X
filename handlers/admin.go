package handlers

// TODO: Define admin-only handlers here (e.g., create/delete rooms)
	

import (
	"CHATX/models"
	"net/http"
)

// HandleCreateRoom handles room creation by admin
func HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
		return
	}

	roleCookie, err := r.Cookie("role")
	if err != nil || roleCookie.Value != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	roomName := r.FormValue("room_name")
	if roomName == "" {
		http.Error(w, "Room name required", http.StatusBadRequest)
		return
	}

	err = models.CreateRoom(roomName)
	if err != nil {
		http.Error(w, "Could not create room", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/hub", http.StatusSeeOther)
}
