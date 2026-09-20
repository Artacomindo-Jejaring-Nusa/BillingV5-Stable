package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) domain.ChatRepository {
	return &chatRepository{db: db}
}

func detectExactBrand(brandHint string, p *domain.Pelanggan) string {
	combined := strings.ToUpper(brandHint)
	if p != nil {
		if p.HargaLayanan != nil && p.HargaLayanan.Brand != "" {
			combined += " " + strings.ToUpper(p.HargaLayanan.Brand)
		}
		if p.BrandDefault != nil && *p.BrandDefault != "" {
			combined += " " + strings.ToUpper(*p.BrandDefault)
		}
		if p.IDBrand != nil && *p.IDBrand != "" {
			combined += " " + strings.ToUpper(*p.IDBrand)
		}
		combined += " " + strings.ToUpper(p.Alamat)
	}

	if strings.Contains(combined, "NAGRAK") {
		return "JELANTIK NAGRAK"
	}
	if strings.Contains(combined, "JELANTIK") || strings.Contains(combined, "AJN-02") {
		return "JELANTIK"
	}
	return "JAKINET"
}

func (r *chatRepository) GetOrCreateRoomByPelangganID(ctx context.Context, pelangganID uint64, brand string) (*domain.ChatRoom, error) {
	var room domain.ChatRoom
	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("Pelanggan.HargaLayanan").
		Preload("Pelanggan.Langganan").
		Where("pelanggan_id = ? AND status = ?", pelangganID, "open").
		First(&room).Error

	if err == nil {
		// Update brand jika ada identifikasi brand lebih akurat
		exact := detectExactBrand(room.Brand, room.Pelanggan)
		if room.Brand != exact {
			room.Brand = exact
			_ = r.db.WithContext(ctx).Model(&room).Update("brand", exact).Error
		}
		return &room, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Deteksi brand presisi dari data pelanggan
	var p domain.Pelanggan
	if err := r.db.WithContext(ctx).Preload("HargaLayanan").First(&p, pelangganID).Error; err == nil {
		brand = detectExactBrand(brand, &p)
	} else {
		brand = detectExactBrand(brand, nil)
	}

	// Buat room baru jika belum ada
	now := time.Now()
	newRoom := domain.ChatRoom{
		PelangganID:     pelangganID,
		Brand:           brand,
		Status:          "open",
		LastMessageAt:   &now,
		LastMessageText: "Obrolan dimulai",
	}

	if err := r.db.WithContext(ctx).Create(&newRoom).Error; err != nil {
		return nil, err
	}

	// Reload with relations
	_ = r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("Pelanggan.HargaLayanan").
		Preload("Pelanggan.Langganan").
		First(&newRoom, newRoom.ID)
	return &newRoom, nil
}

func (r *chatRepository) GetRoomByID(ctx context.Context, roomID uint64) (*domain.ChatRoom, error) {
	var room domain.ChatRoom
	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("Pelanggan.HargaLayanan").
		Preload("Pelanggan.Langganan").
		Preload("AssignedAdmin").
		First(&room, roomID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

func (r *chatRepository) ListRooms(ctx context.Context, filter domain.ChatRoomFilter) ([]domain.ChatRoom, int64, error) {
	var rooms []domain.ChatRoom
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.ChatRoom{})

	if filter.Brand != "" {
		query = query.Where("brand = ?", filter.Brand)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Search != "" {
		query = query.Joins("LEFT JOIN pelanggan ON pelanggan.id = chat_rooms.pelanggan_id").
			Where("pelanggan.nama LIKE ? OR pelanggan.no_telp LIKE ? OR pelanggan.email LIKE ?",
				"%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	err := query.
		Preload("Pelanggan").
		Preload("Pelanggan.HargaLayanan").
		Preload("Pelanggan.Langganan").
		Preload("AssignedAdmin").
		Order("chat_rooms.last_message_at DESC").
		Limit(filter.PageSize).
		Offset(offset).
		Find(&rooms).Error

	if err != nil {
		return nil, 0, err
	}

	return rooms, total, nil
}

func (r *chatRepository) SaveMessage(ctx context.Context, msg *domain.ChatMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *chatRepository) GetMessagesByRoomID(ctx context.Context, roomID uint64, limit, offset int) ([]domain.ChatMessage, error) {
	var messages []domain.ChatMessage
	if limit <= 0 {
		limit = 50
	}

	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *chatRepository) UpdateMessageStatus(ctx context.Context, messageID uint64, status domain.ChatMessageStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	now := time.Now()
	if status == domain.ChatStatusDelivered {
		updates["delivered_at"] = &now
	} else if status == domain.ChatStatusRead {
		updates["read_at"] = &now
	}

	return r.db.WithContext(ctx).
		Model(&domain.ChatMessage{}).
		Where("id = ?", messageID).
		Updates(updates).Error
}

func (r *chatRepository) MarkMessagesAsDelivered(ctx context.Context, roomID uint64, recipientType string) error {
	now := time.Now()
	// Ubah semua pesan yang dikirim oleh pihak lain dan masih 'sent' menjadi 'delivered'
	return r.db.WithContext(ctx).
		Model(&domain.ChatMessage{}).
		Where("room_id = ? AND sender_type != ? AND status = ?", roomID, recipientType, domain.ChatStatusSent).
		Updates(map[string]interface{}{
			"status":       domain.ChatStatusDelivered,
			"delivered_at": &now,
		}).Error
}

func (r *chatRepository) MarkMessagesAsRead(ctx context.Context, roomID uint64, readerType string) error {
	now := time.Now()
	// Ubah semua pesan yang dikirim oleh pihak lain dan belum 'read' menjadi 'read'
	return r.db.WithContext(ctx).
		Model(&domain.ChatMessage{}).
		Where("room_id = ? AND sender_type != ? AND status IN ?", roomID, readerType, []domain.ChatMessageStatus{domain.ChatStatusSent, domain.ChatStatusDelivered}).
		Updates(map[string]interface{}{
			"status":  domain.ChatStatusRead,
			"read_at": &now,
		}).Error
}

func (r *chatRepository) UpdateRoomStats(ctx context.Context, roomID uint64, lastMsg string, unreadDeltaAdmin, unreadDeltaCust int) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_message_at":   &now,
		"last_message_text": lastMsg,
		"status":            "open", // Reopen automatically if room was closed
	}

	if unreadDeltaAdmin != 0 {
		updates["unread_count_admin"] = gorm.Expr("unread_count_admin + ?", unreadDeltaAdmin)
	}
	if unreadDeltaCust != 0 {
		updates["unread_count_customer"] = gorm.Expr("unread_count_customer + ?", unreadDeltaCust)
	}

	return r.db.WithContext(ctx).
		Model(&domain.ChatRoom{}).
		Where("id = ?", roomID).
		Updates(updates).Error
}

func (r *chatRepository) UpdateRoomStatus(ctx context.Context, roomID uint64, status string) error {
	return r.db.WithContext(ctx).
		Model(&domain.ChatRoom{}).
		Where("id = ?", roomID).
		Update("status", status).Error
}

func (r *chatRepository) ResetUnreadCount(ctx context.Context, roomID uint64, readerType string) error {
	column := "unread_count_customer"
	if readerType == "admin" {
		column = "unread_count_admin"
	}

	return r.db.WithContext(ctx).
		Model(&domain.ChatRoom{}).
		Where("id = ?", roomID).
		Update(column, 0).Error
}

func (r *chatRepository) ListTemplates(ctx context.Context) ([]domain.QuickReplyTemplate, error) {
	var count int64
	r.db.WithContext(ctx).Model(&domain.QuickReplyTemplate{}).Count(&count)
	if count == 0 {
		// Auto-seed default templates
		defaults := []domain.QuickReplyTemplate{
			{
				Shortcut:  "salam",
				Title:     "Salam",
				Content:   "Halo, selamat datang di layanan Customer Care Artacom. Ada yang bisa kami bantu?",
				Icon:      "mdi-hand-wave-outline",
				SortOrder: 1,
			},
			{
				Shortcut:  "cekteknis",
				Title:     "Cek Teknis",
				Content:   "Baik pak/bu, mohon ditunggu sebentar ya. Sedang kami lakukan pengecekan ke tim teknis lapangan.",
				Icon:      "mdi-wrench-clock-outline",
				SortOrder: 2,
			},
			{
				Shortcut:  "lunas",
				Title:     "Lunas & Aktif",
				Content:   "Terima kasih atas konfirmasinya. Tagihan Anda telah terverifikasi dan layanan internet sudah aktif normal kembali.",
				Icon:      "mdi-check-decagram-outline",
				SortOrder: 3,
			},
			{
				Shortcut:  "fotomodem",
				Title:     "Foto Modem",
				Content:   "Bisa tolong difotokan lampu indikator (PON / LOS / Internet) yang menyala pada perangkat modem router Anda?",
				Icon:      "mdi-camera-outline",
				SortOrder: 4,
			},
			{
				Shortcut:  "restart",
				Title:     "Restart Modem",
				Content:   "Bisa dicoba untuk mematikan modem router selama 1-2 menit, lalu hidupkan kembali dan periksa koneksinya?",
				Icon:      "mdi-restart",
				SortOrder: 5,
			},
			{
				Shortcut:  "tutup",
				Title:     "Tutup & Terima Kasih",
				Content:   "Terima kasih telah menghubungi Customer Care Artacom. Jika tidak ada hal lain yang ditanyakan, percakapan ini akan kami tutup. Selamat beraktivitas!",
				Icon:      "mdi-hand-heart-outline",
				SortOrder: 6,
			},
		}
		for _, d := range defaults {
			_ = r.db.WithContext(ctx).Create(&d).Error
		}
	}

	var list []domain.QuickReplyTemplate
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *chatRepository) CreateTemplate(ctx context.Context, tpl *domain.QuickReplyTemplate) error {
	return r.db.WithContext(ctx).Create(tpl).Error
}

func (r *chatRepository) UpdateTemplate(ctx context.Context, tpl *domain.QuickReplyTemplate) error {
	return r.db.WithContext(ctx).Save(tpl).Error
}

func (r *chatRepository) DeleteTemplate(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.QuickReplyTemplate{}, id).Error
}

