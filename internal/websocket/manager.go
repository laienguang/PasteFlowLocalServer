package websocket

import (
	"log"
	"sync"

	"example.com/web-service/internal/config"
)

type ClientManager struct {
	clients   map[string]*CloudClient // map[url]*CloudClient
	clientIDs map[string]string       // map[clientId]url
	hub       *Hub
	mu        sync.RWMutex
}

func NewClientManager(hub *Hub) *ClientManager {
	return &ClientManager{
		clients:   make(map[string]*CloudClient),
		clientIDs: make(map[string]string),
		hub:       hub,
	}
}

func (m *ClientManager) IsConnected(clientId string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.clientIDs[clientId]
	return exists
}

func (m *ClientManager) ConnectToCloud(url string, clientId string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[url]; exists {
		// Already connected or connecting to this URL
		return
	}

	if clientId != "" {
		if _, exists := m.clientIDs[clientId]; exists {
			// Already connected to this ClientID (maybe different URL?)
			return
		}
		m.clientIDs[clientId] = url
	}

	log.Printf("Initiating connection to new cloud server: %s (ClientID: %s)", url, clientId)
	client := NewCloudClient(url, m.hub, func() {
		m.RemoveClient(url)
	})
	m.clients[url] = client
	client.Connect()
	m.logStats()
}

func (m *ClientManager) RemoveClient(url string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[url]; exists {
		delete(m.clients, url)
		log.Printf("Removed client for %s from manager", url)
	}

	// Remove associated clientIDs
	for id, u := range m.clientIDs {
		if u == url {
			delete(m.clientIDs, id)
		}
	}
	m.logStats()
}

func (m *ClientManager) Broadcast(message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, client := range m.clients {
		client.Send(message)
	}
}

func (m *ClientManager) logStats() {
	connectedIDs := make([]string, 0, len(m.clientIDs))
	for id := range m.clientIDs {
		connectedIDs = append(connectedIDs, id)
	}
	log.Printf("[ClientManager] Self ClientID: %s, Total Connections: %d, Connected Cloud ClientIDs: %v", config.ClientID, len(m.clients), connectedIDs)
}
