package store

import (
	"log"
	"sync"

	"example.com/web-service/internal/models"
)

var (
	storeMu        sync.Mutex
	fileStore      []models.FileData
	storedIP       string
	storedPort     int
	fileStoreIndex int64
	nextIndex      int64 = 1
)

// StoreFiles saves the file list, ip, and port to memory and returns the current index
func StoreFiles(files []models.FileData, ip string, port int) int64 {
	storeMu.Lock()
	defer storeMu.Unlock()
	currentIndex := nextIndex
	fileStore = files
	storedIP = ip
	storedPort = port
	fileStoreIndex = currentIndex
	nextIndex++
	log.Printf("Saved files with index: %d, ip: %s, port: %d", currentIndex, ip, port)
	return currentIndex
}

// GetFiles returns the stored files, ip, and port
func GetFiles() ([]models.FileData, string, int) {
	storeMu.Lock()
	defer storeMu.Unlock()
	return fileStore, storedIP, storedPort
}
