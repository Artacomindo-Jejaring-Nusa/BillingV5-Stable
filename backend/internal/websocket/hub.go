package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin (e.g. Vue dev server)
	},
}

var GlobalHub *Hub

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	log.Println("[WebSocket Hub] Running...")
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			log.Println("[WebSocket Hub] Client registered successfully")
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Println("[WebSocket Hub] Client unregistered")
		case message := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.Clients, client)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastNotification(notifType string, data map[string]interface{}) {
	sound := ""
	switch notifType {
	case "new_payment":
		sound = "/api/v1/notifications/sounds/pembayaran_selesai.mp3"
	case "new_customer", "new_customer_for_noc":
		sound = "/api/v1/notifications/sounds/new_pelanggan.mp3"
	case "new_technical_data":
		sound = "/api/v1/notifications/sounds/noc_finance.mp3"
	}

	payload := map[string]interface{}{
		"type":      notifType,
		"timestamp": time.Now().Format(time.RFC3339),
		"data":      data,
		"sound":     sound,
		"_source":   instanceID,
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[WebSocket Hub] Failed to marshal notification payload: %v", err)
		return
	}
	h.Broadcast <- bytes

	PublishToRedis(payload)
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(512)
	// Simple ping-pong or read loop to detect disconnection
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket Client] Error: %v", err)
			}
			break
		}
		// Handle ping frames sent as text from client
		if string(message) == "ping" {
			c.Send <- []byte("pong")
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		message, ok := <-c.Send
		if !ok {
			// Hub closed the channel
			_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("[WebSocket Client] Failed to write message: %v", err)
			return
		}
	}
}
