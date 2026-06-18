package domain

import "context"

// PelangganRepository defines database operations for Pelanggan
type PelangganRepository interface {
	GetAll(ctx context.Context, limit, offset int, search, connectionStatus string) ([]Pelanggan, int64, error)
	GetByID(ctx context.Context, id uint64) (*Pelanggan, error)
	Create(ctx context.Context, pelanggan *Pelanggan) error
	Update(ctx context.Context, pelanggan *Pelanggan) error
	Delete(ctx context.Context, id uint64) error
	GetByEmail(ctx context.Context, email string) (*Pelanggan, error)
	GetByEmails(ctx context.Context, emails []string) ([]Pelanggan, error)
	GetByNoKtp(ctx context.Context, noKtp string) (*Pelanggan, error)
	GetByNoTelp(ctx context.Context, noTelp string) (*Pelanggan, error)
	GetUniqueLocations(ctx context.Context) ([]string, error)
}

// PelangganUsecase defines business logic operations for Pelanggan
type PelangganUsecase interface {
	FetchAll(ctx context.Context, skip, limit int, search, connectionStatus string) ([]Pelanggan, int64, error)
	GetByID(ctx context.Context, id uint64) (*Pelanggan, error)
	Store(ctx context.Context, pelanggan *Pelanggan) error
	Update(ctx context.Context, id uint64, pelanggan *Pelanggan) error
	Delete(ctx context.Context, id uint64) error
	GetUniqueLocations(ctx context.Context) ([]string, error)
	Export(ctx context.Context, format string) ([]byte, string, error)
	ImportFromCSV(ctx context.Context, csvContent string) (int, error)
}
