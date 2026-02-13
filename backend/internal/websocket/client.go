package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Client represents a websocket client
type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	send        chan []byte
	executionID string
}

// ClientMessage represents a message from the client
type ClientMessage struct {
	Type string `json:"type"`
	Data struct {
		StepID    string `json:"step_id,omitempty"`
		NewOutput string `json:"new_output,omitempty"`
	} `json:"data"`
}

// ServeWS handles websocket requests
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	executionID := r.PathValue("id")
	if executionID == "" {
		// Try to get from URL query for older Go versions
		executionID = r.URL.Query().Get("id")
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan []byte, 256),
		executionID: executionID,
	}
	hub.register <- client

	// Start read and write pumps
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Read error: %v", err)
			}
			break
		}

		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			log.Printf("[WebSocket] Invalid message format: %v", err)
			continue
		}

		// Handle client messages (for replay/modify functionality)
		c.handleMessage(clientMsg)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Batch queued messages
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

// handleMessage processes messages from the client
func (c *Client) handleMessage(msg ClientMessage) {
	switch msg.Type {
	case "modify_step":
		// Handle step modification during replay
		log.Printf("[WebSocket] Modify step request: step=%s", msg.Data.StepID)
		// This would trigger the replay engine to modify and recalculate

	case "ping":
		// Respond to ping
		response, _ := json.Marshal(map[string]string{"type": "pong"})
		c.send <- response

	default:
		log.Printf("[WebSocket] Unknown message type: %s", msg.Type)
	}
}
