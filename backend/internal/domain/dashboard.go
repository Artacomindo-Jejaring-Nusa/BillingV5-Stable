package domain

import (
	"context"
)

type BrandRevenueItem struct {
	Brand   string  `json:"brand"`
	Revenue float64 `json:"revenue"`
}

type RevenueSummary struct {
	Total     float64            `json:"total"`
	Periode   string             `json:"periode"`
	Breakdown []BrandRevenueItem `json:"breakdown"`
}

type StatCard struct {
	Title       string      `json:"title"`
	Value       interface{} `json:"value"` // can be string or int
	Description string      `json:"description"`
}

type ChartData struct {
	Labels []string `json:"labels"`
	Data   []int    `json:"data"`
}

type InvoiceSummary struct {
	Labels     []string `json:"labels"`
	Total      []int    `json:"total"`
	Lunas      []int    `json:"lunas"`
	Menunggu   []int    `json:"menunggu"`
	Kadaluarsa []int    `json:"kadaluarsa"`
	Otomatis   []int    `json:"otomatis"`
	Manual     []int    `json:"manual"`
	Reinvoice  []int    `json:"reinvoice"`
}

type DashboardData struct {
	RevenueSummary           *RevenueSummary `json:"revenue_summary"`
	StatCards                []StatCard      `json:"stat_cards"`
	LokasiChart              *ChartData      `json:"lokasi_chart"`
	PaketChart               *ChartData      `json:"paket_chart"`
	GrowthChart              *ChartData      `json:"growth_chart"`
	InvoiceSummaryChart      *InvoiceSummary `json:"invoice_summary_chart"`
	StatusLanggananChart     *ChartData      `json:"status_langganan_chart"`
	PelangganPerAlamatChart  *ChartData      `json:"pelanggan_per_alamat_chart"`
	LoyalitasPembayaranChart *ChartData      `json:"loyalitas_pembayaran_chart"`
}

type SidebarBadgeResponse struct {
	SuspendedCount     int `json:"suspended_count"`
	UnpaidInvoiceCount int `json:"unpaid_invoice_count"`
	StoppedCount       int `json:"stopped_count"`
	TotalInvoiceCount  int `json:"total_invoice_count"`
	OpenTicketsCount   int `json:"open_tickets_count"`
}

type BreakdownItem struct {
	Nama   string `json:"nama"`
	Jumlah int    `json:"jumlah"`
}

type PaketDetail struct {
	TotalPelanggan  int             `json:"total_pelanggan"`
	BreakdownLokasi []BreakdownItem `json:"breakdown_lokasi"`
	BreakdownBrand  []BreakdownItem `json:"breakdown_brand"`
}

type LoyalitasUserDetail struct {
	ID          uint64 `json:"id"`
	Nama        string `json:"nama"`
	IDPelanggan string `json:"id_pelanggan"`
	Alamat      string `json:"alamat"`
	NoTelp      string `json:"no_telp"`
}

type InvoiceGenerationMonitorResponse struct {
	TargetDate      string  `json:"target_date"`
	TotalShouldHave int     `json:"total_should_have"`
	TotalGenerated  int     `json:"total_generated"`
	TotalSkipped    int     `json:"total_skipped"`
	SuccessRate     float64 `json:"success_rate"`
	Status          string  `json:"status"`
	StatusColor     string  `json:"status_color"`
	StatusIcon      string  `json:"status_icon"`
	Message         string  `json:"message"`
	DetailURL       string  `json:"detail_url"`
}

type FutureInvoiceProjectionResponse struct {
	TargetDate           string  `json:"target_date"`
	EstimatedCustomers   int     `json:"estimated_customers"`
	TotalActiveCustomers int     `json:"total_active_customers"`
	DaysUntil            int     `json:"days_until"`
	GenerationDate       string  `json:"generation_date"`
	GenerationDaysUntil  int     `json:"generation_days_until"`
	SystemStatus         string  `json:"system_status"`
	IsFuture             bool    `json:"is_future"`
	PercentageOfActive   float64 `json:"percentage_of_active"`
}

type MainStatsData struct {
	PelangganAktif            int     `json:"pelanggan_aktif"`
	PelangganBaruBulanIni     int     `json:"pelanggan_baru_bulan_ini"`
	PelangganBerhentiBulanIni int     `json:"pelanggan_berhenti_bulan_ini"`
	PelangganJakiNetAktif     int     `json:"pelanggan_jakinet_aktif"`
	PendapatanJakiNetBulanIni float64 `json:"pendapatan_jakinet_bulan_ini"`
}

type DashboardPelangganResponse struct {
	MainStats    *MainStatsData `json:"main_stats"`
	GrowthChart  *ChartData     `json:"growth_chart"`
	RevenueChart *ChartData     `json:"revenue_chart"`
}

// DashboardRepository defines database queries for the dashboard.
type DashboardRepository interface {
	GetRevenueSummary(ctx context.Context) (*RevenueSummary, error)
	GetPelangganStatCards(ctx context.Context) ([]StatCard, error)
	GetLoyaltyChart(ctx context.Context) (*ChartData, error)
	GetLokasiChart(ctx context.Context) (*ChartData, error)
	GetPaketChart(ctx context.Context) (*ChartData, error)
	GetGrowthChart(ctx context.Context) (*ChartData, error)
	GetInvoiceSummaryChart(ctx context.Context) (*InvoiceSummary, error)
	GetStatusLanggananChart(ctx context.Context) (*ChartData, error)
	GetPelangganPerAlamatChart(ctx context.Context) (*ChartData, error)
	GetLoyaltyUsersBySegment(ctx context.Context, segmen string) ([]LoyalitasUserDetail, error)
	GetSidebarBadges(ctx context.Context) (*SidebarBadgeResponse, error)
	GetPaketDetails(ctx context.Context) (map[string]PaketDetail, error)
	GetInvoiceGenerationMonitor(ctx context.Context, targetDate string) (*InvoiceGenerationMonitorResponse, error)
	GetFutureInvoiceProjection(ctx context.Context, targetDate string) (*FutureInvoiceProjectionResponse, error)
	GetMainStats(ctx context.Context) (*MainStatsData, error)
	GetGrowthChartData(ctx context.Context, months int) (*ChartData, error)
	GetRevenueChartData(ctx context.Context, months int) (*ChartData, error)
}

// DashboardUsecase defines business logic for dashboard data.
type DashboardUsecase interface {
	GetDashboardData(ctx context.Context, userPermissions map[string]bool) (*DashboardData, error)
	GetLoyaltyUsersBySegment(ctx context.Context, segmen string) ([]LoyalitasUserDetail, error)
	GetSidebarBadges(ctx context.Context) (*SidebarBadgeResponse, error)
	GetPaketDetails(ctx context.Context) (map[string]PaketDetail, error)
	GetInvoiceGenerationMonitor(ctx context.Context, targetDate string, userRole string) (*InvoiceGenerationMonitorResponse, error)
	GetFutureInvoiceProjection(ctx context.Context, targetDate string, userRole string) (*FutureInvoiceProjectionResponse, error)
}
