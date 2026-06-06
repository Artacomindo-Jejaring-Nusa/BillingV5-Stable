package usecase

import (
	"context"
	"time"

	"billing-backend/internal/domain"
)

type systemUsecase struct {
	repo domain.SystemRepository
}

func NewSystemUsecase(r domain.SystemRepository) domain.SystemUsecase {
	return &systemUsecase{repo: r}
}

func (u *systemUsecase) LogActivity(ctx context.Context, userID uint64, action string, details string) error {
	log := &domain.ActivityLog{
		UserID:    userID,
		Action:    action,
		Details:   &details,
		Timestamp: time.Now(),
	}
	return u.repo.CreateActivityLog(ctx, log)
}

func (u *systemUsecase) LogSystem(ctx context.Context, level string, message string) error {
	log := &domain.SystemLog{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
	}
	return u.repo.CreateSystemLog(ctx, log)
}

func (u *systemUsecase) GetSetting(ctx context.Context, key string) (string, error) {
	setting, err := u.repo.GetSettingByKey(ctx, key)
	if err != nil {
		return "", err
	}
	if setting == nil || setting.SettingValue == nil {
		return "", nil
	}
	return *setting.SettingValue, nil
}

func (u *systemUsecase) SetSetting(ctx context.Context, key string, value string) error {
	setting, err := u.repo.GetSettingByKey(ctx, key)
	if err != nil {
		return err
	}
	if setting == nil {
		setting = &domain.SystemSetting{
			SettingKey:   key,
			SettingValue: &value,
		}
	} else {
		setting.SettingValue = &value
	}
	return u.repo.UpdateSetting(ctx, setting)
}

func (u *systemUsecase) FetchSKAll(ctx context.Context) ([]domain.SyaratKetentuan, error) {
	return u.repo.GetSyaratKetentuanAll(ctx)
}

func (u *systemUsecase) GetSKByID(ctx context.Context, id uint64) (*domain.SyaratKetentuan, error) {
	return u.repo.GetSyaratKetentuanByID(ctx, id)
}

func (u *systemUsecase) CreateSK(ctx context.Context, sk *domain.SyaratKetentuan) error {
	sk.CreatedAt = time.Now()
	return u.repo.CreateSyaratKetentuan(ctx, sk)
}

func (u *systemUsecase) UpdateSK(ctx context.Context, id uint64, sk *domain.SyaratKetentuan) error {
	existing, err := u.repo.GetSyaratKetentuanByID(ctx, id)
	if err != nil {
		return err
	}
	existing.Judul = sk.Judul
	existing.Konten = sk.Konten
	existing.Tipe = sk.Tipe
	existing.Versi = sk.Versi
	return u.repo.UpdateSyaratKetentuan(ctx, existing)
}

func (u *systemUsecase) DeleteSK(ctx context.Context, id uint64) error {
	return u.repo.DeleteSyaratKetentuan(ctx, id)
}

func (u *systemUsecase) FetchActivityLogs(ctx context.Context, filters domain.ActivityLogFilters) ([]domain.ActivityLog, int64, error) {
	if filters.Limit <= 0 {
		filters.Limit = 10
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	return u.repo.GetActivityLogs(ctx, filters)
}
