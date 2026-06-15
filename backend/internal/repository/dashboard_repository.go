package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) domain.DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) GetRevenueSummary(ctx context.Context) (*domain.RevenueSummary, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var endOfMonth time.Time
	if now.Month() == 12 {
		endOfMonth = time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, now.Location())
	} else {
		endOfMonth = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	}

	type Result struct {
		Brand        string
		TotalRevenue float64
	}
	var results []Result

	err := r.db.WithContext(ctx).Table("invoices").
		Select("harga_layanan.brand, SUM(invoices.total_harga) as total_revenue").
		Joins("LEFT JOIN pelanggan ON invoices.pelanggan_id = pelanggan.id").
		Joins("LEFT JOIN harga_layanan ON pelanggan.id_brand = harga_layanan.id_brand").
		Where("invoices.status_invoice = ?", "Lunas").
		Where("harga_layanan.brand IS NOT NULL").
		Where("invoices.paid_at >= ?", startOfMonth).
		Where("invoices.paid_at < ?", endOfMonth).
		Where("invoices.deleted_at IS NULL").
		Group("harga_layanan.brand").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var breakdown []domain.BrandRevenueItem
	var total float64
	for _, row := range results {
		breakdown = append(breakdown, domain.BrandRevenueItem{
			Brand:   row.Brand,
			Revenue: row.TotalRevenue,
		})
		total += row.TotalRevenue
	}

	nextMonth := now.AddDate(0, 1, 0)
	monthsID := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	periodeStr := fmt.Sprintf("%s %d", monthsID[nextMonth.Month()-1], nextMonth.Year())

	return &domain.RevenueSummary{
		Total:     total,
		Periode:   periodeStr,
		Breakdown: breakdown,
	}, nil
}

func (r *dashboardRepository) GetPelangganStatCards(ctx context.Context) ([]domain.StatCard, error) {
	type Result struct {
		Brand string
		Count int
	}
	var results []Result

	err := r.db.WithContext(ctx).Table("harga_layanan").
		Select("harga_layanan.brand, COUNT(pelanggan.id) as count").
		Joins("LEFT JOIN pelanggan ON harga_layanan.id_brand = pelanggan.id_brand AND pelanggan.deleted_at IS NULL").
		Group("harga_layanan.brand").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, row := range results {
		counts[strings.ToLower(row.Brand)] = row.Count
	}

	return []domain.StatCard{
		{
			Title:       "Jumlah Pelanggan Jakinet",
			Value:       counts["jakinet"],
			Description: "Total Pelanggan Jakinet",
		},
		{
			Title:       "Jumlah Pelanggan Jelantik",
			Value:       counts["jelantik"],
			Description: "Total Pelanggan Jelantik",
		},
		{
			Title:       "Pelanggan Jelantik Nagrak",
			Value:       counts["jelantik nagrak"],
			Description: "Total Pelanggan Rusun Nagrak",
		},
	}, nil
}

func (r *dashboardRepository) GetLoyaltyChart(ctx context.Context) (*domain.ChartData, error) {
	var totalActive int64
	err := r.db.WithContext(ctx).Table("langganan").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("langganan.status = ?", "Aktif").
		Count(&totalActive).Error
	if err != nil {
		return nil, err
	}

	var outstandingCount int64
	err = r.db.WithContext(ctx).Table("langganan").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("langganan.status = ?", "Aktif").
		Where("langganan.pelanggan_id IN (SELECT DISTINCT pelanggan_id FROM invoices WHERE status_invoice IN (?, ?, ?) AND deleted_at IS NULL)", "Belum Bayar", "Expired", "Expired").
		Count(&outstandingCount).Error
	if err != nil {
		return nil, err
	}

	var everLateCount int64
	err = r.db.WithContext(ctx).Table("langganan").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("langganan.status = ?", "Aktif").
		Where("langganan.pelanggan_id IN (SELECT DISTINCT pelanggan_id FROM invoices WHERE paid_at > tgl_jatuh_tempo AND deleted_at IS NULL)").
		Where("langganan.pelanggan_id NOT IN (SELECT DISTINCT pelanggan_id FROM invoices WHERE status_invoice IN (?, ?, ?) AND deleted_at IS NULL)", "Belum Bayar", "Expired", "Expired").
		Count(&everLateCount).Error
	if err != nil {
		return nil, err
	}

	setiaCount := totalActive - outstandingCount - everLateCount
	if setiaCount < 0 {
		setiaCount = 0
	}

	return &domain.ChartData{
		Labels: []string{"Setia On-Time", "Lunas (Tapi Telat)", "Menunggak"},
		Data: []int{
			int(setiaCount),
			int(everLateCount),
			int(outstandingCount),
		},
	}, nil
}

func (r *dashboardRepository) GetLokasiChart(ctx context.Context) (*domain.ChartData, error) {
	type Result struct {
		Alamat string
		Count  int
	}
	var results []Result

	err := r.db.WithContext(ctx).Table("pelanggan").
		Select("alamat, COUNT(id) as count").
		Where("alamat IS NOT NULL AND alamat != ''").
		Where("deleted_at IS NULL").
		Group("alamat").
		Order("count DESC").
		Limit(15).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var labels []string
	var data []int
	for _, row := range results {
		labels = append(labels, row.Alamat)
		data = append(data, row.Count)
	}

	return &domain.ChartData{
		Labels: labels,
		Data:   data,
	}, nil
}

func (r *dashboardRepository) GetPaketChart(ctx context.Context) (*domain.ChartData, error) {
	type Result struct {
		Kecepatan string
		Count     int
	}
	var results []Result

	err := r.db.WithContext(ctx).Table("paket_layanan").
		Select("paket_layanan.kecepatan, COUNT(langganan.id) as count").
		Joins("LEFT JOIN langganan ON paket_layanan.id = langganan.paket_layanan_id").
		Joins("LEFT JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("langganan.status = ?", "Aktif").
		Group("paket_layanan.kecepatan").
		Order("paket_layanan.kecepatan").
		Limit(10).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var labels []string
	var data []int
	for _, row := range results {
		labels = append(labels, fmt.Sprintf("%s Mbps", row.Kecepatan))
		data = append(data, row.Count)
	}

	return &domain.ChartData{
		Labels: labels,
		Data:   data,
	}, nil
}

func (r *dashboardRepository) GetGrowthChart(ctx context.Context) (*domain.ChartData, error) {
	twoYearsAgo := time.Now().AddDate(-2, 0, 0)

	type Result struct {
		Year   int
		Month  int
		Jumlah int
	}
	var results []Result

	err := r.db.WithContext(ctx).Table("pelanggan").
		Select("YEAR(tgl_instalasi) as year, MONTH(tgl_instalasi) as month, COUNT(id) as jumlah").
		Where("tgl_instalasi >= ?", twoYearsAgo).
		Where("deleted_at IS NULL").
		Group("YEAR(tgl_instalasi), MONTH(tgl_instalasi)").
		Order("year ASC, month ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	monthsShort := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	var labels []string
	var data []int
	for _, row := range results {
		if row.Month >= 1 && row.Month <= 12 {
			labels = append(labels, fmt.Sprintf("%s %d", monthsShort[row.Month-1], row.Year))
			data = append(data, row.Jumlah)
		}
	}

	return &domain.ChartData{
		Labels: labels,
		Data:   data,
	}, nil
}

func (r *dashboardRepository) GetInvoiceSummaryChart(ctx context.Context) (*domain.InvoiceSummary, error) {
	sixMonthsAgo := time.Now().AddDate(0, 0, -180)

	type Result struct {
		Year       int
		Month      int
		Total      int
		Lunas      int
		Menunggu   int
		Expired    int
		Otomatis   int
		Manual     int
		Reinvoice  int
	}
	var results []Result

	err := r.db.WithContext(ctx).Table("invoices").
		Select(`
			YEAR(tgl_invoice) as year, 
			MONTH(tgl_invoice) as month, 
			COUNT(id) as total,
			SUM(CASE WHEN status_invoice = 'Lunas' THEN 1 ELSE 0 END) as lunas,
			SUM(CASE WHEN status_invoice = 'Belum Bayar' THEN 1 ELSE 0 END) as menunggu,
			SUM(CASE WHEN status_invoice IN ('Expired', 'Expired') THEN 1 ELSE 0 END) as expired,
			SUM(CASE WHEN invoice_type = 'automatic' THEN 1 ELSE 0 END) as otomatis,
			SUM(CASE WHEN invoice_type = 'manual' THEN 1 ELSE 0 END) as manual,
			SUM(CASE WHEN is_reinvoice = 1 THEN 1 ELSE 0 END) as reinvoice
		`).
		Where("tgl_invoice >= ?", sixMonthsAgo).
		Where("deleted_at IS NULL").
		Group("YEAR(tgl_invoice), MONTH(tgl_invoice)").
		Order("year ASC, month ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	monthsShort := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	var labels []string
	var total []int
	var lunas []int
	var menunggu []int
	var expired []int
	var otomatis []int
	var manual []int
	var reinvoice []int

	if len(results) == 0 {
		now := time.Now()
		for i := 5; i >= 0; i-- {
			t := now.AddDate(0, -i, 0)
			labels = append(labels, fmt.Sprintf("%s %d", monthsShort[t.Month()-1], t.Year()))
			total = append(total, 0)
			lunas = append(lunas, 0)
			menunggu = append(menunggu, 0)
			expired = append(expired, 0)
			otomatis = append(otomatis, 0)
			manual = append(manual, 0)
			reinvoice = append(reinvoice, 0)
		}
	} else {
		for _, row := range results {
			if row.Month >= 1 && row.Month <= 12 {
				labels = append(labels, fmt.Sprintf("%s %d", monthsShort[row.Month-1], row.Year))
				total = append(total, row.Total)
				lunas = append(lunas, row.Lunas)
				menunggu = append(menunggu, row.Menunggu)
				expired = append(expired, row.Expired)
				otomatis = append(otomatis, row.Otomatis)
				manual = append(manual, row.Manual)
				reinvoice = append(reinvoice, row.Reinvoice)
			}
		}
	}

	return &domain.InvoiceSummary{
		Labels:     labels,
		Total:      total,
		Lunas:      lunas,
		Menunggu:   menunggu,
		Expired: expired,
		Otomatis:   otomatis,
		Manual:     manual,
		Reinvoice:  reinvoice,
	}, nil
}

func (r *dashboardRepository) GetStatusLanggananChart(ctx context.Context) (*domain.ChartData, error) {
	type Result struct {
		Status string
		Jumlah int
	}
	var results []Result

	err := r.db.WithContext(ctx).Table("langganan").
		Select("langganan.status, COUNT(langganan.id) as jumlah").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Group("langganan.status").
		Order("langganan.status").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var labels []string
	var data []int
	for _, row := range results {
		labels = append(labels, row.Status)
		data = append(data, row.Jumlah)
	}

	return &domain.ChartData{
		Labels: labels,
		Data:   data,
	}, nil
}

func (r *dashboardRepository) GetPelangganPerAlamatChart(ctx context.Context) (*domain.ChartData, error) {
	type Result struct {
		Alamat string
		Jumlah int
	}
	var results []Result

	err := r.db.WithContext(ctx).Table("pelanggan").
		Select("pelanggan.alamat, COUNT(pelanggan.id) as jumlah").
		Joins("JOIN langganan ON pelanggan.id = langganan.pelanggan_id").
		Where("langganan.status = ?", "Aktif").
		Where("pelanggan.deleted_at IS NULL").
		Group("pelanggan.alamat").
		Order("jumlah DESC").
		Limit(20).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var labels []string
	var data []int
	for _, row := range results {
		labels = append(labels, row.Alamat)
		data = append(data, row.Jumlah)
	}

	return &domain.ChartData{
		Labels: labels,
		Data:   data,
	}, nil
}

func (r *dashboardRepository) GetLoyaltyUsersBySegment(ctx context.Context, segmen string) ([]domain.LoyalitasUserDetail, error) {
	type QueryResult struct {
		ID               uint64
		Nama             string
		Alamat           *string
		NoTelp           *string
		IDPelanggan      *string
		OutstandingCount int64
		LateCount        int64
	}
	var rawData []QueryResult

	err := r.db.WithContext(ctx).Table("pelanggan").
		Select(`
			pelanggan.id,
			pelanggan.nama,
			pelanggan.alamat,
			pelanggan.no_telp,
			data_teknis.id_pelanggan,
			SUM(CASE WHEN invoices.status_invoice IN ('Belum Bayar', 'Expired', 'Expired') THEN 1 ELSE 0 END) as outstanding_count,
			SUM(CASE WHEN invoices.paid_at > invoices.tgl_jatuh_tempo THEN 1 ELSE 0 END) as late_count
		`).
		Joins("JOIN langganan ON pelanggan.id = langganan.pelanggan_id").
		Joins("LEFT JOIN invoices ON pelanggan.id = invoices.pelanggan_id AND invoices.deleted_at IS NULL").
		Joins("LEFT JOIN data_teknis ON pelanggan.id = data_teknis.pelanggan_id").
		Where("langganan.status = ?", "Aktif").
		Where("pelanggan.deleted_at IS NULL").
		Group("pelanggan.id, pelanggan.nama, pelanggan.alamat, pelanggan.no_telp, data_teknis.id_pelanggan").
		Scan(&rawData).Error

	if err != nil {
		return nil, err
	}

	var filtered []domain.LoyalitasUserDetail
	for _, row := range rawData {
		isOutstanding := row.OutstandingCount > 0
		isEverLate := row.LateCount > 0

		include := false
		if segmen == "Menunggak" && isOutstanding {
			include = true
		} else if segmen == "Lunas (Tapi Telat)" && !isOutstanding && isEverLate {
			include = true
		} else if segmen == "Setia On-Time" && !isOutstanding && !isEverLate {
			include = true
		}

		if include {
			idPel := fmt.Sprintf("PLG-%04d", row.ID)
			if row.IDPelanggan != nil && *row.IDPelanggan != "" {
				idPel = *row.IDPelanggan
			}
			alamat := "Alamat tidak tersedia"
			if row.Alamat != nil && *row.Alamat != "" {
				alamat = *row.Alamat
			}
			noTelp := "Nomor tidak tersedia"
			if row.NoTelp != nil && *row.NoTelp != "" {
				noTelp = *row.NoTelp
			}

			filtered = append(filtered, domain.LoyalitasUserDetail{
				ID:          row.ID,
				Nama:        row.Nama,
				IDPelanggan: idPel,
				Alamat:      alamat,
				NoTelp:      noTelp,
			})
		}
	}

	return filtered, nil
}

func (r *dashboardRepository) GetSidebarBadges(ctx context.Context) (*domain.SidebarBadgeResponse, error) {
	var suspendedCount int64
	if err := r.db.WithContext(ctx).Table("langganan").Where("status = ?", "Suspended").Count(&suspendedCount).Error; err != nil {
		return nil, err
	}

	var stoppedCount int64
	if err := r.db.WithContext(ctx).Table("langganan").Where("status = ?", "Berhenti").Count(&stoppedCount).Error; err != nil {
		return nil, err
	}

	var totalInvoiceCount int64
	if err := r.db.WithContext(ctx).Table("invoices").Where("deleted_at IS NULL").Count(&totalInvoiceCount).Error; err != nil {
		return nil, err
	}

	var openTicketsCount int64
	if err := r.db.WithContext(ctx).Table("trouble_ticket").Where("status = ?", "Open").Count(&openTicketsCount).Error; err != nil {
		return nil, err
	}

	today := time.Now()
	activeCutoff := time.Date(today.Year(), today.Month()-1, 1, 0, 0, 0, 0, today.Location())

	var unpaidCount int64
	err := r.db.WithContext(ctx).Table("invoices").
		Where("status_invoice = ?", "Belum Bayar").
		Where("tgl_invoice >= ?", activeCutoff).
		Where("deleted_at IS NULL").
		Count(&unpaidCount).Error
	if err != nil {
		return nil, err
	}

	return &domain.SidebarBadgeResponse{
		SuspendedCount:     int(suspendedCount),
		UnpaidInvoiceCount: int(unpaidCount),
		StoppedCount:       int(stoppedCount),
		TotalInvoiceCount:  int(totalInvoiceCount),
		OpenTicketsCount:   int(openTicketsCount),
	}, nil
}

func (r *dashboardRepository) GetPaketDetails(ctx context.Context) (map[string]domain.PaketDetail, error) {
	type QueryResult struct {
		Kecepatan string
		Alamat    string
		Brand     string
		Jumlah    int
	}
	var rawData []QueryResult

	err := r.db.WithContext(ctx).Table("paket_layanan").
		Select("paket_layanan.kecepatan, pelanggan.alamat, harga_layanan.brand, COUNT(pelanggan.id) as jumlah").
		Joins("JOIN langganan ON paket_layanan.id = langganan.paket_layanan_id").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Joins("JOIN harga_layanan ON pelanggan.id_brand = harga_layanan.id_brand").
		Group("paket_layanan.kecepatan, pelanggan.alamat, harga_layanan.brand").
		Order("paket_layanan.kecepatan ASC, jumlah DESC").
		Scan(&rawData).Error

	if err != nil {
		return nil, err
	}

	paketRaw := make(map[string]*struct {
		TotalPelanggan int
		LokasiMap      map[string]int
		BrandMap       map[string]int
	})

	for _, row := range rawData {
		if row.Alamat == "" || row.Brand == "" {
			continue
		}

		paketKey := fmt.Sprintf("%s Mbps", row.Kecepatan)
		if _, ok := paketRaw[paketKey]; !ok {
			paketRaw[paketKey] = &struct {
				TotalPelanggan int
				LokasiMap      map[string]int
				BrandMap       map[string]int
			}{
				LokasiMap: make(map[string]int),
				BrandMap:  make(map[string]int),
			}
		}

		paketRaw[paketKey].TotalPelanggan += row.Jumlah
		paketRaw[paketKey].LokasiMap[row.Alamat] += row.Jumlah
		paketRaw[paketKey].BrandMap[row.Brand] += row.Jumlah
	}

	finalResponse := make(map[string]domain.PaketDetail)
	for key, details := range paketRaw {
		var breakdownLokasi []domain.BreakdownItem
		for name, jml := range details.LokasiMap {
			breakdownLokasi = append(breakdownLokasi, domain.BreakdownItem{Nama: name, Jumlah: jml})
		}
		// Sort locations descending
		for i := 0; i < len(breakdownLokasi); i++ {
			for j := i + 1; j < len(breakdownLokasi); j++ {
				if breakdownLokasi[i].Jumlah < breakdownLokasi[j].Jumlah {
					breakdownLokasi[i], breakdownLokasi[j] = breakdownLokasi[j], breakdownLokasi[i]
				}
			}
		}

		var breakdownBrand []domain.BreakdownItem
		for name, jml := range details.BrandMap {
			breakdownBrand = append(breakdownBrand, domain.BreakdownItem{Nama: name, Jumlah: jml})
		}
		// Sort brands descending
		for i := 0; i < len(breakdownBrand); i++ {
			for j := i + 1; j < len(breakdownBrand); j++ {
				if breakdownBrand[i].Jumlah < breakdownBrand[j].Jumlah {
					breakdownBrand[i], breakdownBrand[j] = breakdownBrand[j], breakdownBrand[i]
				}
			}
		}

		finalResponse[key] = domain.PaketDetail{
			TotalPelanggan:  details.TotalPelanggan,
			BreakdownLokasi: breakdownLokasi,
			BreakdownBrand:  breakdownBrand,
		}
	}

	return finalResponse, nil
}

func (r *dashboardRepository) GetInvoiceGenerationMonitor(ctx context.Context, targetDate string) (*domain.InvoiceGenerationMonitorResponse, error) {
	var targetDateObj time.Time
	var err error
	if targetDate != "" {
		targetDateObj, err = time.Parse("2006-01-02", targetDate)
		if err != nil {
			return nil, fmt.Errorf("invalid target_date format: %w", err)
		}
	} else {
		today := time.Now()
		m := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
		var mPlus1 time.Time
		if m.Month() == 12 {
			mPlus1 = time.Date(m.Year()+1, 1, 1, 0, 0, 0, 0, today.Location())
		} else {
			mPlus1 = time.Date(m.Year(), m.Month()+1, 1, 0, 0, 0, 0, today.Location())
		}

		genDateMPlus1 := mPlus1.AddDate(0, 0, -5)
		if today.After(genDateMPlus1) || today.Equal(genDateMPlus1) {
			targetDateObj = mPlus1
		} else {
			targetDateObj = m
		}
	}

	targetYear, targetMonth := targetDateObj.Year(), targetDateObj.Month()
	startOfMonth := time.Date(targetYear, targetMonth, 1, 0, 0, 0, 0, targetDateObj.Location())
	var endOfMonth time.Time
	if targetMonth == 12 {
		endOfMonth = time.Date(targetYear+1, 1, 1, 0, 0, 0, 0, targetDateObj.Location()).Add(-24 * time.Hour)
	} else {
		endOfMonth = time.Date(targetYear, targetMonth+1, 1, 0, 0, 0, 0, targetDateObj.Location()).Add(-24 * time.Hour)
	}

	var totalShouldHave int64
	err = r.db.WithContext(ctx).Table("langganan").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("COALESCE(langganan.tgl_jatuh_tempo_pembayaran, langganan.tgl_jatuh_tempo) BETWEEN ? AND ? AND langganan.status = ?", startOfMonth, endOfMonth, "Aktif").
		Count(&totalShouldHave).Error
	if err != nil {
		return nil, err
	}

	var totalGenerated int64
	err = r.db.WithContext(ctx).Table("invoices").
		Select("COUNT(DISTINCT pelanggan_id)").
		Where("DATE(tgl_jatuh_tempo) BETWEEN ? AND ?", startOfMonth.Format("2006-01-02"), endOfMonth.Format("2006-01-02")).
		Where("deleted_at IS NULL").
		Row().Scan(&totalGenerated)
	if err != nil {
		totalGenerated = 0
	}

	var totalSkipped int64
	err = r.db.WithContext(ctx).Table("langganan").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Joins("LEFT JOIN invoices ON langganan.pelanggan_id = invoices.pelanggan_id AND invoices.deleted_at IS NULL AND invoices.tgl_jatuh_tempo BETWEEN ? AND ?", startOfMonth, endOfMonth).
		Where("COALESCE(langganan.tgl_jatuh_tempo_pembayaran, langganan.tgl_jatuh_tempo) BETWEEN ? AND ?", startOfMonth, endOfMonth).
		Where("langganan.status = ?", "Aktif").
		Where("invoices.id IS NULL").
		Count(&totalSkipped).Error
	if err != nil {
		totalSkipped = 0
	}

	today := time.Now()
	generationDate := targetDateObj.AddDate(0, 0, -5)
	generationHour := 13
	isUpcoming := today.Before(generationDate) || (today.Equal(generationDate) && today.Hour() < generationHour)

	if !isUpcoming {
		totalShouldHave = totalGenerated + totalSkipped
	}

	successRate := 100.0
	if totalShouldHave > 0 {
		successRate = float64(totalGenerated) / float64(totalShouldHave) * 100.0
	}

	now := time.Now()
	todayDateOnly := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	status := "HEALTHY"
	statusColor := "success"
	statusIcon := "✅"
	message := "✅ Semua invoice berhasil di-generate"

	if totalSkipped == 0 {
		status, statusColor, statusIcon = "HEALTHY", "success", "✅"
		message = "✅ Semua invoice berhasil di-generate"
	} else if todayDateOnly.Before(generationDate) || (todayDateOnly.Equal(generationDate) && now.Hour() < generationHour) {
		status, statusColor, statusIcon = "UPCOMING", "info", "🕒"
		totalGenerated = 0
		totalSkipped = 0
		successRate = 0.0

		monthsID := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
		genMonthName := monthsID[generationDate.Month()-1]
		genDateStr := fmt.Sprintf("%d %s %d", generationDate.Day(), genMonthName, generationDate.Year())

		if todayDateOnly.Equal(generationDate) {
			message = fmt.Sprintf("🕒 Menunggu jadwal otomatis hari ini jam %d:00 WIB", generationHour)
		} else {
			message = fmt.Sprintf("🕒 Menunggu jadwal generate otomatis pada %s (H-5)", genDateStr)
		}
	} else if totalSkipped <= 5 {
		status, statusColor, statusIcon = "NEEDS_ATTENTION", "warning", "⚠️"
		message = fmt.Sprintf("⚠️ %d pelanggan terlewat", totalSkipped)
	} else {
		status, statusColor, statusIcon = "CRITICAL", "error", "🔴"
		message = fmt.Sprintf("🔴 %d pelanggan terlewat", totalSkipped)
	}

	return &domain.InvoiceGenerationMonitorResponse{
		TargetDate:      targetDateObj.Format("2006-01-02"),
		TotalShouldHave: int(totalShouldHave),
		TotalGenerated:  int(totalGenerated),
		TotalSkipped:    int(totalSkipped),
		SuccessRate:     successRate,
		Status:          status,
		StatusColor:     statusColor,
		StatusIcon:      statusIcon,
		Message:         message,
		DetailURL:       fmt.Sprintf("/invoices/skipped-invoice-generation?target_date=%s", targetDateObj.Format("2006-01-02")),
	}, nil
}

func (r *dashboardRepository) GetFutureInvoiceProjection(ctx context.Context, targetDate string) (*domain.FutureInvoiceProjectionResponse, error) {
	var targetDateObj time.Time
	var err error
	today := time.Now()
	todayOnly := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	if targetDate != "" {
		targetDateObj, err = time.Parse("2006-01-02", targetDate)
		if err != nil {
			return nil, fmt.Errorf("invalid target_date format: %w", err)
		}
	} else {
		m := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
		var mPlus1 time.Time
		if m.Month() == 12 {
			mPlus1 = time.Date(m.Year()+1, 1, 1, 0, 0, 0, 0, today.Location())
		} else {
			mPlus1 = time.Date(m.Year(), m.Month()+1, 1, 0, 0, 0, 0, today.Location())
		}

		genDateMPlus1 := mPlus1.AddDate(0, 0, -5)
		if todayOnly.After(genDateMPlus1) || todayOnly.Equal(genDateMPlus1) {
			if mPlus1.Month() == 12 {
				targetDateObj = time.Date(mPlus1.Year()+1, 1, 1, 0, 0, 0, 0, today.Location())
			} else {
				targetDateObj = time.Date(mPlus1.Year(), mPlus1.Month()+1, 1, 0, 0, 0, 0, today.Location())
			}
		} else {
			targetDateObj = mPlus1
		}
	}

	daysUntil := int(targetDateObj.Sub(todayOnly).Hours() / 24)
	if daysUntil < 0 {
		daysUntil = 0
	}

	var estimatedCustomers int64
	err = r.db.WithContext(ctx).Table("langganan").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("DATE(COALESCE(langganan.tgl_jatuh_tempo_pembayaran, langganan.tgl_jatuh_tempo)) = ? AND langganan.status = ?", targetDateObj.Format("2006-01-02"), "Aktif").
		Count(&estimatedCustomers).Error
	if err != nil {
		return nil, err
	}

	var totalActive int64
	err = r.db.WithContext(ctx).Table("langganan").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("langganan.status = ?", "Aktif").
		Count(&totalActive).Error
	if err != nil {
		return nil, err
	}

	projectionDate := targetDateObj.AddDate(0, 0, -5)
	generationDaysUntil := int(projectionDate.Sub(todayOnly).Hours() / 24)
	if generationDaysUntil < 0 {
		generationDaysUntil = 0
	}

	systemStatus := "Persiapan"
	now := time.Now()
	generationHour := 13

	if todayOnly.After(projectionDate) || (todayOnly.Equal(projectionDate) && now.Hour() >= generationHour) {
		targetYear, targetMonth := targetDateObj.Year(), targetDateObj.Month()
		startOfMonth := time.Date(targetYear, targetMonth, 1, 0, 0, 0, 0, targetDateObj.Location())
		var endOfMonth time.Time
		if targetMonth == 12 {
			endOfMonth = time.Date(targetYear+1, 1, 1, 0, 0, 0, 0, targetDateObj.Location()).Add(-24 * time.Hour)
		} else {
			endOfMonth = time.Date(targetYear, targetMonth+1, 1, 0, 0, 0, 0, targetDateObj.Location()).Add(-24 * time.Hour)
		}

		var totalGenerated int64
		err = r.db.WithContext(ctx).Table("invoices").
			Select("COUNT(DISTINCT pelanggan_id)").
			Where("tgl_jatuh_tempo BETWEEN ? AND ?", startOfMonth, endOfMonth).
			Where("deleted_at IS NULL").
			Row().Scan(&totalGenerated)
		if err != nil {
			totalGenerated = 0
		}

		if totalGenerated > 0 {
			if float64(totalGenerated) >= (float64(estimatedCustomers) * 0.9) {
				systemStatus = "Selesai"
			} else {
				systemStatus = "Sebagian Selesai"
			}
		} else {
			systemStatus = "Terlewat"
		}
	} else if todayOnly.Equal(projectionDate) && now.Hour() < generationHour {
		systemStatus = "Menunggu Jadwal"
	} else {
		if daysUntil > 30 {
			systemStatus = "Siap"
		} else {
			systemStatus = "Persiapan"
		}
	}

	percentage := 0.0
	if totalActive > 0 {
		percentage = float64(estimatedCustomers) / float64(totalActive) * 100.0
	}

	return &domain.FutureInvoiceProjectionResponse{
		TargetDate:           targetDateObj.Format("2006-01-02"),
		EstimatedCustomers:   int(estimatedCustomers),
		TotalActiveCustomers: int(totalActive),
		DaysUntil:            daysUntil,
		GenerationDate:       projectionDate.Format("2006-01-02"),
		GenerationDaysUntil:  generationDaysUntil,
		SystemStatus:         systemStatus,
		IsFuture:             daysUntil > 0,
		PercentageOfActive:   percentage,
	}, nil
}

func (r *dashboardRepository) GetMainStats(ctx context.Context) (*domain.MainStatsData, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var endOfMonth time.Time
	if now.Month() == 12 {
		endOfMonth = time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, now.Location())
	} else {
		endOfMonth = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	}

	var pelangganAktif int64
	r.db.WithContext(ctx).Table("langganan").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("langganan.status = ?", "Aktif").
		Where("langganan.deleted_at IS NULL").
		Count(&pelangganAktif)

	var pelangganBaru int64
	r.db.WithContext(ctx).Table("langganan").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("langganan.status = ?", "Aktif").
		Where("langganan.tgl_mulai_langganan >= ?", startOfMonth).
		Where("langganan.deleted_at IS NULL").
		Count(&pelangganBaru)

	var pelangganBerhenti int64
	r.db.WithContext(ctx).Table("langganan").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("langganan.status = ?", "Berhenti").
		Where("langganan.tgl_berhenti >= ?", startOfMonth).
		Where("langganan.deleted_at IS NULL").
		Count(&pelangganBerhenti)

	var pelangganJakinet int64
	r.db.WithContext(ctx).Table("langganan").
		Joins("JOIN pelanggan ON langganan.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("langganan.status = ?", "Aktif").
		Where("pelanggan.id_brand = ?", "ajn-01").
		Where("langganan.deleted_at IS NULL").
		Count(&pelangganJakinet)

	var pendapatanJakinet float64
	r.db.WithContext(ctx).Table("invoices").
		Select("COALESCE(SUM(invoices.total_harga), 0)").
		Joins("JOIN pelanggan ON invoices.pelanggan_id = pelanggan.id AND pelanggan.deleted_at IS NULL").
		Where("invoices.status_invoice = ?", "Lunas").
		Where("invoices.paid_at >= ?", startOfMonth).
		Where("invoices.paid_at < ?", endOfMonth).
		Where("invoices.deleted_at IS NULL").
		Where("pelanggan.id_brand = ?", "ajn-01").
		Row().Scan(&pendapatanJakinet)

	return &domain.MainStatsData{
		PelangganAktif:            int(pelangganAktif),
		PelangganBaruBulanIni:     int(pelangganBaru),
		PelangganBerhentiBulanIni: int(pelangganBerhenti),
		PelangganJakiNetAktif:     int(pelangganJakinet),
		PendapatanJakiNetBulanIni: pendapatanJakinet,
	}, nil
}

func (r *dashboardRepository) GetGrowthChartData(ctx context.Context, months int) (*domain.ChartData, error) {
	if months <= 0 {
		months = 6
	}

	startDate := time.Now().AddDate(0, -months, 0)
	startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())

	type GrowthResult struct {
		Date  time.Time
		Count int
	}

	var results []GrowthResult

	r.db.WithContext(ctx).Table("langganan").
		Select("DATE(tgl_mulai_langganan) as date, COUNT(*) as count").
		Where("status = ?", "Aktif").
		Where("tgl_mulai_langganan >= ?", startDate).
		Where("deleted_at IS NULL").
		Group("DATE(tgl_mulai_langganan)").
		Order("date ASC").
		Scan(&results)

	var labels []string
	var data []int
	cumulativeCount := 0

	for _, row := range results {
		labels = append(labels, row.Date.Format("2006-01-02"))
		cumulativeCount += row.Count
		data = append(data, cumulativeCount)
	}

	return &domain.ChartData{
		Labels: labels,
		Data:   data,
	}, nil
}

func (r *dashboardRepository) GetRevenueChartData(ctx context.Context, months int) (*domain.ChartData, error) {
	if months <= 0 {
		months = 6
	}

	startDate := time.Now().AddDate(0, -months, 0)
	startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())

	type RevenueResult struct {
		Year  int
		Month int
		Total float64
	}

	var results []RevenueResult

	r.db.WithContext(ctx).Table("invoices").
		Select("YEAR(paid_at) as year, MONTH(paid_at) as month, COALESCE(SUM(total_harga), 0) as total").
		Where("status_invoice = ?", "Lunas").
		Where("paid_at >= ?", startDate).
		Where("deleted_at IS NULL").
		Group("YEAR(paid_at), MONTH(paid_at)").
		Order("year ASC, month ASC").
		Scan(&results)

	monthNames := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

	var labels []string
	var data []int

	for _, row := range results {
		if row.Month >= 1 && row.Month <= 12 {
			labels = append(labels, monthNames[row.Month-1])
			data = append(data, int(row.Total))
		}
	}

	return &domain.ChartData{
		Labels: labels,
		Data:   data,
	}, nil
}
