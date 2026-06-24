package domain

import "context"

// NotificationRepository defines database operations for WhatsApp outbox messages.
type NotificationRepository interface {
	// CreateOutbox inserts a new outbox message with PENDING status.
	CreateOutbox(ctx context.Context, msg *WhatsAppOutbox) error

	// UpdateOutbox updates an existing outbox message (status, retry_count, etc).
	UpdateOutbox(ctx context.Context, msg *WhatsAppOutbox) error

	// GetPendingAndFailed retrieves messages that need to be sent or retried.
	// Returns messages with status PENDING or FAILED where retry_count < max_retries,
	// ordered by created_at ASC (oldest first).
	GetPendingAndFailed(ctx context.Context, limit int) ([]WhatsAppOutbox, error)

	// GetOutboxLogs retrieves outbox logs for a specific reference (for audit trail).
	GetOutboxLogs(ctx context.Context, refType string, refID uint64) ([]WhatsAppOutbox, error)
}

// NotificationUsecase defines the business logic for WhatsApp notification delivery.
type NotificationUsecase interface {
	// SendWhatsApp enqueues a WA message to the outbox and attempts immediate delivery.
	// If the immediate attempt fails, the message remains in the outbox for retry by the cron job.
	SendWhatsApp(ctx context.Context, phone, message string, refType string, refID uint64, createdBy *uint64) error

	// RetryFailedMessages processes all pending/failed messages that haven't exceeded max retries.
	// This is called by the scheduler cron job.
	RetryFailedMessages(ctx context.Context) error
}
