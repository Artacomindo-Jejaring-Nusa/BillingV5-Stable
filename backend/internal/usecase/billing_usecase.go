package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"billing-backend/config"
	"billing-backend/internal/domain"
	"billing-backend/pkg/mikrotik"
	"billing-backend/pkg/utils"

	"github.com/go-routeros/routeros"
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



// Invoice Logic

func (u *billingUsecase) FetchInvoices(ctx context.Context, page, pageSize int) ([]domain.Invoice, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	invoices, total, err := u.invoiceRepo.GetAll(ctx, pageSize, offset)
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
		return errors.New("invoice number and valid total_harga are required")
	}

	// 1. Load pelanggan
	pelanggan, err := u.pelangganRepo.GetByID(ctx, invoice.PelangganID)
	if err != nil {
		return err
	}
	if pelanggan == nil {
		return errors.New("pelanggan not found")
	}

	// 2. Load Brand/HargaLayanan
	if pelanggan.IDBrand == nil || *pelanggan.IDBrand == "" {
		return errors.New("data brand pelanggan tidak ditemukan")
	}
	brand, err := u.brandRepo.GetByID(ctx, *pelanggan.IDBrand)
	if err != nil || brand == nil {
		return errors.New("data brand pelanggan tidak ditemukan")
	}
	pelanggan.HargaLayanan = brand

	// 3. Load DataTeknis
	var dataTeknis *domain.DataTeknis
	dt, err := u.dataTeknisRepo.GetByPelangganID(ctx, pelanggan.ID)
	if err == nil && dt != nil {
		dataTeknis = dt
	}
	if dataTeknis == nil {
		return errors.New("data teknis pelanggan tidak ditemukan")
	}
	pelanggan.DataTeknis = dataTeknis

	// Fill historical/required invoice details
	invoice.IDPelanggan = dataTeknis.IDPelanggan
	invoice.Brand = brand.Brand
	invoice.NoTelp = pelanggan.NoTelp
	invoice.Email = pelanggan.Email

	// Find active subscription if any to get package info
	var activeLangganan *domain.Langganan
	for i := range pelanggan.Langganan {
		if pelanggan.Langganan[i].Status == "Aktif" {
			activeLangganan = &pelanggan.Langganan[i]
			break
		}
	}

	var paket *domain.PaketLayanan
	if activeLangganan != nil {
		paket, _ = u.paketRepo.GetByID(ctx, activeLangganan.PaketLayananID)
	}

	// Calculate tax (PPN)
	pajakPersen := brand.Pajak
	hargaDasar := invoice.TotalHarga / (1.0 + (pajakPersen / 100.0))
	hargaDasar = math.Round(hargaDasar)
	pajak := invoice.TotalHarga - hargaDasar

	// Description
	var deskripsiXendit string
	if invoice.ReinvoiceReason != nil && *invoice.ReinvoiceReason != "" {
		deskripsiXendit = *invoice.ReinvoiceReason
	} else {
		deskripsiXendit = fmt.Sprintf("Tagihan internet manual - %s", invoice.InvoiceNumber)
		if paket != nil {
			deskripsiXendit = fmt.Sprintf("Biaya berlangganan internet up to %d Mbps - %s", paket.Kecepatan, invoice.InvoiceNumber)
		}
	}

	noTelpXendit := utils.NormalizePhoneForXendit(pelanggan.NoTelp)

	// Call Xendit API
	xenditResponse, err := u.createXenditInvoice(ctx, invoice, pelanggan, paket, deskripsiXendit, pajak, noTelpXendit)
	if err != nil {
		return fmt.Errorf("failed to create Xendit invoice: %w", err)
	}

	shortURL, _ := xenditResponse["short_url"].(string)
	if shortURL == "" {
		shortURL, _ = xenditResponse["invoice_url"].(string)
	}
	xenditID, _ := xenditResponse["id"].(string)
	xenditExtID, _ := xenditResponse["external_id"].(string)

	invoice.PaymentLink = &shortURL
	invoice.XenditID = &xenditID
	invoice.XenditExternalID = &xenditExtID
	invoice.XenditStatus = "PENDING"
	invoice.StatusInvoice = "Belum Dibayar"
	invoice.InvoiceType = "manual"

	return u.invoiceRepo.Create(ctx, invoice)
}

func (u *billingUsecase) UpdateInvoiceStatus(ctx context.Context, id uint64, status string) error {
	invoice, err := u.invoiceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	invoice.StatusInvoice = status
	return u.invoiceRepo.Update(ctx, invoice)
}

func (u *billingUsecase) GetInvoiceSummary(ctx context.Context) (*domain.InvoiceSummaryStats, error) {
	return u.invoiceRepo.GetInvoiceSummary(ctx)
}

// Langganan Logic

func (u *billingUsecase) FetchLangganan(ctx context.Context, page, pageSize int, search, status string, forInvoiceSelection bool) ([]domain.Langganan, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	return u.langgananRepo.GetAll(ctx, pageSize, offset, search, status, forInvoiceSelection)
}

func (u *billingUsecase) GetNewUserLangganans(ctx context.Context) ([]domain.Langganan, error) {
	langganans, err := u.langgananRepo.GetNewUserLangganans(ctx)
	if err != nil {
		return nil, err
	}

	today := time.Now()
	for i := range langganans {
		lang := &langganans[i]
		if lang.Pelanggan != nil && lang.MetodePembayaran != "Prorate" {
			clusterToCheck := strings.TrimSpace(lang.Pelanggan.Alamat)
			if clusterToCheck != "" {
				diskon, err := u.diskonRepo.GetActiveForCluster(ctx, clusterToCheck, today)
				if err == nil && diskon != nil {
					if lang.HargaAwal != nil {
						subtotal := *lang.HargaAwal
						diskonAmount := math.Floor((subtotal * diskon.PersentaseDiskon / 100.0) + 0.5)
						finalHarga := subtotal - diskonAmount
						lang.HargaAwal = &finalHarga
					}
				}
			}
		}
	}

	return langganans, nil
}

func (u *billingUsecase) GetLangganan(ctx context.Context, id uint64) (*domain.Langganan, error) {
	return u.langgananRepo.GetByID(ctx, id)
}

func (u *billingUsecase) CreateLangganan(ctx context.Context, langganan *domain.Langganan) error {
	if langganan.PelangganID == 0 || langganan.PaketLayananID == 0 {
		return errors.New("pelanggan_id and paket_layanan_id are required")
	}

	// 1. Validasi pelanggan ada
	pelanggan, err := u.pelangganRepo.GetByID(ctx, langganan.PelangganID)
	if err != nil {
		return errors.New("pelanggan not found")
	}
	if pelanggan == nil {
		return errors.New("pelanggan not found")
	}

	// Ambil Brand/HargaLayanan
	if pelanggan.IDBrand == nil || *pelanggan.IDBrand == "" {
		return errors.New("Data Brand pelanggan tidak ditemukan.")
	}
	brand, err := u.brandRepo.GetByID(ctx, *pelanggan.IDBrand)
	if err != nil || brand == nil {
		return errors.New("Data Brand pelanggan tidak ditemukan.")
	}

	// 2. VALIDASI UTAMA: Cek apakah pelanggan sudah punya data teknis
	if pelanggan.DataTeknis == nil {
		return fmt.Errorf("Langganan tidak dapat dibuat. Pelanggan '%s' belum memiliki data teknis. Tim NOC harus menambahkan data teknis terlebih dahulu sebelum membuat langganan.", pelanggan.Nama)
	}

	// 3. Validasi paket layanan ada
	paket, err := u.paketRepo.GetByID(ctx, langganan.PaketLayananID)
	if err != nil || paket == nil {
		return errors.New("Paket Layanan tidak ditemukan.")
	}

	var startDate time.Time
	if langganan.TglMulaiLangganan != nil {
		startDate = *langganan.TglMulaiLangganan
	} else {
		startDate = time.Now()
	}

	hargaPaket := paket.Harga
	pajakPersen := brand.Pajak
	var hargaAwalFinal float64
	var tglJatuhTempoFinal time.Time

	if langganan.MetodePembayaran == "Prorate" {
		lastDay := time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())
		lastDayNum := lastDay.Day()
		remainingDays := lastDayNum - startDate.Day() + 1
		if remainingDays < 0 {
			remainingDays = 0
		}

		hargaPerHari := hargaPaket / float64(lastDayNum)
		proratedPriceBeforeTax := hargaPerHari * float64(remainingDays)
		hargaProrateFinal := proratedPriceBeforeTax * (1 + (pajakPersen / 100))

		if langganan.SertakanBulanDepan {
			hargaNormalFull := hargaPaket * (1 + (pajakPersen / 100))
			hargaAwalFinal = hargaProrateFinal + hargaNormalFull
		} else {
			hargaAwalFinal = hargaProrateFinal
		}

		tglJatuhTempoFinal = lastDay
	} else {
		hargaAwalFinal = hargaPaket * (1 + (pajakPersen / 100))
		nextMonth := startDate.AddDate(0, 1, 0)
		tglJatuhTempoFinal = time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, startDate.Location())
	}

	roundedHargaAwal := math.Round(hargaAwalFinal)
	langganan.HargaAwal = &roundedHargaAwal
	langganan.TglJatuhTempo = &tglJatuhTempoFinal

	if langganan.Status == "" {
		langganan.Status = "Aktif"
	}

	return u.langgananRepo.Create(ctx, langganan)
}

func (u *billingUsecase) UpdateLangganan(ctx context.Context, id uint64, langganan *domain.Langganan) error {
	existing, err := u.langgananRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("langganan not found")
	}

	// Jika status diubah menjadi "Berhenti"
	if langganan.Status == "Berhenti" && existing.Status != "Berhenti" {
		hariIni := time.Now()
		existing.TglBerhenti = &hariIni

		// Parse existing riwayat
		var riwayatList []map[string]interface{}
		if existing.RiwayatTglBerhenti != nil && *existing.RiwayatTglBerhenti != "" {
			_ = json.Unmarshal([]byte(*existing.RiwayatTglBerhenti), &riwayatList)
		}

		// Add new entry
		alasan := ""
		if langganan.AlasanBerhenti != nil {
			alasan = *langganan.AlasanBerhenti
		}
		newEntry := map[string]interface{}{
			"tanggal":   hariIni.Format("2006-01-02"),
			"alasan":    alasan,
			"timestamp": time.Now().Format(time.RFC3339),
		}
		riwayatList = append(riwayatList, newEntry)

		// Marshal back to string
		jsonBytes, _ := json.Marshal(riwayatList)
		riwayatStr := string(jsonBytes)
		existing.RiwayatTglBerhenti = &riwayatStr
		existing.AlasanBerhenti = langganan.AlasanBerhenti
	} else if langganan.Status != "Berhenti" && existing.Status == "Berhenti" {
		// Jika status diubah dari "Berhenti" ke status lain, kosongkan tgl_berhenti tapi RIWAYAT TETAP DISIMPAN
		existing.TglBerhenti = nil
	}

	// Update fields (Partial update / PATCH behavior)
	if langganan.PelangganID != 0 {
		existing.PelangganID = langganan.PelangganID
	}
	if langganan.PaketLayananID != 0 {
		existing.PaketLayananID = langganan.PaketLayananID
	}
	if langganan.Status != "" {
		existing.Status = langganan.Status
	}
	if langganan.TglJatuhTempo != nil {
		existing.TglJatuhTempo = langganan.TglJatuhTempo
	}
	if langganan.TglInvoiceTerakhir != nil {
		existing.TglInvoiceTerakhir = langganan.TglInvoiceTerakhir
	}
	if langganan.TglMulaiLangganan != nil {
		existing.TglMulaiLangganan = langganan.TglMulaiLangganan
	}
	if langganan.MetodePembayaran != "" {
		existing.MetodePembayaran = langganan.MetodePembayaran
	}
	if langganan.HargaAwal != nil {
		existing.HargaAwal = langganan.HargaAwal
	}
	if langganan.StatusModem != nil {
		existing.StatusModem = langganan.StatusModem
	}
	if langganan.WhatsappStatus != nil {
		existing.WhatsappStatus = langganan.WhatsappStatus
	}
	if langganan.LastWhatsappSent != nil {
		existing.LastWhatsappSent = langganan.LastWhatsappSent
	}

	return u.langgananRepo.Update(ctx, existing)
}

func (u *billingUsecase) DeleteLangganan(ctx context.Context, id uint64) error {
	existing, err := u.langgananRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("langganan not found")
	}
	return u.langgananRepo.Delete(ctx, id)
}

func (u *billingUsecase) CalculatePrice(ctx context.Context, req *domain.LanggananCalculateRequest) (*domain.LanggananCalculateResponse, error) {
	pelanggan, err := u.pelangganRepo.GetByID(ctx, req.PelangganID)
	if err != nil {
		return nil, errors.New("Data Brand pelanggan tidak ditemukan.")
	}
	if pelanggan == nil {
		return nil, errors.New("Data Brand pelanggan tidak ditemukan.")
	}
	if pelanggan.IDBrand == nil || *pelanggan.IDBrand == "" {
		return nil, errors.New("Data Brand pelanggan tidak ditemukan.")
	}

	brand, err := u.brandRepo.GetByID(ctx, *pelanggan.IDBrand)
	if err != nil || brand == nil {
		return nil, errors.New("Data Brand pelanggan tidak ditemukan.")
	}

	paket, err := u.paketRepo.GetByID(ctx, req.PaketLayananID)
	if err != nil || paket == nil {
		return nil, errors.New("Paket Layanan tidak ditemukan.")
	}

	var startDate time.Time
	if req.TglMulai != nil {
		startDate = *req.TglMulai
	} else {
		startDate = time.Now()
	}

	hargaPaket := paket.Harga
	pajakPersen := brand.Pajak
	var hargaAwalFinal float64
	var tglJatuhTempoFinal time.Time

	if req.MetodePembayaran == "Prorate" {
		lastDay := time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())
		lastDayNum := lastDay.Day()
		remainingDays := lastDayNum - startDate.Day() + 1
		if remainingDays < 0 {
			remainingDays = 0
		}

		hargaPerHari := hargaPaket / float64(lastDayNum)
		proratedPriceBeforeTax := hargaPerHari * float64(remainingDays)
		hargaAwalFinal = proratedPriceBeforeTax * (1 + (pajakPersen / 100))
		tglJatuhTempoFinal = lastDay
	} else {
		hargaAwalFinal = hargaPaket * (1 + (pajakPersen / 100))
		nextMonth := startDate.AddDate(0, 1, 0)
		tglJatuhTempoFinal = time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, startDate.Location())
	}

	return &domain.LanggananCalculateResponse{
		HargaAwal:     math.Round(hargaAwalFinal),
		TglJatuhTempo: tglJatuhTempoFinal,
	}, nil
}

func (u *billingUsecase) CalculateProratePlusFull(ctx context.Context, req *domain.LanggananCalculateRequest) (*domain.LanggananCalculateProratePlusFullResponse, error) {
	pelanggan, err := u.pelangganRepo.GetByID(ctx, req.PelangganID)
	if err != nil {
		return nil, errors.New("Data Brand pelanggan tidak ditemukan.")
	}
	if pelanggan == nil {
		return nil, errors.New("Data Brand pelanggan tidak ditemukan.")
	}
	if pelanggan.IDBrand == nil || *pelanggan.IDBrand == "" {
		return nil, errors.New("Data Brand pelanggan tidak ditemukan.")
	}

	brand, err := u.brandRepo.GetByID(ctx, *pelanggan.IDBrand)
	if err != nil || brand == nil {
		return nil, errors.New("Data Brand pelanggan tidak ditemukan.")
	}

	paket, err := u.paketRepo.GetByID(ctx, req.PaketLayananID)
	if err != nil || paket == nil {
		return nil, errors.New("Paket Layanan tidak ditemukan.")
	}

	var startDate time.Time
	if req.TglMulai != nil {
		startDate = *req.TglMulai
	} else {
		startDate = time.Now()
	}

	hargaPaket := paket.Harga
	pajakPersen := brand.Pajak

	hargaNormalFull := hargaPaket * (1 + (pajakPersen / 100))

	lastDay := time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())
	lastDayNum := lastDay.Day()
	remainingDays := lastDayNum - startDate.Day() + 1
	if remainingDays < 0 {
		remainingDays = 0
	}

	hargaPerHari := hargaPaket / float64(lastDayNum)
	proratedPriceBeforeTax := hargaPerHari * float64(remainingDays)
	hargaProrateFinal := proratedPriceBeforeTax * (1 + (pajakPersen / 100))

	hargaTotalFinal := hargaProrateFinal + hargaNormalFull

	return &domain.LanggananCalculateProratePlusFullResponse{
		HargaProrate:   math.Round(hargaProrateFinal),
		HargaNormal:    math.Round(hargaNormalFull),
		HargaTotalAwal: math.Round(hargaTotalFinal),
		TglJatuhTempo:  lastDay,
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

	pajakMentah := hargaPaket * (pajakPersen / 100)
	pajakAmount := math.Floor(pajakMentah + 0.5)

	subtotalSebelumDiskon := hargaPaket + pajakAmount

	diskonMentah := subtotalSebelumDiskon * (persentaseDiskon / 100)
	diskonAmount := math.Floor(diskonMentah + 0.5)

	hargaFinal := subtotalSebelumDiskon - diskonAmount

	detailPerhitungan := fmt.Sprintf(
		"📊 Rincian Perhitungan Diskon:\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"📦 Harga Paket     : Rp %s\n"+
			"📈 Pajak (%.0f%%)    : Rp %s\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"💰 Subtotal        : Rp %s\n"+
			"🏷️  Diskon (%.0f%%)   : -Rp %s\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"✅ Harga Final     : Rp %s\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"💬 Catatan:\n"+
			"• Pajak dihitung dari harga paket\n"+
			"• Diskon dihitung dari subtotal (harga + pajak)\n"+
			"• Pembulatan menggunakan floor(x + 0.5)",
		formatRupiah(hargaPaket),
		pajakPersen,
		formatRupiah(pajakAmount),
		formatRupiah(subtotalSebelumDiskon),
		persentaseDiskon,
		formatRupiah(diskonAmount),
		formatRupiah(hargaFinal),
	)

	return &domain.DiskonCalculationResponse{
		NamaPaket:             paket.NamaPaket,
		NamaBrand:             brand.Brand,
		HargaPaket:            math.Round(hargaPaket),
		PajakPersen:           pajakPersen,
		PajakAmount:           math.Round(pajakAmount),
		SubtotalSebelumDiskon: math.Round(subtotalSebelumDiskon),
		PersentaseDiskon:      persentaseDiskon,
		DiskonAmount:          math.Round(diskonAmount),
		HargaFinal:            math.Round(hargaFinal),
		DetailPerhitungan:     detailPerhitungan,
	}, nil
}

func formatRupiah(val float64) string {
	str := fmt.Sprintf("%.0f", val)
	var result []rune
	for i, r := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, r)
	}
	return string(result)
}

func (u *billingUsecase) ProcessXenditCallback(ctx context.Context, xCallbackToken string, payload map[string]interface{}, idempotencyKey string) error {
	externalID, _ := payload["external_id"].(string)
	if externalID == "" {
		return errors.New("external_id not found in payload")
	}

	var brandPrefix string
	if parts := strings.Split(externalID, "/"); len(parts) > 0 {
		brandPrefix = strings.ToLower(parts[0])
	}

	var correctToken string
	artacomToken := u.cfg.XenditCallbackTokenArtacomindo
	jelantikToken := u.cfg.XenditCallbackTokenJelantik

	if artacomToken != "" && xCallbackToken == artacomToken {
		if brandPrefix != "" && brandPrefix != "jakinet" && brandPrefix != "nagrak" && brandPrefix != "artacom" && brandPrefix != "jelantiknagrak" {
			return errors.New("invalid brand for this token")
		}
		correctToken = artacomToken
	} else if jelantikToken != "" && xCallbackToken == jelantikToken {
		if brandPrefix != "" && brandPrefix != "jelantik" {
			return errors.New("invalid brand for this token")
		}
		correctToken = jelantikToken
	}

	if correctToken == "" {
		return errors.New("invalid callback token")
	}

	xenditID, _ := payload["id"].(string)
	xenditStatus, _ := payload["status"].(string)

	// Check duplicate callback
	existingLog, err := u.invoiceRepo.GetCallbackLog(ctx, xenditID, externalID, idempotencyKey)
	if err != nil {
		return err
	}

	// Query invoice with all relations loaded
	invoice, err := u.invoiceRepo.GetInvoiceWithRelations(ctx, externalID)
	if err != nil {
		return err
	}
	if invoice == nil {
		return errors.New("invoice not found")
	}

	// If duplicate and already paid
	if existingLog != nil && invoice.StatusInvoice == "Lunas" {
		return nil
	}

	if invoice.StatusInvoice == "Lunas" {
		return nil
	}

	// Create or update callback log
	payloadBytes, _ := json.Marshal(payload)
	payloadStr := string(payloadBytes)
	if len(payloadStr) > 1000 {
		payloadStr = payloadStr[:997] + "..."
	}

	var idKey *string
	if idempotencyKey != "" {
		idKey = &idempotencyKey
	}

	logEntry := &domain.PaymentCallbackLog{
		IdempotencyKey: idKey,
		XenditID:       xenditID,
		ExternalID:     externalID,
		CallbackData:   &payloadStr,
		Status:         xenditStatus,
		ProcessedAt:    time.Now(),
		CreatedAt:      time.Now(),
	}

	_ = u.invoiceRepo.CreateCallbackLog(ctx, logEntry)

	var paidAmount float64
	if pa, ok := payload["paid_amount"].(float64); ok {
		paidAmount = pa
	} else if paStr, ok := payload["paid_amount"].(string); ok {
		if paVal, err := strconv.ParseFloat(paStr, 64); err == nil {
			paidAmount = paVal
		}
	}

	var paidAt time.Time = time.Now()
	if patStr, ok := payload["paid_at"].(string); ok && patStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, patStr); err == nil {
			paidAt = parsedTime
		} else if parsedTime, err := time.Parse("2006-01-02T15:04:05.999Z07:00", patStr); err == nil {
			paidAt = parsedTime
		} else if parsedTime, err := time.Parse("2006-01-02T15:04:05", patStr); err == nil {
			paidAt = parsedTime
		}
	}

	if xenditStatus == "PAID" {
		if err := u.processSuccessfulPayment(ctx, invoice, paidAmount, paidAt); err != nil {
			return err
		}
	} else if xenditStatus == "EXPIRED" {
		invoice.StatusInvoice = "Expired"
		if err := u.invoiceRepo.Update(ctx, invoice); err != nil {
			return err
		}
	}

	return nil
}

func (u *billingUsecase) executeRouterOS(ctx context.Context, serverID uint64, op func(*routeros.Client) error) error {
	server, err := u.mikrotikRepo.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to fetch Mikrotik server: %w", err)
	}

	if !server.IsActive {
		return fmt.Errorf("mikrotik server %s is inactive", server.Name)
	}

	decryptedPassword := ""
	if server.Password != "" {
		decryptedPassword = utils.GlobalEncryptionService.Decrypt(server.Password)
	}

	client, err := mikrotik.GlobalPool.GetConnection(server.HostIP, server.Port, server.Username, decryptedPassword)
	if err != nil {
		return fmt.Errorf("failed to connect to Mikrotik router: %w", err)
	}
	defer mikrotik.GlobalPool.ReturnConnection(client, server.HostIP, server.Port)

	return op(client)
}

func (u *billingUsecase) triggerMikrotikUpdate(ctx context.Context, oldName string, data *domain.DataTeknis, newStatus string) error {
	if data.MikrotikServerID == nil {
		return errors.New("mikrotik_server_id is nil")
	}
	return u.executeRouterOS(ctx, *data.MikrotikServerID, func(client *routeros.Client) error {
		profile := ""
		disabled := "no"
		if newStatus == "Aktif" {
			if data.ProfilePppoe != nil {
				profile = *data.ProfilePppoe
			}
			disabled = "no"
		} else if newStatus == "Suspended" {
			profile = "SUSPENDED"
			disabled = "yes"
		} else {
			if data.ProfilePppoe != nil {
				profile = *data.ProfilePppoe
			}
			disabled = "no"
		}

		ip := ""
		if data.IPPelanggan != nil {
			ip = *data.IPPelanggan
		}

		err := mikrotik.UpdatePPPoESecret(client, oldName, data.IDPelanggan, data.PasswordPppoe, profile, ip, disabled)
		if err != nil {
			return err
		}

		if newStatus == "Suspended" {
			_ = mikrotik.RemoveActiveConnection(client, data.IDPelanggan)
		}
		return nil
	})
}

func (u *billingUsecase) logSystem(ctx context.Context, level string, message string) {
	log := &domain.SystemLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}
	_ = u.systemRepo.CreateSystemLog(ctx, log)
}

func (u *billingUsecase) processSuccessfulPayment(ctx context.Context, invoice *domain.Invoice, paidAmount float64, paidAt time.Time) error {
	if invoice.StatusInvoice == "Lunas" {
		return nil
	}

	invoice.StatusInvoice = "Lunas"
	invoice.PaidAmount = &paidAmount
	invoice.PaidAt = &paidAt
	if err := u.invoiceRepo.Update(ctx, invoice); err != nil {
		return err
	}

	var originalLanggananStatus string
	var rollbackLangganan *domain.Langganan
	var rollbackDataTeknis *domain.DataTeknis
	var pelanggan *domain.Pelanggan

	if invoice.Pelanggan != nil {
		pelanggan = invoice.Pelanggan
		if len(pelanggan.Langganan) > 0 {
			rollbackLangganan = &pelanggan.Langganan[0]
			originalLanggananStatus = rollbackLangganan.Status
			rollbackDataTeknis = pelanggan.DataTeknis
		}
	}

	if rollbackLangganan != nil {
		isSuspendedOrInactive := originalLanggananStatus == "Suspended" || originalLanggananStatus == ""
		var nextDueDate time.Time

		if rollbackLangganan.MetodePembayaran == "Prorate" {
			rollbackLangganan.MetodePembayaran = "Otomatis"
			currentDueDate := invoice.TglJatuhTempo

			var hargaPaket float64
			if rollbackLangganan.PaketLayanan != nil {
				hargaPaket = rollbackLangganan.PaketLayanan.Harga
			} else {
				paket, err := u.paketRepo.GetByID(ctx, rollbackLangganan.PaketLayananID)
				if err == nil && paket != nil {
					hargaPaket = paket.Harga
				}
			}

			var pajakPersen float64
			if pelanggan != nil {
				if pelanggan.HargaLayanan != nil {
					pajakPersen = pelanggan.HargaLayanan.Pajak
				} else if pelanggan.IDBrand != nil {
					brand, err := u.brandRepo.GetByID(ctx, *pelanggan.IDBrand)
					if err == nil && brand != nil {
						pajakPersen = brand.Pajak
					}
				}
			}

			hargaNormalFull := hargaPaket * (1.0 + (pajakPersen / 100.0))

			if invoice.TotalHarga > (hargaNormalFull + 1.0) {
				if currentDueDate.Day() == 1 {
					nextDueDate = time.Date(currentDueDate.Year(), currentDueDate.Month()+1, 1, 0, 0, 0, 0, currentDueDate.Location())
				} else {
					nextDueDate = time.Date(currentDueDate.Year(), currentDueDate.Month()+2, 1, 0, 0, 0, 0, currentDueDate.Location())
				}
			} else {
				if currentDueDate.Day() == 1 {
					nextDueDate = currentDueDate
				} else {
					nextDueDate = time.Date(currentDueDate.Year(), currentDueDate.Month()+1, 1, 0, 0, 0, 0, currentDueDate.Location())
				}
			}

			roundedHargaNormal := math.Round(hargaNormalFull)
			rollbackLangganan.HargaAwal = &roundedHargaNormal
		} else {
			currentDueDate := invoice.TglJatuhTempo
			nextDueDate = time.Date(currentDueDate.Year(), currentDueDate.Month()+1, 1, 0, 0, 0, 0, currentDueDate.Location())
		}

		rollbackLangganan.Status = "Aktif"
		rollbackLangganan.TglJatuhTempo = &nextDueDate
		today := time.Now()
		rollbackLangganan.TglInvoiceTerakhir = &today

		if err := u.langgananRepo.Update(ctx, rollbackLangganan); err != nil {
			// Rollback invoice status
			invoice.StatusInvoice = "Belum Dibayar"
			invoice.PaidAmount = nil
			invoice.PaidAt = nil
			_ = u.invoiceRepo.Update(ctx, invoice)
			return err
		}

		// Mikrotik activation
		if isSuspendedOrInactive && rollbackDataTeknis != nil {
			err := u.triggerMikrotikUpdate(ctx, rollbackDataTeknis.IDPelanggan, rollbackDataTeknis, "Aktif")
			if err != nil {
				// Mark sync pending
				rollbackDataTeknis.MikrotikSyncPending = true
				_ = u.dataTeknisRepo.Update(ctx, rollbackDataTeknis)
			} else {
				if rollbackDataTeknis.MikrotikSyncPending {
					rollbackDataTeknis.MikrotikSyncPending = false
					_ = u.dataTeknisRepo.Update(ctx, rollbackDataTeknis)
				}
			}
		}
	}

	return nil
}

func (u *billingUsecase) createXenditInvoice(ctx context.Context, invoice *domain.Invoice, pelanggan *domain.Pelanggan, paket *domain.PaketLayanan, deskripsiXendit string, pajak float64, noTelpXendit string) (map[string]interface{}, error) {
	targetKeyName := pelanggan.HargaLayanan.XenditKeyName
	apiKeys := u.cfg.XENDIT_API_KEYS()
	apiKey := apiKeys[targetKeyName]

	if targetKeyName == "ajn-01" {
		if k, ok := apiKeys["JAKINET"]; ok && k != "" {
			apiKey = k
		}
	} else if targetKeyName == "ajn-02" {
		if k, ok := apiKeys["JELANTIK"]; ok && k != "" {
			apiKey = k
		}
	} else if targetKeyName == "ajn-03" {
		if k, ok := apiKeys["JAKINET"]; ok && k != "" {
			apiKey = k
		}
	} else if apiKey == "" && targetKeyName != "" {
		lowerKey := strings.ToLower(targetKeyName)
		if strings.Contains(lowerKey, "jelantik") && strings.Contains(lowerKey, "nagrak") {
			if k, ok := apiKeys["JAKINET"]; ok && k != "" {
				apiKey = k
			}
		} else if strings.Contains(lowerKey, "jelantik") {
			if k, ok := apiKeys["JELANTIK"]; ok && k != "" {
				apiKey = k
			}
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("Xendit API Key for brand '%s' not found", targetKeyName)
	}

	encodedKey := base64.StdEncoding.EncodeToString([]byte(apiKey + ":"))

	hargaDasar := invoice.TotalHarga - pajak
	if hargaDasar < 0 {
		return nil, errors.New("harga dasar tidak boleh negatif")
	}

	itemName := "Biaya Layanan Internet"
	if paket != nil {
		itemName = fmt.Sprintf("Biaya berlangganan internet up to %d Mbps", paket.Kecepatan)
	}

	payload := map[string]interface{}{
		"external_id":      invoice.InvoiceNumber,
		"amount":           invoice.TotalHarga,
		"description":      deskripsiXendit,
		"invoice_duration": 86400 * 10,
		"customer": map[string]interface{}{
			"given_names":   pelanggan.Nama,
			"email":         pelanggan.Email,
			"mobile_number": noTelpXendit,
		},
		"currency":       "IDR",
		"with_short_url": true,
		"customer_notification_preference": map[string]interface{}{
			"invoice_created":  []string{"whatsapp", "email"},
			"invoice_reminder": []string{"whatsapp", "email"},
			"invoice_paid":     []string{"whatsapp", "email"},
		},
		"business_profile": map[string]interface{}{
			"business_name":     "Artacomindo Jejaring Nusa",
			"business_address":  "Indonesia",
			"business_contact":  "+628986937819",
			"business_industry": "Telecommunications",
		},
		"items": []map[string]interface{}{
			{
				"name":        itemName,
				"price":       hargaDasar,
				"quantity":    1,
				"description": deskripsiXendit,
				"currency":    "IDR",
				"type":        "PRODUCT",
			},
		},
		"fees": []map[string]interface{}{
			{
				"type":  "Tax",
				"value": pajak,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	reqURL := u.cfg.XenditApiUrl
	if reqURL == "" {
		reqURL = "https://api.xendit.co/v2/invoices"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+encodedKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return nil, fmt.Errorf("Xendit API returned status %d: %v", resp.StatusCode, errData)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func (u *billingUsecase) GenerateManualInvoice(ctx context.Context, langgananID uint64) (*domain.Invoice, error) {
	langganan, err := u.langgananRepo.GetByID(ctx, langgananID)
	if err != nil {
		return nil, err
	}
	if langganan == nil {
		return nil, errors.New("langganan not found")
	}

	pelanggan := langganan.Pelanggan
	if pelanggan == nil {
		p, err := u.pelangganRepo.GetByID(ctx, langganan.PelangganID)
		if err != nil || p == nil {
			return nil, fmt.Errorf("pelanggan not found for langganan ID %d", langgananID)
		}
		langganan.Pelanggan = p
		pelanggan = p
	}

	paket := langganan.PaketLayanan
	if paket == nil {
		pk, err := u.paketRepo.GetByID(ctx, langganan.PaketLayananID)
		if err != nil || pk == nil {
			return nil, fmt.Errorf("paket not found for langganan ID %d", langgananID)
		}
		langganan.PaketLayanan = pk
		paket = pk
	}

	var brand *domain.HargaLayanan
	if pelanggan.IDBrand != nil {
		br, err := u.brandRepo.GetByID(ctx, *pelanggan.IDBrand)
		if err == nil {
			brand = br
		}
	}
	if brand == nil {
		return nil, fmt.Errorf("brand not found for pelanggan ID %d", pelanggan.ID)
	}
	pelanggan.HargaLayanan = brand

	var dataTeknis *domain.DataTeknis
	dt, err := u.dataTeknisRepo.GetByPelangganID(ctx, pelanggan.ID)
	if err == nil && dt != nil {
		dataTeknis = dt
	}
	if dataTeknis == nil {
		return nil, fmt.Errorf("data teknis not found for pelanggan ID %d", pelanggan.ID)
	}
	pelanggan.DataTeknis = dataTeknis

	hargaDasar := paket.Harga
	pajakPersen := brand.Pajak
	pajakMentah := hargaDasar * (pajakPersen / 100.0)
	pajak := math.Floor(pajakMentah + 0.5)
	totalHargaSebelumDiskon := hargaDasar + pajak
	totalHarga := totalHargaSebelumDiskon

	var diskonID *uint64
	var diskonPersen *float64
	var diskonAmount *float64

	clusterToCheck := strings.TrimSpace(pelanggan.Alamat)
	if langganan.MetodePembayaran == "Prorate" {
		// Prorate users do not get discount
	} else if clusterToCheck != "" {
		today := time.Now()
		diskon, err := u.diskonRepo.GetActiveForCluster(ctx, clusterToCheck, today)
		if err == nil && diskon != nil {
			diskonID = &diskon.ID
			persen := diskon.PersentaseDiskon
			diskonPersen = &persen
			amt := math.Floor((totalHargaSebelumDiskon * persen / 100.0) + 0.5)
			diskonAmount = &amt
			totalHarga = totalHargaSebelumDiskon - amt
		}
	}

	nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9]`)
	namaPelangganSingkat := strings.ToUpper(nonAlphanumericRegex.ReplaceAllString(pelanggan.Nama, ""))
	alamatSingkat := utils.GenerateAlamatSingkat(pelanggan.Alamat, pelanggan.Blok, pelanggan.Unit, 10)
	brandSingkat := strings.ToUpper(nonAlphanumericRegex.ReplaceAllString(brand.Brand, ""))

	dueDate := *langganan.TglJatuhTempo
	namingDate := dueDate
	if langganan.MetodePembayaran == "Prorate" && namingDate.Day() == 1 {
		namingDate = namingDate.AddDate(0, 0, -1)
	}

	monthNames := []string{
		"", "JANUARY", "FEBRUARY", "MARCH", "APRIL", "MAY", "JUNE",
		"JULY", "AUGUST", "SEPTEMBER", "OCTOBER", "NOVEMBER", "DECEMBER",
	}
	bulanTahun := fmt.Sprintf("%s-%d", monthNames[namingDate.Month()], namingDate.Year())

	idPfx := ""
	if len(dataTeknis.IDPelanggan) > 3 {
		idPfx = dataTeknis.IDPelanggan[len(dataTeknis.IDPelanggan)-3:]
	} else {
		idPfx = dataTeknis.IDPelanggan
	}

	nomorInvoiceBaru := fmt.Sprintf("%s/ftth/%s/%s/%s/%s", brandSingkat, namaPelangganSingkat, bulanTahun, alamatSingkat, idPfx)

	actualDueDate := *langganan.TglJatuhTempo
	if langganan.MetodePembayaran == "Prorate" && actualDueDate.Day() == 1 {
		actualDueDate = actualDueDate.AddDate(0, 0, -1)
	}

	startOfMonth := time.Date(actualDueDate.Year(), actualDueDate.Month(), 1, 0, 0, 0, 0, actualDueDate.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	existingInvoice, err := u.invoiceRepo.GetInvoiceByPelangganAndDueDateRange(ctx, pelanggan.ID, startOfMonth, endOfMonth)
	if err != nil {
		return nil, err
	}
	if existingInvoice != nil {
		return nil, fmt.Errorf("invoice untuk periode ini sudah ada (%s)", existingInvoice.InvoiceNumber)
	}

	newInvoice := &domain.Invoice{
		InvoiceNumber:      nomorInvoiceBaru,
		PelangganID:        pelanggan.ID,
		IDPelanggan:        dataTeknis.IDPelanggan,
		Brand:              brand.Brand,
		TotalHarga:         totalHarga,
		NoTelp:             pelanggan.NoTelp,
		Email:              pelanggan.Email,
		TglInvoice:         time.Now(),
		TglJatuhTempo:      actualDueDate,
		StatusInvoice:      "Belum Dibayar",
		InvoiceType:        "manual", // manual invoice
		DiskonID:           diskonID,
		DiskonPersen:       diskonPersen,
		DiskonAmount:       diskonAmount,
		HargaSebelumDiskon: &totalHargaSebelumDiskon,
	}

	// 🔒 COMMIT AWAL
	if err := u.invoiceRepo.Create(ctx, newInvoice); err != nil {
		return nil, err
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Manual Invoice basic %s saved to DB. Proceeding to Xendit...", newInvoice.InvoiceNumber))

	deskripsiXendit := ""
	jatuhTempoStrLengkap := actualDueDate.Format("02/01/2006")

	if langganan.MetodePembayaran == "Prorate" {
		hargaNormalFull := paket.Harga * (1.0 + (brand.Pajak / 100.0))
		if newInvoice.TotalHarga > (hargaNormalFull + 1.0) {
			invoiceDate := newInvoice.TglInvoice
			jatuhTempoDate := newInvoice.TglJatuhTempo

			var periodeBundleEnd time.Time
			if jatuhTempoDate.Day() == 1 {
				periodeBundleEnd = jatuhTempoDate.AddDate(0, 0, -1)
			} else {
				periodeBundleEnd = jatuhTempoDate
			}

			periodeProrateStart := invoiceDate
			nextMonth := time.Date(invoiceDate.Year(), invoiceDate.Month()+1, 1, 0, 0, 0, 0, invoiceDate.Location())
			periodeProrateEnd := nextMonth.AddDate(0, 0, -1)

			periodeProrateStr := periodeProrateEnd.Format("January 2006")
			periodeBerikutnyaStr := periodeBundleEnd.Format("January 2006")

			deskripsiXendit = fmt.Sprintf("Biaya internet up to %d Mbps. Periode Prorate %d-%d %s + Periode %s",
				paket.Kecepatan, periodeProrateStart.Day(), periodeProrateEnd.Day(), periodeProrateStr, periodeBerikutnyaStr)
		} else {
			invoiceDate := newInvoice.TglInvoice
			jatuhTempoDate := newInvoice.TglJatuhTempo

			periodeStart := invoiceDate
			var periodeEnd time.Time
			if jatuhTempoDate.Day() == 1 {
				periodeEnd = jatuhTempoDate.AddDate(0, 0, -1)
			} else {
				periodeEnd = jatuhTempoDate
			}

			diff := periodeEnd.Sub(periodeStart)
			if diff.Hours()/24 >= 27 {
				periodeStart = time.Date(periodeEnd.Year(), periodeEnd.Month(), 1, 0, 0, 0, 0, periodeEnd.Location())
			}

			periodDesc := ""
			if periodeStart.Month() == periodeEnd.Month() && periodeStart.Year() == periodeEnd.Year() {
				periodeStr := periodeEnd.Format("January 2006")
				periodDesc = fmt.Sprintf("Periode Tgl %d-%d %s", periodeStart.Day(), periodeEnd.Day(), periodeStr)
			} else if periodeStart.Year() == periodeEnd.Year() {
				startMonth := periodeStart.Format("January")
				endMonthYear := periodeEnd.Format("January 2006")
				periodDesc = fmt.Sprintf("Periode Tgl %d %s - %d %s", periodeStart.Day(), startMonth, periodeEnd.Day(), endMonthYear)
			} else {
				startFull := periodeStart.Format("02 January 2006")
				endFull := periodeEnd.Format("02 January 2006")
				periodDesc = fmt.Sprintf("Periode Tgl %s - %s", startFull, endFull)
			}

			deskripsiXendit = fmt.Sprintf("Biaya berlangganan internet up to %d Mbps, %s", paket.Kecepatan, periodDesc)
		}
	} else {
		deskripsiXendit = fmt.Sprintf("Biaya berlangganan internet up to %d Mbps jatuh tempo pembayaran tanggal %s",
			paket.Kecepatan, jatuhTempoStrLengkap)
	}

	noTelpXendit := utils.NormalizePhoneForXendit(pelanggan.NoTelp)

	xenditResponse, err := u.createXenditInvoice(ctx, newInvoice, pelanggan, paket, deskripsiXendit, pajak, noTelpXendit)
	if err != nil {
		u.logSystem(ctx, "ERROR", fmt.Sprintf("Xendit API failed for Langganan ID %d (Manual): %v", langgananID, err))
		newInvoice.XenditStatus = "failed"
		xenditErrMsg := err.Error()
		newInvoice.XenditErrorMessage = &xenditErrMsg
		_ = u.invoiceRepo.Update(ctx, newInvoice)
		return nil, err
	}

	shortURL, _ := xenditResponse["short_url"].(string)
	if shortURL == "" {
		shortURL, _ = xenditResponse["invoice_url"].(string)
	}
	xenditID, _ := xenditResponse["id"].(string)
	xenditExtID, _ := xenditResponse["external_id"].(string)

	newInvoice.PaymentLink = &shortURL
	newInvoice.XenditID = &xenditID
	newInvoice.XenditExternalID = &xenditExtID
	newInvoice.XenditStatus = "PENDING"

	// 🔒 COMMIT KEDUA
	if err := u.invoiceRepo.Update(ctx, newInvoice); err != nil {
		return nil, err
	}

	// Reload pelanggan relationships before returning (like in automatic script)
	newInvoice.Pelanggan = pelanggan
	newInvoice.Pelanggan.HargaLayanan = brand
	newInvoice.Pelanggan.DataTeknis = dataTeknis

	u.logSystem(ctx, "INFO", fmt.Sprintf("Manual Invoice %s generated successfully with payment link", newInvoice.InvoiceNumber))
	return newInvoice, nil
}

func (u *billingUsecase) generateSingleInvoice(ctx context.Context, langgananID uint64) (bool, error) {
	langganan, err := u.langgananRepo.GetByID(ctx, langgananID)
	if err != nil {
		return false, err
	}
	if langganan == nil {
		return false, errors.New("langganan not found")
	}

	pelanggan := langganan.Pelanggan
	if pelanggan == nil {
		p, err := u.pelangganRepo.GetByID(ctx, langganan.PelangganID)
		if err != nil || p == nil {
			return false, fmt.Errorf("pelanggan not found for langganan ID %d", langgananID)
		}
		langganan.Pelanggan = p
		pelanggan = p
	}

	paket := langganan.PaketLayanan
	if paket == nil {
		pk, err := u.paketRepo.GetByID(ctx, langganan.PaketLayananID)
		if err != nil || pk == nil {
			return false, fmt.Errorf("paket not found for langganan ID %d", langgananID)
		}
		langganan.PaketLayanan = pk
		paket = pk
	}

	var brand *domain.HargaLayanan
	if pelanggan.IDBrand != nil {
		br, err := u.brandRepo.GetByID(ctx, *pelanggan.IDBrand)
		if err == nil {
			brand = br
		}
	}
	if brand == nil {
		return false, fmt.Errorf("brand not found for pelanggan ID %d", pelanggan.ID)
	}
	pelanggan.HargaLayanan = brand

	var dataTeknis *domain.DataTeknis
	dt, err := u.dataTeknisRepo.GetByPelangganID(ctx, pelanggan.ID)
	if err == nil && dt != nil {
		dataTeknis = dt
	}
	if dataTeknis == nil {
		return false, fmt.Errorf("data teknis not found for pelanggan ID %d", pelanggan.ID)
	}
	pelanggan.DataTeknis = dataTeknis

	hargaDasar := paket.Harga
	pajakPersen := brand.Pajak
	pajakMentah := hargaDasar * (pajakPersen / 100.0)
	pajak := math.Floor(pajakMentah + 0.5)
	totalHargaSebelumDiskon := hargaDasar + pajak
	totalHarga := totalHargaSebelumDiskon

	var diskonID *uint64
	var diskonPersen *float64
	var diskonAmount *float64

	clusterToCheck := strings.TrimSpace(pelanggan.Alamat)
	if langganan.MetodePembayaran == "Prorate" {
		// Prorate users do not get discount
	} else if clusterToCheck != "" {
		today := time.Now()
		diskon, err := u.diskonRepo.GetActiveForCluster(ctx, clusterToCheck, today)
		if err == nil && diskon != nil {
			diskonID = &diskon.ID
			persen := diskon.PersentaseDiskon
			diskonPersen = &persen
			amt := math.Floor((totalHargaSebelumDiskon * persen / 100.0) + 0.5)
			diskonAmount = &amt
			totalHarga = totalHargaSebelumDiskon - amt
		}
	}

	nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9]`)
	namaPelangganSingkat := strings.ToUpper(nonAlphanumericRegex.ReplaceAllString(pelanggan.Nama, ""))
	alamatSingkat := utils.GenerateAlamatSingkat(pelanggan.Alamat, pelanggan.Blok, pelanggan.Unit, 10)
	brandSingkat := strings.ToUpper(nonAlphanumericRegex.ReplaceAllString(brand.Brand, ""))

	dueDate := *langganan.TglJatuhTempo
	namingDate := dueDate
	if langganan.MetodePembayaran == "Prorate" && namingDate.Day() == 1 {
		namingDate = namingDate.AddDate(0, 0, -1)
	}

	monthNames := []string{
		"", "JANUARY", "FEBRUARY", "MARCH", "APRIL", "MAY", "JUNE",
		"JULY", "AUGUST", "SEPTEMBER", "OCTOBER", "NOVEMBER", "DECEMBER",
	}
	bulanTahun := fmt.Sprintf("%s-%d", monthNames[namingDate.Month()], namingDate.Year())

	idPfx := ""
	if len(dataTeknis.IDPelanggan) > 3 {
		idPfx = dataTeknis.IDPelanggan[len(dataTeknis.IDPelanggan)-3:]
	} else {
		idPfx = dataTeknis.IDPelanggan
	}

	nomorInvoiceBaru := fmt.Sprintf("%s/ftth/%s/%s/%s/%s", brandSingkat, namaPelangganSingkat, bulanTahun, alamatSingkat, idPfx)

	actualDueDate := *langganan.TglJatuhTempo
	if langganan.MetodePembayaran == "Prorate" && actualDueDate.Day() == 1 {
		actualDueDate = actualDueDate.AddDate(0, 0, -1)
	}

	startOfMonth := time.Date(actualDueDate.Year(), actualDueDate.Month(), 1, 0, 0, 0, 0, actualDueDate.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	existingInvoice, err := u.invoiceRepo.GetInvoiceByPelangganAndDueDateRange(ctx, pelanggan.ID, startOfMonth, endOfMonth)
	if err != nil {
		return false, err
	}
	if existingInvoice != nil {
		u.logSystem(ctx, "WARNING", fmt.Sprintf("Invoice for %s period %s already exists (%s). Skipping.", namaPelangganSingkat, bulanTahun, existingInvoice.InvoiceNumber))
		return false, nil
	}

	newInvoice := &domain.Invoice{
		InvoiceNumber:      nomorInvoiceBaru,
		PelangganID:        pelanggan.ID,
		IDPelanggan:        dataTeknis.IDPelanggan,
		Brand:              brand.Brand,
		TotalHarga:         totalHarga,
		NoTelp:             pelanggan.NoTelp,
		Email:              pelanggan.Email,
		TglInvoice:         time.Now(),
		TglJatuhTempo:      actualDueDate,
		StatusInvoice:      "Belum Dibayar",
		InvoiceType:        "automatic",
		DiskonID:           diskonID,
		DiskonPersen:       diskonPersen,
		DiskonAmount:       diskonAmount,
		HargaSebelumDiskon: &totalHargaSebelumDiskon,
	}

	// 🔒 COMMIT AWAL (First Commit)
	if err := u.invoiceRepo.Create(ctx, newInvoice); err != nil {
		return false, err
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Invoice basic %s saved to DB. Proceeding to Xendit...", newInvoice.InvoiceNumber))

	deskripsiXendit := ""
	jatuhTempoStrLengkap := actualDueDate.Format("02/01/2006")

	if langganan.MetodePembayaran == "Prorate" {
		hargaNormalFull := paket.Harga * (1.0 + (brand.Pajak / 100.0))
		if newInvoice.TotalHarga > (hargaNormalFull + 1.0) {
			invoiceDate := newInvoice.TglInvoice
			jatuhTempoDate := newInvoice.TglJatuhTempo

			var periodeBundleEnd time.Time
			if jatuhTempoDate.Day() == 1 {
				periodeBundleEnd = jatuhTempoDate.AddDate(0, 0, -1)
			} else {
				periodeBundleEnd = jatuhTempoDate
			}

			periodeProrateStart := invoiceDate
			nextMonth := time.Date(invoiceDate.Year(), invoiceDate.Month()+1, 1, 0, 0, 0, 0, invoiceDate.Location())
			periodeProrateEnd := nextMonth.AddDate(0, 0, -1)

			periodeProrateStr := periodeProrateEnd.Format("January 2006")
			periodeBerikutnyaStr := periodeBundleEnd.Format("January 2006")

			deskripsiXendit = fmt.Sprintf("Biaya internet up to %d Mbps. Periode Prorate %d-%d %s + Periode %s",
				paket.Kecepatan, periodeProrateStart.Day(), periodeProrateEnd.Day(), periodeProrateStr, periodeBerikutnyaStr)
		} else {
			invoiceDate := newInvoice.TglInvoice
			jatuhTempoDate := newInvoice.TglJatuhTempo

			periodeStart := invoiceDate
			var periodeEnd time.Time
			if jatuhTempoDate.Day() == 1 {
				periodeEnd = jatuhTempoDate.AddDate(0, 0, -1)
			} else {
				periodeEnd = jatuhTempoDate
			}

			diff := periodeEnd.Sub(periodeStart)
			if diff.Hours()/24 >= 27 {
				periodeStart = time.Date(periodeEnd.Year(), periodeEnd.Month(), 1, 0, 0, 0, 0, periodeEnd.Location())
			}

			periodDesc := ""
			if periodeStart.Month() == periodeEnd.Month() && periodeStart.Year() == periodeEnd.Year() {
				periodeStr := periodeEnd.Format("January 2006")
				periodDesc = fmt.Sprintf("Periode Tgl %d-%d %s", periodeStart.Day(), periodeEnd.Day(), periodeStr)
			} else if periodeStart.Year() == periodeEnd.Year() {
				startMonth := periodeStart.Format("January")
				endMonthYear := periodeEnd.Format("January 2006")
				periodDesc = fmt.Sprintf("Periode Tgl %d %s - %d %s", periodeStart.Day(), startMonth, periodeEnd.Day(), endMonthYear)
			} else {
				startFull := periodeStart.Format("02 January 2006")
				endFull := periodeEnd.Format("02 January 2006")
				periodDesc = fmt.Sprintf("Periode Tgl %s - %s", startFull, endFull)
			}

			deskripsiXendit = fmt.Sprintf("Biaya berlangganan internet up to %d Mbps, %s", paket.Kecepatan, periodDesc)
		}
	} else {
		deskripsiXendit = fmt.Sprintf("Biaya berlangganan internet up to %d Mbps jatuh tempo pembayaran tanggal %s",
			paket.Kecepatan, jatuhTempoStrLengkap)
	}

	noTelpXendit := utils.NormalizePhoneForXendit(pelanggan.NoTelp)

	xenditResponse, err := u.createXenditInvoice(ctx, newInvoice, pelanggan, paket, deskripsiXendit, pajak, noTelpXendit)
	if err != nil {
		u.logSystem(ctx, "ERROR", fmt.Sprintf("Xendit API failed for Langganan ID %d: %v", langgananID, err))
		newInvoice.XenditStatus = "failed"
		xenditErrMsg := err.Error()
		newInvoice.XenditErrorMessage = &xenditErrMsg
		_ = u.invoiceRepo.Update(ctx, newInvoice)
		return true, nil
	}

	shortURL, _ := xenditResponse["short_url"].(string)
	if shortURL == "" {
		shortURL, _ = xenditResponse["invoice_url"].(string)
	}
	xenditID, _ := xenditResponse["id"].(string)
	xenditExtID, _ := xenditResponse["external_id"].(string)

	newInvoice.PaymentLink = &shortURL
	newInvoice.XenditID = &xenditID
	newInvoice.XenditExternalID = &xenditExtID
	newInvoice.XenditStatus = "PENDING"

	// 🔒 COMMIT KEDUA (Update Xendit Data)
	if err := u.invoiceRepo.Update(ctx, newInvoice); err != nil {
		return false, err
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Invoice %s generated successfully with payment link", newInvoice.InvoiceNumber))
	return true, nil
}

func (u *billingUsecase) getPaidInvoiceIDsSince(ctx context.Context, days int) ([]string, error) {
	startDate := time.Now().AddDate(0, 0, -days).UTC().Format(time.RFC3339)
	apiKeys := u.cfg.XENDIT_API_KEYS()
	var allPaidIDs []string
	uniqueAPIKeys := make(map[string]string)
	for brandName, apiKey := range apiKeys {
		if apiKey != "" {
			if _, exists := uniqueAPIKeys[apiKey]; !exists {
				uniqueAPIKeys[apiKey] = brandName
			}
		}
	}

	for apiKey, brandName := range uniqueAPIKeys {
		u.logSystem(ctx, "INFO", fmt.Sprintf("Checking paid invoices for brand key associated with: %s", brandName))
		encodedKey := base64.StdEncoding.EncodeToString([]byte(apiKey + ":"))

		baseURL := "https://api.xendit.co/v2/invoices"
		pageSize := 100
		lastInvoiceID := ""
		pageCount := 0
		maxPages := 20
		var brandPaidIDs []string

		client := &http.Client{Timeout: 30 * time.Second}

		for pageCount < maxPages {
			pageCount++
			reqURL := fmt.Sprintf("%s?statuses[]=PAID&paid_after=%s&limit=%d", baseURL, startDate, pageSize)
			if lastInvoiceID != "" {
				reqURL = fmt.Sprintf("%s&after_id=%s", reqURL, lastInvoiceID)
			}

			req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
			if err != nil {
				u.logSystem(ctx, "ERROR", fmt.Sprintf("Failed to create request for brand %s: %v", brandName, err))
				break
			}

			req.Header.Set("Authorization", "Basic "+encodedKey)
			resp, err := client.Do(req)
			if err != nil {
				u.logSystem(ctx, "ERROR", fmt.Sprintf("Network error fetching brand %s: %v", brandName, err))
				break
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				u.logSystem(ctx, "ERROR", fmt.Sprintf("Xendit API error for brand %s, status %d", brandName, resp.StatusCode))
				resp.Body.Close()
				break
			}

			var responseData struct {
				Data []struct {
					ID         string `json:"id"`
					ExternalID string `json:"external_id"`
				} `json:"data"`
			}

			err = json.NewDecoder(resp.Body).Decode(&responseData)
			resp.Body.Close()
			if err != nil {
				u.logSystem(ctx, "ERROR", fmt.Sprintf("Failed to decode response for brand %s: %v", brandName, err))
				break
			}

			if len(responseData.Data) == 0 {
				break
			}

			for _, item := range responseData.Data {
				if item.ExternalID != "" {
					brandPaidIDs = append(brandPaidIDs, item.ExternalID)
				}
			}

			if len(responseData.Data) < pageSize {
				break
			}

			lastInvoiceID = responseData.Data[len(responseData.Data)-1].ID
			if lastInvoiceID == "" {
				break
			}
		}

		allPaidIDs = append(allPaidIDs, brandPaidIDs...)
		u.logSystem(ctx, "INFO", fmt.Sprintf("Found %d paid invoices for brand %s (%d pages).", len(brandPaidIDs), brandName, pageCount))
	}

	return allPaidIDs, nil
}

func (u *billingUsecase) GenerateInvoices(ctx context.Context) error {
	u.logSystem(ctx, "INFO", "Scheduler 'job_generate_invoices' started")

	targetDueDate := time.Now().AddDate(0, 0, 5)
	startOfDay := time.Date(targetDueDate.Year(), targetDueDate.Month(), targetDueDate.Day(), 0, 0, 0, 0, targetDueDate.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	langganans, err := u.langgananRepo.GetActiveByDueDateRange(ctx, startOfDay, endOfDay)
	if err != nil {
		u.logSystem(ctx, "ERROR", fmt.Sprintf("Scheduler 'job_generate_invoices' failed to query active langganan: %v", err))
		return err
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Found %d active subscriptions for due date %s", len(langganans), startOfDay.Format("2006-01-02")))

	successCount := 0
	for _, langganan := range langganans {
		success, err := u.generateSingleInvoice(ctx, langganan.ID)
		if err != nil {
			u.logSystem(ctx, "ERROR", fmt.Sprintf("Failed to generate invoice for langganan ID %d: %v", langganan.ID, err))
		} else if success {
			successCount++
		}
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Scheduler 'job_generate_invoices' completed. Successfully generated %d invoices.", successCount))
	return nil
}

func (u *billingUsecase) AutoSuspend(ctx context.Context) error {
	u.logSystem(ctx, "INFO", "Scheduler 'job_suspend_services' started")

	currentDate := time.Now()
	isMainSuspendDay := currentDate.Day() == 5
	isRetroactiveDay := currentDate.Day() >= 6 && currentDate.Day() <= 10

	if !(isMainSuspendDay || isRetroactiveDay) {
		u.logSystem(ctx, "INFO", fmt.Sprintf("Scheduler 'job_suspend_services' completed. Day of month: %d (suspend only runs between 5-10).", currentDate.Day()))
		return nil
	}

	targetDueDate := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())
	endOfPrevMonth := targetDueDate.AddDate(0, 0, -1)

	langganans, err := u.langgananRepo.GetActiveOverdueForSuspend(ctx, targetDueDate, endOfPrevMonth)
	if err != nil {
		u.logSystem(ctx, "ERROR", fmt.Sprintf("Scheduler 'job_suspend_services' failed to query overdue subscriptions: %v", err))
		return err
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Found %d active subscriptions overdue for suspend", len(langganans)))

	suspendedCount := 0
	invoicesOverdueCount := 0

	for _, langganan := range langganans {
		isPaid, err := u.invoiceRepo.HasPaidInvoiceForPeriod(ctx, langganan.PelangganID, targetDueDate, endOfPrevMonth)
		if err != nil {
			continue
		}
		if isPaid {
			u.logSystem(ctx, "INFO", fmt.Sprintf("SKIP SUSPEND: Pelanggan %s already paid for period", langganan.Pelanggan.Nama))
			continue
		}

		u.logSystem(ctx, "WARNING", fmt.Sprintf("SUSPEND: Suspending service for Langganan ID: %d - Pelanggan: %s", langganan.ID, langganan.Pelanggan.Nama))

		dataTeknis := langganan.Pelanggan.DataTeknis
		mikrotikSuccess := false
		var mikrotikErrorMsg string
		originalStatus := langganan.Status

		if dataTeknis != nil {
			langganan.Status = "Suspended"
			err = u.triggerMikrotikUpdate(ctx, dataTeknis.IDPelanggan, dataTeknis, "Suspended")
			if err != nil {
				mikrotikErrorMsg = err.Error()
				u.logSystem(ctx, "ERROR", fmt.Sprintf("Mikrotik suspend FAILED for Langganan ID: %d. Error: %s", langganan.ID, err.Error()))
				dataTeknis.MikrotikSyncPending = true
				_ = u.dataTeknisRepo.Update(ctx, dataTeknis)
			} else {
				mikrotikSuccess = true
				u.logSystem(ctx, "INFO", fmt.Sprintf("Mikrotik suspend SUCCESS for Langganan ID: %d", langganan.ID))
			}
		} else {
			u.logSystem(ctx, "WARNING", fmt.Sprintf("Data Teknis not found for langganan ID %d, suspending only in DB.", langganan.ID))
		}

		langganan.Status = "Suspended"
		dbSuccess := false
		if err := u.langgananRepo.Update(ctx, &langganan); err == nil {
			dbSuccess = true
			rows, _ := u.invoiceRepo.UpdateStatusForUnpaidInvoices(ctx, langganan.PelangganID, targetDueDate, endOfPrevMonth, "Expired")
			invoicesOverdueCount += int(rows)
			suspendedCount++
			u.logSystem(ctx, "INFO", fmt.Sprintf("Database update SUCCESS for Langganan ID: %d. Status = Suspended.", langganan.ID))
		} else {
			u.logSystem(ctx, "ERROR", fmt.Sprintf("CRITICAL: Database update FAILED for Langganan ID %d. Error: %v", langganan.ID, err))
			if mikrotikSuccess && dataTeknis != nil {
				langganan.Status = originalStatus
				_ = u.triggerMikrotikUpdate(ctx, dataTeknis.IDPelanggan, dataTeknis, originalStatus)
				u.logSystem(ctx, "INFO", fmt.Sprintf("Rollback Mikrotik success for Pelanggan: %s", langganan.Pelanggan.Nama))
			}
		}

		if dbSuccess {
			if mikrotikSuccess {
				u.logSystem(ctx, "INFO", fmt.Sprintf("SUSPEND COMPLETE (Mikrotik + DB) for: %s", langganan.Pelanggan.Nama))
			} else {
				u.logSystem(ctx, "WARNING", fmt.Sprintf("SUSPEND PARTIAL (DB only, Mikrotik failed) for: %s. Details: %s", langganan.Pelanggan.Nama, mikrotikErrorMsg))
			}
		}
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Scheduler 'job_suspend_services' completed. Suspended %d services and expired %d invoices.", suspendedCount, invoicesOverdueCount))
	return nil
}

func (u *billingUsecase) VerifyPayments(ctx context.Context) error {
	u.logSystem(ctx, "INFO", "Scheduler 'job_verify_payments' started")

	paidInvoiceIDs, err := u.getPaidInvoiceIDsSince(ctx, 3)
	if err != nil {
		u.logSystem(ctx, "ERROR", fmt.Sprintf("Scheduler 'job_verify_payments' failed to query Xendit: %v", err))
		return err
	}

	if len(paidInvoiceIDs) == 0 {
		u.logSystem(ctx, "INFO", "Scheduler 'job_verify_payments' completed. No new paid invoices from Xendit.")
		return nil
	}

	invoices, err := u.invoiceRepo.GetUnpaidByExternalIDs(ctx, paidInvoiceIDs)
	if err != nil {
		u.logSystem(ctx, "ERROR", fmt.Sprintf("Scheduler 'job_verify_payments' failed to query database: %v", err))
		return err
	}

	processedCount := 0
	for _, invoice := range invoices {
		paidAmount := invoice.TotalHarga
		paidAt := time.Now()

		u.logSystem(ctx, "INFO", fmt.Sprintf("[VERIFY] Reconciling missed payment for Invoice: %s", invoice.InvoiceNumber))
		err = u.processSuccessfulPayment(ctx, &invoice, paidAmount, paidAt)
		if err != nil {
			u.logSystem(ctx, "ERROR", fmt.Sprintf("Failed to reconcile invoice %s: %v", invoice.InvoiceNumber, err))
		} else {
			processedCount++
		}
	}

	u.logSystem(ctx, "INFO", fmt.Sprintf("Scheduler 'job_verify_payments' completed. Reconciled %d payments.", processedCount))
	return nil
}


