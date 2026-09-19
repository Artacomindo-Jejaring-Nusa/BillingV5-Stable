package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"billing-backend/internal/domain"
	"billing-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatUsecase domain.ChatUsecase
}

func NewChatHandler(r *gin.RouterGroup, chatUsecase domain.ChatUsecase, authMiddleware gin.HandlerFunc) *ChatHandler {
	handler := &ChatHandler{chatUsecase: chatUsecase}

	chatGroup := r.Group("/chat")
	{
		// WebSocket connection
		chatGroup.GET("/ws", handler.HandleWebSocket)

		// Customer & Common APIs
		chatGroup.GET("/customer/room", handler.GetCustomerRoom)
		chatGroup.GET("/messages/:room_id", handler.GetMessages)
		chatGroup.POST("/messages/:room_id/read", handler.MarkRead)
		chatGroup.POST("/upload", handler.UploadMedia)

		// Admin APIs (protected by JWT middleware)
		chatGroup.GET("/rooms", authMiddleware, handler.ListRooms)
		chatGroup.POST("/rooms/:room_id/close", authMiddleware, handler.CloseRoom)
		chatGroup.POST("/rooms/:room_id/reopen", authMiddleware, handler.ReopenRoom)
	}

	return handler
}

// HandleWebSocket upgrades HTTP connection to bidirectional Chat WebSocket
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	if websocket.GlobalChatHub == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Chat WebSocket Hub is not initialized"})
		return
	}

	role := c.DefaultQuery("role", "customer")
	name := c.DefaultQuery("name", "Pelanggan")
	brand := c.DefaultQuery("brand", "Jakinet")

	var roomID uint64
	var senderID uint64
	var isListening bool

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if role == "customer" {
		pelangganIDStr := c.Query("pelanggan_id")
		pid, err := strconv.ParseUint(pelangganIDStr, 10, 64)
		if err != nil || pid == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pelanggan_id wajib diisi untuk role customer"})
			return
		}
		senderID = pid

		// Ambil atau buat chat room untuk pelanggan ini
		room, err := h.chatUsecase.GetCustomerRoom(ctx, pid, brand)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menginisialisasi chat room: " + err.Error()})
			return
		}
		roomID = room.ID
		if room.Pelanggan != nil && room.Pelanggan.Nama != "" {
			name = room.Pelanggan.Nama
		}
	} else {
		// Admin
		adminIDStr := c.DefaultQuery("user_id", "1")
		aid, _ := strconv.ParseUint(adminIDStr, 10, 64)
		senderID = aid

		roomIDStr := c.Query("room_id")
		if roomIDStr != "" {
			rid, _ := strconv.ParseUint(roomIDStr, 10, 64)
			roomID = rid
		} else {
			isListening = true // Admin mendengar seluruh room yang aktif
		}
	}

	conn, err := websocket.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[Chat WebSocket] Upgrade failed: %v", err)
		return
	}

	clientID := fmt.Sprintf("%s_%d_%d", role, senderID, time.Now().UnixNano())
	client := &websocket.ChatClient{
		ID:          clientID,
		Hub:         websocket.GlobalChatHub,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		RoomID:      roomID,
		Role:        role,
		SenderID:    senderID,
		SenderName:  name,
		IsListening: isListening,
	}

	client.Hub.RegisterClient(client)

	// Kirim info room saat pertama kali terkoneksi
	client.SendDirect("room_connected", map[string]interface{}{
		"room_id":   roomID,
		"role":      role,
		"sender_id": senderID,
		"name":      name,
		"brand":     brand,
		"connected": true,
	})

	go client.WritePump()
	go client.ReadPump()
}

// GetCustomerRoom mengambil atau membuat room berdasarkan pelanggan_id
func (h *ChatHandler) GetCustomerRoom(c *gin.Context) {
	pelangganIDStr := c.Query("pelanggan_id")
	pid, err := strconv.ParseUint(pelangganIDStr, 10, 64)
	if err != nil || pid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pelanggan_id tidak valid"})
		return
	}

	brand := c.DefaultQuery("brand", "Jakinet")

	room, err := h.chatUsecase.GetCustomerRoom(c.Request.Context(), pid, brand)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": room})
}

// GetMessages mengambil riwayat pesan dalam room
func (h *ChatHandler) GetMessages(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	rid, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil || rid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id tidak valid"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.chatUsecase.GetRoomMessages(c.Request.Context(), rid, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// MarkRead menandai pesan di room telah dibaca
func (h *ChatHandler) MarkRead(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	rid, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil || rid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id tidak valid"})
		return
	}

	readerType := c.DefaultQuery("reader_type", "customer")

	if err := h.chatUsecase.MarkRead(c.Request.Context(), rid, readerType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Beritahu juga via WebSocket
	if websocket.GlobalChatHub != nil {
		websocket.GlobalChatHub.BroadcastToRoom(rid, "room_read", map[string]interface{}{
			"room_id": rid,
			"read_by": readerType,
			"read_at": time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Pesan telah ditandai dibaca"})
}

// ListRooms mengambil daftar seluruh chat room untuk Admin CS
func (h *ChatHandler) ListRooms(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	brand := c.Query("brand")
	status := c.Query("status")
	search := c.Query("search")

	filter := domain.ChatRoomFilter{
		Brand:    brand,
		Status:   status,
		Search:   search,
		Page:     page,
		PageSize: pageSize,
	}

	rooms, total, err := h.chatUsecase.ListActiveRooms(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      rooms,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CloseRoom closes a chat room (marks conversation completed)
func (h *ChatHandler) CloseRoom(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	rid, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil || rid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id tidak valid"})
		return
	}

	if err := h.chatUsecase.UpdateRoomStatus(c.Request.Context(), rid, "closed"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if websocket.GlobalChatHub != nil {
		websocket.GlobalChatHub.BroadcastToRoom(rid, "room_status_update", map[string]interface{}{
			"room_id": rid,
			"status":  "closed",
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Percakapan telah ditutup"})
}

// ReopenRoom reopens a closed chat room
func (h *ChatHandler) ReopenRoom(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	rid, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil || rid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id tidak valid"})
		return
	}

	if err := h.chatUsecase.UpdateRoomStatus(c.Request.Context(), rid, "open"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if websocket.GlobalChatHub != nil {
		websocket.GlobalChatHub.BroadcastToRoom(rid, "room_status_update", map[string]interface{}{
			"room_id": rid,
			"status":  "open",
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Percakapan telah dibuka kembali"})
}

// UploadMedia handles file/image upload for chat messages
func (h *ChatHandler) UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada file yang diunggah"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File harus memiliki ekstensi yang valid"})
		return
	}

	uniqueFilename := fmt.Sprintf("chat_%d%s", time.Now().UnixNano(), ext)
	dir := "./uploads/chat"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat direktori upload"})
		return
	}

	filePath := filepath.Join(dir, uniqueFilename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file: " + err.Error()})
		return
	}

	fileInfo, err := os.Stat(filePath)
	var size int64
	if err == nil {
		size = fileInfo.Size()
	}

	contentType := file.Header.Get("Content-Type")
	fileURL := fmt.Sprintf("/static/uploads/chat/%s", uniqueFilename)

	c.JSON(http.StatusOK, gin.H{
		"file_url":     fileURL,
		"filename":     file.Filename,
		"content_type": contentType,
		"size":         size,
	})
}

