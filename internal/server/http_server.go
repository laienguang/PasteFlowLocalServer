package server

import (
	"fmt"
	"log"
	"net/http"

	"example.com/web-service/internal/api"
	"example.com/web-service/internal/config"
	"example.com/web-service/internal/websocket"
)

func StartHTTP(hub *websocket.Hub, manager *websocket.ClientManager) {
	// Setup HTTP routes
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request: /hello")
		fmt.Fprintf(w, "hello")
	})
	http.HandleFunc("/api/copyFileInfoToCloud", api.HandleCopyFileInfoToCloud(hub, manager))
	http.HandleFunc("/download", api.HandleDownload)
	http.HandleFunc("/udp/send", api.HandleUDPSend)

	// WebSocket route
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	log.Printf("Server running at http://localhost:%d", config.HttpPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.HttpPort), nil); err != nil {
		log.Fatal(err)
	}
}
