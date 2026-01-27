package websocket

import (
	"log"
	"sync"

	"example.com/web-service/internal/config"
)

type Hub struct {
	clients    map[string]*Client // map[ClientID]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if client.ClientID != "" {
				if oldClient, ok := h.clients[client.ClientID]; ok {
					// Close old connection
					close(oldClient.send)
					delete(h.clients, client.ClientID)
				}
				h.clients[client.ClientID] = client
			}
			h.logStats()
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if client.ClientID != "" {
				if currentClient, ok := h.clients[client.ClientID]; ok && currentClient == client {
					delete(h.clients, client.ClientID)
					close(client.send)
				}
			}
			h.logStats()
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.ClientID)
					h.logStats()
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) logStats() {
	clientIDs := make([]string, 0, len(h.clients))
	for id := range h.clients {
		clientIDs = append(clientIDs, id)
	}
	log.Printf("[Hub] Self ClientID: %s, Total Clients: %d, Connected ClientIDs: %v", config.ClientID, len(h.clients), clientIDs)
}

func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}
