package domain

import (
	"time"
)

// ActivityLog represents the user activity logs.
type ActivityLog struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64     `gorm:"index;not null" json:"user_id"`
	Timestamp time.Time  `gorm:"type:datetime(6);default:CURRENT_TIMESTAMP(6);index" json:"timestamp"`
	Action    string     `gorm:"type:varchar(255);not null" json:"action"`
	Details   *string    `gorm:"type:text" json:"details"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user"`
}

// TableName overrides the default table name for ActivityLog
func (ActivityLog) TableName() string {
	return "activity_logs"
}

// SystemLog represents general system error or info logs.
type SystemLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Timestamp time.Time `gorm:"type:datetime(6);default:CURRENT_TIMESTAMP(6);index" json:"timestamp"`
	Level     string    `gorm:"type:varchar(50);not null" json:"level"`
	Message   string    `gorm:"type:text;not null" json:"message"`
}

// TableName overrides the default table name for SystemLog
func (SystemLog) TableName() string {
	return "system_logs"
}

// SystemSetting stores system-wide key-value configurations.
type SystemSetting struct {
	ID           uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	SettingKey   string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"setting_key"`
	SettingValue *string `gorm:"type:varchar(500)" json:"setting_value"`
}

// TableName overrides the default table name for SystemSetting
func (SystemSetting) TableName() string {
	return "system_settings"
}

// SyaratKetentuan stores terms and conditions content.
type SyaratKetentuan struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Judul     string    `gorm:"type:varchar(255);not null" json:"judul"`
	Konten    string    `gorm:"type:text;not null" json:"konten"`
	Tipe      *string   `gorm:"type:varchar(50);default:'Ketentuan'" json:"tipe"`
	Versi     *string   `gorm:"type:varchar(50)" json:"versi"`
	CreatedAt time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName overrides the default table name for SyaratKetentuan
func (SyaratKetentuan) TableName() string {
	return "syarat_ketentuan"
}

// TokenBlacklist stores revoked JWT tokens.
type TokenBlacklist struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Jti           string     `gorm:"type:varchar(36);uniqueIndex;not null" json:"jti"`
	UserID        uint64     `gorm:"index;not null" json:"user_id"`
	TokenType     string     `gorm:"type:varchar(50);index;not null" json:"token_type"`
	ExpiresAt     time.Time  `gorm:"type:datetime;index;not null" json:"expires_at"`
	CreatedAt     time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index;not null" json:"created_at"`
	Revoked       bool       `gorm:"default:false;index;not null" json:"revoked"`
	RevokedAt     *time.Time `gorm:"type:datetime;index" json:"revoked_at"`
	RevokedReason *string    `gorm:"type:varchar(255);index" json:"revoked_reason"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user"`
}

// TableName overrides the default table name for TokenBlacklist
func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}

// ActivityLogFilters represents filters for retrieving activity logs.
type ActivityLogFilters struct {
	Limit    int
	Offset   int
	Search   string
	UserID   *uint64
	Action   string
	DateFrom string
	DateTo   string
}

