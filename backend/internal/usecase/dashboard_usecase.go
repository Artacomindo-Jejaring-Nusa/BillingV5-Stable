package usecase

import (
	"context"
	"errors"
	"sync"

	"billing-backend/config"
	"billing-backend/internal/domain"
)

type dashboardUsecase struct {
	repo   domain.DashboardRepository
	config *config.Config
}

func NewDashboardUsecase(repo domain.DashboardRepository, cfg *config.Config) domain.DashboardUsecase {
	return &dashboardUsecase{
		repo:   repo,
		config: cfg,
	}
}

func (u *dashboardUsecase) GetDashboardData(ctx context.Context, userPermissions map[string]bool) (*domain.DashboardData, error) {
	data := &domain.DashboardData{
		RevenueSummary:           nil,
		StatCards:                []domain.StatCard{},
		LokasiChart:              &domain.ChartData{Labels: []string{}, Data: []int{}},
		PaketChart:               &domain.ChartData{Labels: []string{}, Data: []int{}},
		GrowthChart:              &domain.ChartData{Labels: []string{}, Data: []int{}},
		InvoiceSummaryChart:      &domain.InvoiceSummary{Labels: []string{}, Total: []int{}, Lunas: []int{}, Menunggu: []int{}, Kadaluarsa: []int{}},
		StatusLanggananChart:     &domain.ChartData{Labels: []string{}, Data: []int{}},
		PelangganPerAlamatChart:  &domain.ChartData{Labels: []string{}, Data: []int{}},
		LoyalitasPembayaranChart: &domain.ChartData{Labels: []string{}, Data: []int{}},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 1. Revenue Summary
	if userPermissions["view_widget_pendapatan_bulanan"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := u.repo.GetRevenueSummary(ctx)
			if err == nil && res != nil {
				mu.Lock()
				data.RevenueSummary = res
				mu.Unlock()
			}
		}()
	}

	// 2. Stat Cards & Loyalty Chart
	if userPermissions["view_widget_statistik_pelanggan"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cards, err := u.repo.GetPelangganStatCards(ctx)
			if err == nil {
				mu.Lock()
				data.StatCards = append(data.StatCards, cards...)
				mu.Unlock()
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			chart, err := u.repo.GetLoyaltyChart(ctx)
			if err == nil && chart != nil {
				mu.Lock()
				data.LoyalitasPembayaranChart = chart
				mu.Unlock()
			}
		}()
	}

	// 3. Lokasi Chart
	if userPermissions["view_widget_pelanggan_per_lokasi"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			chart, err := u.repo.GetLokasiChart(ctx)
			if err == nil && chart != nil {
				mu.Lock()
				data.LokasiChart = chart
				mu.Unlock()
			}
		}()
	}

	// 4. Paket Chart
	if userPermissions["view_widget_pelanggan_per_paket"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			chart, err := u.repo.GetPaketChart(ctx)
			if err == nil && chart != nil {
				mu.Lock()
				data.PaketChart = chart
				mu.Unlock()
			}
		}()
	}

	// 5. Growth Chart
	if userPermissions["view_widget_tren_pertumbuhan"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			chart, err := u.repo.GetGrowthChart(ctx)
			if err == nil && chart != nil {
				mu.Lock()
				data.GrowthChart = chart
				mu.Unlock()
			}
		}()
	}

	// 6. Invoice Summary (always fetched)
	wg.Add(1)
	go func() {
		defer wg.Done()
		chart, err := u.repo.GetInvoiceSummaryChart(ctx)
		if err == nil && chart != nil {
			mu.Lock()
			data.InvoiceSummaryChart = chart
			mu.Unlock()
		}
	}()

	// 7. Status Langganan
	if userPermissions["view_widget_status_langganan"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			chart, err := u.repo.GetStatusLanggananChart(ctx)
			if err == nil && chart != nil {
				mu.Lock()
				data.StatusLanggananChart = chart
				mu.Unlock()
			}
		}()
	}

	// 8. Pelanggan Per Alamat
	if userPermissions["view_widget_alamat_aktif"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			chart, err := u.repo.GetPelangganPerAlamatChart(ctx)
			if err == nil && chart != nil {
				mu.Lock()
				data.PelangganPerAlamatChart = chart
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return data, nil
}

func (u *dashboardUsecase) GetLoyaltyUsersBySegment(ctx context.Context, segmen string) ([]domain.LoyalitasUserDetail, error) {
	return u.repo.GetLoyaltyUsersBySegment(ctx, segmen)
}

func (u *dashboardUsecase) GetSidebarBadges(ctx context.Context) (*domain.SidebarBadgeResponse, error) {
	return u.repo.GetSidebarBadges(ctx)
}

func (u *dashboardUsecase) GetPaketDetails(ctx context.Context) (map[string]domain.PaketDetail, error) {
	return u.repo.GetPaketDetails(ctx)
}

func (u *dashboardUsecase) GetInvoiceGenerationMonitor(ctx context.Context, targetDate string, userRole string) (*domain.InvoiceGenerationMonitorResponse, error) {
	if !u.config.CanAccessWidget("invoice_generation_monitor", userRole) {
		return nil, errors.New("akses ditolak: anda tidak memiliki izin untuk mengakses widget ini")
	}
	return u.repo.GetInvoiceGenerationMonitor(ctx, targetDate)
}

func (u *dashboardUsecase) GetFutureInvoiceProjection(ctx context.Context, targetDate string, userRole string) (*domain.FutureInvoiceProjectionResponse, error) {
	if !u.config.CanAccessWidget("future_invoice_projection", userRole) {
		return nil, errors.New("akses ditolak: anda tidak memiliki izin untuk mengakses widget ini")
	}
	return u.repo.GetFutureInvoiceProjection(ctx, targetDate)
}
