package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"billing-backend/internal/domain"
	"billing-backend/pkg/logger"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type JobConfig struct {
	Key         string
	Name        string
	Description string
	DefaultCron string
	Func        func(ctx context.Context) error
}

type JobStatus struct {
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Schedule    string    `json:"schedule"`
	Enabled     bool      `json:"enabled"`
	LastRun     time.Time `json:"last_run"`
	NextRun     time.Time `json:"next_run"`
	IsRunning   bool      `json:"is_running"`
}

type SchedulerManager struct {
	db            *gorm.DB
	systemUsecase domain.SystemUsecase
	cronInst      *cron.Cron
	jobs          map[string]*JobConfig

	// State management
	mu          sync.RWMutex
	entryIDs    map[string]cron.EntryID
	lastRuns    map[string]time.Time
	runningJobs map[string]bool
}

func NewSchedulerManager(db *gorm.DB, su domain.SystemUsecase, bu domain.BillingUsecase) *SchedulerManager {
	mgr := &SchedulerManager{
		db:            db,
		systemUsecase: su,
		jobs:          make(map[string]*JobConfig),
		entryIDs:      make(map[string]cron.EntryID),
		lastRuns:      make(map[string]time.Time),
		runningJobs:   make(map[string]bool),
	}

	mgr.jobs["generate_invoices"] = &JobConfig{
		Key:         "generate_invoices",
		Name:        "Generate Invoices Otomatis",
		Description: "Membuat invoice tagihan bulanan otomatis untuk pelanggan aktif.",
		DefaultCron: "0 12 * * *",
		Func:        bu.GenerateInvoices,
	}

	mgr.jobs["suspend_services"] = &JobConfig{
		Key:         "suspend_services",
		Name:        "Suspended Otomatis",
		Description: "Melakukan isolasi/suspend otomatis ke pelanggan yang menunggak dan sinkronisasi ke Mikrotik.",
		DefaultCron: "0 0 * * *",
		Func:        bu.AutoSuspend,
	}

	mgr.jobs["verify_payments"] = &JobConfig{
		Key:         "verify_payments",
		Name:        "Verifikasi Pembayaran Otomatis",
		Description: "Memeriksa status pembayaran invoice secara otomatis ke payment gateway (Xendit).",
		DefaultCron: "0 * * * *",
		Func:        bu.VerifyPayments,
	}

	mgr.jobs["archive_old_invoices"] = &JobConfig{
		Key:         "archive_old_invoices",
		Name:        "Arsip Invoice Lama (> 3 Bulan)",
		Description: "Memindahkan invoice yang sudah lunas/expired lebih dari 3 bulan ke tabel arsip untuk menjaga performa database.",
		DefaultCron: "0 2 * * *", // Runs daily at 02:00 AM
		Func:        bu.ArchiveOldInvoices,
	}

	mgr.jobs["retry_mikrotik_sync"] = &JobConfig{
		Key:         "retry_mikrotik_sync",
		Name:        "Retry Sinkronisasi Mikrotik",
		Description: "Mengulangi sinkronisasi status modem ke Mikrotik yang tertunda karena kendala koneksi.",
		DefaultCron: "*/5 * * * *",
		Func:        bu.RetryFailedMikrotikSync,
	}

	return mgr
}

func (m *SchedulerManager) getSettingOrSetDefault(ctx context.Context, key, defaultValue string) string {
	val, err := m.systemUsecase.GetSetting(ctx, key)
	if err != nil || val == "" {
		_ = m.systemUsecase.SetSetting(ctx, key, defaultValue)
		return defaultValue
	}
	return val
}

func (m *SchedulerManager) Start(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for global enable/disable
	globalEnabledStr := m.getSettingOrSetDefault(ctx, "scheduler_global_enabled", "true")
	if globalEnabledStr == "false" {
		logger.Warn("SchedulerManager: Global scheduler is DISABLED. No jobs will be scheduled.")
		if m.cronInst != nil {
			m.cronInst.Stop()
		}
		return
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		logger.Warn("SchedulerManager: Failed to load location Asia/Jakarta, falling back to UTC/Local: %v", err)
		m.cronInst = cron.New()
	} else {
		m.cronInst = cron.New(cron.WithLocation(loc))
	}

	for key, config := range m.jobs {
		scheduleKey := fmt.Sprintf("cron_%s_schedule", key)
		enabledKey := fmt.Sprintf("cron_%s_enabled", key)

		schedule := m.getSettingOrSetDefault(ctx, scheduleKey, config.DefaultCron)
		enabledStr := m.getSettingOrSetDefault(ctx, enabledKey, "true")
		enabled := enabledStr == "true"

		if !enabled {
			continue
		}

		jobKey := key
		jobFunc := config.Func

		entryID, err := m.cronInst.AddFunc(schedule, func() {
			m.executeJob(jobKey, jobFunc)
		})
		if err != nil {
			logger.Error("SchedulerManager: Failed to schedule job %s with pattern %s: %v", jobKey, schedule, err)
			continue
		}
		m.entryIDs[jobKey] = entryID
	}

	m.cronInst.Start()
	logger.Info("SchedulerManager: Started successfully")
}

func (m *SchedulerManager) executeJob(key string, jobFunc func(ctx context.Context) error) {
	m.mu.Lock()
	if m.runningJobs[key] {
		m.mu.Unlock()
		logger.Warn("SchedulerManager: Job %s is already running, skipping", key)
		return
	}
	m.runningJobs[key] = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.runningJobs[key] = false
		m.lastRuns[key] = time.Now()
		m.mu.Unlock()
	}()

	logger.Info("SchedulerManager: Starting job %s...", key)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := jobFunc(ctx); err != nil {
		logger.Error("SchedulerManager: Job %s failed: %v", key, err)
	} else {
		logger.Info("SchedulerManager: Job %s completed successfully", key)
	}
}

func (m *SchedulerManager) Reload(ctx context.Context) {
	m.mu.Lock()
	if m.cronInst != nil {
		logger.Info("SchedulerManager: Stopping existing cron scheduler...")
		stopCtx := m.cronInst.Stop()
		select {
		case <-stopCtx.Done():
			logger.Info("SchedulerManager: Cron scheduler stopped gracefully")
		case <-time.After(10 * time.Second):
			logger.Warn("SchedulerManager: Cron scheduler shutdown timed out")
		}
	}
	m.entryIDs = make(map[string]cron.EntryID)
	m.mu.Unlock()

	m.Start(ctx)
}

func (m *SchedulerManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cronInst != nil {
		stopCtx := m.cronInst.Stop()
		select {
		case <-stopCtx.Done():
			logger.Info("SchedulerManager: Stopped gracefully")
		case <-time.After(10 * time.Second):
			logger.Warn("SchedulerManager: Shutdown timed out")
		}
	}
}

func (m *SchedulerManager) GetJobsStatus(ctx context.Context) []JobStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var statuses []JobStatus
	for key, config := range m.jobs {
		scheduleKey := fmt.Sprintf("cron_%s_schedule", key)
		enabledKey := fmt.Sprintf("cron_%s_enabled", key)

		scheduleVal, _ := m.systemUsecase.GetSetting(ctx, scheduleKey)
		if scheduleVal == "" {
			scheduleVal = config.DefaultCron
		}
		enabledVal, _ := m.systemUsecase.GetSetting(ctx, enabledKey)
		enabled := enabledVal != "false" // Default to true if not set

		var nextRun time.Time
		if entryID, ok := m.entryIDs[key]; ok && m.cronInst != nil {
			entry := m.cronInst.Entry(entryID)
			nextRun = entry.Next
		}

		statuses = append(statuses, JobStatus{
			Key:         key,
			Name:        config.Name,
			Description: config.Description,
			Schedule:    scheduleVal,
			Enabled:     enabled,
			LastRun:     m.lastRuns[key],
			NextRun:     nextRun,
			IsRunning:   m.runningJobs[key],
		})
	}
	return statuses
}

func (m *SchedulerManager) UpdateJob(ctx context.Context, key, schedule string, enabled bool) error {
	m.mu.Lock()
	_, exists := m.jobs[key]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("job %s not found", key)
	}

	if enabled {
		if _, err := cron.ParseStandard(schedule); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
	}

	scheduleKey := fmt.Sprintf("cron_%s_schedule", key)
	enabledKey := fmt.Sprintf("cron_%s_enabled", key)

	enabledStr := "true"
	if !enabled {
		enabledStr = "false"
	}

	if err := m.systemUsecase.SetSetting(ctx, scheduleKey, schedule); err != nil {
		return err
	}
	if err := m.systemUsecase.SetSetting(ctx, enabledKey, enabledStr); err != nil {
		return err
	}

	go m.Reload(context.Background())

	return nil
}

func (m *SchedulerManager) RunJobNow(ctx context.Context, key string) error {
	m.mu.RLock()
	config, exists := m.jobs[key]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("job %s not found", key)
	}

	m.mu.RLock()
	isRunning := m.runningJobs[key]
	m.mu.RUnlock()

	if isRunning {
		return fmt.Errorf("job %s is already running", key)
	}

	go m.executeJob(key, config.Func)

	return nil
}

func (m *SchedulerManager) ToggleGlobal(ctx context.Context, enabled bool) error {
	enabledStr := "true"
	if !enabled {
		enabledStr = "false"
	}

	if err := m.systemUsecase.SetSetting(ctx, "scheduler_global_enabled", enabledStr); err != nil {
		return err
	}

	go m.Reload(context.Background())
	return nil
}

func (m *SchedulerManager) IsGlobalEnabled(ctx context.Context) bool {
	val := m.getSettingOrSetDefault(ctx, "scheduler_global_enabled", "true")
	return val == "true"
}
