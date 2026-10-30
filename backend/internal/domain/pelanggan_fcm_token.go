package domain

import (
	"time"

	"gorm.io/gorm"
)

type PelangganFcmToken struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PelangganID uint64         `gorm:"not null;index" json:"pelanggan_id"`
	Token       string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"token"`
	DeviceType  string         `gorm:"type:varchar(50);default:'android'" json:"device_type"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Pelanggan Pelanggan `gorm:"foreignKey:PelangganID" json:"-"`
}

func (PelangganFcmToken) TableName() string {
	return "pelanggan_fcm_tokens"
}

type SavePelangganFcmTokenRequest struct {
	PelangganID uint64 `json:"pelanggan_id"`
	Identifier  string `json:"identifier"`
	FCMToken    string `json:"fcm_token" binding:"required"`
	DeviceType  string `json:"device_type"`
}

