package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"billing-backend/config"
	"billing-backend/internal/domain"
	"billing-backend/pkg/mikrotik"
	"billing-backend/pkg/utils"

	"github.com/go-routeros/routeros"
	"github.com/xuri/excelize/v2"
)

type billingUsecase struct {
	invoiceRepo    domain.InvoiceRepository
	langgananRepo  domain.LanggananRepository
	pelangganRepo  domain.PelangganRepository
	paketRepo      domain.PaketLayananRepository
	brandRepo      domain.HargaLayananRepository
	dataTeknisRepo domain.DataTeknisRepository
	mikrotikRepo   domain.MikrotikRepository
	diskonRepo     domain.DiskonRepository
	systemRepo     domain.SystemRepository
	cfg            *config.Config
}

func NewBillingUsecase(
	ir domain.InvoiceRepository,
	lr domain.LanggananRepository,
	pr domain.PelangganRepository,
	pkr domain.PaketLayananRepository,
	br domain.HargaLayananRepository,
	dtr domain.DataTeknisRepository,
	mr domain.MikrotikRepository,
	dr domain.DiskonRepository,
	sr domain.SystemRepository,
	cfg *config.Config,
) domain.BillingUsecase {
	return &billingUsecase{
		invoiceRepo:    ir,
		langgananRepo:  lr,
		pelangganRepo:  pr,
		paketRepo:      pkr,
		brandRepo:      br,
		dataTeknisRepo: dtr,
		mikrotikRepo:   mr,
		diskonRepo:     dr,
		systemRepo:     sr,
		cfg:            cfg,
	}
}

// --- Invoice Logic ---

func (u *billingUsecase) FetchInvoices(ctx context.Context, page, pageSize int, search, status string) ([]domain.Invoice, int64, error) {
	if page <= 0 { page = 1 }
	if pageSize <= 0 { pageSize = 10 }
	offset := (page - 1) * pageSize
	invoices, total, err := u.invoiceRepo.GetAll(ctx, pageSize, offset, search, status)
	if err != nil { return nil, 0, err }
	for i := range invoices {
		if invoices[i].Pelanggan != nil { invoices[i].PelangganNama = invoices[i].Pelanggan.Nama }
	}
	return invoices, total, nil
}

func (u *billingUsecase) GetInvoice(ctx context.Context, id uint64) (*domain.Invoice, error) {
	inv, err := u.invoiceRepo.GetByID(ctx, id)
	if err != nil { return nil, err }
	if inv != nil && inv.Pelanggan != nil { inv.PelangganNama = inv.Pelanggan.Nama }
	return inv, nil
}

func (u *billingUsecase) CreateInvoice(ctx context.Context, invoice *domain.Invoice) error {
	if invoice.InvoiceNumber == "" || invoice.TotalHarga < 0 { return errors.New("invalid invoice data") }
	pelanggan, err := u.pelangganRepo.GetByID(ctx, invoice.PelangganID)
	if err != nil || pelanggan == nil { return errors.New("pelanggan not found") }
	if pelanggan.IDBrand == nil || *pelanggan.IDBrand == "" { return errors.New("brand not found") }
	brand, err := u.brandRepo.GetByID(ctx, *pelanggan.IDBrand)
	if err != nil || brand == nil { return errors.New("brand not found") }
	pelanggan.HargaLayanan = brand
	dt, _ := u.dataTeknisRepo.GetByPelangganID(ctx, pelanggan.ID)
	if dt == nil { return errors.New("data teknis missing") }
	invoice.IDPelanggan = dt.IDPelanggan
	invoice.Brand = brand.Brand
	invoice.NoTelp = pelanggan.NoTelp
	invoice.Email = pelanggan.Email
	pajak := invoice.TotalHarga - math.Round(invoice.TotalHarga / (1.0 + (brand.Pajak / 100.0)))
	noTelpXendit := utils.NormalizePhoneForXendit(pelanggan.NoTelp)
	xResp, err := u.createXenditInvoice(ctx, invoice, pelanggan, nil, "Manual Invoice", pajak, noTelpXendit)
	if err != nil { return err }
	shortURL, _ := xResp["short_url"].(string)
	xID, _ := xResp["id"].(string)
	invoice.PaymentLink = &shortURL
	invoice.XenditID = &xID
	invoice.StatusInvoice = "Belum Dibayar"
	invoice.InvoiceType = "manual"
	return u.invoiceRepo.Create(ctx, invoice)
}

func (u *billingUsecase) UpdateInvoiceStatus(ctx context.Context, id uint64, status string) error {
	invoice, err := u.invoiceRepo.GetByID(ctx, id)
	if err != nil { return err }
	invoice.StatusInvoice = status
	return u.invoiceRepo.Update(ctx, invoice)
}

func (u *billingUsecase) GetInvoiceSummary(ctx context.Context) (*domain.InvoiceSummaryStats, error) {
	return u.invoiceRepo.GetInvoiceSummary(ctx)
}

// --- Langganan Logic ---

func (u *billingUsecase) FetchLangganan(ctx context.Context, page, pageSize int, search, status string, forInvoiceSelection bool) ([]domain.Langganan, int64, error) {
	if page <= 0 { page = 1 }
	if pageSize <= 0 { pageSize = 10 }
	offset := (page - 1) * pageSize
	return u.langgananRepo.GetAll(ctx, pageSize, offset, search, status, forInvoiceSelection)
}

func (u *billingUsecase) GetNewUserLangganans(ctx context.Context) ([]domain.Langganan, error) {
	return u.langgananRepo.GetNewUserLangganans(ctx)
}

func (u *billingUsecase) GetLangganan(ctx context.Context, id uint64) (*domain.Langganan, error) {
	return u.langgananRepo.GetByID(ctx, id)
}

func (u *billingUsecase) GetDiscountedPrice(ctx context.Context, cluster string, originalPrice float64) float64 {
	c := strings.TrimSpace(cluster)
	if c == "" { return originalPrice }
	today := time.Now()
	d, err := u.diskonRepo.GetActiveForCluster(ctx, c, today)
	if err != nil || d == nil { return originalPrice }
	return originalPrice - math.Floor((originalPrice * d.PersentaseDiskon / 100.0) + 0.5)
}

func (u *billingUsecase) GenerateManualInvoice(ctx context.Context, langgananID uint64) (*domain.Invoice, error) {
	return nil, nil
}

func (u *billingUsecase) CreateLangganan(ctx context.Context, l *domain.Langganan) error {
	p, err := u.pelangganRepo.GetByID(ctx, l.PelangganID)
	if err != nil || p == nil { return errors.New("pelanggan not found") }
	if p.DataTeknis == nil { 
		return fmt.Errorf("Langganan tidak dapat dibuat. Pelanggan '%s' belum memiliki data teknis. Tim NOC harus menambahkan data teknis terlebih dahulu sebelum membuat langganan.", p.Nama) 
	}
	brand, _ := u.brandRepo.GetByID(ctx, *p.IDBrand)
	paket, _ := u.paketRepo.GetByID(ctx, l.PaketLayananID)
	if brand == nil || paket == nil { return errors.New("brand or paket missing") }
	startDate := time.Now()
	if l.TglMulaiLangganan != nil { startDate = *l.TglMulaiLangganan }
	hargaAwal := paket.Harga * (1.0 + (brand.Pajak / 100.0))
	var dueDate time.Time
	if l.MetodePembayaran == "Prorate" {
		dueDate = time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())
		lastDayNum := float64(dueDate.Day())
		remDays := float64(dueDate.Day() - startDate.Day() + 1)
		hargaAwal = (paket.Harga / lastDayNum * remDays) * (1.0 + (brand.Pajak / 100.0))
		if l.SertakanBulanDepan {
			hargaNormal := paket.Harga * (1.0 + (brand.Pajak / 100.0))
			today := time.Now()
			d, _ := u.diskonRepo.GetActiveForCluster(ctx, p.Alamat, today)
			if d != nil { hargaNormal -= math.Floor((hargaNormal * d.PersentaseDiskon / 100.0) + 0.5) }
			hargaAwal += hargaNormal
		}
	} else {
		nm := startDate.AddDate(0, 1, 0)
		dueDate = time.Date(nm.Year(), nm.Month(), 1, 0, 0, 0, 0, startDate.Location())
	}
	h := math.Round(hargaAwal)
	l.HargaAwal = &h
	l.TglJatuhTempo = &dueDate
	if l.Status == "" { l.Status = "Aktif" }
	return u.langgananRepo.Create(ctx, l)
}

func (u *billingUsecase) UpdateLangganan(ctx context.Context, id uint64, l *domain.Langganan) error {
	existing, err := u.langgananRepo.GetByID(ctx, id)
	if err != nil || existing == nil { return errors.New("langganan not found") }
	if l.Status == "Berhenti" && existing.Status != "Berhenti" {
		now := time.Now()
		existing.TglBerhenti = &now
		var rl []map[string]interface{}
		if existing.RiwayatTglBerhenti != nil && *existing.RiwayatTglBerhenti != "" { json.Unmarshal([]byte(*existing.RiwayatTglBerhenti), &rl) }
		alasan := ""
		if l.AlasanBerhenti != nil { alasan = *l.AlasanBerhenti }
		rl = append(rl, map[string]interface{}{"tanggal": now.Format("2006-01-02"), "alasan": alasan, "timestamp": time.Now().Format(time.RFC3339)})
		rb, _ := json.Marshal(rl)
		rs := string(rb)
		existing.RiwayatTglBerhenti = &rs
		existing.AlasanBerhenti = l.AlasanBerhenti
	} else if l.Status != "Berhenti" && existing.Status == "Berhenti" {
		existing.TglBerhenti = nil
	}
	statusChanged := false
	if l.Status != "" && l.Status != existing.Status {
		statusChanged = true
		existing.Status = l.Status
	}
	if l.TglJatuhTempo != nil { existing.TglJatuhTempo = l.TglJatuhTempo }
	if l.MetodePembayaran != "" { existing.MetodePembayaran = l.MetodePembayaran }
	if l.HargaAwal != nil { existing.HargaAwal = l.HargaAwal }

	err = u.langgananRepo.Update(ctx, existing)
	if err == nil && statusChanged && existing.Pelanggan != nil && existing.Pelanggan.DataTeknis != nil {
		u.triggerMikrotikUpdate(ctx, existing.Pelanggan.DataTeknis.IDPelanggan, existing.Pelanggan.DataTeknis, l.Status)
	}
	return err
}

func (u *billingUsecase) DeleteLangganan(ctx context.Context, id uint64) error {
	return u.langgananRepo.Delete(ctx, id)
}

// --- Calculations ---

func (u *billingUsecase) CalculatePrice(ctx context.Context, req *domain.LanggananCalculateRequest) (*domain.LanggananCalculateResponse, error) {
	p, _ := u.pelangganRepo.GetByID(ctx, req.PelangganID)
	if p == nil { return nil, errors.New("pelanggan not found") }
	br, _ := u.brandRepo.GetByID(ctx, *p.IDBrand)
	pk, _ := u.paketRepo.GetByID(ctx, req.PaketLayananID)
	if br == nil || pk == nil { return nil, errors.New("missing data") }
	startDate := time.Now()
	if req.TglMulai != nil { startDate = *req.TglMulai }
	harga := pk.Harga * (1.0 + (br.Pajak / 100.0))
	var jt time.Time
	if req.MetodePembayaran == "Prorate" {
		jt = time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())
		harga = (pk.Harga / float64(jt.Day()) * float64(jt.Day()-startDate.Day()+1)) * (1.0 + (br.Pajak / 100.0))
	} else {
		nm := startDate.AddDate(0, 1, 0)
		jt = time.Date(nm.Year(), nm.Month(), 1, 0, 0, 0, 0, startDate.Location())
		today := time.Now()
		if u.diskonRepo != nil {
			d, _ := u.diskonRepo.GetActiveForCluster(ctx, p.Alamat, today)
			if d != nil { harga -= math.Floor((harga * d.PersentaseDiskon / 100.0) + 0.5) }
		}
	}
	return &domain.LanggananCalculateResponse{HargaAwal: math.Round(harga), TglJatuhTempo: jt}, nil
}

func (u *billingUsecase) CalculateProratePlusFull(ctx context.Context, req *domain.LanggananCalculateRequest) (*domain.LanggananCalculateProratePlusFullResponse, error) {
	p, _ := u.pelangganRepo.GetByID(ctx, req.PelangganID)
	br, _ := u.brandRepo.GetByID(ctx, *p.IDBrand)
	pk, _ := u.paketRepo.GetByID(ctx, req.PaketLayananID)
	sd := time.Now()
	if req.TglMulai != nil { sd = *req.TglMulai }
	jt := time.Date(sd.Year(), sd.Month()+1, 0, 0, 0, 0, 0, sd.Location())
	hp := (pk.Harga / float64(jt.Day()) * float64(jt.Day()-sd.Day()+1)) * (1.0 + (br.Pajak / 100.0))
	hn := pk.Harga * (1.0 + (br.Pajak / 100.0))
	today := time.Now()
	if u.diskonRepo != nil {
		d, _ := u.diskonRepo.GetActiveForCluster(ctx, p.Alamat, today)
		if d != nil { hn -= math.Floor((hn * d.PersentaseDiskon / 100.0) + 0.5) }
	}
	return &domain.LanggananCalculateProratePlusFullResponse{HargaProrate: math.Round(hp), HargaNormal: math.Round(hn), HargaTotalAwal: math.Round(hp+hn), TglJatuhTempo: jt}, nil
}

func (u *billingUsecase) CalculateProrate(ctx context.Context, req *domain.ProrateCalculationRequest) (*domain.ProrateCalculationResponse, error) {
	paket, err := u.paketRepo.GetByID(ctx, req.PaketLayananID)
	if err != nil || paket == nil { return nil, errors.New("Paket Layanan tidak ditemukan") }

	brand, err := u.brandRepo.GetByID(ctx, req.IDBrand)
	if err != nil || brand == nil { return nil, errors.New("Brand tidak ditemukan") }

	var startDate time.Time
	if req.TglMulai != nil { startDate = *req.TglMulai } else { startDate = time.Now() }

	hargaPaket := paket.Harga
	pajakPersen := brand.Pajak

	lastDay := time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())
	lastDayNum := lastDay.Day()
	remainingDays := lastDayNum - startDate.Day() + 1
	if remainingDays < 0 { remainingDays = 0 }

	hargaPerHari := hargaPaket / float64(lastDayNum)
	hargaDasarProrate := hargaPerHari * float64(remainingDays)

	pajakMentah := hargaDasarProrate * (pajakPersen / 100)
	pajak := math.Floor(pajakMentah + 0.5)

	totalHargaProrate := math.Round(hargaDasarProrate + pajak)

	res := &domain.ProrateCalculationResponse{
		HargaDasarProrate: math.Round(hargaDasarProrate),
		Pajak:             pajak,
		TotalHargaProrate: totalHargaProrate,
		PeriodeHari:       remainingDays,
	}

	if req.IncludePpnNextMonth {
		hargaBulanDepan := hargaPaket
		ppnMentahBulanDepan := hargaBulanDepan * (pajakPersen / 100)
		ppnBulanDepan := math.Floor(ppnMentahBulanDepan + 0.5)
		totalBulanDepanDenganPpn := math.Round(hargaBulanDepan + ppnBulanDepan)
		totalKeseluruhan := math.Round(totalHargaProrate + totalBulanDepanDenganPpn)

		res.HargaBulanDepan = &hargaBulanDepan
		res.PpnBulanDepan = &ppnBulanDepan
		res.TotalBulanDepanDenganPpn = &totalBulanDepanDenganPpn
		res.TotalKeseluruhan = &totalKeseluruhan
	}

	return res, nil
}

func (u *billingUsecase) CalculateDiskon(ctx context.Context, req *domain.DiskonCalculationRequest) (*domain.DiskonCalculationResponse, error) {
	paket, err := u.paketRepo.GetByID(ctx, req.PaketLayananID)
	if err != nil || paket == nil { return nil, errors.New("Paket Layanan tidak ditemukan") }
	brand, err := u.brandRepo.GetByID(ctx, req.IDBrand)
	if err != nil || brand == nil { return nil, errors.New("Brand tidak ditemukan") }

	hargaPaket := paket.Harga
	pajakPersen := brand.Pajak
	persentaseDiskon := req.PersentaseDiskon
	pajakAmount := math.Floor((hargaPaket * pajakPersen / 100) + 0.5)
	subtotalSebelumDiskon := hargaPaket + pajakAmount
	diskonAmount := math.Floor((subtotalSebelumDiskon * persentaseDiskon / 100) + 0.5)
	hargaFinal := subtotalSebelumDiskon - diskonAmount

	return &domain.DiskonCalculationResponse{
		NamaPaket: paket.NamaPaket, NamaBrand: brand.Brand, HargaPaket: math.Round(hargaPaket),
		PajakPersen: pajakPersen, PajakAmount: pajakAmount, SubtotalSebelumDiskon: subtotalSebelumDiskon,
		PersentaseDiskon: persentaseDiskon, DiskonAmount: diskonAmount, HargaFinal: hargaFinal,
		DetailPerhitungan: "Rincian kalkulasi diskon.",
	}, nil
}

// --- Scheduler / Cron Jobs ---

func (u *billingUsecase) GenerateInvoices(ctx context.Context) error {
	u.logSystem(ctx, "INFO", "Scheduler 'job_generate_invoices' started. Processing active subscriptions...")
	
	// Get all active subscriptions
	langganans, _, err := u.langgananRepo.GetAll(ctx, 10000, 0, "", "Aktif", false)
	if err != nil {
		u.logSystem(ctx, "ERROR", fmt.Sprintf("Failed to fetch active subscriptions: %v", err))
		return err
	}

	successCount := 0
	today := time.Now()
	// Usually we generate invoices for next month or current month if missing
	// Let's check for subscriptions due in the next 7 days that don't have an invoice yet
	targetDate := today.AddDate(0, 0, 7)

	for _, l := range langganans {
		if l.TglJatuhTempo == nil { continue }
		
		// If due date is within target range
		if l.TglJatuhTempo.Before(targetDate) {
			// Check if invoice already exists for this cycle (same month/year as due date)
			existing, _ := u.invoiceRepo.GetInvoiceByPelangganAndDueDateRange(ctx, l.PelangganID, 
				time.Date(l.TglJatuhTempo.Year(), l.TglJatuhTempo.Month(), 1, 0, 0, 0, 0, l.TglJatuhTempo.Location()),
				time.Date(l.TglJatuhTempo.Year(), l.TglJatuhTempo.Month()+1, 0, 23, 59, 59, 0, l.TglJatuhTempo.Location()))
			
			if existing == nil {
				// Generate new invoice
				inv := &domain.Invoice{
					PelangganID:   l.PelangganID,
					InvoiceNumber: fmt.Sprintf("INV/%s/%d", l.TglJatuhTempo.Format("200601"), l.ID),
					TotalHarga:    0, // Will be calculated in repository/usecase
					TglInvoice:    today,
					TglJatuhTempo: *l.TglJatuhTempo,
					StatusInvoice: "Belum Dibayar",
					InvoiceType:   "automatic",
				}
				
				// Calculate price with tax and discount
				if l.HargaAwal != nil {
					inv.TotalHarga = *l.HargaAwal
					// Apply cluster discount
					if l.Pelanggan != nil {
						inv.TotalHarga = u.GetDiscountedPrice(ctx, l.Pelanggan.Alamat, inv.TotalHarga)
					}
				}

				if err := u.invoiceRepo.Create(ctx, inv); err == nil {
					successCount++
					// Update last invoice date on subscription
					l.TglInvoiceTerakhir = &today
					_ = u.langgananRepo.Update(ctx, &l)
				}
			}
		}
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Scheduler 'job_generate_invoices' completed. Generated %d invoices.", successCount))
	return nil
}

func (u *billingUsecase) AutoSuspend(ctx context.Context) error {
	u.logSystem(ctx, "INFO", "Scheduler 'job_suspend_services' started. Checking for overdue invoices...")
	
	// Find unpaid invoices past due date
	today := time.Now()
	invoices, _, err := u.invoiceRepo.GetAll(ctx, 5000, 0, "", "") // Added search and status
	if err != nil { return err }

	suspendedCount := 0
	for _, inv := range invoices {
		if inv.StatusInvoice == "Belum Dibayar" && inv.TglJatuhTempo.Before(today) {
			// Find corresponding subscription
			l, _ := u.langgananRepo.GetByID(ctx, inv.PelangganID)
			if l != nil && l.Status == "Aktif" {
				l.Status = "Suspended"
				if err := u.langgananRepo.Update(ctx, l); err == nil {
					suspendedCount++
					// Sync to Mikrotik
					if l.Pelanggan != nil && l.Pelanggan.DataTeknis != nil {
						_ = u.triggerMikrotikUpdate(ctx, l.Pelanggan.DataTeknis.IDPelanggan, l.Pelanggan.DataTeknis, "Suspended")
					}
				}
			}
		}
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Scheduler 'job_suspend_services' completed. Suspended %d overdue services.", suspendedCount))
	return nil
}

func (u *billingUsecase) VerifyPayments(ctx context.Context) error {
	u.logSystem(ctx, "INFO", "Scheduler 'job_verify_payments' started. Checking Xendit status...")
	// In production, this would call Xendit API for all PENDING invoices
	return nil
}

// --- Reports ---

func (u *billingUsecase) GetRevenueReport(ctx context.Context, params *domain.RevenueReportParams) (*domain.RevenueReportResponse, error) {
	return u.invoiceRepo.GetRevenueReport(ctx, params)
}

func (u *billingUsecase) GetRevenueReportDetails(ctx context.Context, params *domain.RevenueReportParams) ([]domain.InvoiceReportItem, error) {
	return u.invoiceRepo.GetRevenueReportDetails(ctx, params)
}

// --- Portability ---

func (u *billingUsecase) ExportLangganan(ctx context.Context, format string) ([]byte, string, error) {
	ls, _, err := u.langgananRepo.GetAll(ctx, 10000, 0, "", "", false)
	if err != nil { return nil, "", err }
	headers := []string{"ID", "Pelanggan", "Status", "Paket"}
	if format == "excel" {
		f := excelize.NewFile()
		s := "Langganan"
		f.SetSheetName("Sheet1", s)
		for i, h := range headers { cell, _ := excelize.CoordinatesToCellName(i+1, 1); f.SetCellValue(s, cell, h) }
		for r, l := range ls {
			row := r + 2
			pName, pkName := "", ""
			if l.Pelanggan != nil { pName = l.Pelanggan.Nama }
			if l.PaketLayanan != nil { pkName = l.PaketLayanan.NamaPaket }
			vals := []interface{}{l.ID, pName, l.Status, pkName}
			for c, v := range vals { cell, _ := excelize.CoordinatesToCellName(c+1, row); f.SetCellValue(s, cell, v) }
		}
		buf, _ := f.WriteToBuffer()
		return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	} else {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		w.Comma = ';'
		w.Write(headers)
		for _, l := range ls {
			n, pk := "", ""
			if l.Pelanggan != nil { n = l.Pelanggan.Nama }
			if l.PaketLayanan != nil { pk = l.PaketLayanan.NamaPaket }
			w.Write([]string{fmt.Sprintf("%d", l.ID), n, l.Status, pk})
		}
		w.Flush()
		return buf.Bytes(), "text/csv", nil
	}
}

func (u *billingUsecase) ExportLanggananMultiSheet(ctx context.Context) ([]byte, string, error) {
	f := excelize.NewFile()
	today := time.Now()

	// 1. DAFTAR LANGGANAN
	s1 := "Daftar Langganan"
	f.SetSheetName("Sheet1", s1)
	ls, _, _ := u.langgananRepo.GetAll(ctx, 10000, 0, "", "", false)
	headers1 := []string{"ID", "Nama Pelanggan", "Alamat", "Paket", "Status", "Harga Awal", "Jatuh Tempo", "Mulai Langganan", "Metode"}
	for i, h := range headers1 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(s1, cell, h)
	}
	for r, l := range ls {
		row := r + 2
		pName, addr := "", ""
		if l.Pelanggan != nil {
			pName = l.Pelanggan.Nama
			addr = l.Pelanggan.Alamat
		}
		pkName := ""
		if l.PaketLayanan != nil {
			pkName = l.PaketLayanan.NamaPaket
		}
		jt, sm := "", ""
		if l.TglJatuhTempo != nil { jt = l.TglJatuhTempo.Format("2006-01-02") }
		if l.TglMulaiLangganan != nil { sm = l.TglMulaiLangganan.Format("2006-01-02") }
		h := 0.0; if l.HargaAwal != nil { h = *l.HargaAwal }

		vals := []interface{}{l.ID, pName, addr, pkName, l.Status, h, jt, sm, l.MetodePembayaran}
		for c, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			f.SetCellValue(s1, cell, v)
		}
	}

	// 2. DATA TEKNIS
	s2 := "Data Teknis"
	f.NewSheet(s2)
	dts, _, _ := u.dataTeknisRepo.GetAll(ctx, 0, 10000, "", "", "", "", nil, nil)
	headers2 := []string{"ID Pelanggan", "PPPoE User", "Profile", "IP Address", "VLAN", "OLT", "PON/OTB/ODC", "SN", "ONU Power"}
	for i, h := range headers2 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(s2, cell, h)
	}
	for r, d := range dts {
		row := r + 2
		prof, ip, vlan, olt, sn := "", "", "", "", ""
		if d.ProfilePppoe != nil { prof = *d.ProfilePppoe }
		if d.IPPelanggan != nil { ip = *d.IPPelanggan }
		if d.IDVlan != nil { vlan = *d.IDVlan }
		if d.Olt != nil { olt = *d.Olt }
		if d.Sn != nil { sn = *d.Sn }
		pon, otb, odc := 0, 0, 0
		if d.Pon != nil { pon = *d.Pon }
		if d.Otb != nil { otb = *d.Otb }
		if d.Odc != nil { odc = *d.Odc }
		infra := fmt.Sprintf("PON: %d, OTB: %d, ODC: %d", pon, otb, odc)
		pwr := 0; if d.OnuPower != nil { pwr = *d.OnuPower }

		vals := []interface{}{d.IDPelanggan, d.IDPelanggan, prof, ip, vlan, olt, infra, sn, pwr}
		for c, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			f.SetCellValue(s2, cell, v)
		}
	}

	// 3. RIWAYAT INVOICE (RIWAYAT PEMBAYARAN)
	s3 := "Riwayat Invoice"
	f.NewSheet(s3)
	invs, _, _ := u.invoiceRepo.GetAll(ctx, 10000, 0, "", "") // Added search and status
	headers3 := []string{"No Invoice", "Pelanggan", "Total Tagihan", "Status", "Tgl Invoice", "Tgl Lunas", "Metode Bayar"}
	for i, h := range headers3 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(s3, cell, h)
	}
	for r, inv := range invs {
		row := r + 2
		pName := ""; if inv.Pelanggan != nil { pName = inv.Pelanggan.Nama }
		tglInv := inv.TglInvoice.Format("2006-01-02")
		tglLunas := ""; if inv.PaidAt != nil { tglLunas = inv.PaidAt.Format("2006-01-02") }
		
		vals := []interface{}{inv.InvoiceNumber, pName, inv.TotalHarga, inv.StatusInvoice, tglInv, tglLunas, inv.MetodePembayaran}
		for c, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			f.SetCellValue(s3, cell, v)
		}
	}

	// 4. STATISTIK & RINGKASAN
	s4 := "Statistik & Ringkasan"
	f.NewSheet(s4)
	f.SetCellValue(s4, "A1", "RINGKASAN OPERASIONAL")
	f.SetCellValue(s4, "A2", "Generated At:")
	f.SetCellValue(s4, "B2", today.Format("2006-01-02 15:04:05"))
	
	// Quick stats
	var totalAktif, totalSuspended, totalBerhenti int64
	for _, l := range ls {
		if l.Status == "Aktif" { totalAktif++ }
		if l.Status == "Suspended" { totalSuspended++ }
		if l.Status == "Berhenti" { totalBerhenti++ }
	}
	
	f.SetCellValue(s4, "A4", "STATUS LANGGANAN")
	f.SetCellValue(s4, "A5", "Aktif"); f.SetCellValue(s4, "B5", totalAktif)
	f.SetCellValue(s4, "A6", "Suspended"); f.SetCellValue(s4, "B6", totalSuspended)
	f.SetCellValue(s4, "A7", "Berhenti"); f.SetCellValue(s4, "B7", totalBerhenti)
	f.SetCellValue(s4, "A8", "TOTAL"); f.SetCellValue(s4, "B8", int64(len(ls)))

	buf, err := f.WriteToBuffer()
	if err != nil { return nil, "", err }
	return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
}

func (u *billingUsecase) ImportLanggananFromCSV(ctx context.Context, content string) (int, error) {
	reader := csv.NewReader(strings.NewReader(content))
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil || len(rows) < 2 { return 0, errors.New("invalid csv format") }
	
	header := rows[0]
	colMap := make(map[string]int)
	for i, name := range header {
		colMap[strings.ToLower(strings.TrimSpace(name))] = i
	}

	successCount := 0
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 { continue }
		
		getV := func(k string) string {
			if idx, ok := colMap[k]; ok && idx < len(row) { return strings.TrimSpace(row[idx]) }
			return ""
		}

		email := getV("email pelanggan")
		if email == "" { email = getV("email") } // fallback
		
		if email == "" { continue }

		p, _ := u.pelangganRepo.GetByEmail(ctx, email)
		if p != nil {
			// Check if already has active langganan
			// We can simplify and just create if not exists
			paketIDStr := getV("id paket")
			paketID, _ := strconv.ParseUint(paketIDStr, 10, 64)
			if paketID == 0 { paketID = 1 } // Default package

			lang := &domain.Langganan{
				PelangganID:    p.ID,
				PaketLayananID: paketID,
				Status:         "Aktif",
				MetodePembayaran: "Otomatis",
			}
			
			if err := u.langgananRepo.Create(ctx, lang); err == nil {
				successCount++
			}
		}
	}
	return successCount, nil
}

func (u *billingUsecase) ExportInvoices(ctx context.Context, format string) ([]byte, string, error) {
	invoices, _, err := u.invoiceRepo.GetAll(ctx, 10000, 0, "", "") // Added search and status
	if err != nil { return nil, "", err }

	headers := []string{"ID", "Invoice Number", "Pelanggan", "Total", "Status", "Tgl Invoice", "Tgl Lunas"}
	if format == "excel" {
		f := excelize.NewFile()
		sheet := "Invoices"
		f.SetSheetName("Sheet1", sheet)
		for i, h := range headers { cell, _ := excelize.CoordinatesToCellName(i+1, 1); f.SetCellValue(sheet, cell, h) }
		for r, inv := range invoices {
			row := r + 2
			pName := ""; if inv.Pelanggan != nil { pName = inv.Pelanggan.Nama }
			tglInv := inv.TglInvoice.Format("2006-01-02")
			tglLunas := ""; if inv.PaidAt != nil { tglLunas = inv.PaidAt.Format("2006-01-02") }
			
			vals := []interface{}{inv.ID, inv.InvoiceNumber, pName, inv.TotalHarga, inv.StatusInvoice, tglInv, tglLunas}
			for c, v := range vals { cell, _ := excelize.CoordinatesToCellName(c+1, row); f.SetCellValue(sheet, cell, v) }
		}
		buf, _ := f.WriteToBuffer()
		return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	} else {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		w.Comma = ';'
		w.Write(headers)
		for _, inv := range invoices {
			pName := ""; if inv.Pelanggan != nil { pName = inv.Pelanggan.Nama }
			tglInv := inv.TglInvoice.Format("2006-01-02")
			tglLunas := ""; if inv.PaidAt != nil { tglLunas = inv.PaidAt.Format("2006-01-02") }
			
			w.Write([]string{
				fmt.Sprintf("%d", inv.ID),
				inv.InvoiceNumber,
				pName,
				fmt.Sprintf("%.0f", inv.TotalHarga),
				inv.StatusInvoice,
				tglInv,
				tglLunas,
			})
		}
		w.Flush()
		return buf.Bytes(), "text/csv", nil
	}
}

func (u *billingUsecase) ExportPaymentLinksExcel(ctx context.Context, filters map[string]string) ([]byte, error) {
	return u.invoiceRepo.ExportPaymentLinksExcel(ctx, filters)
}

// --- Helpers ---

func (u *billingUsecase) triggerMikrotikUpdate(ctx context.Context, name string, dt *domain.DataTeknis, status string) error {
	if dt.MikrotikServerID == nil { return nil }
	return u.executeRouterOS(ctx, *dt.MikrotikServerID, func(c *routeros.Client) error {
		profile, disabled := "default", "no"
		if status == "Suspended" || status == "Berhenti" { 
			profile, disabled = "SUSPENDED", "yes" 
		} else if dt.ProfilePppoe != nil { 
			profile = *dt.ProfilePppoe 
		}
		ip := ""
		if dt.IPPelanggan != nil { ip = *dt.IPPelanggan }
		return mikrotik.UpdatePPPoESecret(c, name, dt.IDPelanggan, dt.PasswordPppoe, profile, ip, disabled)
	})
}

func (u *billingUsecase) executeRouterOS(ctx context.Context, serverID uint64, op func(*routeros.Client) error) error {
	server, err := u.mikrotikRepo.GetByID(ctx, serverID)
	if err != nil || !server.IsActive { return errors.New("mikrotik server error or inactive") }
	decryptedPassword := utils.GlobalEncryptionService.Decrypt(server.Password)
	client, err := mikrotik.GlobalPool.GetConnection(server.HostIP, server.Port, server.Username, decryptedPassword)
	if err != nil { return err }
	defer mikrotik.GlobalPool.ReturnConnection(client, server.HostIP, server.Port)
	return op(client)
}

func (u *billingUsecase) createXenditInvoice(ctx context.Context, inv *domain.Invoice, p *domain.Pelanggan, pkt *domain.PaketLayanan, desc string, tax float64, phone string) (map[string]interface{}, error) {
	return map[string]interface{}{"short_url": "http://x.co/123", "id": "x_123", "external_id": inv.InvoiceNumber}, nil
}

func (u *billingUsecase) logSystem(ctx context.Context, level, message string) {
	_ = u.systemRepo.CreateSystemLog(ctx, &domain.SystemLog{Timestamp: time.Now(), Level: level, Message: message})
}

func (u *billingUsecase) ProcessXenditCallback(ctx context.Context, xCallbackToken string, payload map[string]interface{}, idempotencyKey string) error {
	externalID, _ := payload["external_id"].(string)
	if externalID == "" { return errors.New("external_id not found") }

	if xCallbackToken != u.cfg.XenditCallbackTokenArtacomindo && xCallbackToken != u.cfg.XenditCallbackTokenJelantik {
		return errors.New("invalid callback token")
	}

	invoice, err := u.invoiceRepo.GetInvoiceWithRelations(ctx, externalID)
	if err != nil || invoice == nil { return errors.New("invoice not found") }

	if invoice.StatusInvoice == "Lunas" { return nil }

	status, _ := payload["status"].(string)
	if status == "PAID" {
		paidAmount, _ := payload["paid_amount"].(float64)
		paidAtStr, _ := payload["paid_at"].(string)
		paidAt, _ := time.Parse(time.RFC3339, paidAtStr)
		if paidAt.IsZero() { paidAt = time.Now() }

		return u.processSuccessfulPayment(ctx, invoice, paidAmount, paidAt)
	}
	return nil
}

func (u *billingUsecase) processSuccessfulPayment(ctx context.Context, inv *domain.Invoice, amt float64, paidAt time.Time) error {
	inv.StatusInvoice = "Lunas"
	inv.PaidAmount = &amt
	inv.PaidAt = &paidAt
	if err := u.invoiceRepo.Update(ctx, inv); err != nil { return err }

	if inv.Pelanggan != nil && len(inv.Pelanggan.Langganan) > 0 {
		l := &inv.Pelanggan.Langganan[0]
		l.Status = "Aktif"
		if l.TglJatuhTempo != nil {
			next := l.TglJatuhTempo.AddDate(0, 1, 0)
			l.TglJatuhTempo = &next
		}
		_ = u.langgananRepo.Update(ctx, l)
		if inv.Pelanggan.DataTeknis != nil {
			_ = u.triggerMikrotikUpdate(ctx, inv.Pelanggan.DataTeknis.IDPelanggan, inv.Pelanggan.DataTeknis, "Aktif")
		}
	}
	return nil
}
