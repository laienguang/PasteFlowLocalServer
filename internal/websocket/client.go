package websocket

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"example.com/web-service/internal/config"
	"github.com/gorilla/websocket"
)

const (
	ReconnectCount = 3
)

type CloudClient struct {
	conn      *websocket.Conn
	hub       *Hub
	serverURL string
	send      chan []byte
	done      chan struct{}
	onFailure func()
}

func NewCloudClient(serverURL string, hub *Hub, onFailure func()) *CloudClient {
	return &CloudClient{
		serverURL: serverURL,
		hub:       hub,
		send:      make(chan []byte, 256),
		done:      make(chan struct{}),
		onFailure: onFailure,
	}
}

func (c *CloudClient) Connect() {
	// Reconnection loop
	go func() {
		retryCount := 0
		for {
			u := url.URL{Scheme: "ws", Host: c.serverURL, Path: "/ws"}
			log.Printf("Connecting to cloud: %s", u.String())

			header := http.Header{}
			header.Add("X-Client-ID", config.ClientID)

			conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
			if err != nil {
				retryCount++
				log.Printf("Cloud connection failed (attempt %d/%d): %v. Retrying in 5 seconds...", retryCount, ReconnectCount, err)
				if retryCount >= ReconnectCount {
					log.Printf("Max retries reached for %s. Removing client.", c.serverURL)
					if c.onFailure != nil {
						c.onFailure()
					}
					return
				}
				time.Sleep(5 * time.Second)
				continue
			}

			log.Println("Connected to cloud server")
			c.conn = conn
			retryCount = 0

			// Handle reading from cloud
			go c.readPump()
			// Handle writing to cloud
			c.writePump() // This blocks until disconnected

			log.Println("Disconnected from cloud server. Reconnecting...")
			time.Sleep(1 * time.Second)
		}
	}()
}

func (c *CloudClient) readPump() {
	defer func() {
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Cloud read error: %v", err)
			return
		}
		log.Printf("Received from cloud: %s", message)

		HandleMessage(message)
	}
}

func (c *CloudClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *CloudClient) Send(message []byte) {
	select {
	case c.send <- message:
	default:
		log.Println("Cloud send buffer full, dropping message")
	}
}
