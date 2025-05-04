package main

import (
	"CHATX/config"
	"CHATX/handlers"
	"net/http"
)

func main() {
	config.ConnectDB()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/register", handlers.RegisterUser)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/hub", handlers.HandleHub)
	http.HandleFunc("/create-room", handlers.HandleCreateRoom)
	http.HandleFunc("/join", handlers.HandleJoinRoom)
	http.HandleFunc("/ws", handlers.HandleWebSocket)
	http.HandleFunc("/api/delete-message", handlers.HandleAPIDeleteMessage)
	http.HandleFunc("/admin/dashboard", handlers.HandleAdminDashboard)
	http.HandleFunc("/ws-dashboard", handlers.HandleDashboardWebSocket)
	http.HandleFunc("/api/delete-room/", handlers.HandleDeleteRoom)
	http.ListenAndServe(":8080", nil)
}
