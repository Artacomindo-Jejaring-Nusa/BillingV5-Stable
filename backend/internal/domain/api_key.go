package domain

import (
	"time"

	"gorm.io/gorm"
)

type APIKey struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Prefix    string         `gorm:"type:varchar(16);not null" json:"prefix"`
	TokenHash string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`
	RoleID    uint64         `gorm:"not null" json:"role_id"`
	IsActive  bool           `gorm:"default:true;not null" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Role Role `gorm:"foreignKey:RoleID" json:"role"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

type APIKeyResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	RoleID    uint64    `json:"role_id"`
	RoleName  string    `json:"role_name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAPIKeyRequest struct {
	Name   string `json:"name" binding:"required"`
	RoleID uint64 `json:"role_id" binding:"required"`
}

type CreateAPIKeyResponse struct {
	APIKeyResponse
	RawKey string `json:"raw_key"` // Shown only once during generation
}
