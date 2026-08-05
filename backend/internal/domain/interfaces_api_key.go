package domain

import "context"

type APIKeyRepository interface {
	GetByID(ctx context.Context, id uint64) (*APIKey, error)
	GetByHash(ctx context.Context, hash string) (*APIKey, error)
	GetAll(ctx context.Context) ([]APIKey, error)
	Create(ctx context.Context, apiKey *APIKey) error
	Update(ctx context.Context, apiKey *APIKey) error
	Delete(ctx context.Context, id uint64) error
}

type APIKeyUsecase interface {
	FetchAll(ctx context.Context) ([]APIKey, error)
	CreateAPIKey(ctx context.Context, name string, roleID uint64) (*APIKey, string, error) // Returns model and raw key
	ToggleAPIKey(ctx context.Context, id uint64) (*APIKey, error)
	DeleteAPIKey(ctx context.Context, id uint64) error
}
