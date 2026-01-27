package discovery

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"example.com/web-service/internal/config"
	"example.com/web-service/internal/websocket"
)

const (
	BroadcastInterval = 1 * time.Second
	BroadcastCount    = 3
	BroadcastAddr     = "255.255.255.255"
)

type DiscoveryMessage struct {
	ClientID string `json:"clientId"`
	Port     int    `json:"port"`
	IP       string `json:"ip"` // Optional, receiver can use remote addr
}

type CloudServerInfo struct {
	WSUrl string `json:"wsUrl"`
}

func Start(manager *websocket.ClientManager) {
	// Start listening for UDP responses/broadcasts
	go listenForCloudServers(manager)

	// Send broadcasts
	go sendBroadcasts()
}

func sendBroadcasts() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", BroadcastAddr, config.UdpPort))
	if err != nil {
		log.Printf("Discovery: Failed to resolve broadcast address: %v", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Printf("Discovery: Failed to dial UDP: %v", err)
		return
	}
	defer conn.Close()

	msg := DiscoveryMessage{
		ClientID: config.ClientID,
		Port:     config.HttpPort, // Local Server's HTTP/WS port
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Discovery: Failed to marshal message: %v", err)
		return
	}

	for i := 0; i < BroadcastCount; i++ {
		log.Printf("Discovery: Sending broadcast %d/%d: %s", i+1, BroadcastCount, string(data))
		_, err := conn.Write(data)
		if err != nil {
			log.Printf("Discovery: Failed to send broadcast: %v", err)
		}
		time.Sleep(BroadcastInterval)
	}
}

func listenForCloudServers(manager *websocket.ClientManager) {
	// We reuse the UDP port logic or listen on a specific port?
	// Assuming cloud server replies to the source port or broadcasts to the same port.
	// Since main.go already starts `server.StartUDP()` which listens on config.UdpPort (8001),
	// we have a conflict if we try to listen on 8001 here again.
	//
	// SOLUTION: We should integrate this logic into `server.StartUDP()` OR
	// modify `server.StartUDP()` to handle discovery messages.
	//
	// However, if the requirement is "CloudServerURL from received UDP",
	// it implies we are listening.
	//
	// Given the user instruction "local-server simultaneously acts as websocket client and server",
	// and "CloudServerURL obtained from received UDP",
	// let's assume `server.StartUDP` is the place receiving these messages.
	//
	// We will Modify `internal/server/udp_server.go` to handle Cloud Server discovery logic instead of here.
	// This file will only handle the broadcasting part if we want to keep it separate,
	// or we can move the broadcast logic to `server` package too.
	//
	// Let's keep `sendBroadcasts` here and call it.
	// But `listenForCloudServers` is redundant if `server.StartUDP` does the job.
	//
	// Wait, `server.StartUDP` listens on 8001.
	// The broadcast is sent to 8001.
	// So `server.StartUDP` WILL receive its own broadcast (which it ignores).
	// It will also receive broadcasts/responses from Cloud Server.
}
