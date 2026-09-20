package domain

import (
	"time"
)

// ChatMessageStatus defines the delivery status of a chat message.
type ChatMessageStatus string

const (
	ChatStatusPending   ChatMessageStatus = "pending"   // 🕒 Saat pesan diproses di client
	ChatStatusSent      ChatMessageStatus = "sent"      // ✔️ Ceklis 1: Tersimpan di server DB
	ChatStatusDelivered ChatMessageStatus = "delivered" // ✔️✔️ Ceklis 2 Abu-abu: Diterima device lawan bicara
	ChatStatusRead      ChatMessageStatus = "read"      // ✔️✔️ Ceklis 2 Biru: Dibaca oleh lawan bicara
)

// ChatRoom represents an active or historical conversation session between Customer and Support.
type ChatRoom struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PelangganID         uint64     `gorm:"index;not null" json:"pelanggan_id"`
	Brand               string     `gorm:"type:varchar(50);not null;index" json:"brand"` // 'Jakinet' / 'Jelantik'
	Status              string     `gorm:"type:varchar(20);default:'open';index" json:"status"` // 'open', 'closed'
	AssignedAdminID     *uint64    `gorm:"index" json:"assigned_admin_id"`
	LastMessageAt       *time.Time `gorm:"type:datetime;index" json:"last_message_at"`
	LastMessageText     string     `gorm:"type:varchar(500)" json:"last_message_text"`
	UnreadCountAdmin    int        `gorm:"default:0" json:"unread_count_admin"`
	UnreadCountCustomer int        `gorm:"default:0" json:"unread_count_customer"`
	CreatedAt           time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`

	// Relational data
	Pelanggan     *Pelanggan `gorm:"foreignKey:PelangganID" json:"pelanggan,omitempty"`
	AssignedAdmin *User      `gorm:"foreignKey:AssignedAdminID" json:"assigned_admin,omitempty"`
}

func (ChatRoom) TableName() string {
	return "chat_rooms"
}

// ChatMessage represents a single message exchanged in a ChatRoom.
type ChatMessage struct {
	ID            uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	RoomID        uint64            `gorm:"index;not null" json:"room_id"`
	SenderType    string            `gorm:"type:varchar(20);not null;index" json:"sender_type"` // 'customer', 'admin', 'system'
	SenderID      uint64            `gorm:"index;not null" json:"sender_id"`
	SenderName    string            `gorm:"type:varchar(100);not null" json:"sender_name"`
	Message       string            `gorm:"type:text;not null" json:"message"`
	MessageType   string            `gorm:"type:varchar(20);default:'text'" json:"message_type"` // 'text', 'image', 'file'
	AttachmentURL *string           `gorm:"type:varchar(255)" json:"attachment_url,omitempty"`
	Status        ChatMessageStatus `gorm:"type:varchar(20);default:'sent';index;not null" json:"status"`
	DeliveredAt   *time.Time        `gorm:"type:datetime" json:"delivered_at,omitempty"`
	ReadAt        *time.Time        `gorm:"type:datetime" json:"read_at,omitempty"`
	TempID        string            `gorm:"-" json:"temp_id,omitempty"` // Temporary client-side UUID for optimistic UI
	CreatedAt     time.Time         `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index" json:"created_at"`

	// Relational data
	Room *ChatRoom `gorm:"foreignKey:RoomID" json:"-"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

// ChatEvent defines WebSocket event payloads.
type ChatEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// Filter parameter for Chat Rooms
type ChatRoomFilter struct {
	Brand    string `json:"brand"`
	Status   string `json:"status"`
	Search   string `json:"search"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// QuickReplyTemplate represents a canned response / shortcut for Live Chat
type QuickReplyTemplate struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Shortcut  string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"shortcut"` // e.g. "salam", "cek", "modem"
	Title     string    `gorm:"type:varchar(100);not null" json:"title"`              // e.g. "Salam Pembuka"
	Content   string    `gorm:"type:text;not null" json:"content"`                     // Message text
	Icon      string    `gorm:"type:varchar(50);default:'mdi-message-text-outline'" json:"icon"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
}

func (QuickReplyTemplate) TableName() string {
	return "quick_reply_templates"
}

