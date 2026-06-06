package repository

import (
	"context"
	"errors"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type systemRepository struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) domain.SystemRepository {
	return &systemRepository{db: db}
}

func (r *systemRepository) CreateActivityLog(ctx context.Context, log *domain.ActivityLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *systemRepository) CreateSystemLog(ctx context.Context, log *domain.SystemLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *systemRepository) GetSettingByKey(ctx context.Context, key string) (*domain.SystemSetting, error) {
	var setting domain.SystemSetting
	err := r.db.WithContext(ctx).Where("setting_key = ?", key).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &setting, nil
}

func (r *systemRepository) UpdateSetting(ctx context.Context, setting *domain.SystemSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

func (r *systemRepository) GetSyaratKetentuanAll(ctx context.Context) ([]domain.SyaratKetentuan, error) {
	var sk []domain.SyaratKetentuan
	err := r.db.WithContext(ctx).Order("id desc").Find(&sk).Error
	return sk, err
}

func (r *systemRepository) GetSyaratKetentuanByID(ctx context.Context, id uint64) (*domain.SyaratKetentuan, error) {
	var sk domain.SyaratKetentuan
	err := r.db.WithContext(ctx).First(&sk, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("terms and conditions not found")
		}
		return nil, err
	}
	return &sk, nil
}

func (r *systemRepository) CreateSyaratKetentuan(ctx context.Context, sk *domain.SyaratKetentuan) error {
	return r.db.WithContext(ctx).Create(sk).Error
}

func (r *systemRepository) UpdateSyaratKetentuan(ctx context.Context, sk *domain.SyaratKetentuan) error {
	return r.db.WithContext(ctx).Save(sk).Error
}

func (r *systemRepository) DeleteSyaratKetentuan(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.SyaratKetentuan{}, id).Error
}

func (r *systemRepository) GetActivityLogs(ctx context.Context, filters domain.ActivityLogFilters) ([]domain.ActivityLog, int64, error) {
	var logs []domain.ActivityLog
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.ActivityLog{})

	db = db.Joins("LEFT JOIN users ON users.id = activity_logs.user_id")

	if filters.Search != "" {
		db = db.Where("activity_logs.action LIKE ? OR activity_logs.details LIKE ? OR users.name LIKE ?", "%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	if filters.UserID != nil {
		db = db.Where("activity_logs.user_id = ?", *filters.UserID)
	}

	if filters.Action != "" {
		db = db.Where("activity_logs.action LIKE ?", filters.Action+"%")
	}

	if filters.DateFrom != "" {
		db = db.Where("DATE(activity_logs.timestamp) >= ?", filters.DateFrom)
	}

	if filters.DateTo != "" {
		db = db.Where("DATE(activity_logs.timestamp) <= ?", filters.DateTo)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	db = db.Select("activity_logs.*").Preload("User")
	if filters.Limit > 0 {
		db = db.Limit(filters.Limit)
	}
	if filters.Offset >= 0 {
		db = db.Offset(filters.Offset)
	}

	err := db.Order("activity_logs.timestamp desc, activity_logs.id desc").Find(&logs).Error
	return logs, total, err
}
