package main

import (
	"CHATX/config"
	"CHATX/handlers"
	"net/http"
)

func main() {
	config.ConnectDB()

	http.HandleFunc("/register", handlers.HandleRegister)
	http.HandleFunc("/login", handlers.HandleLogin)
	http.HandleFunc("/hub", handlers.HandleHub)
	http.HandleFunc("/create-room", handlers.HandleCreateRoom)
	http.HandleFunc("/join", handlers.HandleJoinRoom)
	http.HandleFunc("/ws", handlers.HandleWebSocket)
	http.ListenAndServe(":8080", nil)
}
