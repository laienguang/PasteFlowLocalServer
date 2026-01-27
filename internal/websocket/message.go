package websocket

import (
	"encoding/json"
	"log"

	"example.com/web-service/internal/models"
	"example.com/web-service/internal/store"
)

// Message represents a generic WebSocket message
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// HandleMessage parses and processes incoming WebSocket messages
func HandleMessage(message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err == nil {
		log.Printf("Parsed Message - Type: %s, Data: %+v", msg.Type, msg.Data)
		if msg.Type == "copyFileInfoToCloud" {
			if dataBytes, err := json.Marshal(msg.Data); err == nil {
				var payload models.CopyFileInfoData
				if err := json.Unmarshal(dataBytes, &payload); err == nil {
					store.StoreFiles(payload.Files, payload.IP, payload.Port)
				} else {
					log.Printf("Failed to parse copyFileInfoToCloud data: %v", err)
				}
			}
		}
	} else {
		log.Printf("Failed to parse message as generic Message: %v", err)
	}
}
