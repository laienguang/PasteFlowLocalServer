package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"example.com/web-service/internal/config"
	"example.com/web-service/internal/websocket"
)

type DiscoveryMessage struct {
	ClientID string `json:"clientId"`
	Port     int    `json:"port"`  // HTTP/WS Port
	WSUrl    string `json:"wsUrl"` // Optional, if provided directly
}

// StartUDP starts the UDP server.
func StartUDP(manager *websocket.ClientManager) {
	addr := net.UDPAddr{
		Port: config.UdpPort,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("UDP server error: %v", err) // Use Fatalf to exit if UDP server fails
	}
	defer conn.Close()

	log.Printf("UDP server listening %s", conn.LocalAddr().String())

	buf := make([]byte, 65535) // Max UDP packet size
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("UDP read error: %v", err)
			continue
		}

		msg := buf[:n]

		// Try to parse as Discovery Message
		var discoveryMsg DiscoveryMessage
		if err := json.Unmarshal(msg, &discoveryMsg); err == nil {
			// 1. Check if it's myself
			if discoveryMsg.ClientID == config.ClientID {
				continue
			}

			log.Printf("Received UDP message from %s: %s", remoteAddr.String(), string(msg))

			// 2. Determine WebSocket URL
			var targetUrl string
			if discoveryMsg.WSUrl != "" {
				targetUrl = discoveryMsg.WSUrl
			} else if discoveryMsg.Port > 0 {
				// Construct URL from remote IP and Port
				// remoteAddr.IP can be IPv4 or IPv6.
				// For IPv6, we need brackets.
				ip := remoteAddr.IP.String()
				// Basic check for IPv6
				if remoteAddr.IP.To4() == nil {
					ip = fmt.Sprintf("[%s]", ip)
				}
				targetUrl = fmt.Sprintf("%s:%d", ip, discoveryMsg.Port)
			}

			if targetUrl != "" {
				// 3. Check if already connected
				if manager.IsConnected(discoveryMsg.ClientID) {
					continue
				}

				log.Printf("Discovered Peer: %s (ClientID: %s)", targetUrl, discoveryMsg.ClientID)
				manager.ConnectToCloud(targetUrl, discoveryMsg.ClientID)
				continue
			}
		}

		if remoteAddr.Port == config.UdpPort {
			// Ignore other broadcasts that we couldn't parse or process
			continue
		}

		// Fallback to Echo logic (only for non-discovery messages)
		response := []byte(fmt.Sprintf("Echo: %s", string(msg)))
		_, err = conn.WriteToUDP(response, remoteAddr)
		if err != nil {
			log.Printf("UDP send error: %v", err)
		}
	}
}
