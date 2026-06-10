package domain

import (
	"context"
	"time"
)

// InvoiceRepository defines database operations for Invoice
type InvoiceRepository interface {
	GetAll(ctx context.Context, limit, offset int, search, status string) ([]Invoice, int64, error)
	GetByID(ctx context.Context, id uint64) (*Invoice, error)
	GetByInvoiceNumber(ctx context.Context, invNumber string) (*Invoice, error)
	Create(ctx context.Context, invoice *Invoice) error
	Update(ctx context.Context, invoice *Invoice) error
	
	// Payment Webhooks & Callbacks
	GetCallbackLog(ctx context.Context, xenditID, externalID, idempotencyKey string) (*PaymentCallbackLog, error)
	CreateCallbackLog(ctx context.Context, log *PaymentCallbackLog) error
	GetInvoiceWithRelations(ctx context.Context, externalID string) (*Invoice, error)

	// Helper queries for cron jobs
	GetUnpaidByExternalIDs(ctx context.Context, externalIDs []string) ([]Invoice, error)
	HasPaidInvoiceForPeriod(ctx context.Context, pelangganID uint64, targetDueDate, endOfPrevMonth time.Time) (bool, error)
	UpdateStatusForUnpaidInvoices(ctx context.Context, pelangganID uint64, targetDueDate, endOfPrevMonth time.Time, newStatus string) (int64, error)
	GetInvoiceByPelangganAndDueDateRange(ctx context.Context, pelangganID uint64, start, end time.Time) (*Invoice, error)
	GetInvoiceSummary(ctx context.Context) (*InvoiceSummaryStats, error)
	GetRevenueReport(ctx context.Context, params *RevenueReportParams) (*RevenueReportResponse, error)
	GetRevenueReportDetails(ctx context.Context, params *RevenueReportParams) ([]InvoiceReportItem, error)
	ExportPaymentLinksExcel(ctx context.Context, filters map[string]string) ([]byte, error)
}

// LanggananRepository defines database operations for Langganan
type LanggananRepository interface {
	GetAll(ctx context.Context, limit, offset int, search, status string, forInvoiceSelection bool) ([]Langganan, int64, error)
	GetByID(ctx context.Context, id uint64) (*Langganan, error)
	Create(ctx context.Context, langganan *Langganan) error
	Update(ctx context.Context, langganan *Langganan) error
	Delete(ctx context.Context, id uint64) error

	// Helper queries for cron jobs
	GetActiveByDueDateRange(ctx context.Context, start, end time.Time) ([]Langganan, error)
	GetActiveOverdueForSuspend(ctx context.Context, targetDueDate, endOfPrevMonth time.Time) ([]Langganan, error)
	GetNewUserLangganans(ctx context.Context) ([]Langganan, error)
}

// BillingUsecase defines business logic for Invoices and Subscriptions
type BillingUsecase interface {
	// Invoice
	FetchInvoices(ctx context.Context, page, pageSize int, search, status string) ([]Invoice, int64, error)
	GetInvoice(ctx context.Context, id uint64) (*Invoice, error)
	CreateInvoice(ctx context.Context, invoice *Invoice) error
	UpdateInvoiceStatus(ctx context.Context, id uint64, status string) error
	GetInvoiceSummary(ctx context.Context) (*InvoiceSummaryStats, error)
	GenerateManualInvoice(ctx context.Context, langgananID uint64) (*Invoice, error)
	
	// Xendit Webhook Callback
	ProcessXenditCallback(ctx context.Context, xCallbackToken string, payload map[string]interface{}, idempotencyKey string) error

	// Langganan
	FetchLangganan(ctx context.Context, page, pageSize int, search, status string, forInvoiceSelection bool) ([]Langganan, int64, error)
	GetNewUserLangganans(ctx context.Context) ([]Langganan, error)
	GetLangganan(ctx context.Context, id uint64) (*Langganan, error)
	CreateLangganan(ctx context.Context, langganan *Langganan) error
	UpdateLangganan(ctx context.Context, id uint64, langganan *Langganan) error
	DeleteLangganan(ctx context.Context, id uint64) error

	// Calculations
	CalculatePrice(ctx context.Context, req *LanggananCalculateRequest) (*LanggananCalculateResponse, error)
	CalculateProratePlusFull(ctx context.Context, req *LanggananCalculateRequest) (*LanggananCalculateProratePlusFullResponse, error)
	CalculateProrate(ctx context.Context, req *ProrateCalculationRequest) (*ProrateCalculationResponse, error)
	CalculateDiskon(ctx context.Context, req *DiskonCalculationRequest) (*DiskonCalculationResponse, error)
	GetDiscountedPrice(ctx context.Context, cluster string, originalPrice float64) float64

	// Cron Jobs
	GenerateInvoices(ctx context.Context) error
	AutoSuspend(ctx context.Context) error
	VerifyPayments(ctx context.Context) error

	// Reports
	GetRevenueReport(ctx context.Context, params *RevenueReportParams) (*RevenueReportResponse, error)
	GetRevenueReportDetails(ctx context.Context, params *RevenueReportParams) ([]InvoiceReportItem, error)

	// Portability
	ExportLangganan(ctx context.Context, format string) ([]byte, string, error)
	ExportLanggananMultiSheet(ctx context.Context) ([]byte, string, error)
	ImportLanggananFromCSV(ctx context.Context, csvContent string) (int, error)
	ExportInvoices(ctx context.Context, format string) ([]byte, string, error)
	ExportPaymentLinksExcel(ctx context.Context, filters map[string]string) ([]byte, error)
	ArchiveOldInvoices(ctx context.Context) error
}



