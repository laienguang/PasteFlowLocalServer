package server

import (
	"fmt"
	"log"
	"net"
	"example.com/web-service/internal/config"
)

// StartUDP starts the UDP server.
func StartUDP() {
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

		msg := string(buf[:n])
		log.Printf("UDP server got: %s from %s", msg, remoteAddr.String())

		if remoteAddr.Port == config.UdpPort {
			// In Go, ReadFromUDP returns the remote address.
			// If we receive a message from ourselves (e.g. broadcast or explicit send to localhost),
			// checking the port might help if we sent it from the bound port.
			// However, usually outgoing UDP packets use a random ephemeral port unless explicitly bound.
			// The Node.js logic checks if rinfo.port === udpPort.
			// We'll keep this check to match the logic, though it might only trigger if the sender bound to 8001.
			log.Println("Ignoring message from self")
			continue
		}

		response := []byte(fmt.Sprintf("Echo: %s", msg))
		_, err = conn.WriteToUDP(response, remoteAddr)
		if err != nil {
			log.Printf("UDP send error: %v", err)
		} else {
			log.Printf("UDP response sent to %s", remoteAddr.String())
		}
	}
}
