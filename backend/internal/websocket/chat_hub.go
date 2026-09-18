package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"billing-backend/internal/domain"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512 KB
)

type ChatClient struct {
	ID          string // Unique client session ID for deduplication
	Hub         *ChatHub
	Conn        *websocket.Conn
	Send        chan []byte
	RoomID      uint64
	Role        string // 'customer' or 'admin'
	SenderID    uint64 // PelangganID or UserID
	SenderName  string
	IsListening bool // if true, admin is listening to all rooms
}

type ChatHub struct {
	chatUsecase    domain.ChatUsecase
	rooms          map[uint64]map[*ChatClient]bool
	adminListeners map[*ChatClient]bool
	register       chan *ChatClient
	unregister     chan *ChatClient
	mu             sync.RWMutex
}

var GlobalChatHub *ChatHub

func NewChatHub(chatUsecase domain.ChatUsecase) *ChatHub {
	return &ChatHub{
		chatUsecase:    chatUsecase,
		rooms:          make(map[uint64]map[*ChatClient]bool),
		adminListeners: make(map[*ChatClient]bool),
		register:       make(chan *ChatClient),
		unregister:     make(chan *ChatClient),
	}
}

func (h *ChatHub) Run() {
	log.Println("[WebSocket ChatHub] Running...")
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if client.RoomID > 0 {
				if h.rooms[client.RoomID] == nil {
					h.rooms[client.RoomID] = make(map[*ChatClient]bool)
				}
				h.rooms[client.RoomID][client] = true
			}
			if client.Role == "admin" && client.IsListening {
				h.adminListeners[client] = true
			}
			h.mu.Unlock()

			// Saat customer atau admin connect ke room, tandai pesan yang belum terkirim sebagai 'delivered'
			if client.RoomID > 0 {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_ = h.chatUsecase.MarkDelivered(ctx, client.RoomID, client.Role)
				cancel()

				// Beritahu pihak lawan bicara bahwa pesan sudah sampai (delivered - ceklis 2 abu-abu)
				h.BroadcastToRoomExcept(client.RoomID, client, "message_status_update", map[string]interface{}{
					"room_id":      client.RoomID,
					"status":       domain.ChatStatusDelivered,
					"delivered_at": time.Now(),
				})
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if client.RoomID > 0 && h.rooms[client.RoomID] != nil {
				delete(h.rooms[client.RoomID], client)
				if len(h.rooms[client.RoomID]) == 0 {
					delete(h.rooms, client.RoomID)
				}
			}
			delete(h.adminListeners, client)
			close(client.Send)
			h.mu.Unlock()
		}
	}
}

func (h *ChatHub) RegisterClient(client *ChatClient) {
	h.register <- client
}

func (h *ChatHub) UnregisterClient(client *ChatClient) {
	h.unregister <- client
}

// BroadcastToRoom mengirim pesan ke seluruh peserta di dalam room tertentu dan publish ke Redis
func (h *ChatHub) BroadcastToRoom(roomID uint64, event string, data interface{}) {
	h.broadcastLocalInternal(roomID, event, data, "")
	PublishChatToRedis(roomID, event, data, "")
}

// BroadcastToRoomExcept mengirim pesan ke seluruh peserta di room kecuali pengirim dan publish ke Redis
func (h *ChatHub) BroadcastToRoomExcept(roomID uint64, except *ChatClient, event string, data interface{}) {
	var excID string
	if except != nil {
		excID = except.ID
	}
	h.broadcastLocalInternal(roomID, event, data, excID)
	PublishChatToRedis(roomID, event, data, excID)
}

// BroadcastLocal menyiarkan pesan ke klien lokal di node ini (digunakan saat menerima event dari Redis Pub/Sub)
func (h *ChatHub) BroadcastLocal(roomID uint64, event string, data interface{}, exceptClientID string) {
	h.broadcastLocalInternal(roomID, event, data, exceptClientID)
}

func (h *ChatHub) broadcastLocalInternal(roomID uint64, event string, data interface{}, exceptClientID string) {
	payload, err := json.Marshal(domain.ChatEvent{
		Event: event,
		Data:  data,
	})
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// Kirim ke peserta di dalam room
	if roomID > 0 {
		if clients, ok := h.rooms[roomID]; ok {
			for c := range clients {
				if exceptClientID != "" && c.ID == exceptClientID {
					continue
				}
				select {
				case c.Send <- payload:
				default:
				}
			}
		}
	}

	// Kirim juga ke admin listener global
	for admin := range h.adminListeners {
		if exceptClientID != "" && admin.ID == exceptClientID {
			continue
		}
		select {
		case admin.Send <- payload:
		default:
		}
	}
}

// SendDirectToClient mengirim langsung ke satu client spesifik
func (c *ChatClient) SendDirect(event string, data interface{}) {
	payload, err := json.Marshal(domain.ChatEvent{
		Event: event,
		Data:  data,
	})
	if err != nil {
		return
	}
	select {
	case c.Send <- payload:
	default:
	}
}

// ReadPump membaca pesan yang masuk dari socket client
func (c *ChatClient) ReadPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(messageBytes, &raw); err != nil {
			continue
		}

		event, _ := raw["event"].(string)
		data, _ := raw["data"].(map[string]interface{})

		switch event {
		case "send_message":
			// Pelanggan atau Admin mengirim pesan
			targetRoomID := c.RoomID
			if rID, ok := data["room_id"].(float64); ok && rID > 0 {
				targetRoomID = uint64(rID)
			}
			if targetRoomID == 0 {
				continue
			}

			msgText, _ := data["message"].(string)
			tempID, _ := data["temp_id"].(string)
			msgType, _ := data["message_type"].(string)
			if msgType == "" {
				msgType = "text"
			}
			var attachURL *string
			if urlStr, ok := data["attachment_url"].(string); ok && urlStr != "" {
				attachURL = &urlStr
			}

			if msgText == "" && attachURL == nil {
				continue
			}

			chatMsg := &domain.ChatMessage{
				RoomID:        targetRoomID,
				SenderType:    c.Role,
				SenderID:      c.SenderID,
				SenderName:    c.SenderName,
				Message:       msgText,
				MessageType:   msgType,
				AttachmentURL: attachURL,
				TempID:        tempID,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			savedMsg, err := c.Hub.chatUsecase.SendMessage(ctx, chatMsg)
			cancel()

			if err != nil {
				c.SendDirect("message_error", map[string]interface{}{
					"temp_id": tempID,
					"error":   err.Error(),
				})
				continue
			}

			// 1. Balas ke Pengirim: ACK SENT (Ceklis 1 Abu-abu)
			c.SendDirect("ack_sent", map[string]interface{}{
				"id":         savedMsg.ID,
				"temp_id":    tempID,
				"room_id":    savedMsg.RoomID,
				"status":     domain.ChatStatusSent,
				"created_at": savedMsg.CreatedAt,
			})

			// 2. Cek apakah lawan bicara sedang online di dalam room
			hasActivePeer := false
			c.Hub.mu.RLock()
			if clients, ok := c.Hub.rooms[targetRoomID]; ok {
				for peer := range clients {
					if peer != c {
						hasActivePeer = true
						break
					}
				}
			}
			c.Hub.mu.RUnlock()

			// Jika ada lawan bicara online, status otomatis naik ke 'delivered'
			if hasActivePeer {
				ctxDelivered, cancelDelivered := context.WithTimeout(context.Background(), 2*time.Second)
				_ = c.Hub.chatUsecase.MarkDelivered(ctxDelivered, targetRoomID, c.Role)
				cancelDelivered()

				savedMsg.Status = domain.ChatStatusDelivered
				now := time.Now()
				savedMsg.DeliveredAt = &now

				// Update pengirim ke Ceklis 2 Abu-abu
				c.SendDirect("message_status_update", map[string]interface{}{
					"id":           savedMsg.ID,
					"temp_id":      tempID,
					"status":       domain.ChatStatusDelivered,
					"delivered_at": savedMsg.DeliveredAt,
				})
			}

			// 3. Broadcast ke lawan bicara: NEW MESSAGE
			c.Hub.BroadcastToRoomExcept(targetRoomID, c, "new_message", savedMsg)

		case "ack_delivered":
			// Device lawan bicara mengonfirmasi telah menerima pesan (Ceklis 2 Abu-abu)
			targetRoomID := c.RoomID
			if rID, ok := data["room_id"].(float64); ok && rID > 0 {
				targetRoomID = uint64(rID)
			}
			msgIDFloat, ok := data["message_id"].(float64)
			if ok && targetRoomID > 0 {
				msgID := uint64(msgIDFloat)
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				_ = c.Hub.chatUsecase.MarkDelivered(ctx, targetRoomID, c.Role)
				cancel()

				c.Hub.BroadcastToRoom(targetRoomID, "message_status_update", map[string]interface{}{
					"id":           msgID,
					"status":       domain.ChatStatusDelivered,
					"delivered_at": time.Now(),
				})
			}

		case "read_room":
			// Pengguna sedang aktif membuka layar chatroom (Ceklis 2 Biru)
			targetRoomID := c.RoomID
			if rID, ok := data["room_id"].(float64); ok && rID > 0 {
				targetRoomID = uint64(rID)
			}
			if targetRoomID > 0 {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_ = c.Hub.chatUsecase.MarkRead(ctx, targetRoomID, c.Role)
				cancel()

				// Beritahu seluruh peserta bahwa semua pesan di room ini sudah DIBACA
				c.Hub.BroadcastToRoom(targetRoomID, "room_read", map[string]interface{}{
					"room_id": targetRoomID,
					"read_by": c.Role,
					"read_at": time.Now(),
				})
			}

		case "typing":
			// Live typing indicator
			targetRoomID := c.RoomID
			if rID, ok := data["room_id"].(float64); ok && rID > 0 {
				targetRoomID = uint64(rID)
			}
			if targetRoomID > 0 {
				isTyping, _ := data["is_typing"].(bool)
				c.Hub.BroadcastToRoomExcept(targetRoomID, c, "typing_indicator", map[string]interface{}{
					"room_id":   targetRoomID,
					"sender":    c.SenderName,
					"is_typing": isTyping,
				})
			}
		}
	}
}

// WritePump menulis pesan ke websocket connection
func (c *ChatClient) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Drain queued messages
			n := len(c.Send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
