package repository

import (
	"context"
	"errors"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type mikrotikRepository struct {
	db *gorm.DB
}

func NewMikrotikRepository(db *gorm.DB) domain.MikrotikRepository {
	return &mikrotikRepository{db: db}
}

func (r *mikrotikRepository) GetAll(ctx context.Context) ([]domain.MikrotikServer, error) {
	var servers []domain.MikrotikServer
	err := r.db.WithContext(ctx).Order("id desc").Find(&servers).Error
	return servers, err
}

func (r *mikrotikRepository) GetByID(ctx context.Context, id uint64) (*domain.MikrotikServer, error) {
	var server domain.MikrotikServer
	err := r.db.WithContext(ctx).First(&server, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mikrotik server not found")
		}
		return nil, err
	}
	return &server, nil
}

func (r *mikrotikRepository) Create(ctx context.Context, server *domain.MikrotikServer) error {
	return r.db.WithContext(ctx).Create(server).Error
}

func (r *mikrotikRepository) Update(ctx context.Context, server *domain.MikrotikServer) error {
	return r.db.WithContext(ctx).Save(server).Error
}

func (r *mikrotikRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.MikrotikServer{}, id).Error
}

func (r *mikrotikRepository) GetByName(ctx context.Context, name string) (*domain.MikrotikServer, error) {
	var server domain.MikrotikServer
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&server).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

func (r *mikrotikRepository) GetByNames(ctx context.Context, names []string) ([]domain.MikrotikServer, error) {
	var list []domain.MikrotikServer
	err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&list).Error
	return list, err
}

type oltRepository struct {
	db *gorm.DB
}

func NewOLTRepository(db *gorm.DB) domain.OLTRepository {
	return &oltRepository{db: db}
}

func (r *oltRepository) GetAll(ctx context.Context) ([]domain.OLT, error) {
	var olts []domain.OLT
	err := r.db.WithContext(ctx).Preload("MikrotikServer").Order("id desc").Find(&olts).Error
	return olts, err
}

func (r *oltRepository) GetByID(ctx context.Context, id uint64) (*domain.OLT, error) {
	var olt domain.OLT
	err := r.db.WithContext(ctx).Preload("MikrotikServer").First(&olt, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("OLT not found")
		}
		return nil, err
	}
	return &olt, nil
}

func (r *oltRepository) Create(ctx context.Context, olt *domain.OLT) error {
	return r.db.WithContext(ctx).Create(olt).Error
}

func (r *oltRepository) Update(ctx context.Context, olt *domain.OLT) error {
	return r.db.WithContext(ctx).Save(olt).Error
}

func (r *oltRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.OLT{}, id).Error
}

type odpRepository struct {
	db *gorm.DB
}

func NewODPRepository(db *gorm.DB) domain.ODPRepository {
	return &odpRepository{db: db}
}

func (r *odpRepository) GetAll(ctx context.Context) ([]domain.ODP, error) {
	var odps []domain.ODP
	err := r.db.WithContext(ctx).
		Preload("Olt").
		Preload("ParentOdp").
		Order("id desc").
		Find(&odps).Error
	if err != nil {
		return nil, err
	}

	// Calculate PortTerpakai for each ODP
	for i := range odps {
		var count int64
		r.db.WithContext(ctx).Model(&domain.DataTeknis{}).Where("odp_id = ?", odps[i].ID).Count(&count)
		odps[i].PortTerpakai = int(count)
	}

	return odps, nil
}

func (r *odpRepository) GetByID(ctx context.Context, id uint64) (*domain.ODP, error) {
	var odp domain.ODP
	err := r.db.WithContext(ctx).Preload("Olt").Preload("ParentOdp").First(&odp, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ODP not found")
		}
		return nil, err
	}

	var count int64
	r.db.WithContext(ctx).Model(&domain.DataTeknis{}).Where("odp_id = ?", odp.ID).Count(&count)
	odp.PortTerpakai = int(count)

	return &odp, nil
}

func (r *odpRepository) Create(ctx context.Context, odp *domain.ODP) error {
	return r.db.WithContext(ctx).Create(odp).Error
}

func (r *odpRepository) Update(ctx context.Context, odp *domain.ODP) error {
	return r.db.WithContext(ctx).Save(odp).Error
}

func (r *odpRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.ODP{}, id).Error
}

