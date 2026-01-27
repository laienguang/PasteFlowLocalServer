package api

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"

	"example.com/web-service/internal/config"
	"example.com/web-service/internal/models"
	"example.com/web-service/internal/store"
	"example.com/web-service/internal/websocket"
)

// GetLocalIP returns the non-loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func HandleCopyFileInfoToCloud(hub *websocket.Hub, manager *websocket.ClientManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		EnableCORS(w)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Println("Received copyFileInfoToCloud request")

		// Expecting JSON: { "files": [ ... ] }
		var payload struct {
			Files []models.FileData `json:"files"`
		}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&payload); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		if payload.Files == nil {
			// Try to handle if array is passed directly, though Node code suggests object wrapper
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		log.Printf("Files to copy file info to cloud: %+v", payload.Files)

		// Get local IP
		localIP := GetLocalIP()

		// Save files to memory with auto-increment index
		currentIndex := store.StoreFiles(payload.Files, localIP, config.HttpPort)

		// Broadcast to local clients and cloud servers
		msg := websocket.Message{
			Type: "copyFileInfoToCloud",
			Data: models.CopyFileInfoData{
				Files: payload.Files,
				Index: currentIndex,
				IP:    localIP,
				Port:  config.HttpPort,
			},
		}
		msgBytes, err := json.Marshal(msg)
		if err == nil {
			if hub != nil {
				hub.Broadcast(msgBytes)
			}
			if manager != nil {
				manager.Broadcast(msgBytes)
			}
		} else {
			log.Printf("Error marshaling broadcast message: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Copy file info to cloud received successfully",
			"count":   len(payload.Files),
		})
	}
}

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "Missing file path", http.StatusBadRequest)
		return
	}

	log.Printf("Received request: /download?path=%s", filePath)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}
