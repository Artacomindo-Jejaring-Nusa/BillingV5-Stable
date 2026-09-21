package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"billing-backend/internal/domain"
	"billing-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatUsecase          domain.ChatUsecase
	troubleTicketUsecase domain.TroubleTicketUsecase
}

func NewChatHandler(r *gin.RouterGroup, chatUsecase domain.ChatUsecase, troubleTicketUsecase domain.TroubleTicketUsecase, authMiddleware gin.HandlerFunc) *ChatHandler {
	handler := &ChatHandler{
		chatUsecase:          chatUsecase,
		troubleTicketUsecase: troubleTicketUsecase,
	}

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
		chatGroup.GET("/rooms/counts", authMiddleware, handler.GetRoomCounts)
		chatGroup.POST("/rooms/:room_id/close", authMiddleware, handler.CloseRoom)
		chatGroup.POST("/rooms/:room_id/reopen", authMiddleware, handler.ReopenRoom)
		chatGroup.POST("/rooms/:room_id/assign", authMiddleware, handler.AssignRoom)
		chatGroup.POST("/rooms/:room_id/unassign", authMiddleware, handler.UnassignRoom)
		chatGroup.POST("/rooms/:room_id/auto-fill-ticket", authMiddleware, handler.AutoFillTicket)

		// Quick Reply Templates
		chatGroup.GET("/templates", handler.ListTemplates)
		chatGroup.POST("/templates", authMiddleware, handler.CreateTemplate)
		chatGroup.PUT("/templates/:id", authMiddleware, handler.UpdateTemplate)
		chatGroup.DELETE("/templates/:id", authMiddleware, handler.DeleteTemplate)
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
	assignment := c.Query("assignment")
	adminID, _ := strconv.ParseUint(c.DefaultQuery("admin_id", "0"), 10, 64)

	filter := domain.ChatRoomFilter{
		Brand:      brand,
		Status:     status,
		Search:     search,
		Assignment: assignment,
		AdminID:    adminID,
		Page:       page,
		PageSize:   pageSize,
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

// ListTemplates returns all quick reply templates
func (h *ChatHandler) ListTemplates(c *gin.Context) {
	templates, err := h.chatUsecase.ListTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar template: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": templates})
}

type CreateTemplateRequest struct {
	Shortcut  string `json:"shortcut" binding:"required"`
	Title     string `json:"title"`
	Content   string `json:"content" binding:"required"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
}

// CreateTemplate creates a new quick reply template
func (h *ChatHandler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: " + err.Error()})
		return
	}

	tpl := domain.QuickReplyTemplate{
		Shortcut:  req.Shortcut,
		Title:     req.Title,
		Content:   req.Content,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
	}

	if err := h.chatUsecase.CreateTemplate(c.Request.Context(), &tpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan template: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Template berhasil dibuat", "data": tpl})
}

// UpdateTemplate updates an existing quick reply template
func (h *ChatHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID template tidak valid"})
		return
	}

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: " + err.Error()})
		return
	}

	tpl := domain.QuickReplyTemplate{
		ID:        id,
		Shortcut:  req.Shortcut,
		Title:     req.Title,
		Content:   req.Content,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
	}

	if err := h.chatUsecase.UpdateTemplate(c.Request.Context(), &tpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui template: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template berhasil diperbarui", "data": tpl})
}

// DeleteTemplate deletes a quick reply template
func (h *ChatHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID template tidak valid"})
		return
	}

	if err := h.chatUsecase.DeleteTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus template: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template berhasil dihapus"})
}

// AssignRoom assigns a chat room to a specific admin/CS agent
func (h *ChatHandler) AssignRoom(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	rid, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil || rid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id tidak valid"})
		return
	}

	var req struct {
		AdminID uint64 `json:"admin_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin_id wajib diisi"})
		return
	}

	if err := h.chatUsecase.AssignRoom(c.Request.Context(), rid, req.AdminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload room with admin relation for broadcast
	room, _ := h.chatUsecase.GetRoomByID(c.Request.Context(), rid)

	if websocket.GlobalChatHub != nil {
		websocket.GlobalChatHub.BroadcastToRoom(rid, "room_assigned", map[string]interface{}{
			"room_id":  rid,
			"admin_id": req.AdminID,
			"room":     room,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Room berhasil di-assign", "data": room})
}

// UnassignRoom removes the admin assignment from a chat room
func (h *ChatHandler) UnassignRoom(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	rid, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil || rid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id tidak valid"})
		return
	}

	if err := h.chatUsecase.UnassignRoom(c.Request.Context(), rid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if websocket.GlobalChatHub != nil {
		websocket.GlobalChatHub.BroadcastToRoom(rid, "room_unassigned", map[string]interface{}{
			"room_id": rid,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Room telah di-unassign"})
}

// GetRoomCounts returns aggregate counts for inbox filter tabs
func (h *ChatHandler) GetRoomCounts(c *gin.Context) {
	adminID, _ := strconv.ParseUint(c.DefaultQuery("admin_id", "0"), 10, 64)
	brand := c.Query("brand")

	counts, err := h.chatUsecase.GetRoomCounts(c.Request.Context(), adminID, brand)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": counts})
}

// AutoFillTicket analyzes recent chat messages and auto-fills trouble ticket fields
func (h *ChatHandler) AutoFillTicket(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	rid, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil || rid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id tidak valid"})
		return
	}

	// Ambil room data untuk info pelanggan
	room, err := h.chatUsecase.GetRoomByID(c.Request.Context(), rid)
	if err != nil || room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room tidak ditemukan"})
		return
	}

	// Ambil riwayat pesan terakhir dari customer
	messages, err := h.chatUsecase.GetRoomMessages(c.Request.Context(), rid, 20, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil riwayat pesan"})
		return
	}

	// Gabungkan pesan dari customer untuk analisis
	var customerTexts []string
	for _, msg := range messages {
		if msg.SenderType == "customer" && msg.Message != "" {
			customerTexts = append(customerTexts, msg.Message)
		}
	}
	combined := strings.ToLower(strings.Join(customerTexts, " "))

	// Deteksi kategori berdasarkan keyword matching
	category := detectCategory(combined)
	priority := detectPriority(combined, category)
	title := detectTitle(combined, category)
	description := buildDescription(customerTexts)

	result := gin.H{
		"pelanggan_id": room.PelangganID,
		"category":     category,
		"priority":     priority,
		"title":        title,
		"description":  description,
	}

	if room.Pelanggan != nil {
		result["pelanggan_nama"] = room.Pelanggan.Nama
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func detectCategory(text string) string {
	categoryRules := []struct {
		keywords []string
		category string
	}{
		{[]string{"los", "merah", "lampu merah", "signal", "sinyal"}, "no_connection"},
		{[]string{"putus", "kabel", "potong", "kabel putus"}, "cable_issue"},
		{[]string{"lemot", "lambat", "lelet", "slow", "lag", "pelan"}, "slow_connection"},
		{[]string{"hidup mati", "kadang", "naik turun", "intermittent", "putus nyambung"}, "intermittent"},
		{[]string{"modem", "router", "onu", "device"}, "onu_issue"},
		{[]string{"olt", "server"}, "olt_issue"},
		{[]string{"mikrotik", "pppoe", "login"}, "mikrotik_issue"},
		{[]string{"hardware", "perangkat", "alat"}, "hardware_issue"},
	}

	for _, rule := range categoryRules {
		for _, kw := range rule.keywords {
			if strings.Contains(text, kw) {
				return rule.category
			}
		}
	}
	return "other"
}

func detectPriority(text string, category string) string {
	highKeywords := []string{"mati total", "tidak bisa", "ga bisa", "urgent", "darurat", "los merah"}
	for _, kw := range highKeywords {
		if strings.Contains(text, kw) {
			return "high"
		}
	}
	if category == "no_connection" || category == "cable_issue" {
		return "high"
	}
	if category == "slow_connection" || category == "intermittent" {
		return "medium"
	}
	return "low"
}

func detectTitle(text string, category string) string {
	titleMap := map[string]string{
		"no_connection":   "Internet Mati Total / Tidak Ada Koneksi",
		"cable_issue":     "Kabel Fiber Optic Bermasalah",
		"slow_connection": "Koneksi Internet Lambat",
		"intermittent":    "Koneksi Internet Tidak Stabil / Putus-Nyambung",
		"onu_issue":       "Perangkat Modem/ONU Bermasalah",
		"olt_issue":       "Gangguan Perangkat OLT/Server",
		"mikrotik_issue":  "Kendala Konfigurasi Mikrotik/PPPoE",
		"hardware_issue":  "Kerusakan Perangkat/Hardware",
	}

	if t, ok := titleMap[category]; ok {
		return t
	}
	return "Kendala Layanan Internet"
}

func buildDescription(customerTexts []string) string {
	if len(customerTexts) == 0 {
		return ""
	}
	// Ambil maksimal 5 pesan terakhir yang relevan
	start := 0
	if len(customerTexts) > 5 {
		start = len(customerTexts) - 5
	}
	var lines []string
	for _, t := range customerTexts[start:] {
		trimmed := strings.TrimSpace(t)
		if len(trimmed) > 200 {
			trimmed = trimmed[:200] + "..."
		}
		lines = append(lines, "- "+trimmed)
	}
	return "Laporan pelanggan via Live Chat:\n" + strings.Join(lines, "\n")
}
