package domain

import "context"

// PelangganRepository defines database operations for Pelanggan
type PelangganRepository interface {
	GetAll(ctx context.Context, limit, offset int) ([]Pelanggan, int64, error)
	GetByID(ctx context.Context, id uint64) (*Pelanggan, error)
	Create(ctx context.Context, pelanggan *Pelanggan) error
	Update(ctx context.Context, pelanggan *Pelanggan) error
	Delete(ctx context.Context, id uint64) error
	GetByEmail(ctx context.Context, email string) (*Pelanggan, error)
	GetByEmails(ctx context.Context, emails []string) ([]Pelanggan, error)
	GetByNoKtp(ctx context.Context, noKtp string) (*Pelanggan, error)
}

// PelangganUsecase defines business logic operations for Pelanggan
type PelangganUsecase interface {
	FetchAll(ctx context.Context, page, pageSize int) ([]Pelanggan, int64, error)
	GetByID(ctx context.Context, id uint64) (*Pelanggan, error)
	Store(ctx context.Context, pelanggan *Pelanggan) error
	Update(ctx context.Context, id uint64, pelanggan *Pelanggan) error
	Delete(ctx context.Context, id uint64) error
}
