package handlers

import (
	"CHATX/models"
	"html/template"
	"net/http"
)

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
