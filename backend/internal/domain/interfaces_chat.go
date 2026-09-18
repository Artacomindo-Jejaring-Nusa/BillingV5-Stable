package domain

import "context"

// ChatRepository defines the contract for persisting chat rooms and messages.
type ChatRepository interface {
	GetOrCreateRoomByPelangganID(ctx context.Context, pelangganID uint64, brand string) (*ChatRoom, error)
	GetRoomByID(ctx context.Context, roomID uint64) (*ChatRoom, error)
	ListRooms(ctx context.Context, filter ChatRoomFilter) ([]ChatRoom, int64, error)
	SaveMessage(ctx context.Context, msg *ChatMessage) error
	GetMessagesByRoomID(ctx context.Context, roomID uint64, limit, offset int) ([]ChatMessage, error)
	UpdateMessageStatus(ctx context.Context, messageID uint64, status ChatMessageStatus) error
	MarkMessagesAsDelivered(ctx context.Context, roomID uint64, recipientType string) error
	MarkMessagesAsRead(ctx context.Context, roomID uint64, readerType string) error
	UpdateRoomStats(ctx context.Context, roomID uint64, lastMsg string, unreadDeltaAdmin, unreadDeltaCust int) error
	ResetUnreadCount(ctx context.Context, roomID uint64, readerType string) error
}

// ChatUsecase defines business operations for managing chat and status updates.
type ChatUsecase interface {
	GetCustomerRoom(ctx context.Context, pelangganID uint64, brand string) (*ChatRoom, error)
	GetRoomMessages(ctx context.Context, roomID uint64, limit, offset int) ([]ChatMessage, error)
	ListActiveRooms(ctx context.Context, filter ChatRoomFilter) ([]ChatRoom, int64, error)
	SendMessage(ctx context.Context, msg *ChatMessage) (*ChatMessage, error)
	MarkDelivered(ctx context.Context, roomID uint64, recipientType string) error
	MarkRead(ctx context.Context, roomID uint64, readerType string) error
}
