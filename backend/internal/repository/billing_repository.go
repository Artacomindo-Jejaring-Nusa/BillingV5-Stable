package repository

import (
	"context"
	"errors"
	"time"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) domain.InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) GetAll(ctx context.Context, limit, offset int) ([]domain.Invoice, int64, error) {
	var invoices []domain.Invoice
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Invoice{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Limit(limit).
		Offset(offset).
		Order("id desc").
		Find(&invoices).Error

	return invoices, total, err
}

func (r *invoiceRepository) GetByID(ctx context.Context, id uint64) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := r.db.WithContext(ctx).Preload("Pelanggan").First(&inv, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}
	return &inv, nil
}

func (r *invoiceRepository) GetByInvoiceNumber(ctx context.Context, invNumber string) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := r.db.WithContext(ctx).Preload("Pelanggan").Where("invoice_number = ?", invNumber).First(&inv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}
	return &inv, nil
}

func (r *invoiceRepository) Create(ctx context.Context, invoice *domain.Invoice) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

func (r *invoiceRepository) Update(ctx context.Context, invoice *domain.Invoice) error {
	return r.db.WithContext(ctx).Omit("Pelanggan").Save(invoice).Error
}

func (r *invoiceRepository) GetCallbackLog(ctx context.Context, xenditID, externalID, idempotencyKey string) (*domain.PaymentCallbackLog, error) {
	var callbackLog domain.PaymentCallbackLog
	query := r.db.WithContext(ctx)

	if idempotencyKey != "" {
		err := query.Where("idempotency_key = ?", idempotencyKey).First(&callbackLog).Error
		if err == nil {
			return &callbackLog, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	if xenditID != "" {
		err := query.Where("xendit_id = ?", xenditID).First(&callbackLog).Error
		if err == nil {
			return &callbackLog, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	if externalID != "" {
		err := query.Where("external_id = ?", externalID).First(&callbackLog).Error
		if err == nil {
			return &callbackLog, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	return nil, nil // Not found, no error
}

func (r *invoiceRepository) CreateCallbackLog(ctx context.Context, log *domain.PaymentCallbackLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *invoiceRepository) GetInvoiceWithRelations(ctx context.Context, externalID string) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("Pelanggan.HargaLayanan").
		Preload("Pelanggan.Langganan").
		Preload("Pelanggan.Langganan.PaketLayanan").
		Preload("Pelanggan.DataTeknis").
		Where("LOWER(xendit_external_id) = LOWER(?) OR LOWER(invoice_number) = LOWER(?)", externalID, externalID).
		First(&inv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}
	return &inv, nil
}

func (r *invoiceRepository) GetUnpaidByExternalIDs(ctx context.Context, externalIDs []string) ([]domain.Invoice, error) {
	var invoices []domain.Invoice
	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("Pelanggan.HargaLayanan").
		Preload("Pelanggan.Langganan").
		Preload("Pelanggan.Langganan.PaketLayanan").
		Preload("Pelanggan.DataTeknis").
		Where("status_invoice != ?", "Lunas").
		Where("xendit_external_id IN ? OR invoice_number IN ?", externalIDs, externalIDs).
		Find(&invoices).Error
	return invoices, err
}

func (r *invoiceRepository) HasPaidInvoiceForPeriod(ctx context.Context, pelangganID uint64, targetDueDate, endOfPrevMonth time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Invoice{}).
		Where("pelanggan_id = ?", pelangganID).
		Where("status_invoice = ?", "Lunas").
		Where("tgl_jatuh_tempo IN ? OR tgl_invoice >= ?", []time.Time{targetDueDate, endOfPrevMonth}, targetDueDate).
		Count(&count).Error
	return count > 0, err
}

func (r *invoiceRepository) UpdateStatusForUnpaidInvoices(ctx context.Context, pelangganID uint64, targetDueDate, endOfPrevMonth time.Time, newStatus string) (int64, error) {
	res := r.db.WithContext(ctx).Model(&domain.Invoice{}).
		Where("pelanggan_id = ?", pelangganID).
		Where("tgl_jatuh_tempo IN ?", []time.Time{targetDueDate, endOfPrevMonth}).
		Where("status_invoice IN ?", []string{"Belum Dibayar", "Expired"}).
		Update("status_invoice", newStatus)
	return res.RowsAffected, res.Error
}

func (r *invoiceRepository) GetInvoiceByPelangganAndDueDateRange(ctx context.Context, pelangganID uint64, start, end time.Time) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := r.db.WithContext(ctx).
		Where("pelanggan_id = ?", pelangganID).
		Where("tgl_jatuh_tempo BETWEEN ? AND ?", start, end).
		First(&inv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

func (r *invoiceRepository) GetInvoiceSummary(ctx context.Context) (*domain.InvoiceSummaryStats, error) {
	var summary domain.InvoiceSummaryStats

	// Count automatic invoices
	if err := r.db.WithContext(ctx).Model(&domain.Invoice{}).Where("invoice_type = ?", "automatic").Count(&summary.InvoiceTypes.Automatic).Error; err != nil {
		return nil, err
	}

	// Count manual invoices
	if err := r.db.WithContext(ctx).Model(&domain.Invoice{}).Where("invoice_type = ?", "manual").Count(&summary.InvoiceTypes.Manual).Error; err != nil {
		return nil, err
	}

	// Count total reinvoices
	if err := r.db.WithContext(ctx).Model(&domain.Invoice{}).Where("is_reinvoice = ?", true).Count(&summary.TotalReinvoice).Error; err != nil {
		return nil, err
	}

	return &summary, nil
}

// Langganan Repository
type langgananRepository struct {
	db *gorm.DB
}

func NewLanggananRepository(db *gorm.DB) domain.LanggananRepository {
	return &langgananRepository{db: db}
}

func (r *langgananRepository) GetAll(ctx context.Context, limit, offset int, search, status string, forInvoiceSelection bool) ([]domain.Langganan, int64, error) {
	var langganans []domain.Langganan
	var total int64

	dbCount := r.db.WithContext(ctx).Model(&domain.Langganan{}).
		Joins("JOIN pelanggans ON pelanggans.id = langganans.pelanggan_id AND pelanggans.deleted_at IS NULL").
		Joins("LEFT JOIN data_teknis ON data_teknis.pelanggan_id = pelanggans.id")

	dbFind := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("PaketLayanan").
		Preload("PaketLayanan.HargaLayanan").
		Joins("JOIN pelanggans ON pelanggans.id = langganans.pelanggan_id AND pelanggans.deleted_at IS NULL").
		Joins("LEFT JOIN data_teknis ON data_teknis.pelanggan_id = pelanggans.id")

	if forInvoiceSelection {
		dbCount = dbCount.Where("langganans.status != ?", "Berhenti")
		dbFind = dbFind.Where("langganans.status != ?", "Berhenti")
	}

	if status != "" {
		dbCount = dbCount.Where("langganans.status = ?", status)
		dbFind = dbFind.Where("langganans.status = ?", status)
	}

	if search != "" {
		searchTerm := "%" + search + "%"
		dbCount = dbCount.Where("pelanggans.nama LIKE ? OR data_teknis.id_pelanggan LIKE ? OR pelanggans.no_telp LIKE ? OR pelanggans.email LIKE ?", searchTerm, searchTerm, searchTerm, searchTerm)
		dbFind = dbFind.Where("pelanggans.nama LIKE ? OR data_teknis.id_pelanggan LIKE ? OR pelanggans.no_telp LIKE ? OR pelanggans.email LIKE ?", searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if err := dbCount.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := dbFind.
		Limit(limit).
		Offset(offset).
		Order("langganans.id desc").
		Find(&langganans).Error

	return langganans, total, err
}

func (r *langgananRepository) GetByID(ctx context.Context, id uint64) (*domain.Langganan, error) {
	var lng domain.Langganan
	err := r.db.WithContext(ctx).Preload("Pelanggan").Preload("PaketLayanan").First(&lng, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("langganan not found")
		}
		return nil, err
	}
	return &lng, nil
}

func (r *langgananRepository) Create(ctx context.Context, langganan *domain.Langganan) error {
	return r.db.WithContext(ctx).Create(langganan).Error
}

func (r *langgananRepository) Update(ctx context.Context, langganan *domain.Langganan) error {
	return r.db.WithContext(ctx).Omit("Pelanggan", "PaketLayanan").Save(langganan).Error
}

func (r *langgananRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Langganan{}, id).Error
}

func (r *langgananRepository) GetActiveByDueDateRange(ctx context.Context, start, end time.Time) ([]domain.Langganan, error) {
	var langganans []domain.Langganan
	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("Pelanggan.HargaLayanan").
		Preload("Pelanggan.DataTeknis").
		Preload("PaketLayanan").
		Where("status = ?", "Aktif").
		Where("tgl_jatuh_tempo >= ? AND tgl_jatuh_tempo < ?", start, end).
		Find(&langganans).Error
	return langganans, err
}

func (r *langgananRepository) GetActiveOverdueForSuspend(ctx context.Context, targetDueDate, endOfPrevMonth time.Time) ([]domain.Langganan, error) {
	var langganans []domain.Langganan

	paidSubquery := r.db.Model(&domain.Invoice{}).
		Select("pelanggan_id").
		Where("status_invoice = ?", "Lunas").
		Where("tgl_jatuh_tempo IN ? OR tgl_invoice >= ?", []time.Time{targetDueDate, endOfPrevMonth}, targetDueDate)

	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("Pelanggan.DataTeknis").
		Joins("JOIN invoices ON invoices.pelanggan_id = langganan.pelanggan_id").
		Where("langganan.status = ?", "Aktif").
		Where("invoices.tgl_jatuh_tempo IN ?", []time.Time{targetDueDate, endOfPrevMonth}).
		Where("invoices.status_invoice IN ?", []string{"Belum Dibayar", "Expired"}).
		Where("langganan.pelanggan_id NOT IN (?)", paidSubquery).
		Group("langganan.id").
		Find(&langganans).Error

	return langganans, err
}

func (r *langgananRepository) GetNewUserLangganans(ctx context.Context) ([]domain.Langganan, error) {
	var langganans []domain.Langganan

	// Subquery to check if customer already has archived invoices
	archivedSubquery := r.db.Model(&domain.InvoiceArchive{}).Select("pelanggan_id")

	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("Pelanggan.HargaLayanan").
		Preload("Pelanggan.DataTeknis").
		Preload("PaketLayanan").
		Joins("JOIN pelanggans ON pelanggans.id = langganans.pelanggan_id AND pelanggans.deleted_at IS NULL").
		Joins("LEFT JOIN invoices ON invoices.pelanggan_id = pelanggans.id").
		Where("langganans.status = ?", "Aktif").
		Where("invoices.id IS NULL").
		Where("langganans.pelanggan_id NOT IN (?)", archivedSubquery).
		Order("langganans.id desc").
		Find(&langganans).Error

	return langganans, err
}

