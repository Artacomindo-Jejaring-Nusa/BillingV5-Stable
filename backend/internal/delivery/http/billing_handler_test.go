package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"billing-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type mockBillingUsecase struct {
	domain.BillingUsecase
}

func (m *mockBillingUsecase) FetchLangganan(ctx context.Context, page, pageSize int, search, status string, forInvoiceSelection bool) ([]domain.Langganan, int64, error) {
	return []domain.Langganan{}, 0, nil
}

func (m *mockBillingUsecase) ExportLangganan(ctx context.Context, format string) ([]byte, string, error) {
	return []byte("csv data"), "text/csv", nil
}

func (m *mockBillingUsecase) FetchInvoices(ctx context.Context, page, pageSize int) ([]domain.Invoice, int64, error) {
	return []domain.Invoice{}, 0, nil
}

func (m *mockBillingUsecase) GetInvoiceSummary(ctx context.Context) (*domain.InvoiceSummaryStats, error) {
	return &domain.InvoiceSummaryStats{}, nil
}

func (m *mockBillingUsecase) ExportInvoices(ctx context.Context, format string) ([]byte, string, error) {
	return []byte("csv data"), "text/csv", nil
}

func (m *mockBillingUsecase) ExportPaymentLinksExcel(ctx context.Context, filters map[string]string) ([]byte, error) {
	return []byte("excel data"), nil
}

func (m *mockBillingUsecase) ExportLanggananMultiSheet(ctx context.Context) ([]byte, string, error) {
	return []byte("excel data"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
}

func (m *mockBillingUsecase) GetRevenueReport(ctx context.Context, params *domain.RevenueReportParams) (*domain.RevenueReportResponse, error) {
	return &domain.RevenueReportResponse{}, nil
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	return r
}

func TestBillingHandler_FetchInvoices(t *testing.T) {
	router := setupRouter()
	mockUsecase := &mockBillingUsecase{}
	authMiddleware := func(c *gin.Context) { c.Next() }
	NewBillingHandler(router.Group("/api/v1"), mockUsecase, authMiddleware)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/invoices", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBillingHandler_FetchLangganan(t *testing.T) {
	router := setupRouter()
	mockUsecase := &mockBillingUsecase{}
	authMiddleware := func(c *gin.Context) { c.Next() }
	NewBillingHandler(router.Group("/api/v1"), mockUsecase, authMiddleware)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/langganan", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBillingHandler_ExportInvoices(t *testing.T) {
	router := setupRouter()
	mockUsecase := &mockBillingUsecase{}
	authMiddleware := func(c *gin.Context) { c.Next() }
	NewBillingHandler(router.Group("/api/v1"), mockUsecase, authMiddleware)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/invoices/export?format=csv", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBillingHandler_GetRevenueReport(t *testing.T) {
	router := setupRouter()
	mockUsecase := &mockBillingUsecase{}
	authMiddleware := func(c *gin.Context) { c.Next() }
	NewBillingHandler(router.Group("/api/v1"), mockUsecase, authMiddleware)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/reports/revenue", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
