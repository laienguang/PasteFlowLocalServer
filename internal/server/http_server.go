package server

import (
	"fmt"
	"log"
	"net/http"

	"example.com/web-service/internal/api"
	"example.com/web-service/internal/config"
)

func StartHTTP() {
	// Setup HTTP routes
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request: /hello")
		fmt.Fprintf(w, "hello")
	})
	http.HandleFunc("/api/copyFileInfoToCloud", api.HandleCopyFileInfoToCloud)
	http.HandleFunc("/download", api.HandleDownload)
	http.HandleFunc("/udp/send", api.HandleUDPSend)

	log.Printf("Server running at http://localhost:%d", config.HttpPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.HttpPort), nil); err != nil {
		log.Fatal(err)
	}
}
