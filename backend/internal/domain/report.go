package domain

import (
	"time"
)

type BillStat struct {
	Count       int     `json:"count"`
	Nominal     float64 `json:"nominal"`
	Diskon      float64 `json:"diskon"`
	BiayaPasang float64 `json:"biaya_pasang"`
	Total       float64 `json:"total"`
}

type TaxStat struct {
	Ppn        float64 `json:"ppn"`
	Bhp        float64 `json:"bhp"`
	Uso        float64 `json:"uso"`
	TotalPajak float64 `json:"total_pajak"`
}

type PaymentMethodStat struct {
	Method      string  `json:"method"`
	Count       int     `json:"count"`
	TotalAmount float64 `json:"total_amount"`
	Pajak       float64 `json:"pajak"`
	Diskon      float64 `json:"diskon"`
}

type RevenueReportResponse struct {
	TotalPendapatan  float64 `json:"total_pendapatan"`
	TotalInvoices    int     `json:"total_invoices"`
	FinancialSummary struct {
		TotalPemasukan   float64 `json:"total_pemasukan"`
		TotalPengeluaran float64 `json:"total_pengeluaran"`
		SaldoAkhir      float64 `json:"saldo_akhir"`
	} `json:"financial_summary"`
	BillingSummary struct {
		TotalTagihan BillStat `json:"total_tagihan"`
		Lunas       BillStat `json:"lunas"`
		Pending     BillStat `json:"pending"`
		Expired     BillStat `json:"expired"`
	} `json:"billing_summary"`
	TaxSummary struct {
		Lunas   TaxStat `json:"lunas"`
		Pending TaxStat `json:"pending"`
		Expired TaxStat `json:"expired"`
		Total   TaxStat `json:"total"`
	} `json:"tax_summary"`
	PaymentMethods []PaymentMethodStat `json:"payment_methods"`
}

type InvoiceReportItem struct {
	ID            uint64    `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	PelangganNama string    `json:"pelanggan_nama"`
	Alamat        string    `json:"alamat"`
	TotalHarga    float64   `json:"total_harga"`
	StatusInvoice string    `json:"status_invoice"`
	TglInvoice    time.Time `json:"tgl_invoice"`
	TglLunas      *time.Time `json:"tgl_lunas"`
	Metode        string    `json:"metode"`
	Brand         string    `json:"brand"`
}

type RevenueReportParams struct {
	StartDate string
	EndDate   string
	Alamat    string
	IDBrand   string
	Limit     int
	Skip      int
}

// MLCustomerRecord represents customer payment behavioral features for ML
type MLCustomerRecord struct {
	CustomerID             uint64  `json:"customer_id"`
	CustomerName           string  `json:"customer_name"`
	NoTelp                 string  `json:"no_telp"`
	Brand                  string  `json:"brand"`
	MonthlyBill            float64 `json:"monthly_bill"`
	TotalInvoices          int     `json:"total_invoices"`
	PaidInvoices           int     `json:"paid_invoices"`
	ExpiredInvoices        int     `json:"expired_invoices"`
	AvgPaymentDayOfMonth   float64 `json:"avg_payment_day_of_month"`
	DaysToPayAvg           float64 `json:"days_to_pay_avg"`
	OnTimeRatio            float64 `json:"on_time_ratio"`
	GracePeriodRatio       float64 `json:"grace_period_ratio"`
	LateRatio              float64 `json:"late_ratio"`
	RecentStatus           string  `json:"recent_status"`
}

// MLRevenueRequest payload sent to Python ML service
type MLRevenueRequest struct {
	CycleDate   string             `json:"cycle_date"`
	TargetMonth string             `json:"target_month"`
	Customers   []MLCustomerRecord `json:"customers"`
}

// MLClusterSummary represents aggregated statistics for each cluster
type MLClusterSummary struct {
	Key          string  `json:"key"`
	Name         string  `json:"name"`
	Count        int     `json:"count"`
	Percentage   float64 `json:"percentage"`
	TotalNominal float64 `json:"total_nominal"`
	Color        string  `json:"color"`
	RiskLevel    string  `json:"risk_level"`
}

// MLCashFlowPhase represents a payment collection phase
type MLCashFlowPhase struct {
	Phase           string  `json:"phase"`
	Description     string  `json:"description"`
	EstimatedAmount float64 `json:"estimated_amount"`
	Percentage      float64 `json:"percentage"`
	TargetDays      string  `json:"target_days"`
	StatusClass     string  `json:"status_class"`
}

// MLClassifiedCustomer represents each customer's ML classification result
type MLClassifiedCustomer struct {
	CustomerID      uint64  `json:"customer_id"`
	CustomerName    string  `json:"customer_name"`
	NoTelp          string  `json:"no_telp"`
	Brand           string  `json:"brand"`
	MonthlyBill     float64 `json:"monthly_bill"`
	ClusterKey      string  `json:"cluster_key"`
	ClusterName     string  `json:"cluster_name"`
	ClusterColor    string  `json:"cluster_color"`
	RiskScore       int     `json:"risk_score"`
	RiskLevel       string  `json:"risk_level"`
	AvgPaymentDay   float64 `json:"avg_payment_day"`
	DaysToPayAvg    float64 `json:"days_to_pay_avg"`
	PaidRatio       float64 `json:"paid_ratio"`
	Recommendation  string  `json:"recommendation"`
}

// MLRevenueInsightResponse represents the full response for ML Revenue Report
type MLRevenueInsightResponse struct {
	Status    string `json:"status"`
	CycleInfo struct {
		InvoiceIssuanceDate     string `json:"invoice_issuance_date"`
		DueDate                 string `json:"due_date"`
		GracePeriodCutoff       string `json:"grace_period_cutoff"`
		TotalActiveSubscribers  int    `json:"total_active_subscribers"`
	} `json:"cycle_info"`
	Summary struct {
		TotalCustomers          int     `json:"total_customers"`
		ProjectedBilling        float64 `json:"projected_billing"`
		ProjectedEarlyCollection float64 `json:"projected_early_collection"`
		ProjectedGraceCollection float64 `json:"projected_grace_collection"`
		ProjectedAtRisk         float64 `json:"projected_at_risk"`
		ProjectedRecoveryRate   float64 `json:"projected_recovery_rate"`
	} `json:"summary"`
	ClustersSummary    []MLClusterSummary     `json:"clusters_summary"`
	CashFlowForecast   []MLCashFlowPhase      `json:"cash_flow_forecast"`
	ClassifiedCustomers []MLClassifiedCustomer `json:"classified_customers"`
	ActionableInsights []string               `json:"actionable_insights"`
}
