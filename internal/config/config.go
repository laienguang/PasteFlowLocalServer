package config

import "github.com/google/uuid"

const (
	HttpPort = 8000
	UdpPort  = 8001
)

var (
	ClientID = uuid.New().String()
)
