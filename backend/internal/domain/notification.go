package domain

import (
	"time"
)

// WAOutboxStatus represents the delivery status of a WhatsApp outbox message.
type WAOutboxStatus string

const (
	WAStatusPending   WAOutboxStatus = "PENDING"
	WAStatusSending   WAOutboxStatus = "SENDING"
	WAStatusSuccess   WAOutboxStatus = "SUCCESS"
	WAStatusFailed    WAOutboxStatus = "FAILED"
	WAStatusAbandoned WAOutboxStatus = "ABANDONED"
)

// WhatsAppOutbox stores outbound WhatsApp messages with delivery tracking.
// Implements the Transactional Outbox Pattern to ensure no message is lost.
type WhatsAppOutbox struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PhoneNo    string         `gorm:"type:varchar(20);not null;index" json:"phone_no"`
	Message    string         `gorm:"type:text;not null" json:"message"`
	Status     WAOutboxStatus `gorm:"type:varchar(20);default:'PENDING';index;not null" json:"status"`
	RetryCount int            `gorm:"default:0;not null" json:"retry_count"`
	MaxRetries int            `gorm:"default:5;not null" json:"max_retries"`
	LastError  *string        `gorm:"type:text" json:"last_error"`

	// Context references for audit trail
	RefType   *string `gorm:"type:varchar(50);index" json:"ref_type"`  // e.g. "trouble_ticket"
	RefID     *uint64 `gorm:"index" json:"ref_id"`                    // e.g. ticket.ID
	CreatedBy *uint64 `gorm:"index" json:"created_by"`

	SentAt    *time.Time `gorm:"type:datetime" json:"sent_at"`
	CreatedAt time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index;not null" json:"created_at"`
	UpdatedAt time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
}

// TableName overrides the default table name for WhatsAppOutbox
func (WhatsAppOutbox) TableName() string {
	return "whatsapp_outbox"
}
