package repository

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type chatRepository struct {
	db     *gorm.DB
	roomMu sync.Mutex
}

func NewChatRepository(db *gorm.DB) domain.ChatRepository {
	repo := &chatRepository{db: db}
	go repo.ConsolidateDuplicateRooms(context.Background())
	return repo
}

// ConsolidateDuplicateRooms mengonsolidasikan chat room ganda milik pelanggan yang sama
// (misal akibat pemisahan sesi 48 jam sebelumnya) agar seluruh histori pesan menyatu kembali.
func (r *chatRepository) ConsolidateDuplicateRooms(ctx context.Context) {
	type DupCust struct {
		PelangganID uint64 `gorm:"column:pelanggan_id"`
		Cnt         int    `gorm:"column:cnt"`
	}
	var dups []DupCust
	err := r.db.WithContext(ctx).
		Model(&domain.ChatRoom{}).
		Select("pelanggan_id, COUNT(*) as cnt").
		Group("pelanggan_id").
		Having("cnt > 1").
		Scan(&dups).Error
	if err != nil || len(dups) == 0 {
		return
	}

	for _, d := range dups {
		if d.PelangganID == 0 {
			continue
		}
		var rooms []domain.ChatRoom
		if err := r.db.WithContext(ctx).
			Where("pelanggan_id = ?", d.PelangganID).
			Order("last_message_at DESC, id DESC").
			Find(&rooms).Error; err != nil || len(rooms) <= 1 {
			continue
		}

		targetRoomID := rooms[0].ID
		var oldIDs []uint64
		for i := 1; i < len(rooms); i++ {
			oldIDs = append(oldIDs, rooms[i].ID)
		}

		// Pindahkan semua pesan historis ke room utama
		_ = r.db.WithContext(ctx).
			Model(&domain.ChatMessage{}).
			Where("room_id IN ?", oldIDs).
			Update("room_id", targetRoomID).Error

		// Hapus room duplikat yang sudah dikonsolidasikan
		_ = r.db.WithContext(ctx).
			Where("id IN ?", oldIDs).
			Delete(&domain.ChatRoom{}).Error
	}
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
	r.roomMu.Lock()
	defer r.roomMu.Unlock()

	var rooms []domain.ChatRoom
	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("Pelanggan.HargaLayanan").
		Preload("Pelanggan.Langganan").
		Where("pelanggan_id = ?", pelangganID).
		Order("last_message_at DESC, id DESC").
		Find(&rooms).Error

	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Jika pelanggan sudah memiliki obrolan (baik open maupun closed/resolved)
	if len(rooms) > 0 {
		latestRoom := rooms[0]

		// Jika ada duplikasi room, konsolidasikan seluruh pesan lama ke dalam latestRoom
		if len(rooms) > 1 {
			var dupIDs []uint64
			for i := 1; i < len(rooms); i++ {
				dupIDs = append(dupIDs, rooms[i].ID)
			}
			_ = r.db.WithContext(ctx).
				Model(&domain.ChatMessage{}).
				Where("room_id IN ?", dupIDs).
				Update("room_id", latestRoom.ID).Error

			_ = r.db.WithContext(ctx).
				Where("id IN ?", dupIDs).
				Delete(&domain.ChatRoom{}).Error
		}

		// Histori percakapan bersifat abadi (tidak dihapus atau dipisah setelah 2 hari)
		exact := detectExactBrand(latestRoom.Brand, latestRoom.Pelanggan)
		if latestRoom.Brand != exact {
			latestRoom.Brand = exact
			_ = r.db.WithContext(ctx).Model(&latestRoom).Update("brand", exact).Error
		}
		return &latestRoom, nil
	}

	// Deteksi brand presisi dari data pelanggan jika percakapan pertama kali dibuat
	var p domain.Pelanggan
	if err := r.db.WithContext(ctx).Preload("HargaLayanan").First(&p, pelangganID).Error; err == nil {
		brand = detectExactBrand(brand, &p)
	} else {
		brand = detectExactBrand(brand, nil)
	}

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

	if filter.Brand != "" && filter.Brand != "ALL" {
		query = query.Where("chat_rooms.brand = ?", filter.Brand)
	}
	if filter.Status != "" && filter.Status != "all" && filter.Status != "ALL" {
		query = query.Where("chat_rooms.status = ?", filter.Status)
	}

	// Assignment-based filtering
	switch filter.Assignment {
	case "unassigned":
		query = query.Where("chat_rooms.assigned_admin_id IS NULL OR chat_rooms.assigned_admin_id = 0")
		if filter.Status == "" {
			query = query.Where("chat_rooms.status = 'open'")
		}
	case "assigned":
		query = query.Where("chat_rooms.assigned_admin_id IS NOT NULL AND chat_rooms.assigned_admin_id > 0")
		if filter.Status == "" {
			query = query.Where("chat_rooms.status = 'open'")
		}
	case "mine":
		if filter.AdminID > 0 {
			query = query.Where("chat_rooms.assigned_admin_id = ?", filter.AdminID)
			if filter.Status == "" {
				query = query.Where("chat_rooms.status = 'open'")
			}
		}
	case "closed", "resolved":
		query = query.Where("chat_rooms.status = 'closed'")
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
		limit = 200
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

func (r *chatRepository) AssignRoom(ctx context.Context, roomID uint64, adminID uint64) error {
	return r.db.WithContext(ctx).
		Model(&domain.ChatRoom{}).
		Where("id = ?", roomID).
		Updates(map[string]interface{}{
			"assigned_admin_id": adminID,
			"status":            "open",
		}).Error
}

func (r *chatRepository) UnassignRoom(ctx context.Context, roomID uint64) error {
	return r.db.WithContext(ctx).
		Model(&domain.ChatRoom{}).
		Where("id = ?", roomID).
		Update("assigned_admin_id", nil).Error
}

func (r *chatRepository) GetRoomCounts(ctx context.Context, adminID uint64, brand string) (*domain.RoomCounts, error) {
	counts := &domain.RoomCounts{}
	brandCond := r.brandClause(brand)

	// All open
	r.db.WithContext(ctx).Model(&domain.ChatRoom{}).Where(brandCond).Where("status = 'open'").Count(&counts.All)
	// Unassigned & open
	r.db.WithContext(ctx).Model(&domain.ChatRoom{}).Where(brandCond).Where("status = 'open' AND (assigned_admin_id IS NULL OR assigned_admin_id = 0)").Count(&counts.Unassigned)
	// Assigned & open
	r.db.WithContext(ctx).Model(&domain.ChatRoom{}).Where(brandCond).Where("status = 'open' AND assigned_admin_id IS NOT NULL AND assigned_admin_id > 0").Count(&counts.Assigned)
	// Mine
	if adminID > 0 {
		r.db.WithContext(ctx).Model(&domain.ChatRoom{}).Where(brandCond).Where("status = 'open' AND assigned_admin_id = ?", adminID).Count(&counts.Mine)
	}
	// Closed / Resolved
	r.db.WithContext(ctx).Model(&domain.ChatRoom{}).Where(brandCond).Where("status = 'closed'").Count(&counts.Closed)

	return counts, nil
}

func (r *chatRepository) brandClause(brand string) string {
	if brand != "" && brand != "ALL" {
		return "brand = '" + brand + "'"
	}
	return "1 = 1"
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

