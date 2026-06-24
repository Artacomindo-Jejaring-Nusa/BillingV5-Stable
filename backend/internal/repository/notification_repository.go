package repository

import (
	"context"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) CreateOutbox(ctx context.Context, msg *domain.WhatsAppOutbox) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *notificationRepository) UpdateOutbox(ctx context.Context, msg *domain.WhatsAppOutbox) error {
	return r.db.WithContext(ctx).Save(msg).Error
}

func (r *notificationRepository) GetPendingAndFailed(ctx context.Context, limit int) ([]domain.WhatsAppOutbox, error) {
	var messages []domain.WhatsAppOutbox
	err := r.db.WithContext(ctx).
		Where("status IN ? AND retry_count < max_retries", []string{
			string(domain.WAStatusPending),
			string(domain.WAStatusFailed),
		}).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *notificationRepository) GetOutboxLogs(ctx context.Context, refType string, refID uint64) ([]domain.WhatsAppOutbox, error) {
	var messages []domain.WhatsAppOutbox
	err := r.db.WithContext(ctx).
		Where("ref_type = ? AND ref_id = ?", refType, refID).
		Order("created_at DESC").
		Find(&messages).Error
	return messages, err
}
