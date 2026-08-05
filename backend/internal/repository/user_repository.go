package repository

import (
	"context"
	"errors"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id uint64) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Preload("Role.Permissions").First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Preload("Role.Permissions").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context, limit, offset int, search string, roleID *uint64) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.User{})

	// Apply search filter (name or email)
	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("name LIKE ? OR email LIKE ?", searchTerm, searchTerm)
	}

	// Apply role filter
	if roleID != nil {
		query = query.Where("role_id = ?", *roleID)
	}

	// Get total count before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and fetch with preloaded relationships
	if err := query.Preload("Role.Permissions").
		Offset(offset).Limit(limit).
		Order("id DESC").
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, id).Error
}

func (r *userRepository) GetByResetToken(ctx context.Context, email, token string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Where("email = ? AND reset_token = ?", email, token).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) SaveFcmToken(ctx context.Context, userID uint64, token string, deviceType string) error {
	var fcmToken domain.UserFcmToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&fcmToken).Error
	if err == nil {
		fcmToken.UserID = userID
		if deviceType != "" {
			fcmToken.DeviceType = deviceType
		}
		return r.db.WithContext(ctx).Save(&fcmToken).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fcmToken = domain.UserFcmToken{
			UserID:     userID,
			Token:      token,
			DeviceType: deviceType,
		}
		if fcmToken.DeviceType == "" {
			fcmToken.DeviceType = "web"
		}
		return r.db.WithContext(ctx).Create(&fcmToken).Error
	}
	return err
}

func (r *userRepository) DeleteFcmToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&domain.UserFcmToken{}).Error
}
