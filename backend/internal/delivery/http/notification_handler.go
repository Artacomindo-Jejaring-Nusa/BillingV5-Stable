package http

import (
	"log"
	"net/http"

	"billing-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct{}

func NewNotificationHandler(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	notificationGroup := r.Group("/notifications")
	{
		// These might need auth, but for sound files we can keep it public or use the group
		notificationGroup.GET("/unread", authMiddleware, GetUnread)
		notificationGroup.POST("/:id/mark-as-read", authMiddleware, MarkAsRead)
		notificationGroup.POST("/mark-all-as-read", authMiddleware, MarkAllAsRead)
		
		// Route to serve notification sounds
		notificationGroup.StaticFS("/sounds", gin.Dir("./Notif Sound", true))
	}
}

func GetUnread(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"notifications": []interface{}{},
	})
}

func MarkAsRead(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func MarkAllAsRead(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

// HandleWebSocket provides a real WebSocket connection upgrade handler using gorilla/websocket.
func HandleWebSocket(c *gin.Context) {
	if websocket.GlobalHub == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket Hub is not initialized"})
		return
	}

	conn, err := websocket.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade failed: %v", err)
		return
	}

	client := &websocket.Client{
		Hub:  websocket.GlobalHub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	client.Hub.Register <- client

	// Start read/write pumps in separate goroutines
	go client.WritePump()
	go client.ReadPump()
}
