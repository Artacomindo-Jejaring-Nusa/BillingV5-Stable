package domain

import "context"

type SystemRepository interface {
	CreateActivityLog(ctx context.Context, log *ActivityLog) error
	CreateSystemLog(ctx context.Context, log *SystemLog) error
	GetSettingByKey(ctx context.Context, key string) (*SystemSetting, error)
	UpdateSetting(ctx context.Context, setting *SystemSetting) error
	GetActivityLogs(ctx context.Context, filters ActivityLogFilters) ([]ActivityLog, int64, error)
	
	// Syarat & Ketentuan (SK)
	GetSyaratKetentuanAll(ctx context.Context) ([]SyaratKetentuan, error)
	GetSyaratKetentuanByID(ctx context.Context, id uint64) (*SyaratKetentuan, error)
	CreateSyaratKetentuan(ctx context.Context, sk *SyaratKetentuan) error
	UpdateSyaratKetentuan(ctx context.Context, sk *SyaratKetentuan) error
	DeleteSyaratKetentuan(ctx context.Context, id uint64) error
}

type SystemUsecase interface {
	LogActivity(ctx context.Context, userID uint64, action string, details string) error
	LogSystem(ctx context.Context, level string, message string) error
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key string, value string) error
	FetchActivityLogs(ctx context.Context, filters ActivityLogFilters) ([]ActivityLog, int64, error)
	
	// SK
	FetchSKAll(ctx context.Context) ([]SyaratKetentuan, error)
	GetSKByID(ctx context.Context, id uint64) (*SyaratKetentuan, error)
	CreateSK(ctx context.Context, sk *SyaratKetentuan) error
	UpdateSK(ctx context.Context, id uint64, sk *SyaratKetentuan) error
	DeleteSK(ctx context.Context, id uint64) error
}
