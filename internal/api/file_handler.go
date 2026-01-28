package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

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

func HandlePasteFileFromCloud(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Path string `json:"path"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if payload.Path == "" {
		http.Error(w, "Missing path", http.StatusBadRequest)
		return
	}

	// Check if destination directory exists and is a directory
	info, err := os.Stat(payload.Path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Destination path does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to access destination path: %v", err), http.StatusInternalServerError)
		return
	}
	if !info.IsDir() {
		http.Error(w, "Destination path is not a directory", http.StatusBadRequest)
		return
	}

	files, storedIP, storedPort := store.GetFiles()
	localIP := GetLocalIP()

	log.Printf("PasteFileFromCloud: Dest=%s, StoredIP=%s, LocalIP=%s, FilesCount=%d (Async started)", payload.Path, storedIP, localIP, len(files))

	go func() {
		var successCount int
		var failureCount int

		for _, file := range files {
			destPath := filepath.Join(payload.Path, file.Name)

			if storedIP == localIP {
				// Local copy
				if err := copyFile(file.Path, destPath); err != nil {
					log.Printf("Failed to copy local file %s: %v", file.Path, err)
					failureCount++
				} else {
					successCount++
				}
			} else {
				// Remote download
				downloadURL := fmt.Sprintf("http://%s:%d/download?path=%s", storedIP, storedPort, url.QueryEscape(file.Path))
				if err := downloadFile(downloadURL, destPath); err != nil {
					log.Printf("Failed to download remote file %s: %v", downloadURL, err)
					failureCount++
				} else {
					successCount++
				}
			}
		}
		log.Printf("Paste operation finished: success=%d, failure=%d", successCount, failureCount)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Paste operation started",
	})
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func downloadFile(url, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, resp.Body)
	return err
}
