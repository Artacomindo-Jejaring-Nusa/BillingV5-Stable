package domain

import (
	"context"
	"time"
)

// HargaLayananRepository defines the DB contracts for Brand/Provider
type HargaLayananRepository interface {
	Create(ctx context.Context, brand *HargaLayanan) error
	GetAll(ctx context.Context) ([]HargaLayanan, error)
	GetByID(ctx context.Context, idBrand string) (*HargaLayanan, error)
	Update(ctx context.Context, brand *HargaLayanan) error
	Delete(ctx context.Context, idBrand string) error
}

// HargaLayananUsecase defines the business contracts for Brand/Provider
type HargaLayananUsecase interface {
	Create(ctx context.Context, brand *HargaLayanan) (*HargaLayanan, error)
	GetAll(ctx context.Context) ([]HargaLayanan, error)
	GetByID(ctx context.Context, idBrand string) (*HargaLayanan, error)
	Update(ctx context.Context, idBrand string, updates map[string]interface{}) (*HargaLayanan, error)
	Delete(ctx context.Context, idBrand string) error
}

// PaketLayananRepository defines the DB contracts for Internet Packages
type PaketLayananRepository interface {
	Create(ctx context.Context, paket *PaketLayanan) error
	GetAll(ctx context.Context) ([]PaketLayanan, error)
	GetByID(ctx context.Context, id uint64) (*PaketLayanan, error)
	Update(ctx context.Context, paket *PaketLayanan) error
	Delete(ctx context.Context, id uint64) error
}

// PaketLayananUsecase defines the business contracts for Internet Packages
type PaketLayananUsecase interface {
	Create(ctx context.Context, paket *PaketLayanan) (*PaketLayanan, error)
	GetAll(ctx context.Context) ([]PaketLayanan, error)
	GetByID(ctx context.Context, id uint64) (*PaketLayanan, error)
	Update(ctx context.Context, id uint64, updates map[string]interface{}) (*PaketLayanan, error)
	Delete(ctx context.Context, id uint64) error
}

// DiskonRepository defines the DB contracts for Cluster Discount management
type DiskonRepository interface {
	Create(ctx context.Context, diskon *Diskon) error
	GetAll(ctx context.Context, limit, offset int, cluster string, isActive *bool) ([]Diskon, int64, error)
	GetByID(ctx context.Context, id uint64) (*Diskon, error)
	GetActiveForCluster(ctx context.Context, cluster string, checkDate time.Time) (*Diskon, error)
	Update(ctx context.Context, diskon *Diskon) error
	Delete(ctx context.Context, id uint64) error
	GetClusterList(ctx context.Context) ([]string, error)
}

// DiskonUsecase defines the business contracts for Cluster Discount management
type DiskonUsecase interface {
	Create(ctx context.Context, diskon *Diskon) (*Diskon, error)
	GetAll(ctx context.Context, page, pageSize int, cluster string, isActive *bool) ([]Diskon, int64, error)
	GetByID(ctx context.Context, id uint64) (*Diskon, error)
	GetActiveForCluster(ctx context.Context, cluster string, checkDate *time.Time) (*Diskon, error)
	Update(ctx context.Context, id uint64, updates map[string]interface{}) (*Diskon, error)
	Delete(ctx context.Context, id uint64) error
	Activate(ctx context.Context, id uint64) (*Diskon, error)
	GetClusterList(ctx context.Context) ([]string, error)
}
