package repository

import (
	"context"
	"time"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type tokenBlacklistRepository struct {
	db *gorm.DB
}

// NewTokenBlacklistRepository creates a new token blacklist repository
func NewTokenBlacklistRepository(db *gorm.DB) domain.TokenBlacklistRepository {
	return &tokenBlacklistRepository{db: db}
}

func (r *tokenBlacklistRepository) Blacklist(ctx context.Context, jti string, userID uint64, tokenType string, expiresAt time.Time, reason string) error {
	now := time.Now()
	entry := domain.TokenBlacklist{
		Jti:           jti,
		UserID:        userID,
		TokenType:     tokenType,
		ExpiresAt:     expiresAt,
		Revoked:       true,
		RevokedAt:     &now,
		RevokedReason: &reason,
		CreatedAt:     now,
	}
	return r.db.WithContext(ctx).Create(&entry).Error
}

func (r *tokenBlacklistRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.TokenBlacklist{}).
		Where("jti = ?", jti).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *tokenBlacklistRepository) BlacklistAllForUser(ctx context.Context, userID uint64) error {
	// Mark all active tokens for this user as blacklisted
	// This is used for "logout from all devices" functionality
	now := time.Now()
	reason := "User logout from all devices"
	entry := domain.TokenBlacklist{
		Jti:           "all_revoked_" + time.Now().Format("20060102150405"),
		UserID:        userID,
		TokenType:     "all",
		ExpiresAt:     now.Add(7 * 24 * time.Hour), // Future expiry
		Revoked:       true,
		RevokedAt:     &now,
		RevokedReason: &reason,
		CreatedAt:     now,
	}
	return r.db.WithContext(ctx).Create(&entry).Error
}
