package usecase

import (
	"context"
	"fmt"
	"time"

	"billing-backend/config"
	"billing-backend/internal/domain"
	"billing-backend/pkg/logger"
	"billing-backend/pkg/utils"
)

type notificationUsecase struct {
	repo       domain.NotificationRepository
	systemRepo domain.SystemRepository
	cfg        *config.Config
}

// NewNotificationUsecase creates a new notification usecase
func NewNotificationUsecase(repo domain.NotificationRepository, sr domain.SystemRepository, cfg *config.Config) domain.NotificationUsecase {
	return &notificationUsecase{
		repo:       repo,
		systemRepo: sr,
		cfg:        cfg,
	}
}

// SendWhatsApp enqueues a WA message to the outbox and attempts immediate delivery.
// Pattern: Write-first, then send (Transactional Outbox).
func (u *notificationUsecase) SendWhatsApp(ctx context.Context, phone, message string, refType string, refID uint64, createdBy *uint64) error {
	// 1. Create outbox record (PENDING)
	outbox := &domain.WhatsAppOutbox{
		PhoneNo:    phone,
		Message:    message,
		Status:     domain.WAStatusPending,
		RetryCount: 0,
		MaxRetries: 5,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if refType != "" {
		outbox.RefType = &refType
	}
	if refID > 0 {
		outbox.RefID = &refID
	}
	if createdBy != nil {
		outbox.CreatedBy = createdBy
	}

	if err := u.repo.CreateOutbox(ctx, outbox); err != nil {
		logger.Error("NotificationUsecase: Failed to create outbox record: %v", err)
		return fmt.Errorf("failed to enqueue WA message: %w", err)
	}

	logger.Info("NotificationUsecase: Enqueued WA message ID=%d to %s", outbox.ID, phone)

	// 2. Optimistic send — attempt immediate delivery
	u.attemptSend(ctx, outbox)

	return nil
}

// RetryFailedMessages processes all pending/failed messages.
// Called by the scheduler cron job every 5 minutes.
func (u *notificationUsecase) RetryFailedMessages(ctx context.Context) error {
	messages, err := u.repo.GetPendingAndFailed(ctx, 50) // Process up to 50 messages per run
	if err != nil {
		logger.Error("NotificationUsecase: Failed to query pending/failed messages: %v", err)
		return fmt.Errorf("failed to query outbox: %w", err)
	}

	if len(messages) == 0 {
		return nil
	}

	logger.Info("NotificationUsecase: Processing %d pending/failed WA messages", len(messages))

	successCount := 0
	failCount := 0

	for i := range messages {
		msg := &messages[i]
		u.attemptSend(ctx, msg)

		if msg.Status == domain.WAStatusSuccess {
			successCount++
		} else {
			failCount++
		}

		// Small delay between sends to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	logger.Info("NotificationUsecase: Retry complete — %d success, %d failed", successCount, failCount)
	return nil
}

// attemptSend tries to send a single outbox message via the Watzap API.
func (u *notificationUsecase) attemptSend(ctx context.Context, msg *domain.WhatsAppOutbox) {
	// Mark as SENDING
	msg.Status = domain.WAStatusSending
	msg.UpdatedAt = time.Now()
	_ = u.repo.UpdateOutbox(ctx, msg)

	apiKey := u.cfg.WatzapApiKey
	numberKey := u.cfg.WatzapNumberKey

	// Fetch dynamic credentials from database settings if they exist
	if u.systemRepo != nil {
		if setting, err := u.systemRepo.GetSettingByKey(ctx, "WATZAP_API_KEY"); err == nil && setting != nil && setting.SettingValue != nil && *setting.SettingValue != "" {
			apiKey = *setting.SettingValue
		}
		if setting, err := u.systemRepo.GetSettingByKey(ctx, "WATZAP_NUMBER_KEY"); err == nil && setting != nil && setting.SettingValue != nil && *setting.SettingValue != "" {
			numberKey = *setting.SettingValue
		}
	}

	// Attempt to send
	err := utils.SendWhatsAppMessage(apiKey, numberKey, msg.PhoneNo, msg.Message)

	if err == nil {
		// Success!
		now := time.Now()
		msg.Status = domain.WAStatusSuccess
		msg.SentAt = &now
		msg.UpdatedAt = now
		if updateErr := u.repo.UpdateOutbox(ctx, msg); updateErr != nil {
			logger.Error("NotificationUsecase: Failed to update outbox ID=%d to SUCCESS: %v", msg.ID, updateErr)
		}
		logger.Info("NotificationUsecase: WA message ID=%d sent successfully to %s", msg.ID, msg.PhoneNo)
		return
	}

	// Failed
	msg.RetryCount++
	errStr := err.Error()
	msg.LastError = &errStr
	msg.UpdatedAt = time.Now()

	if msg.RetryCount >= msg.MaxRetries {
		msg.Status = domain.WAStatusAbandoned
		logger.Error("NotificationUsecase: WA message ID=%d ABANDONED after %d retries: %v", msg.ID, msg.RetryCount, err)
	} else {
		msg.Status = domain.WAStatusFailed
		logger.Warn("NotificationUsecase: WA message ID=%d FAILED (attempt %d/%d): %v", msg.ID, msg.RetryCount, msg.MaxRetries, err)
	}

	if updateErr := u.repo.UpdateOutbox(ctx, msg); updateErr != nil {
		logger.Error("NotificationUsecase: Failed to update outbox ID=%d status: %v", msg.ID, updateErr)
	}
}
