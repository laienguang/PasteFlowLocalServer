package main

import (
	"example.com/web-service/internal/discovery"
	"example.com/web-service/internal/lifecycle"
	"example.com/web-service/internal/logger"
	"example.com/web-service/internal/server"
	"example.com/web-service/internal/websocket"
)

func main() {
	logger.Setup()

	// 监听 Stdin，如果关闭（父进程退出），则自动退出
	lifecycle.WatchParentProcess()

	// Initialize WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize Client Manager
	clientManager := websocket.NewClientManager(hub)

	// Start Discovery (Broadcasting)
	// We don't need to listen here because server.StartUDP handles listening.
	// We pass clientManager to server.StartUDP instead.
	discovery.Start(clientManager)

	// Start UDP server in a goroutine (Handles discovery responses too)
	go server.StartUDP(clientManager)

	// Start HTTP server (blocking)
	server.StartHTTP(hub, clientManager)
}
