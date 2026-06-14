package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"billing-backend/config"
	"billing-backend/internal/domain"
	"billing-backend/internal/websocket"
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

// Helper functions for Redis caching using Go generics
func getCache[T any](ctx context.Context, key string) (*T, bool) {
	client := websocket.GetRedisClient()
	if client == nil {
		return nil, false
	}
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}
	var data T
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, false
	}
	return &data, true
}

func setCache[T any](ctx context.Context, key string, data *T, ttl time.Duration) {
	client := websocket.GetRedisClient()
	if client == nil {
		return
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("[Redis Cache] Failed to marshal data for key %s: %v", key, err)
		return
	}
	err = client.Set(ctx, key, string(bytes), ttl).Err()
	if err != nil {
		log.Printf("[Redis Cache] Failed to set key %s: %v", key, err)
	}
}

func (u *dashboardUsecase) GetDashboardData(ctx context.Context, userPermissions map[string]bool) (*domain.DashboardData, error) {
	data := &domain.DashboardData{
		RevenueSummary:           nil,
		StatCards:                []domain.StatCard{},
		LokasiChart:              &domain.ChartData{Labels: []string{}, Data: []int{}},
		PaketChart:               &domain.ChartData{Labels: []string{}, Data: []int{}},
		GrowthChart:              &domain.ChartData{Labels: []string{}, Data: []int{}},
		InvoiceSummaryChart:      &domain.InvoiceSummary{Labels: []string{}, Total: []int{}, Lunas: []int{}, Menunggu: []int{}, Expired: []int{}},
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
			var res *domain.RevenueSummary
			var err error
			if cached, found := getCache[domain.RevenueSummary](ctx, "dashboard:cache:revenue_summary"); found {
				res = cached
			} else {
				res, err = u.repo.GetRevenueSummary(ctx)
				if err == nil && res != nil {
					setCache(ctx, "dashboard:cache:revenue_summary", res, 5*time.Minute)
				}
			}
			if err == nil && res != nil {
				mu.Lock()
				data.RevenueSummary = res
				mu.Unlock()
			}
		}()
	}

	// 2. Stat Cards
	if userPermissions["view_widget_statistik_pelanggan"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var cards []domain.StatCard
			var err error
			if cached, found := getCache[[]domain.StatCard](ctx, "dashboard:cache:stat_cards"); found {
				cards = *cached
			} else {
				cards, err = u.repo.GetPelangganStatCards(ctx)
				if err == nil {
					setCache(ctx, "dashboard:cache:stat_cards", &cards, 5*time.Minute)
				}
			}
			if err == nil {
				mu.Lock()
				data.StatCards = append(data.StatCards, cards...)
				mu.Unlock()
			}
		}()

		// Loyalty Chart
		wg.Add(1)
		go func() {
			defer wg.Done()
			var chart *domain.ChartData
			var err error
			if cached, found := getCache[domain.ChartData](ctx, "dashboard:cache:loyalty_chart"); found {
				chart = cached
			} else {
				chart, err = u.repo.GetLoyaltyChart(ctx)
				if err == nil && chart != nil {
					setCache(ctx, "dashboard:cache:loyalty_chart", chart, 5*time.Minute)
				}
			}
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
			var chart *domain.ChartData
			var err error
			if cached, found := getCache[domain.ChartData](ctx, "dashboard:cache:lokasi_chart"); found {
				chart = cached
			} else {
				chart, err = u.repo.GetLokasiChart(ctx)
				if err == nil && chart != nil {
					setCache(ctx, "dashboard:cache:lokasi_chart", chart, 10*time.Minute)
				}
			}
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
			var chart *domain.ChartData
			var err error
			if cached, found := getCache[domain.ChartData](ctx, "dashboard:cache:paket_chart"); found {
				chart = cached
			} else {
				chart, err = u.repo.GetPaketChart(ctx)
				if err == nil && chart != nil {
					setCache(ctx, "dashboard:cache:paket_chart", chart, 10*time.Minute)
				}
			}
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
			var chart *domain.ChartData
			var err error
			if cached, found := getCache[domain.ChartData](ctx, "dashboard:cache:growth_chart"); found {
				chart = cached
			} else {
				chart, err = u.repo.GetGrowthChart(ctx)
				if err == nil && chart != nil {
					setCache(ctx, "dashboard:cache:growth_chart", chart, 10*time.Minute)
				}
			}
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
		var chart *domain.InvoiceSummary
		var err error
		if cached, found := getCache[domain.InvoiceSummary](ctx, "dashboard:cache:invoice_summary"); found {
			chart = cached
		} else {
			chart, err = u.repo.GetInvoiceSummaryChart(ctx)
			if err == nil && chart != nil {
				setCache(ctx, "dashboard:cache:invoice_summary", chart, 5*time.Minute)
			}
		}
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
			var chart *domain.ChartData
			var err error
			if cached, found := getCache[domain.ChartData](ctx, "dashboard:cache:status_langganan"); found {
				chart = cached
			} else {
				chart, err = u.repo.GetStatusLanggananChart(ctx)
				if err == nil && chart != nil {
					setCache(ctx, "dashboard:cache:status_langganan", chart, 5*time.Minute)
				}
			}
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
			var chart *domain.ChartData
			var err error
			if cached, found := getCache[domain.ChartData](ctx, "dashboard:cache:pelanggan_alamat"); found {
				chart = cached
			} else {
				chart, err = u.repo.GetPelangganPerAlamatChart(ctx)
				if err == nil && chart != nil {
					setCache(ctx, "dashboard:cache:pelanggan_alamat", chart, 10*time.Minute)
				}
			}
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
	key := "dashboard:cache:loyalitas_users:" + segmen
	if cached, found := getCache[[]domain.LoyalitasUserDetail](ctx, key); found {
		return *cached, nil
	}
	res, err := u.repo.GetLoyaltyUsersBySegment(ctx, segmen)
	if err == nil {
		setCache(ctx, key, &res, 5*time.Minute)
	}
	return res, err
}

func (u *dashboardUsecase) GetSidebarBadges(ctx context.Context) (*domain.SidebarBadgeResponse, error) {
	key := "dashboard:cache:sidebar_badges"
	if cached, found := getCache[domain.SidebarBadgeResponse](ctx, key); found {
		return cached, nil
	}
	res, err := u.repo.GetSidebarBadges(ctx)
	if err == nil && res != nil {
		setCache(ctx, key, res, 5*time.Minute)
	}
	return res, err
}

func (u *dashboardUsecase) GetPaketDetails(ctx context.Context) (map[string]domain.PaketDetail, error) {
	key := "dashboard:cache:paket_details"
	if cached, found := getCache[map[string]domain.PaketDetail](ctx, key); found {
		return *cached, nil
	}
	res, err := u.repo.GetPaketDetails(ctx)
	if err == nil {
		setCache(ctx, key, &res, 10*time.Minute)
	}
	return res, err
}

func (u *dashboardUsecase) GetInvoiceGenerationMonitor(ctx context.Context, targetDate string, userRole string) (*domain.InvoiceGenerationMonitorResponse, error) {
	if !u.config.CanAccessWidget("invoice_generation_monitor", userRole) {
		return nil, errors.New("akses ditolak: anda tidak memiliki izin untuk mengakses widget ini")
	}
	key := "dashboard:cache:invoice_monitor:" + targetDate
	if cached, found := getCache[domain.InvoiceGenerationMonitorResponse](ctx, key); found {
		return cached, nil
	}
	res, err := u.repo.GetInvoiceGenerationMonitor(ctx, targetDate)
	if err == nil && res != nil {
		setCache(ctx, key, res, 5*time.Minute)
	}
	return res, err
}

func (u *dashboardUsecase) GetFutureInvoiceProjection(ctx context.Context, targetDate string, userRole string) (*domain.FutureInvoiceProjectionResponse, error) {
	if !u.config.CanAccessWidget("future_invoice_projection", userRole) {
		return nil, errors.New("akses ditolak: anda tidak memiliki izin untuk mengakses widget ini")
	}
	key := "dashboard:cache:future_projection:" + targetDate
	if cached, found := getCache[domain.FutureInvoiceProjectionResponse](ctx, key); found {
		return cached, nil
	}
	res, err := u.repo.GetFutureInvoiceProjection(ctx, targetDate)
	if err == nil && res != nil {
		setCache(ctx, key, res, 5*time.Minute)
	}
	return res, err
}
