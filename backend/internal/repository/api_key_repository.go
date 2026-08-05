package repository

import (
	"context"
	"errors"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type apiKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) domain.APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id uint64) (*domain.APIKey, error) {
	var apiKey domain.APIKey
	err := r.db.WithContext(ctx).Preload("Role").First(&apiKey, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	var apiKey domain.APIKey
	err := r.db.WithContext(ctx).Preload("Role").Where("token_hash = ?", hash).First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) GetAll(ctx context.Context) ([]domain.APIKey, error) {
	var apiKeys []domain.APIKey
	err := r.db.WithContext(ctx).Preload("Role").Order("id DESC").Find(&apiKeys).Error
	if err != nil {
		return nil, err
	}
	return apiKeys, nil
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *domain.APIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

func (r *apiKeyRepository) Update(ctx context.Context, apiKey *domain.APIKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

func (r *apiKeyRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.APIKey{}, id).Error
}
