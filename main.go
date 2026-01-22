package main

import (
	"example.com/web-service/internal/lifecycle"
	"example.com/web-service/internal/logger"
	"example.com/web-service/internal/server"
)

func main() {
	logger.Setup()

	// 监听 Stdin，如果关闭（父进程退出），则自动退出
	lifecycle.WatchParentProcess()

	// Start UDP server in a goroutine
	go server.StartUDP()

	// Start HTTP server (blocking)
	server.StartHTTP()
}
