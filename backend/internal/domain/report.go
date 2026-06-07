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
