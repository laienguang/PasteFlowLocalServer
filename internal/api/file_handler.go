package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type FileData struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size,omitempty"`
}

func HandleCopyFileInfoToCloud(w http.ResponseWriter, r *http.Request) {
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
		Files []FileData `json:"files"`
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Copy file info to cloud received successfully",
		"count":   len(payload.Files),
	})
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
