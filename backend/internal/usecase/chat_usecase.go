package usecase

import (
	"context"
	"errors"
	"strings"
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

func (u *chatUsecase) GetRoomByID(ctx context.Context, roomID uint64) (*domain.ChatRoom, error) {
	if roomID == 0 {
		return nil, errors.New("room_id tidak valid")
	}
	return u.chatRepo.GetRoomByID(ctx, roomID)
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
	if msg == nil || msg.RoomID == 0 || (msg.Message == "" && msg.AttachmentURL == nil) {
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
	if preview == "" && (msg.MessageType == "image" || msg.AttachmentURL != nil) {
		preview = "[Gambar]"
	} else if preview == "" {
		preview = "[Lampiran]"
	}
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

func (u *chatUsecase) UpdateRoomStatus(ctx context.Context, roomID uint64, status string) error {
	if roomID == 0 {
		return errors.New("room_id tidak valid")
	}
	if status != "open" && status != "closed" {
		return errors.New("status harus 'open' atau 'closed'")
	}
	return u.chatRepo.UpdateRoomStatus(ctx, roomID, status)
}

func (u *chatUsecase) AssignRoom(ctx context.Context, roomID uint64, adminID uint64) error {
	if roomID == 0 {
		return errors.New("room_id tidak valid")
	}
	if adminID == 0 {
		return errors.New("admin_id tidak valid")
	}
	return u.chatRepo.AssignRoom(ctx, roomID, adminID)
}

func (u *chatUsecase) UnassignRoom(ctx context.Context, roomID uint64) error {
	if roomID == 0 {
		return errors.New("room_id tidak valid")
	}
	return u.chatRepo.UnassignRoom(ctx, roomID)
}

func (u *chatUsecase) GetRoomCounts(ctx context.Context, adminID uint64, brand string) (*domain.RoomCounts, error) {
	return u.chatRepo.GetRoomCounts(ctx, adminID, brand)
}

func (u *chatUsecase) ListTemplates(ctx context.Context) ([]domain.QuickReplyTemplate, error) {
	return u.chatRepo.ListTemplates(ctx)
}

func (u *chatUsecase) CreateTemplate(ctx context.Context, tpl *domain.QuickReplyTemplate) error {
	if tpl == nil || tpl.Shortcut == "" || tpl.Content == "" {
		return errors.New("shortcut dan konten template tidak boleh kosong")
	}
	// Normalisasi shortcut: hapus awalan slash jika ada, jadikan lowercase
	tpl.Shortcut = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(tpl.Shortcut)), "/")
	if tpl.Title == "" {
		tpl.Title = tpl.Shortcut
	}
	return u.chatRepo.CreateTemplate(ctx, tpl)
}

func (u *chatUsecase) UpdateTemplate(ctx context.Context, tpl *domain.QuickReplyTemplate) error {
	if tpl == nil || tpl.ID == 0 || tpl.Shortcut == "" || tpl.Content == "" {
		return errors.New("id, shortcut, dan konten template tidak boleh kosong")
	}
	tpl.Shortcut = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(tpl.Shortcut)), "/")
	if tpl.Title == "" {
		tpl.Title = tpl.Shortcut
	}
	return u.chatRepo.UpdateTemplate(ctx, tpl)
}

func (u *chatUsecase) DeleteTemplate(ctx context.Context, id uint64) error {
	if id == 0 {
		return errors.New("id template tidak valid")
	}
	return u.chatRepo.DeleteTemplate(ctx, id)
}

