package domain

import (
	"time"

	"gorm.io/gorm"
)

type UserFcmToken struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64         `gorm:"not null;index" json:"user_id"`
	Token      string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"token"`
	DeviceType string         `gorm:"type:varchar(50);default:'web'" json:"device_type"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (UserFcmToken) TableName() string {
	return "user_fcm_tokens"
}

type SaveFcmTokenRequest struct {
	Token      string `json:"token" binding:"required"`
	DeviceType string `json:"device_type"`
}
