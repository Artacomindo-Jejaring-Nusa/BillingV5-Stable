package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"billing-backend/config"
	"billing-backend/internal/domain"
	"billing-backend/internal/websocket"
	"billing-backend/pkg/database"
	"billing-backend/pkg/logger"
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

func (u *billingUsecase) logActivity(ctx context.Context, action string, details string) {
	if u.systemRepo == nil {
		return
	}
	log := &domain.ActivityLog{
		UserID:    utils.GetUserIDFromCtx(ctx),
		Action:    action,
		Details:   &details,
		Timestamp: time.Now(),
	}
	_ = u.systemRepo.CreateActivityLog(ctx, log)
}

// --- Invoice Logic ---

func (u *billingUsecase) FetchInvoices(ctx context.Context, page, pageSize int, search, status string) ([]domain.Invoice, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	invoices, total, err := u.invoiceRepo.GetAll(ctx, pageSize, offset, search, status)
	if err != nil {
		return nil, 0, err
	}
	for i := range invoices {
		if invoices[i].Pelanggan != nil {
			invoices[i].PelangganNama = invoices[i].Pelanggan.Nama
		}
	}
	return invoices, total, nil
}

func (u *billingUsecase) GetInvoice(ctx context.Context, id uint64) (*domain.Invoice, error) {
	inv, err := u.invoiceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv != nil && inv.Pelanggan != nil {
		inv.PelangganNama = inv.Pelanggan.Nama
	}
	return inv, nil
}

func (u *billingUsecase) CreateInvoice(ctx context.Context, invoice *domain.Invoice) error {
	if invoice.InvoiceNumber == "" || invoice.TotalHarga < 0 {
		return errors.New("invalid invoice data")
	}
	pelanggan, err := u.pelangganRepo.GetByID(ctx, invoice.PelangganID)
	if err != nil || pelanggan == nil {
		return errors.New("pelanggan not found")
	}
	if pelanggan.IDBrand == nil || *pelanggan.IDBrand == "" {
		return errors.New("brand not found")
	}
	brand, err := u.brandRepo.GetByID(ctx, *pelanggan.IDBrand)
	if err != nil || brand == nil {
		return errors.New("brand not found")
	}
	pelanggan.HargaLayanan = brand
	dt, _ := u.dataTeknisRepo.GetByPelangganID(ctx, pelanggan.ID)
	if dt == nil {
		return errors.New("data teknis missing")
	}
	invoice.IDPelanggan = dt.IDPelanggan
	invoice.Brand = brand.Brand
	invoice.NoTelp = pelanggan.NoTelp
	invoice.Email = pelanggan.Email
	pajak := invoice.TotalHarga - math.Round(invoice.TotalHarga/(1.0+(brand.Pajak/100.0)))
	noTelpXendit := utils.NormalizePhoneForXendit(pelanggan.NoTelp)
	deskripsi := fmt.Sprintf("Biaya berlangganan internet jatuh tempo pembayaran tanggal %s", invoice.TglJatuhTempo.Format("02/01/2006"))
	xResp, err := u.createXenditInvoice(ctx, invoice, pelanggan, nil, deskripsi, pajak, noTelpXendit)
	if err != nil {
		return err
	}
	shortURL, _ := xResp["short_url"].(string)
	xID, _ := xResp["id"].(string)
	invoice.PaymentLink = &shortURL
	invoice.XenditID = &xID
	invoice.StatusInvoice = "Belum Bayar"
	invoice.InvoiceType = "manual"
	err = u.invoiceRepo.Create(ctx, invoice)
	if err == nil {
		u.logActivity(ctx, "Create Invoice", fmt.Sprintf("Created invoice: %s for Pelanggan ID %d", invoice.InvoiceNumber, invoice.PelangganID))
		websocket.InvalidateDashboardCache(ctx)
	}
	return err
}

func (u *billingUsecase) UpdateInvoiceStatus(ctx context.Context, id uint64, status string) error {
	invoice, err := u.invoiceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	oldStatus := invoice.StatusInvoice
	invoice.StatusInvoice = status
	err = u.invoiceRepo.Update(ctx, invoice)
	if err == nil {
		u.logActivity(ctx, "Update Invoice Status", fmt.Sprintf("Updated invoice %s status from %s to %s", invoice.InvoiceNumber, oldStatus, status))
		websocket.InvalidateDashboardCache(ctx)
		if status == "Lunas" && oldStatus != "Lunas" {
			if websocket.GlobalHub != nil {
				pName := invoice.PelangganNama
				if pName == "" && invoice.Pelanggan != nil {
					pName = invoice.Pelanggan.Nama
				}
				websocket.GlobalHub.BroadcastNotification("new_payment", map[string]interface{}{
					"invoice_number": invoice.InvoiceNumber,
					"pelanggan_nama": pName,
					"amount":         invoice.TotalHarga,
				})
			}
		}
	}
	return err
}

func (u *billingUsecase) GetInvoiceSummary(ctx context.Context) (*domain.InvoiceSummaryStats, error) {
	return u.invoiceRepo.GetInvoiceSummary(ctx)
}

// --- Langganan Logic ---

func (u *billingUsecase) FetchLangganan(ctx context.Context, page, pageSize int, search, status string, forInvoiceSelection bool, sortBy, sortOrder string) ([]domain.Langganan, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return u.langgananRepo.GetAll(ctx, pageSize, offset, search, status, forInvoiceSelection, sortBy, sortOrder)
}

func (u *billingUsecase) GetNewUserLangganans(ctx context.Context) ([]domain.Langganan, error) {
	return u.langgananRepo.GetNewUserLangganans(ctx)
}

func (u *billingUsecase) GetLangganan(ctx context.Context, id uint64) (*domain.Langganan, error) {
	return u.langgananRepo.GetByID(ctx, id)
}

func (u *billingUsecase) GetDiscountedPrice(ctx context.Context, cluster string, originalPrice float64) float64 {
	c := strings.TrimSpace(cluster)
	if c == "" {
		return originalPrice
	}
	today := time.Now()
	d, err := u.diskonRepo.GetActiveForCluster(ctx, c, today)
	if err != nil || d == nil {
		return originalPrice
	}
	return originalPrice - math.Floor((originalPrice*d.PersentaseDiskon/100.0)+0.5)
}

func (u *billingUsecase) GenerateManualInvoice(ctx context.Context, langgananID uint64) (*domain.Invoice, error) {
	langganan, err := u.langgananRepo.GetByID(ctx, langgananID)
	if err != nil || langganan == nil {
		return nil, errors.New("langganan not found")
	}

	pelanggan, err := u.pelangganRepo.GetByID(ctx, langganan.PelangganID)
	if err != nil || pelanggan == nil {
		return nil, errors.New("pelanggan not found")
	}

	if pelanggan.IDBrand == nil || *pelanggan.IDBrand == "" {
		return nil, errors.New("brand not found")
	}

	brand, err := u.brandRepo.GetByID(ctx, *pelanggan.IDBrand)
	if err != nil || brand == nil {
		return nil, errors.New("brand not found")
	}
	pelanggan.HargaLayanan = brand

	dt, _ := u.dataTeknisRepo.GetByPelangganID(ctx, pelanggan.ID)

	now := time.Now()
	harga := langganan.HargaAwal
	if harga == nil {
		paket, err := u.paketRepo.GetByID(ctx, langganan.PaketLayananID)
		if err != nil || paket == nil {
			return nil, errors.New("paket not found")
		}
		h := paket.Harga * (1.0 + (brand.Pajak / 100.0))
		harga = &h
	}

	totalHarga := math.Round(*harga)
	// Prioritaskan tgl_jatuh_tempo_pembayaran (Kotak 3) sebagai jatuh tempo pembayaran invoice
	dueDate := langganan.TglJatuhTempoPembayaran
	if dueDate == nil {
		dueDate = langganan.TglJatuhTempo // Fallback ke perilaku lama
	}
	if dueDate == nil {
		d := now.AddDate(0, 1, 0)
		dueDate = &d
	}

	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	namaSingkat := strings.ToUpper(re.ReplaceAllString(pelanggan.Nama, ""))
	brandSingkat := strings.ToUpper(re.ReplaceAllString(brand.Brand, ""))
	alamatSingkat := utils.GenerateAlamatSingkat(pelanggan.Alamat, pelanggan.Blok, pelanggan.Unit, 10)
	bulanTahun := fmt.Sprintf("%s-%d", strings.ToUpper(dueDate.Month().String()), dueDate.Year())

	idSuffix := "TMP"
	if dt != nil && dt.IDPelanggan != "" {
		if len(dt.IDPelanggan) >= 3 {
			idSuffix = dt.IDPelanggan[len(dt.IDPelanggan)-3:]
		} else {
			idSuffix = dt.IDPelanggan
		}
	}

	invoiceNumber := fmt.Sprintf("%s/ftth/%s/%s/%s/%s", brandSingkat, namaSingkat, bulanTahun, alamatSingkat, idSuffix)
	counter := 0
	for {
		existing, err := u.invoiceRepo.GetByInvoiceNumber(ctx, invoiceNumber)
		if err != nil {
			if err.Error() == "invoice not found" {
				break
			}
			return nil, err
		}
		if existing != nil {
			counter++
			invoiceNumber = fmt.Sprintf("%s/ftth/%s/%s/%s/%s-%d", brandSingkat, namaSingkat, bulanTahun, alamatSingkat, idSuffix, counter)
		} else {
			break
		}
	}

	invoice := &domain.Invoice{
		InvoiceNumber: invoiceNumber,
		PelangganID:   langganan.PelangganID,
		TotalHarga:    totalHarga,
		TglInvoice:    now,
		TglJatuhTempo: *dueDate,
		StatusInvoice: "Belum Bayar",
		InvoiceType:   "manual",
		Brand:         brand.Brand,
		NoTelp:        pelanggan.NoTelp,
		Email:         pelanggan.Email,
	}

	if dt != nil {
		invoice.IDPelanggan = dt.IDPelanggan
	}

	noTelpXendit := utils.NormalizePhoneForXendit(pelanggan.NoTelp)
	pajak := totalHarga - math.Round(totalHarga/(1.0+(brand.Pajak/100.0)))
	// Load Paket Layanan untuk Kecepatan
	paket, _ := u.paketRepo.GetByID(ctx, langganan.PaketLayananID)
	kecepatan := 0
	if paket != nil {
		kecepatan = paket.Kecepatan
	}

	var itemPrefix string
	if kecepatan > 0 {
		itemPrefix = fmt.Sprintf("Biaya berlangganan internet up to %d Mbps", kecepatan)
	} else {
		itemPrefix = "Biaya berlangganan internet"
	}

	var deskripsi string
	if langganan.MetodePembayaran == "Prorate" {
		periodeStart := now
		if langganan.TglMulaiLangganan != nil {
			periodeStart = *langganan.TglMulaiLangganan
		}

		targetEnd := *dueDate
		if langganan.TglJatuhTempo != nil {
			targetEnd = *langganan.TglJatuhTempo
		}

		periodeEnd := targetEnd
		if targetEnd.Day() == 1 {
			periodeEnd = targetEnd.AddDate(0, 0, -1)
		}

		getIndonesianMonth := func(m time.Month) string {
			months := map[time.Month]string{
				time.January:   "Januari",
				time.February:  "Februari",
				time.March:     "Maret",
				time.April:     "April",
				time.May:       "Mei",
				time.June:      "Juni",
				time.July:      "Juli",
				time.August:    "Agustus",
				time.September: "September",
				time.October:   "Oktober",
				time.November:  "November",
				time.December:  "Desember",
			}
			return months[m]
		}

		var periodDesc string
		if periodeStart.Month() == periodeEnd.Month() && periodeStart.Year() == periodeEnd.Year() {
			periodDesc = fmt.Sprintf("Periode Tgl %d-%d %s %d",
				periodeStart.Day(), periodeEnd.Day(),
				getIndonesianMonth(periodeEnd.Month()), periodeEnd.Year())
		} else if periodeStart.Year() == periodeEnd.Year() {
			periodDesc = fmt.Sprintf("Periode Tgl %d %s - %d %s %d",
				periodeStart.Day(), getIndonesianMonth(periodeStart.Month()),
				periodeEnd.Day(), getIndonesianMonth(periodeEnd.Month()), periodeEnd.Year())
		} else {
			periodDesc = fmt.Sprintf("Periode Tgl %d %s %d - %d %s %d",
				periodeStart.Day(), getIndonesianMonth(periodeStart.Month()), periodeStart.Year(),
				periodeEnd.Day(), getIndonesianMonth(periodeEnd.Month()), periodeEnd.Year())
		}

		deskripsi = fmt.Sprintf("%s, %s", itemPrefix, periodDesc)
	} else {
		deskripsi = fmt.Sprintf("%s jatuh tempo pembayaran tanggal %s",
			itemPrefix, dueDate.Format("02/01/2006"))
	}

	xResp, xErr := u.createXenditInvoice(ctx, invoice, pelanggan, nil, deskripsi, pajak, noTelpXendit)
	if xErr == nil {
		shortURL, _ := xResp["short_url"].(string)
		if shortURL == "" {
			shortURL, _ = xResp["invoice_url"].(string)
		}
		xID, _ := xResp["id"].(string)
		if shortURL != "" {
			invoice.PaymentLink = &shortURL
		}
		if xID != "" {
			invoice.XenditID = &xID
		}
		u.logSystem(ctx, "INFO", fmt.Sprintf("Xendit invoice created: %s", shortURL))
	} else {
		u.logSystem(ctx, "ERROR", fmt.Sprintf("Gagal create Xendit invoice: %v", xErr))
	}

	if err := u.invoiceRepo.Create(ctx, invoice); err != nil {
		return nil, err
	}
	u.logActivity(ctx, "Generate Manual Invoice", fmt.Sprintf("Generated manual invoice %s for Langganan ID %d", invoice.InvoiceNumber, langgananID))
	websocket.InvalidateDashboardCache(ctx)
	return invoice, nil
}

func (u *billingUsecase) DeleteInvoice(ctx context.Context, id uint64) error {
	invoice, err := u.invoiceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	err = u.invoiceRepo.Delete(ctx, id)
	if err == nil {
		u.logActivity(ctx, "Delete Invoice", fmt.Sprintf("Deleted invoice %s (ID: %d)", invoice.InvoiceNumber, id))
		websocket.InvalidateDashboardCache(ctx)
	}
	return err
}

func (u *billingUsecase) CreateLangganan(ctx context.Context, l *domain.Langganan) error {
	p, err := u.pelangganRepo.GetByID(ctx, l.PelangganID)
	if err != nil || p == nil {
		return errors.New("pelanggan not found")
	}
	if p.DataTeknis == nil {
		return fmt.Errorf("Langganan tidak dapat dibuat. Pelanggan '%s' belum memiliki data teknis. Tim NOC harus menambahkan data teknis terlebih dahulu sebelum membuat langganan.", p.Nama)
	}
	brand, _ := u.brandRepo.GetByID(ctx, *p.IDBrand)
	paket, _ := u.paketRepo.GetByID(ctx, l.PaketLayananID)
	if brand == nil || paket == nil {
		return errors.New("brand or paket missing")
	}
	startDate := time.Now()
	if l.TglMulaiLangganan != nil {
		startDate = *l.TglMulaiLangganan
	}
	hargaAwal := paket.Harga * (1.0 + (brand.Pajak / 100.0))
	var dueDate time.Time
	var payDate time.Time
	var tglMulai time.Time

	if l.MetodePembayaran == "Prorate" {
		tglMulai = startDate
		dueDate = time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())
		payDate = startDate

		lastDayNum := float64(dueDate.Day())
		remDays := float64(dueDate.Day() - startDate.Day() + 1)
		hargaAwal = (paket.Harga / lastDayNum * remDays) * (1.0 + (brand.Pajak / 100.0))
		if l.SertakanBulanDepan {
			dueDate = time.Date(startDate.Year(), startDate.Month()+2, 0, 0, 0, 0, 0, startDate.Location())
			hargaNormal := paket.Harga * (1.0 + (brand.Pajak / 100.0))
			today := time.Now()
			if u.diskonRepo != nil {
				d, _ := u.diskonRepo.GetActiveForCluster(ctx, p.Alamat, today)
				if d != nil {
					hargaNormal -= math.Floor((hargaNormal * d.PersentaseDiskon / 100.0) + 0.5)
				}
			}
			hargaAwal += hargaNormal
		}
	} else {
		// Otomatis — tanggal jatuh tempo selalu tanggal 1 bulan depan
		// Sesuai logika legacy Python: (start_date + relativedelta(months=1)).replace(day=1)
		tglMulai = startDate
		dueDate = time.Date(startDate.Year(), startDate.Month()+1, 1, 0, 0, 0, 0, startDate.Location())
		payDate = startDate
	}
	h := math.Round(hargaAwal)
	l.HargaAwal = &h
	l.TglMulaiLangganan = &tglMulai
	l.TglJatuhTempo = &dueDate
	l.TglJatuhTempoPembayaran = &payDate
	if l.Status == "" {
		l.Status = "Aktif"
	}
	err = u.langgananRepo.Create(ctx, l)
	if err == nil {
		u.logActivity(ctx, "Create Langganan", fmt.Sprintf("Created subscription for Pelanggan ID %d (ID: %d)", l.PelangganID, l.ID))
		websocket.InvalidateDashboardCache(ctx)
	}
	return err
}

func (u *billingUsecase) UpdateLangganan(ctx context.Context, id uint64, l *domain.Langganan) error {
	existing, err := u.langgananRepo.GetByID(ctx, id)
	if err != nil || existing == nil {
		return errors.New("langganan not found")
	}
	if l.Status == "Berhenti" && existing.Status != "Berhenti" {
		now := time.Now()
		existing.TglBerhenti = &now
		var rl []map[string]interface{}
		if existing.RiwayatTglBerhenti != nil && *existing.RiwayatTglBerhenti != "" {
			json.Unmarshal([]byte(*existing.RiwayatTglBerhenti), &rl)
		}
		alasan := ""
		if l.AlasanBerhenti != nil {
			alasan = *l.AlasanBerhenti
		}
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
	if l.TglJatuhTempo != nil {
		existing.TglJatuhTempo = l.TglJatuhTempo
	}
	if l.TglJatuhTempoPembayaran != nil {
		existing.TglJatuhTempoPembayaran = l.TglJatuhTempoPembayaran
	}
	if l.TglMulaiLangganan != nil {
		existing.TglMulaiLangganan = l.TglMulaiLangganan
	}
	if l.MetodePembayaran != "" {
		existing.MetodePembayaran = l.MetodePembayaran
	}
	if l.HargaAwal != nil {
		existing.HargaAwal = l.HargaAwal
	}

	err = u.langgananRepo.Update(ctx, existing)
	if err == nil {
		u.logActivity(ctx, "Update Langganan", fmt.Sprintf("Updated subscription ID: %d to status: %s", id, existing.Status))
		websocket.InvalidateDashboardCache(ctx)
		if statusChanged && existing.Pelanggan != nil && existing.Pelanggan.DataTeknis != nil {
			u.triggerMikrotikUpdate(ctx, existing.Pelanggan.DataTeknis.IDPelanggan, existing.Pelanggan.DataTeknis, l.Status)
		}
	}
	return err
}

func (u *billingUsecase) DeleteLangganan(ctx context.Context, id uint64) error {
	err := u.langgananRepo.Delete(ctx, id)
	if err == nil {
		u.logActivity(ctx, "Delete Langganan", fmt.Sprintf("Deleted subscription ID: %d", id))
		websocket.InvalidateDashboardCache(ctx)
	}
	return err
}

// --- Calculations ---

func (u *billingUsecase) CalculatePrice(ctx context.Context, req *domain.LanggananCalculateRequest) (*domain.LanggananCalculateResponse, error) {
	p, _ := u.pelangganRepo.GetByID(ctx, req.PelangganID)
	if p == nil {
		return nil, errors.New("pelanggan not found")
	}
	br, _ := u.brandRepo.GetByID(ctx, *p.IDBrand)
	pk, _ := u.paketRepo.GetByID(ctx, req.PaketLayananID)
	if br == nil || pk == nil {
		return nil, errors.New("missing data")
	}
	startDate := time.Now()
	if req.TglMulai != nil {
		startDate = *req.TglMulai
	}
	harga := pk.Harga * (1.0 + (br.Pajak / 100.0))
	var jt time.Time
	var jtp time.Time
	var tml time.Time

	if req.MetodePembayaran == "Prorate" {
		tml = startDate
		jt = time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())
		harga = (pk.Harga / float64(jt.Day()) * float64(jt.Day()-startDate.Day()+1)) * (1.0 + (br.Pajak / 100.0))
		jtp = startDate
	} else {
		// Otomatis — tanggal jatuh tempo selalu tanggal 1 bulan depan
		// Sesuai logika legacy Python: (start_date + relativedelta(months=1)).replace(day=1)
		tml = startDate
		jt = time.Date(startDate.Year(), startDate.Month()+1, 1, 0, 0, 0, 0, startDate.Location())
		jtp = startDate

		today := time.Now()
		if u.diskonRepo != nil {
			d, _ := u.diskonRepo.GetActiveForCluster(ctx, p.Alamat, today)
			if d != nil {
				harga -= math.Floor((harga * d.PersentaseDiskon / 100.0) + 0.5)
			}
		}
	}
	return &domain.LanggananCalculateResponse{
		HargaAwal:               math.Round(harga),
		TglJatuhTempo:           jt,
		TglJatuhTempoPembayaran: &jtp,
		TglMulaiLangganan:       &tml,
	}, nil
}

func (u *billingUsecase) CalculateProratePlusFull(ctx context.Context, req *domain.LanggananCalculateRequest) (*domain.LanggananCalculateProratePlusFullResponse, error) {
	p, _ := u.pelangganRepo.GetByID(ctx, req.PelangganID)
	br, _ := u.brandRepo.GetByID(ctx, *p.IDBrand)
	pk, _ := u.paketRepo.GetByID(ctx, req.PaketLayananID)
	sd := time.Now()
	if req.TglMulai != nil {
		sd = *req.TglMulai
	}

	// End of next month
	jt := time.Date(sd.Year(), sd.Month()+2, 0, 0, 0, 0, 0, sd.Location())

	// Prorate price based on current month remaining days
	currentMonthEnd := time.Date(sd.Year(), sd.Month()+1, 0, 0, 0, 0, 0, sd.Location())
	hp := (pk.Harga / float64(currentMonthEnd.Day()) * float64(currentMonthEnd.Day()-sd.Day()+1)) * (1.0 + (br.Pajak / 100.0))
	hn := pk.Harga * (1.0 + (br.Pajak / 100.0))
	today := time.Now()
	if u.diskonRepo != nil {
		d, _ := u.diskonRepo.GetActiveForCluster(ctx, p.Alamat, today)
		if d != nil {
			hn -= math.Floor((hn * d.PersentaseDiskon / 100.0) + 0.5)
		}
	}
	jtp := sd
	tml := sd
	return &domain.LanggananCalculateProratePlusFullResponse{
		HargaProrate:            math.Round(hp),
		HargaNormal:             math.Round(hn),
		HargaTotalAwal:          math.Round(hp + hn),
		TglJatuhTempo:           jt,
		TglJatuhTempoPembayaran: &jtp,
		TglMulaiLangganan:       &tml,
	}, nil
}

func (u *billingUsecase) CalculateProrate(ctx context.Context, req *domain.ProrateCalculationRequest) (*domain.ProrateCalculationResponse, error) {
	paket, err := u.paketRepo.GetByID(ctx, req.PaketLayananID)
	if err != nil || paket == nil {
		return nil, errors.New("Paket Layanan tidak ditemukan")
	}

	brand, err := u.brandRepo.GetByID(ctx, req.IDBrand)
	if err != nil || brand == nil {
		return nil, errors.New("Brand tidak ditemukan")
	}

	var startDate time.Time
	if req.TglMulai != nil {
		startDate = *req.TglMulai
	} else {
		startDate = time.Now()
	}

	hargaPaket := paket.Harga
	pajakPersen := brand.Pajak

	lastDay := time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())
	lastDayNum := lastDay.Day()
	remainingDays := lastDayNum - startDate.Day() + 1
	if remainingDays < 0 {
		remainingDays = 0
	}

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
	if err != nil || paket == nil {
		return nil, errors.New("Paket Layanan tidak ditemukan")
	}
	brand, err := u.brandRepo.GetByID(ctx, req.IDBrand)
	if err != nil || brand == nil {
		return nil, errors.New("Brand tidak ditemukan")
	}

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

	limit := 500
	offset := 0
	successCount := 0

	// Gunakan zona waktu Asia/Jakarta agar perhitungan tanggal konsisten dengan legacy Python
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	today := time.Now().In(loc)
	// H-5 sebelum tanggal jatuh tempo.
	targetDate := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, loc).AddDate(0, 0, 5)

	for {
		// Set parameter forInvoiceSelection menjadi true agar query selalu membaca dari Master Database (menghindari repl lag)
		langganans, _, err := u.langgananRepo.GetAll(ctx, limit, offset, "", "Aktif", true, "", "")
		if err != nil {
			u.logSystem(ctx, "ERROR", fmt.Sprintf("Failed to fetch active subscriptions: %v", err))
			return err
		}
		u.logSystem(ctx, "INFO", fmt.Sprintf("GenerateInvoices: Found %d active subscriptions in batch", len(langganans)))
		if len(langganans) == 0 {
			break
		}

		for _, l := range langganans {
			targetDue := l.TglJatuhTempoPembayaran
			if targetDue == nil {
				targetDue = l.TglJatuhTempo
			}
			if targetDue == nil {
				u.logSystem(ctx, "INFO", fmt.Sprintf("GenerateInvoices: Langganan ID %d has nil due dates, skipping", l.ID))
				continue
			}

			// Konversi ke zona waktu Asia/Jakarta untuk pembandingan presisi
			subDue := targetDue.In(loc)
			isBefore := subDue.Before(targetDate)
			u.logSystem(ctx, "INFO", fmt.Sprintf("GenerateInvoices: Checking Langganan ID %d (Pelanggan ID %d). subDue: %s, targetDate: %s, isBefore: %t", 
				l.ID, l.PelangganID, subDue.Format("2006-01-02 15:04:05 Z0700"), targetDate.Format("2006-01-02 15:04:05 Z0700"), isBefore))

			// Jika jatuh tempo berada pada atau sebelum targetDate (H-5)
			if isBefore {
				// Check if invoice already exists for this cycle (same month/year as due date)
				existing, err := u.invoiceRepo.GetInvoiceByPelangganAndDueDateRange(ctx, l.PelangganID,
					time.Date(l.TglJatuhTempo.Year(), l.TglJatuhTempo.Month(), 1, 0, 0, 0, 0, l.TglJatuhTempo.Location()),
					time.Date(l.TglJatuhTempo.Year(), l.TglJatuhTempo.Month()+1, 0, 23, 59, 59, 0, l.TglJatuhTempo.Location()))
				if err != nil {
					u.logSystem(ctx, "ERROR", fmt.Sprintf("GenerateInvoices: Gagal mengecek invoice existing untuk pelanggan %d: %v", l.PelangganID, err))
					continue
				}

				if existing != nil {
					u.logSystem(ctx, "INFO", fmt.Sprintf("GenerateInvoices: Langganan ID %d skipped because invoice already exists: ID %d, Number %s", l.ID, existing.ID, existing.InvoiceNumber))
					continue
				}

				if existing == nil {
					// Fetch full customer details to prevent missing values
					pelanggan, err := u.pelangganRepo.GetByID(ctx, l.PelangganID)
					if err != nil || pelanggan == nil {
						u.logSystem(ctx, "ERROR", fmt.Sprintf("GenerateInvoices: Pelanggan tidak ditemukan untuk langganan ID %d: %v", l.ID, err))
						continue
					}

					if pelanggan.IDBrand == nil || *pelanggan.IDBrand == "" {
						u.logSystem(ctx, "ERROR", fmt.Sprintf("GenerateInvoices: IDBrand tidak diset untuk pelanggan ID %d", pelanggan.ID))
						continue
					}

					brand, err := u.brandRepo.GetByID(ctx, *pelanggan.IDBrand)
					if err != nil || brand == nil {
						u.logSystem(ctx, "ERROR", fmt.Sprintf("GenerateInvoices: Brand %s tidak ditemukan untuk pelanggan ID %d: %v", *pelanggan.IDBrand, pelanggan.ID, err))
						continue
					}
					pelanggan.HargaLayanan = brand

					dt, _ := u.dataTeknisRepo.GetByPelangganID(ctx, pelanggan.ID)

					// Prioritaskan tgl_jatuh_tempo_pembayaran (Kotak 3) sebagai jatuh tempo pembayaran invoice
					dueDate := l.TglJatuhTempoPembayaran
					if dueDate == nil {
						dueDate = l.TglJatuhTempo
					}

					// Generate proper Indonesian invoice number format
					re := regexp.MustCompile(`[^a-zA-Z0-9]`)
					namaSingkat := strings.ToUpper(re.ReplaceAllString(pelanggan.Nama, ""))
					brandSingkat := strings.ToUpper(re.ReplaceAllString(brand.Brand, ""))
					alamatSingkat := utils.GenerateAlamatSingkat(pelanggan.Alamat, pelanggan.Blok, pelanggan.Unit, 10)

					namingDate := l.TglJatuhTempo
					if l.MetodePembayaran == "Prorate" && namingDate.Day() == 1 {
						d := namingDate.AddDate(0, 0, -1)
						namingDate = &d
					}
					bulanTahun := fmt.Sprintf("%s-%d", strings.ToUpper(namingDate.Month().String()), namingDate.Year())

					idSuffix := "TMP"
					if dt != nil && dt.IDPelanggan != "" {
						if len(dt.IDPelanggan) >= 3 {
							idSuffix = dt.IDPelanggan[len(dt.IDPelanggan)-3:]
						} else {
							idSuffix = dt.IDPelanggan
						}
					}

					invoiceNumber := fmt.Sprintf("%s/ftth/%s/%s/%s/%s", brandSingkat, namaSingkat, bulanTahun, alamatSingkat, idSuffix)
					
					// Avoid duplicates
					counter := 0
					for {
						existingInv, err := u.invoiceRepo.GetByInvoiceNumber(ctx, invoiceNumber)
						if err != nil {
							if err.Error() == "invoice not found" {
								break
							}
							u.logSystem(ctx, "ERROR", fmt.Sprintf("GenerateInvoices: Gagal mengecek nomor invoice duplicate %s: %v", invoiceNumber, err))
							break
						}
						if existingInv != nil {
							counter++
							invoiceNumber = fmt.Sprintf("%s/ftth/%s/%s/%s/%s-%d", brandSingkat, namaSingkat, bulanTahun, alamatSingkat, idSuffix, counter)
						} else {
							break
						}
					}

					inv := &domain.Invoice{
						InvoiceNumber: invoiceNumber,
						PelangganID:   l.PelangganID,
						TotalHarga:    0,
						TglInvoice:    today,
						TglJatuhTempo: *dueDate,
						StatusInvoice: "Belum Bayar",
						InvoiceType:   "automatic",
						Brand:         brand.Brand,
						NoTelp:        pelanggan.NoTelp,
						Email:         pelanggan.Email,
					}
					if dt != nil {
						inv.IDPelanggan = dt.IDPelanggan
					}

					// Calculate price with tax and discount
					var originalPrice float64
					if l.HargaAwal != nil {
						originalPrice = *l.HargaAwal
					} else {
						paket, err := u.paketRepo.GetByID(ctx, l.PaketLayananID)
						if err == nil && paket != nil {
							originalPrice = paket.Harga * (1.0 + (brand.Pajak / 100.0))
						}
					}

					inv.TotalHarga = originalPrice
					if l.MetodePembayaran == "Prorate" {
						// Prorate users do not get discounts (legacy logic)
					} else {
						discountedPrice := u.GetDiscountedPrice(ctx, pelanggan.Alamat, originalPrice)
						if discountedPrice < originalPrice {
							inv.TotalHarga = discountedPrice
							activeDiscount, _ := u.diskonRepo.GetActiveForCluster(ctx, strings.TrimSpace(pelanggan.Alamat), today)
							if activeDiscount != nil {
								inv.DiskonID = &activeDiscount.ID
								pVal := activeDiscount.PersentaseDiskon
								inv.DiskonPersen = &pVal
								dAmount := math.Floor((originalPrice * pVal / 100.0) + 0.5)
								inv.DiskonAmount = &dAmount
								inv.HargaSebelumDiskon = &originalPrice
							}
						}
					}
					inv.TotalHarga = math.Round(inv.TotalHarga)

					// Calculate tax and format description for Xendit
					noTelpXendit := utils.NormalizePhoneForXendit(pelanggan.NoTelp)
					pajak := inv.TotalHarga - math.Round(inv.TotalHarga/(1.0+(brand.Pajak/100.0)))
					
					var itemPrefix string
					paket, _ := u.paketRepo.GetByID(ctx, l.PaketLayananID)
					if paket != nil && paket.Kecepatan > 0 {
						itemPrefix = fmt.Sprintf("Biaya berlangganan internet up to %d Mbps", paket.Kecepatan)
					} else {
						itemPrefix = "Biaya berlangganan internet"
					}

					var xenditDesc string
					if l.MetodePembayaran == "Prorate" {
						periodeStart := today
						if l.TglMulaiLangganan != nil {
							periodeStart = *l.TglMulaiLangganan
						}
						targetEnd := *dueDate
						if l.TglJatuhTempo != nil {
							targetEnd = *l.TglJatuhTempo
						}
						periodeEnd := targetEnd
						if targetEnd.Day() == 1 {
							periodeEnd = targetEnd.AddDate(0, 0, -1)
						}

						getIndonesianMonth := func(m time.Month) string {
							months := map[time.Month]string{
								time.January:   "Januari",
								time.February:  "Februari",
								time.March:     "Maret",
								time.April:     "April",
								time.May:       "Mei",
								time.June:      "Juni",
								time.July:      "Juli",
								time.August:    "Agustus",
								time.September: "September",
								time.October:   "Oktober",
								time.November:  "November",
								time.December:  "Desember",
							}
							return months[m]
						}

						var periodDesc string
						if periodeStart.Month() == periodeEnd.Month() && periodeStart.Year() == periodeEnd.Year() {
							periodDesc = fmt.Sprintf("Periode Tgl %d-%d %s %d",
								periodeStart.Day(), periodeEnd.Day(),
								getIndonesianMonth(periodeEnd.Month()), periodeEnd.Year())
						} else if periodeStart.Year() == periodeEnd.Year() {
							periodDesc = fmt.Sprintf("Periode Tgl %d %s - %d %s %d",
								periodeStart.Day(), getIndonesianMonth(periodeStart.Month()),
								periodeEnd.Day(), getIndonesianMonth(periodeEnd.Month()), periodeEnd.Year())
						} else {
							periodDesc = fmt.Sprintf("Periode Tgl %d %s %d - %d %s %d",
								periodeStart.Day(), getIndonesianMonth(periodeStart.Month()), periodeStart.Year(),
								periodeEnd.Day(), getIndonesianMonth(periodeEnd.Month()), periodeEnd.Year())
						}
						xenditDesc = fmt.Sprintf("%s, %s", itemPrefix, periodDesc)
					} else {
						xenditDesc = fmt.Sprintf("%s jatuh tempo pembayaran tanggal %s",
							itemPrefix, dueDate.Format("02/01/2006"))
					}

					// Create Xendit invoice
					xResp, xErr := u.createXenditInvoice(ctx, inv, pelanggan, nil, xenditDesc, pajak, noTelpXendit)
					if xErr == nil {
						shortURL, _ := xResp["short_url"].(string)
						if shortURL == "" {
							shortURL, _ = xResp["invoice_url"].(string)
						}
						xID, _ := xResp["id"].(string)
						if shortURL != "" {
							inv.PaymentLink = &shortURL
						}
						if xID != "" {
							inv.XenditID = &xID
						}
					} else {
						u.logSystem(ctx, "ERROR", fmt.Sprintf("GenerateInvoices: Gagal membuat Xendit invoice untuk pelanggan %d: %v", l.PelangganID, xErr))
						errMsg := xErr.Error()
						inv.XenditStatus = "failed"
						inv.XenditErrorMessage = &errMsg
					}

					if err := u.invoiceRepo.Create(ctx, inv); err != nil {
						u.logSystem(ctx, "ERROR", fmt.Sprintf("GenerateInvoices: Gagal menyimpan invoice %s ke DB untuk pelanggan %d: %v", inv.InvoiceNumber, l.PelangganID, err))
					} else {
						successCount++
						l.TglInvoiceTerakhir = &today
						_ = u.langgananRepo.Update(ctx, &l)

						if inv.PaymentLink != nil {
							u.logSystem(ctx, "INFO", fmt.Sprintf("GenerateInvoices: Invoice %s dibuat sukses untuk pelanggan %s dengan payment link: %s", inv.InvoiceNumber, pelanggan.Nama, *inv.PaymentLink))
						} else {
							u.logSystem(ctx, "WARN", fmt.Sprintf("GenerateInvoices: Invoice %s disimpan sebagai draft failed (tanpa link) untuk pelanggan %s", inv.InvoiceNumber, pelanggan.Nama))
						}
					}
				}
			}
		}

		if len(langganans) < limit {
			break
		}
		offset += limit
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Scheduler 'job_generate_invoices' completed. Generated %d invoices.", successCount))
	return nil
}

func (u *billingUsecase) AutoSuspend(ctx context.Context) error {
	u.logSystem(ctx, "INFO", "Scheduler 'job_suspend_services' started. Checking for overdue invoices...")

	db := database.GetDB()
	suspendedCount := 0

	if db != nil {
		today := time.Now()
		limit := 500
		offset := 0

		for {
			var invoices []domain.Invoice
			err := db.WithContext(ctx).
				Preload("Pelanggan").
				Preload("Pelanggan.DataTeknis").
				Preload("Pelanggan.Langganan", "status = ?", "Aktif").
				Where("status_invoice = ? AND tgl_jatuh_tempo < ?", "Belum Bayar", today).
				Limit(limit).
				Offset(offset).
				Order("id desc").
				Find(&invoices).Error

			if err != nil {
				u.logSystem(ctx, "ERROR", fmt.Sprintf("Failed to fetch overdue invoices: %v", err))
				return err
			}

			if len(invoices) == 0 {
				break
			}

			for _, inv := range invoices {
				if inv.Pelanggan == nil || len(inv.Pelanggan.Langganan) == 0 {
					continue
				}

				l := &inv.Pelanggan.Langganan[0]
				l.Status = "Suspended"
				if err := u.langgananRepo.Update(ctx, l); err == nil {
					suspendedCount++
					// Sync to Mikrotik
					if inv.Pelanggan.DataTeknis != nil {
						_ = u.triggerMikrotikUpdate(ctx, inv.Pelanggan.DataTeknis.IDPelanggan, inv.Pelanggan.DataTeknis, "Suspended")
					}
				}
			}

			if len(invoices) < limit {
				break
			}
			offset += limit
		}
	} else {
		// Fallback for unit testing where database connection is mock/nil
		today := time.Now()
		invoices, _, err := u.invoiceRepo.GetAll(ctx, 5000, 0, "", "")
		if err != nil {
			return err
		}

		for _, inv := range invoices {
			if inv.StatusInvoice == "Belum Bayar" && inv.TglJatuhTempo.Before(today) {
				l, _ := u.langgananRepo.GetByID(ctx, inv.PelangganID)
				if l != nil && l.Status == "Aktif" {
					l.Status = "Suspended"
					if err := u.langgananRepo.Update(ctx, l); err == nil {
						suspendedCount++
						if l.Pelanggan != nil && l.Pelanggan.DataTeknis != nil {
							_ = u.triggerMikrotikUpdate(ctx, l.Pelanggan.DataTeknis.IDPelanggan, l.Pelanggan.DataTeknis, "Suspended")
						}
					}
				}
			}
		}
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Scheduler 'job_suspend_services' completed. Suspended %d overdue services.", suspendedCount))
	return nil
}

func (u *billingUsecase) getXenditInvoice(ctx context.Context, xenditID string, pelanggan *domain.Pelanggan) (map[string]interface{}, error) {
	if u.cfg == nil {
		return nil, errors.New("xendit config is nil (unit testing)")
	}
	xenditApiKey := u.cfg.XenditApiKeyJelantik
	if pelanggan != nil && pelanggan.HargaLayanan != nil && strings.ToUpper(pelanggan.HargaLayanan.XenditKeyName) == "JAKINET" {
		xenditApiKey = u.cfg.XenditApiKeyJakinet
	}

	if xenditApiKey == "" {
		return nil, errors.New("xendit API key not configured")
	}

	url := u.cfg.XenditApiUrl + "/" + xenditID
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(xenditApiKey + ":"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("xendit API error (%d): %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (u *billingUsecase) VerifyPayments(ctx context.Context) error {
	u.logSystem(ctx, "INFO", "Scheduler 'job_verify_payments' started. Checking Xendit status...")

	// 1. Fetch pending invoices (Belum Dibayar)
	invoices, _, err := u.invoiceRepo.GetAll(ctx, 1000, 0, "", "Belum Dibayar")
	if err != nil {
		u.logSystem(ctx, "ERROR", fmt.Sprintf("VerifyPayments failed to fetch pending invoices: %v", err))
		return err
	}

	verifiedCount := 0
	for _, inv := range invoices {
		if inv.XenditID == nil || *inv.XenditID == "" {
			continue
		}

		// Preload/get pelanggan full relation if not preloaded (GetAll preloads Pelanggan)
		pelanggan := inv.Pelanggan
		if pelanggan == nil && u.pelangganRepo != nil {
			p, err := u.pelangganRepo.GetByID(ctx, inv.PelangganID)
			if err == nil {
				pelanggan = p
			}
		}

		xResp, err := u.getXenditInvoice(ctx, *inv.XenditID, pelanggan)
		if err != nil {
			u.logSystem(ctx, "ERROR", fmt.Sprintf("VerifyPayments: Gagal fetch status Xendit untuk Invoice %s (Xendit ID: %s): %v", inv.InvoiceNumber, *inv.XenditID, err))
			continue
		}

		status, _ := xResp["status"].(string)
		if status == "PAID" || status == "SETTLED" {
			paidAmount, _ := xResp["amount"].(float64)
			if pAmt, ok := xResp["paid_amount"].(float64); ok && pAmt > 0 {
				paidAmount = pAmt
			}
			
			paidAtStr, _ := xResp["paid_at"].(string)
			paidAt, _ := time.Parse(time.RFC3339, paidAtStr)
			if paidAt.IsZero() {
				paidAt = time.Now()
			}

			u.logSystem(ctx, "INFO", fmt.Sprintf("VerifyPayments: Invoice %s terdeteksi PAID di Xendit. Memproses pembayaran...", inv.InvoiceNumber))
			
			err = u.processSuccessfulPayment(ctx, &inv, paidAmount, paidAt)
			if err != nil {
				u.logSystem(ctx, "ERROR", fmt.Sprintf("VerifyPayments: Gagal memproses pembayaran untuk Invoice %s: %v", inv.InvoiceNumber, err))
				continue
			}
			verifiedCount++
		} else if status == "EXPIRED" {
			u.logSystem(ctx, "INFO", fmt.Sprintf("VerifyPayments: Invoice %s terdeteksi EXPIRED di Xendit. Mengubah status lokal...", inv.InvoiceNumber))
			inv.StatusInvoice = "Expired"
			_ = u.invoiceRepo.Update(ctx, &inv)
		}
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Scheduler 'job_verify_payments' completed. Processed %d missed payments.", verifiedCount))
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
	headers := []string{"ID", "Pelanggan", "Status", "Paket"}
	limit := 1000
	offset := 0

	if format == "excel" {
		f := excelize.NewFile()
		s := "Langganan"
		f.SetSheetName("Sheet1", s)
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(s, cell, h)
		}

		row := 2
		for {
			chunk, _, err := u.langgananRepo.GetAll(ctx, limit, offset, "", "", false, "", "")
			if err != nil {
				return nil, "", err
			}
			if len(chunk) == 0 {
				break
			}

			for _, l := range chunk {
				pName, pkName := "", ""
				if l.Pelanggan != nil {
					pName = l.Pelanggan.Nama
				}
				if l.PaketLayanan != nil {
					pkName = l.PaketLayanan.NamaPaket
				}
				vals := []interface{}{l.ID, pName, l.Status, pkName}
				for c, v := range vals {
					cell, _ := excelize.CoordinatesToCellName(c+1, row)
					f.SetCellValue(s, cell, v)
				}
				row++
			}

			offset += limit
			if len(chunk) < limit {
				break
			}
		}

		buf, _ := f.WriteToBuffer()
		return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	} else {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		w.Comma = ';'
		w.Write(headers)

		for {
			chunk, _, err := u.langgananRepo.GetAll(ctx, limit, offset, "", "", false, "", "")
			if err != nil {
				return nil, "", err
			}
			if len(chunk) == 0 {
				break
			}

			for _, l := range chunk {
				n, pk := "", ""
				if l.Pelanggan != nil {
					n = l.Pelanggan.Nama
				}
				if l.PaketLayanan != nil {
					pk = l.PaketLayanan.NamaPaket
				}
				w.Write([]string{fmt.Sprintf("%d", l.ID), n, l.Status, pk})
			}

			offset += limit
			if len(chunk) < limit {
				break
			}
		}

		w.Flush()
		return buf.Bytes(), "text/csv", nil
	}
}

func (u *billingUsecase) ExportLanggananMultiSheet(ctx context.Context) ([]byte, string, error) {
	f := excelize.NewFile()
	today := time.Now()
	limit := 1000

	// 1. DAFTAR LANGGANAN
	s1 := "Daftar Langganan"
	f.SetSheetName("Sheet1", s1)
	headers1 := []string{"ID", "Nama Pelanggan", "Alamat", "Paket", "Status", "Harga Awal", "Jatuh Tempo", "Mulai Langganan", "Metode"}
	for i, h := range headers1 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(s1, cell, h)
	}

	var totalAktif, totalSuspended, totalBerhenti, totalLangganan int64
	offset1 := 0
	row1 := 2
	for {
		ls, _, err := u.langgananRepo.GetAll(ctx, limit, offset1, "", "", false, "", "")
		if err != nil {
			return nil, "", err
		}
		if len(ls) == 0 {
			break
		}

		for _, l := range ls {
			totalLangganan++
			if l.Status == "Aktif" {
				totalAktif++
			} else if l.Status == "Suspended" {
				totalSuspended++
			} else if l.Status == "Berhenti" {
				totalBerhenti++
			}

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
			if l.TglJatuhTempo != nil {
				jt = l.TglJatuhTempo.Format("2006-01-02")
			}
			if l.TglMulaiLangganan != nil {
				sm = l.TglMulaiLangganan.Format("2006-01-02")
			}
			h := 0.0
			if l.HargaAwal != nil {
				h = *l.HargaAwal
			}

			vals := []interface{}{l.ID, pName, addr, pkName, l.Status, h, jt, sm, l.MetodePembayaran}
			for c, v := range vals {
				cell, _ := excelize.CoordinatesToCellName(c+1, row1)
				f.SetCellValue(s1, cell, v)
			}
			row1++
		}

		offset1 += limit
		if len(ls) < limit {
			break
		}
	}

	// 2. DATA TEKNIS
	s2 := "Data Teknis"
	f.NewSheet(s2)
	headers2 := []string{"ID Pelanggan", "PPPoE User", "Profile", "IP Address", "VLAN", "OLT", "PON/OTB/ODC", "SN", "ONU Power"}
	for i, h := range headers2 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(s2, cell, h)
	}

	offset2 := 0
	row2 := 2
	for {
		dts, _, err := u.dataTeknisRepo.GetAll(ctx, offset2, limit, "", "", "", "", nil, nil)
		if err != nil {
			return nil, "", err
		}
		if len(dts) == 0 {
			break
		}

		for _, d := range dts {
			prof, ip, vlan, olt, sn := "", "", "", "", ""
			if d.ProfilePppoe != nil {
				prof = *d.ProfilePppoe
			}
			if d.IPPelanggan != nil {
				ip = *d.IPPelanggan
			}
			if d.IDVlan != nil {
				vlan = *d.IDVlan
			}
			if d.Olt != nil {
				olt = *d.Olt
			}
			if d.Sn != nil {
				sn = *d.Sn
			}
			pon, otb, odc := 0, 0, 0
			if d.Pon != nil {
				pon = *d.Pon
			}
			if d.Otb != nil {
				otb = *d.Otb
			}
			if d.Odc != nil {
				odc = *d.Odc
			}
			infra := fmt.Sprintf("PON: %d, OTB: %d, ODC: %d", pon, otb, odc)
			pwr := 0
			if d.OnuPower != nil {
				pwr = *d.OnuPower
			}

			vals := []interface{}{d.IDPelanggan, d.IDPelanggan, prof, ip, vlan, olt, infra, sn, pwr}
			for c, v := range vals {
				cell, _ := excelize.CoordinatesToCellName(c+1, row2)
				f.SetCellValue(s2, cell, v)
			}
			row2++
		}

		offset2 += limit
		if len(dts) < limit {
			break
		}
	}

	// 3. RIWAYAT INVOICE (RIWAYAT PEMBAYARAN)
	s3 := "Riwayat Invoice"
	f.NewSheet(s3)
	headers3 := []string{"No Invoice", "Pelanggan", "Total Tagihan", "Status", "Tgl Invoice", "Tgl Lunas", "Metode Bayar"}
	for i, h := range headers3 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(s3, cell, h)
	}

	offset3 := 0
	row3 := 2
	for {
		invs, _, err := u.invoiceRepo.GetAll(ctx, limit, offset3, "", "")
		if err != nil {
			return nil, "", err
		}
		if len(invs) == 0 {
			break
		}

		for _, inv := range invs {
			pName := ""
			if inv.Pelanggan != nil {
				pName = inv.Pelanggan.Nama
			}
			tglInv := inv.TglInvoice.Format("2006-01-02")
			tglLunas := ""
			if inv.PaidAt != nil {
				tglLunas = inv.PaidAt.Format("2006-01-02")
			}

			vals := []interface{}{inv.InvoiceNumber, pName, inv.TotalHarga, inv.StatusInvoice, tglInv, tglLunas, inv.MetodePembayaran}
			for c, v := range vals {
				cell, _ := excelize.CoordinatesToCellName(c+1, row3)
				f.SetCellValue(s3, cell, v)
			}
			row3++
		}

		offset3 += limit
		if len(invs) < limit {
			break
		}
	}

	// 4. STATISTIK & RINGKASAN
	s4 := "Statistik & Ringkasan"
	f.NewSheet(s4)
	f.SetCellValue(s4, "A1", "RINGKASAN OPERASIONAL")
	f.SetCellValue(s4, "A2", "Generated At:")
	f.SetCellValue(s4, "B2", today.Format("2006-01-02 15:04:05"))

	f.SetCellValue(s4, "A4", "STATUS LANGGANAN")
	f.SetCellValue(s4, "A5", "Aktif")
	f.SetCellValue(s4, "B5", totalAktif)
	f.SetCellValue(s4, "A6", "Suspended")
	f.SetCellValue(s4, "B6", totalSuspended)
	f.SetCellValue(s4, "A7", "Berhenti")
	f.SetCellValue(s4, "B7", totalBerhenti)
	f.SetCellValue(s4, "A8", "TOTAL")
	f.SetCellValue(s4, "B8", totalLangganan)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
}

func (u *billingUsecase) ImportLanggananFromCSV(ctx context.Context, content string) (int, error) {
	reader := csv.NewReader(strings.NewReader(content))
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil || len(rows) < 2 {
		return 0, errors.New("invalid csv format")
	}

	header := rows[0]
	colMap := make(map[string]int)
	for i, name := range header {
		colMap[strings.ToLower(strings.TrimSpace(name))] = i
	}

	successCount := 0
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		getV := func(k string) string {
			if idx, ok := colMap[k]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		email := getV("email pelanggan")
		if email == "" {
			email = getV("email")
		} // fallback

		if email == "" {
			continue
		}

		p, _ := u.pelangganRepo.GetByEmail(ctx, email)
		if p != nil {
			// Check if already has active langganan
			// We can simplify and just create if not exists
			paketIDStr := getV("id paket")
			paketID, _ := strconv.ParseUint(paketIDStr, 10, 64)
			if paketID == 0 {
				paketID = 1
			} // Default package

			lang := &domain.Langganan{
				PelangganID:      p.ID,
				PaketLayananID:   paketID,
				Status:           "Aktif",
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
	headers := []string{"ID", "Invoice Number", "Pelanggan", "Total", "Status", "Tgl Invoice", "Tgl Lunas"}
	limit := 1000
	offset := 0

	if format == "excel" {
		f := excelize.NewFile()
		sheet := "Invoices"
		f.SetSheetName("Sheet1", sheet)
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheet, cell, h)
		}

		row := 2
		for {
			invoices, _, err := u.invoiceRepo.GetAll(ctx, limit, offset, "", "")
			if err != nil {
				return nil, "", err
			}
			if len(invoices) == 0 {
				break
			}

			for _, inv := range invoices {
				pName := ""
				if inv.Pelanggan != nil {
					pName = inv.Pelanggan.Nama
				}
				tglInv := inv.TglInvoice.Format("2006-01-02")
				tglLunas := ""
				if inv.PaidAt != nil {
					tglLunas = inv.PaidAt.Format("2006-01-02")
				}

				vals := []interface{}{inv.ID, inv.InvoiceNumber, pName, inv.TotalHarga, inv.StatusInvoice, tglInv, tglLunas}
				for c, v := range vals {
					cell, _ := excelize.CoordinatesToCellName(c+1, row)
					f.SetCellValue(sheet, cell, v)
				}
				row++
			}

			offset += limit
			if len(invoices) < limit {
				break
			}
		}

		buf, _ := f.WriteToBuffer()
		return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	} else {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		w.Comma = ';'
		w.Write(headers)

		for {
			invoices, _, err := u.invoiceRepo.GetAll(ctx, limit, offset, "", "")
			if err != nil {
				return nil, "", err
			}
			if len(invoices) == 0 {
				break
			}

			for _, inv := range invoices {
				pName := ""
				if inv.Pelanggan != nil {
					pName = inv.Pelanggan.Nama
				}
				tglInv := inv.TglInvoice.Format("2006-01-02")
				tglLunas := ""
				if inv.PaidAt != nil {
					tglLunas = inv.PaidAt.Format("2006-01-02")
				}

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

			offset += limit
			if len(invoices) < limit {
				break
			}
		}

		w.Flush()
		return buf.Bytes(), "text/csv", nil
	}
}

func (u *billingUsecase) ExportPaymentLinksExcel(ctx context.Context, filters map[string]string) ([]byte, error) {
	return u.invoiceRepo.ExportPaymentLinksExcel(ctx, filters)
}

func (u *billingUsecase) ArchiveOldInvoices(ctx context.Context) error {
	logger.Info("Starting auto-archiving of old invoices...")

	// Start transaction
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find invoices older than 3 months with status Lunas or Expired
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	var oldInvoices []domain.Invoice

	if err := tx.Where("tgl_invoice < ? AND status_invoice IN ?", threeMonthsAgo, []string{"Lunas", "Expired"}).Find(&oldInvoices).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to fetch old invoices: %w", err)
	}

	if len(oldInvoices) == 0 {
		logger.Info("No invoices to archive at this time.")
		return tx.Commit().Error
	}

	logger.Info("Found %d invoices to archive.", len(oldInvoices))

	// Map and insert into archive table
	var archives []domain.InvoiceArchive
	for _, inv := range oldInvoices {
		archives = append(archives, domain.InvoiceArchive{
			InvoiceNumber:      inv.InvoiceNumber,
			PelangganID:        inv.PelangganID,
			IDPelanggan:        inv.IDPelanggan,
			Brand:              inv.Brand,
			NoTelp:             inv.NoTelp,
			Email:              inv.Email,
			TotalHarga:         inv.TotalHarga,
			TglInvoice:         inv.TglInvoice,
			TglJatuhTempo:      inv.TglJatuhTempo,
			StatusInvoice:      inv.StatusInvoice,
			PaymentLink:        inv.PaymentLink,
			MetodePembayaran:   inv.MetodePembayaran,
			ExpiryDate:         inv.ExpiryDate,
			PaidAmount:         inv.PaidAmount,
			PaidAt:             inv.PaidAt,
			XenditID:           inv.XenditID,
			XenditExternalID:   inv.XenditExternalID,
			IsProcessing:       inv.IsProcessing,
			XenditRetryCount:   inv.XenditRetryCount,
			XenditLastRetry:    inv.XenditLastRetry,
			XenditErrorMessage: inv.XenditErrorMessage,
			XenditStatus:       inv.XenditStatus,
			InvoiceType:        inv.InvoiceType,
			IsReinvoice:        inv.IsReinvoice,
			OriginalInvoiceID:  inv.OriginalInvoiceID,
			ReinvoiceReason:    inv.ReinvoiceReason,
			CreatedAt:          inv.CreatedAt,
			UpdatedAt:          inv.UpdatedAt,
		})
	}

	if err := tx.Create(&archives).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert into archive table: %w", err)
	}

	// Soft delete from main table
	var ids []uint64
	for _, inv := range oldInvoices {
		ids = append(ids, inv.ID)
	}

	if err := tx.Delete(&domain.Invoice{}, ids).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to soft delete archived invoices: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	logger.Info("Successfully archived %d invoices.", len(archives))
	return nil
}

// --- Helpers ---

func (u *billingUsecase) triggerMikrotikUpdate(ctx context.Context, name string, dt *domain.DataTeknis, status string) error {
	if dt.MikrotikServerID == nil {
		return nil
	}
	err := u.executeRouterOS(ctx, *dt.MikrotikServerID, func(c *routeros.Client) error {
		profile, disabled := "default", "no"
		if status == "Suspended" || status == "Berhenti" {
			profile, disabled = "SUSPENDED", "yes"
		} else if dt.ProfilePppoe != nil {
			profile = *dt.ProfilePppoe
		}
		ip := ""
		if dt.IPPelanggan != nil {
			ip = *dt.IPPelanggan
		}
		return mikrotik.UpdatePPPoESecret(c, name, dt.IDPelanggan, dt.PasswordPppoe, profile, ip, disabled)
	})

	if err != nil {
		if !dt.MikrotikSyncPending {
			dt.MikrotikSyncPending = true
			_ = u.dataTeknisRepo.Update(ctx, dt)
		}
		return err
	}
	if dt.MikrotikSyncPending {
		dt.MikrotikSyncPending = false
		_ = u.dataTeknisRepo.Update(ctx, dt)
	}
	return nil
}

func (u *billingUsecase) executeRouterOS(ctx context.Context, serverID uint64, op func(*routeros.Client) error) error {
	server, err := u.mikrotikRepo.GetByID(ctx, serverID)
	if err != nil || !server.IsActive {
		return errors.New("mikrotik server error or inactive")
	}
	decryptedPassword := utils.GlobalEncryptionService.Decrypt(server.Password)
	client, err := mikrotik.GlobalPool.GetConnection(server.HostIP, server.Port, server.Username, decryptedPassword)
	if err != nil {
		return err
	}
	defer mikrotik.GlobalPool.ReturnConnection(client, server.HostIP, server.Port)
	return op(client)
}

func (u *billingUsecase) createXenditInvoice(ctx context.Context, inv *domain.Invoice, p *domain.Pelanggan, pkt *domain.PaketLayanan, desc string, tax float64, phone string) (map[string]interface{}, error) {
	xenditApiKey := u.cfg.XenditApiKeyJelantik
	if p.HargaLayanan != nil && strings.ToUpper(p.HargaLayanan.XenditKeyName) == "JAKINET" {
		xenditApiKey = u.cfg.XenditApiKeyJakinet
	}

	if xenditApiKey == "" {
		return map[string]interface{}{}, errors.New("xendit API key not configured")
	}

	hargaDasar := inv.TotalHarga - tax
	payload := map[string]interface{}{
		"external_id":      inv.InvoiceNumber,
		"amount":           inv.TotalHarga,
		"description":      fmt.Sprintf("Invoice #: %s", inv.InvoiceNumber),
		"invoice_duration": 86400 * 10,
		"customer": map[string]interface{}{
			"given_names":   p.Nama,
			"email":         p.Email,
			"mobile_number": phone,
			"addresses": []map[string]interface{}{
				{
					"country":      "ID",
					"street_line1": p.Alamat,
					"city":         "",
					"province":     "",
					"postal_code":  "",
				},
			},
		},
		"currency":              "IDR",
		"with_short_url":        true,
		"should_send_email":     true,
		"should_send_whatsapp":  true,
		"notification_channels": []string{"whatsapp", "email"},
		"business_profile": map[string]interface{}{
			"business_name":     "Artacomindo Jejaring Nusa",
			"business_address":  "Indonesia",
			"business_contact":  "+628986937819",
			"business_industry": "Telecommunications",
		},
		"items": []map[string]interface{}{
			{
				"name":        desc,
				"price":       math.Round(hargaDasar),
				"quantity":    1,
				"description": desc,
				"currency":    "IDR",
				"type":        "PRODUCT",
			},
		},
		"fees": []map[string]interface{}{
			{"type": "Tax", "value": math.Round(tax)},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u.cfg.XenditApiUrl, bytes.NewReader(jsonPayload))
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(xenditApiKey + ":"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		errMsg, _ := json.Marshal(result)
		return nil, fmt.Errorf("xendit API error (%d): %s", resp.StatusCode, string(errMsg))
	}

	return result, nil
}

func (u *billingUsecase) logSystem(ctx context.Context, level, message string) {
	_ = u.systemRepo.CreateSystemLog(ctx, &domain.SystemLog{Timestamp: time.Now(), Level: level, Message: message})
}

func (u *billingUsecase) ProcessXenditCallback(ctx context.Context, xCallbackToken string, payload map[string]interface{}, idempotencyKey string) error {
	// Serialize callback payload
	jsonData, _ := json.Marshal(payload)
	dataStr := string(jsonData)

	externalID, _ := payload["external_id"].(string)
	xenditID, _ := payload["id"].(string)

	logEntry := &domain.PaymentCallbackLog{
		XenditID:     xenditID,
		ExternalID:   externalID,
		CallbackData: &dataStr,
		Status:       "PROCESSING",
		ProcessedAt:  time.Now(),
		CreatedAt:    time.Now(),
	}
	if idempotencyKey != "" {
		logEntry.IdempotencyKey = &idempotencyKey
	}

	saveLog := func(status string) {
		logEntry.Status = status
		logEntry.ProcessedAt = time.Now()
		_ = u.invoiceRepo.CreateCallbackLog(ctx, logEntry)
	}

	if externalID == "" {
		saveLog("FAILED_MISSING_EXTERNAL_ID")
		return errors.New("external_id not found")
	}

	// Check for duplicate callbacks (idempotency or Xendit ID)
	existingLog, err := u.invoiceRepo.GetCallbackLog(ctx, xenditID, externalID, idempotencyKey)
	if err == nil && existingLog != nil {
		u.logSystem(ctx, "INFO", fmt.Sprintf("Duplicate Xendit callback received for Invoice %s, Xendit ID %s. Skipping.", externalID, xenditID))
		return nil
	}

	if xCallbackToken != u.cfg.XenditCallbackTokenArtacomindo && xCallbackToken != u.cfg.XenditCallbackTokenJelantik {
		saveLog("FAILED_INVALID_TOKEN")
		return errors.New("invalid callback token")
	}

	invoice, err := u.invoiceRepo.GetInvoiceWithRelations(ctx, externalID)
	if err != nil || invoice == nil {
		saveLog("FAILED_INVOICE_NOT_FOUND")
		return errors.New("invoice not found")
	}

	if invoice.StatusInvoice == "Lunas" {
		saveLog("SUCCESS_ALREADY_PAID")
		return nil
	}

	status, _ := payload["status"].(string)
	if status == "PAID" {
		paidAmount, _ := payload["paid_amount"].(float64)
		paidAtStr, _ := payload["paid_at"].(string)
		paidAt, _ := time.Parse(time.RFC3339, paidAtStr)
		if paidAt.IsZero() {
			paidAt = time.Now()
		}

		err = u.processSuccessfulPayment(ctx, invoice, paidAmount, paidAt)
		if err != nil {
			saveLog("FAILED_PROCESSING_PAYMENT")
			return err
		}
		saveLog("SUCCESS")
		return nil
	}

	saveLog("SKIPPED_UNHANDLED_STATUS_" + status)
	return nil
}

func (u *billingUsecase) processSuccessfulPayment(ctx context.Context, inv *domain.Invoice, amt float64, paidAt time.Time) error {
	inv.StatusInvoice = "Lunas"
	inv.PaidAmount = &amt
	inv.PaidAt = &paidAt
	if err := u.invoiceRepo.Update(ctx, inv); err != nil {
		return err
	}

	// Trigger WebSocket notification
	if websocket.GlobalHub != nil {
		pName := inv.PelangganNama
		if pName == "" && inv.Pelanggan != nil {
			pName = inv.Pelanggan.Nama
		}
		websocket.GlobalHub.BroadcastNotification("new_payment", map[string]interface{}{
			"invoice_number": inv.InvoiceNumber,
			"pelanggan_nama": pName,
			"amount":         amt,
		})
	}

	if inv.Pelanggan != nil && len(inv.Pelanggan.Langganan) > 0 {
		l := &inv.Pelanggan.Langganan[0]
		l.Status = "Aktif"
		if l.TglJatuhTempo != nil {
			next := l.TglJatuhTempo.AddDate(0, 1, 0)
			l.TglJatuhTempo = &next
		}
		if l.TglJatuhTempoPembayaran != nil {
			nextPay := l.TglJatuhTempoPembayaran.AddDate(0, 1, 0)
			l.TglJatuhTempoPembayaran = &nextPay
		}
		if l.TglMulaiLangganan != nil {
			nextMulai := l.TglMulaiLangganan.AddDate(0, 1, 0)
			l.TglMulaiLangganan = &nextMulai
		}
		_ = l.TglMulaiLangganan // avoid unused warning just in case
		_ = u.langgananRepo.Update(ctx, l)
		if inv.Pelanggan.DataTeknis != nil {
			_ = u.triggerMikrotikUpdate(ctx, inv.Pelanggan.DataTeknis.IDPelanggan, inv.Pelanggan.DataTeknis, "Aktif")
		}
	}
	u.logActivity(ctx, "Payment Confirmed", fmt.Sprintf("Invoice %s marked as paid (Lunas) for amount %.2f", inv.InvoiceNumber, amt))
	websocket.InvalidateDashboardCache(ctx)
	return nil
}

func (u *billingUsecase) RetryFailedMikrotikSync(ctx context.Context) error {
	pendingList, err := u.dataTeknisRepo.GetPendingSync(ctx)
	if err != nil {
		u.logSystem(ctx, "ERROR", fmt.Sprintf("Failed to fetch pending Mikrotik syncs: %v", err))
		return err
	}
	if len(pendingList) == 0 {
		return nil
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Retrying %d failed Mikrotik syncs...", len(pendingList)))
	successCount := 0
	for i := range pendingList {
		dt := &pendingList[i]
		status := "Aktif"
		if dt.Pelanggan != nil && len(dt.Pelanggan.Langganan) > 0 {
			status = dt.Pelanggan.Langganan[0].Status
		}
		err := u.triggerMikrotikUpdate(ctx, dt.IDPelanggan, dt, status)
		if err == nil {
			successCount++
		}
	}

	if successCount > 0 {
		u.logSystem(ctx, "INFO", fmt.Sprintf("Successfully retried and synced %d/%d pending Mikrotik records.", successCount, len(pendingList)))
	}
	return nil
}
