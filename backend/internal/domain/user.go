package domain

import (
	"time"

	"gorm.io/gorm"
)

// User represents the users table for authentication and system access.
type User struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name              string         `gorm:"type:varchar(191)" json:"name"`
	Email             string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"email"`
	Password          string         `gorm:"type:varchar(191);not null" json:"-"`
	RememberToken     *string        `gorm:"type:varchar(100)" json:"remember_token"`
	EmailVerifiedAt   *time.Time     `gorm:"type:timestamp" json:"email_verified_at"`
	IsActive          bool           `gorm:"default:true;not null" json:"is_active"`
	RevokedBefore     *time.Time     `gorm:"type:datetime" json:"revoked_before"`
	PasswordChangedAt *time.Time     `gorm:"type:datetime" json:"password_changed_at"`
	ResetToken        *string        `gorm:"type:varchar(255)" json:"reset_token"`
	ResetTokenExpires *time.Time     `gorm:"type:datetime" json:"reset_token_expires"`
	CreatedAt         *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
	RoleID            *uint64        `gorm:"index" json:"role_id"`

	// Relationships
	Role               Role             `gorm:"foreignKey:RoleID" json:"role"`
	ActivityLogs       []ActivityLog    `gorm:"foreignKey:UserID" json:"activity_logs,omitempty"`
	BlacklistedTokens  []TokenBlacklist `gorm:"foreignKey:UserID" json:"blacklisted_tokens,omitempty"`
	InventoryHistories []InventoryHistory `gorm:"foreignKey:UserID" json:"inventory_histories,omitempty"`
}

// Role represents the roles table.
type Role struct {
	ID        uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string       `gorm:"type:varchar(191);unique;not null" json:"name"`
	Users     []User       `gorm:"foreignKey:RoleID" json:"users"`
	Permissions []Permission `gorm:"many2many:role_has_permissions;" json:"permissions"`
}

// Permission represents the permissions table.
type Permission struct {
	ID    uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name  string `gorm:"type:varchar(191);unique;not null" json:"name"`
	Roles []Role `gorm:"many2many:role_has_permissions;" json:"roles"`
}

// RoleHasPermission is the join table for Role and Permission. (Handled automatically by many2many, but we can define if needed).
