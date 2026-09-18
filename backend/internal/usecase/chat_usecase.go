package usecase

import (
	"context"
	"errors"
	"time"

	"billing-backend/internal/domain"
)

type chatUsecase struct {
	chatRepo domain.ChatRepository
}

func NewChatUsecase(chatRepo domain.ChatRepository) domain.ChatUsecase {
	return &chatUsecase{chatRepo: chatRepo}
}

func (u *chatUsecase) GetCustomerRoom(ctx context.Context, pelangganID uint64, brand string) (*domain.ChatRoom, error) {
	if pelangganID == 0 {
		return nil, errors.New("pelanggan_id tidak valid")
	}
	if brand == "" {
		brand = "Jakinet"
	}
	return u.chatRepo.GetOrCreateRoomByPelangganID(ctx, pelangganID, brand)
}

func (u *chatUsecase) GetRoomMessages(ctx context.Context, roomID uint64, limit, offset int) ([]domain.ChatMessage, error) {
	if roomID == 0 {
		return nil, errors.New("room_id tidak valid")
	}
	return u.chatRepo.GetMessagesByRoomID(ctx, roomID, limit, offset)
}

func (u *chatUsecase) ListActiveRooms(ctx context.Context, filter domain.ChatRoomFilter) ([]domain.ChatRoom, int64, error) {
	return u.chatRepo.ListRooms(ctx, filter)
}

func (u *chatUsecase) SendMessage(ctx context.Context, msg *domain.ChatMessage) (*domain.ChatMessage, error) {
	if msg == nil || msg.RoomID == 0 || msg.Message == "" {
		return nil, errors.New("pesan tidak boleh kosong")
	}

	msg.Status = domain.ChatStatusSent
	msg.CreatedAt = time.Now()

	if err := u.chatRepo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	// Update Room stats
	unreadAdmin := 0
	unreadCust := 0
	if msg.SenderType == "customer" {
		unreadAdmin = 1
	} else if msg.SenderType == "admin" {
		unreadCust = 1
	}

	preview := msg.Message
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}

	_ = u.chatRepo.UpdateRoomStats(ctx, msg.RoomID, preview, unreadAdmin, unreadCust)

	return msg, nil
}

func (u *chatUsecase) MarkDelivered(ctx context.Context, roomID uint64, recipientType string) error {
	if roomID == 0 {
		return nil
	}
	return u.chatRepo.MarkMessagesAsDelivered(ctx, roomID, recipientType)
}

func (u *chatUsecase) MarkRead(ctx context.Context, roomID uint64, readerType string) error {
	if roomID == 0 {
		return nil
	}
	if err := u.chatRepo.MarkMessagesAsRead(ctx, roomID, readerType); err != nil {
		return err
	}
	return u.chatRepo.ResetUnreadCount(ctx, roomID, readerType)
}
