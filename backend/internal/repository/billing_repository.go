package repository

import (
	"context"
	"errors"
	"math"
	"time"

	"billing-backend/internal/domain"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) domain.InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) GetAll(ctx context.Context, limit, offset int, search, status string) ([]domain.Invoice, int64, error) {
	var invoices []domain.Invoice
	var total int64

	buildFilter := func(tx *gorm.DB) *gorm.DB {
		if status != "" {
			tx = tx.Where("status_invoice = ?", status)
		}
		if search != "" {
			searchTerm := "%" + search + "%"
			tx = tx.Where(
				"invoice_number LIKE ? OR pelanggan_id IN (SELECT id FROM pelanggan WHERE nama LIKE ?)",
				searchTerm, searchTerm,
			)
		}
		return tx
	}

	if err := buildFilter(r.db.WithContext(ctx).Model(&domain.Invoice{})).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := buildFilter(r.db.WithContext(ctx).Model(&domain.Invoice{})).
		Preload("Pelanggan").
		Limit(limit).
		Offset(offset).
		Order("id desc").
		Find(&invoices).Error

	for i := range invoices {
		if invoices[i].Pelanggan != nil {
			invoices[i].PelangganNama = invoices[i].Pelanggan.Nama
		}
	}

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
	err := r.db.WithContext(ctx).Unscoped().Preload("Pelanggan").Where("invoice_number = ?", invNumber).First(&inv).Error
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

func (r *invoiceRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Invoice{}).Error
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
		Where("status_invoice IN ?", []string{"Belum Bayar", "Expired"}).
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

func (r *invoiceRepository) GetRevenueReport(ctx context.Context, params *domain.RevenueReportParams) (*domain.RevenueReportResponse, error) {
	var report domain.RevenueReportResponse

	// Helper to build base query
	buildBaseQuery := func() *gorm.DB {
		q := r.db.WithContext(ctx).Table("invoices").
			Joins("LEFT JOIN pelanggan ON invoices.pelanggan_id = pelanggan.id").
			Where("invoices.deleted_at IS NULL")

		if params.StartDate != "" {
			q = q.Where("invoices.tgl_invoice >= ?", params.StartDate)
		}
		if params.EndDate != "" {
			// Using 23:59:59 to cover the whole end day if it's a datetime column
			q = q.Where("invoices.tgl_invoice <= ?", params.EndDate+" 23:59:59")
		}
		if params.Alamat != "" {
			q = q.Where("pelanggan.alamat = ?", params.Alamat)
		}
		if params.IDBrand != "" {
			q = q.Where("pelanggan.id_brand = ?", params.IDBrand)
		}
		return q
	}

	// 1. Total Invoices
	var totalInvoices int64
	if err := buildBaseQuery().Count(&totalInvoices).Error; err != nil {
		return nil, err
	}
	report.TotalInvoices = int(totalInvoices)
	report.BillingSummary.TotalTagihan.Count = int(totalInvoices)

	// 2. Billing Summary by Status
	type StatusResult struct {
		Status      string
		Count       int
		Nominal     float64
		Diskon      float64
		TotalAmount float64
	}
	var statusResults []StatusResult
	err := buildBaseQuery().
		Select("status_invoice as status, COUNT(*) as count, SUM(COALESCE(harga_sebelum_diskon, total_harga)) as nominal, SUM(COALESCE(diskon_amount, 0)) as diskon, SUM(total_harga) as total_amount").
		Group("status_invoice").
		Scan(&statusResults).Error
	if err != nil {
		return nil, err
	}

	for _, res := range statusResults {
		stat := domain.BillStat{
			Count:   res.Count,
			Nominal: res.Nominal,
			Diskon:  res.Diskon,
			Total:   res.TotalAmount,
		}
		report.BillingSummary.TotalTagihan.Count += res.Count
		report.BillingSummary.TotalTagihan.Nominal += res.Nominal
		report.BillingSummary.TotalTagihan.Diskon += res.Diskon
		report.BillingSummary.TotalTagihan.Total += res.TotalAmount
		
		// Calculate taxes for this status
		tax := domain.TaxStat{}
		if res.TotalAmount > 0 {
			tax.Ppn = math.Floor(res.TotalAmount - (res.TotalAmount / 1.11) + 0.5)
			revenueExclPpn := res.TotalAmount - tax.Ppn
			tax.Bhp = math.Floor(revenueExclPpn * 0.005 + 0.5)
			tax.Uso = math.Floor(revenueExclPpn * 0.0125 + 0.5)
			tax.TotalPajak = tax.Ppn + tax.Bhp + tax.Uso
		}

		switch res.Status {
		case "Lunas":
			report.BillingSummary.Lunas = stat
			report.TotalPendapatan = res.TotalAmount
			report.FinancialSummary.TotalPemasukan = res.TotalAmount
			report.TaxSummary.Lunas = tax
		case "Belum Bayar":
			report.BillingSummary.Pending = stat
			report.TaxSummary.Pending = tax
		case "Expired":
			report.BillingSummary.Expired = stat
			report.TaxSummary.Expired = tax
		}
	}
	report.FinancialSummary.SaldoAkhir = report.FinancialSummary.TotalPemasukan

	// 3. Overall Tax Summary
	// Total tax is the sum of taxes from all statuses (or just Lunas, depending on business rules)
	// The user's output showed total was the sum of something.
	// Let's make Total the sum of Lunas, Pending, and Expired tax for now to match the "TOTAL" line
	report.TaxSummary.Total.Ppn = report.TaxSummary.Lunas.Ppn + report.TaxSummary.Pending.Ppn + report.TaxSummary.Expired.Ppn
	report.TaxSummary.Total.Bhp = report.TaxSummary.Lunas.Bhp + report.TaxSummary.Pending.Bhp + report.TaxSummary.Expired.Bhp
	report.TaxSummary.Total.Uso = report.TaxSummary.Lunas.Uso + report.TaxSummary.Pending.Uso + report.TaxSummary.Expired.Uso
	report.TaxSummary.Total.TotalPajak = report.TaxSummary.Total.Ppn + report.TaxSummary.Total.Bhp + report.TaxSummary.Total.Uso

	// 4. Payment Methods (Only for Lunas)
	type MethodResult struct {
		Method      string
		Count       int
		TotalAmount float64
		Pajak       float64
		Diskon      float64
	}
	var methodResults []MethodResult
	err = buildBaseQuery().
		Select("metode_pembayaran as method, COUNT(*) as count, SUM(total_harga) as total_amount, SUM(diskon_amount) as diskon").
		Where("status_invoice = ?", "Lunas").
		Group("metode_pembayaran").
		Scan(&methodResults).Error
	if err != nil {
		return nil, err
	}

	for _, res := range methodResults {
		method := res.Method
		if method == "" {
			method = "Lainnya"
		}
		// Calculate PPN for this method
		ppn := math.Floor(res.TotalAmount - (res.TotalAmount / 1.11) + 0.5)
		
		report.PaymentMethods = append(report.PaymentMethods, domain.PaymentMethodStat{
			Method:      method,
			Count:       res.Count,
			TotalAmount: res.TotalAmount,
			Pajak:       ppn,
			Diskon:      res.Diskon,
		})
	}

	return &report, nil
}

func (r *invoiceRepository) GetRevenueReportDetails(ctx context.Context, params *domain.RevenueReportParams) ([]domain.InvoiceReportItem, error) {
	var items []domain.InvoiceReportItem

	query := r.db.WithContext(ctx).Table("invoices").
		Select("invoices.id, invoices.invoice_number, pelanggan.nama as pelanggan_nama, pelanggan.alamat, invoices.total_harga, invoices.status_invoice, invoices.tgl_invoice, invoices.paid_at as tgl_lunas, invoices.metode_pembayaran as metode, invoices.brand").
		Joins("LEFT JOIN pelanggan ON invoices.pelanggan_id = pelanggan.id").
		Where("invoices.deleted_at IS NULL")

	if params.StartDate != "" {
		query = query.Where("invoices.tgl_invoice >= ?", params.StartDate)
	}
	if params.EndDate != "" {
		query = query.Where("invoices.tgl_invoice <= ?", params.EndDate+" 23:59:59")
	}
	if params.Alamat != "" {
		query = query.Where("pelanggan.alamat = ?", params.Alamat)
	}
	if params.IDBrand != "" {
		query = query.Where("pelanggan.id_brand = ?", params.IDBrand)
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Skip > 0 {
		query = query.Offset(params.Skip)
	}

	err := query.Order("invoices.tgl_invoice desc").Scan(&items).Error
	return items, err
}

func (r *invoiceRepository) ExportPaymentLinksExcel(ctx context.Context, filters map[string]string) ([]byte, error) {
	var invoices []domain.Invoice
	query := r.db.WithContext(ctx).Preload("Pelanggan").Where("deleted_at IS NULL")

	if s, ok := filters["search"]; ok && s != "" {
		query = query.Joins("Pelanggan").Where("pelanggan.nama LIKE ? OR invoices.invoice_number LIKE ?", "%"+s+"%", "%"+s+"%")
	}
	if s, ok := filters["status_invoice"]; ok && s != "" {
		query = query.Where("status_invoice = ?", s)
	}
	if s, ok := filters["start_date"]; ok && s != "" {
		query = query.Where("tgl_invoice >= ?", s)
	}
	if s, ok := filters["end_date"]; ok && s != "" {
		query = query.Where("tgl_invoice <= ?", s+" 23:59:59")
	}

	if err := query.Order("tgl_invoice desc").Find(&invoices).Error; err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := "Payment Links"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"No", "Pelanggan", "Invoice Number", "Alamat", "Total", "Status", "Link Pembayaran"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for r, inv := range invoices {
		row := r + 2
		pName := ""; if inv.Pelanggan != nil { pName = inv.Pelanggan.Nama }
		addr := ""; if inv.Pelanggan != nil { addr = inv.Pelanggan.Alamat }
		link := ""; if inv.PaymentLink != nil { link = *inv.PaymentLink }

		vals := []interface{}{r + 1, pName, inv.InvoiceNumber, addr, inv.TotalHarga, inv.StatusInvoice, link}
		for c, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Langganan Repository
type langgananRepository struct {
	db *gorm.DB
}

func NewLanggananRepository(db *gorm.DB) domain.LanggananRepository {
	return &langgananRepository{db: db}
}

func (r *langgananRepository) GetAll(ctx context.Context, limit, offset int, search, status string, forInvoiceSelection bool, sortBy, sortOrder string) ([]domain.Langganan, int64, error) {
	var langganans []domain.Langganan
	var total int64

	dbCount := r.db.WithContext(ctx).Model(&domain.Langganan{}).
		Joins("JOIN pelanggan ON pelanggan.id = langganan.pelanggan_id AND pelanggan.deleted_at IS NULL").
		Joins("LEFT JOIN data_teknis ON data_teknis.pelanggan_id = pelanggan.id")

	dbFind := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("PaketLayanan").
		Preload("PaketLayanan.HargaLayanan").
		Joins("JOIN pelanggan ON pelanggan.id = langganan.pelanggan_id AND pelanggan.deleted_at IS NULL").
		Joins("LEFT JOIN data_teknis ON data_teknis.pelanggan_id = pelanggan.id")

	if forInvoiceSelection {
		dbCount = dbCount.Where("langganan.status != ?", "Berhenti")
		dbFind = dbFind.Where("langganan.status != ?", "Berhenti")
	}

	if status != "" {
		dbCount = dbCount.Where("langganan.status = ?", status)
		dbFind = dbFind.Where("langganan.status = ?", status)
	}

	if search != "" {
		searchTerm := "%" + search + "%"
		dbCount = dbCount.Where("pelanggan.nama LIKE ? OR data_teknis.id_pelanggan LIKE ? OR pelanggan.no_telp LIKE ? OR pelanggan.email LIKE ?", searchTerm, searchTerm, searchTerm, searchTerm)
		dbFind = dbFind.Where("pelanggan.nama LIKE ? OR data_teknis.id_pelanggan LIKE ? OR pelanggan.no_telp LIKE ? OR pelanggan.email LIKE ?", searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if err := dbCount.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause from sort parameters with whitelist validation
	orderClause := "langganan.id desc" // default
	allowedSortColumns := map[string]string{
		"pelanggan.nama":     "pelanggan.nama",
		"status":             "langganan.status",
		"harga_final":        "langganan.harga_awal",
		"tgl_jatuh_tempo":    "langganan.tgl_jatuh_tempo",
		"tgl_berhenti":       "langganan.tgl_berhenti",
		"metode_pembayaran":  "langganan.metode_pembayaran",
		"created_at":         "langganan.created_at",
		"id":                 "langganan.id",
	}
	if col, ok := allowedSortColumns[sortBy]; ok {
		direction := "asc"
		if sortOrder == "desc" {
			direction = "desc"
		}
		orderClause = col + " " + direction
	}

	err := dbFind.
		Limit(limit).
		Offset(offset).
		Order(orderClause).
		Find(&langganans).Error

	return langganans, total, err
}

func (r *langgananRepository) GetByID(ctx context.Context, id uint64) (*domain.Langganan, error) {
	var lng domain.Langganan
	err := r.db.WithContext(ctx).Preload("Pelanggan").Preload("Pelanggan.DataTeknis").Preload("PaketLayanan").First(&lng, id).Error
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
		Where("invoices.status_invoice IN ?", []string{"Belum Bayar", "Expired"}).
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
		Joins("JOIN pelanggan ON pelanggan.id = langganan.pelanggan_id AND pelanggan.deleted_at IS NULL").
		Joins("LEFT JOIN invoices ON invoices.pelanggan_id = pelanggan.id").
		Where("langganan.status = ?", "Aktif").
		Where("invoices.id IS NULL").
		Where("langganan.pelanggan_id NOT IN (?)", archivedSubquery).
		Order("langganan.id desc").
		Find(&langganans).Error

	return langganans, err
}

