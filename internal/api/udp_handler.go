package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
)

func HandleUDPSend(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Host    string `json:"host"`
		Port    int    `json:"port"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if reqBody.Host == "" || reqBody.Port == 0 || reqBody.Message == "" {
		http.Error(w, "Missing host, port, or message", http.StatusBadRequest)
		return
	}

	// Send UDP
	remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", reqBody.Host, reqBody.Port))
	if err != nil {
		log.Printf("UDP resolve error: %v", err)
		http.Error(w, "Invalid address", http.StatusInternalServerError)
		return
	}

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		log.Printf("UDP manual send error: %v", err)
		http.Error(w, "Failed to send UDP message", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(reqBody.Message))
	if err != nil {
		log.Printf("UDP manual send error: %v", err)
		http.Error(w, "Failed to send UDP message", http.StatusInternalServerError)
		return
	}

	log.Printf("UDP message sent to %s:%d", reqBody.Host, reqBody.Port)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "UDP message sent successfully",
	})
}
