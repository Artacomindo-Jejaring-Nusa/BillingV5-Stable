package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"billing-backend/internal/domain"
	"billing-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type BillingHandler struct {
	billingUsecase domain.BillingUsecase
}

func NewBillingHandler(r *gin.RouterGroup, bu domain.BillingUsecase, authMiddleware gin.HandlerFunc) {
	handler := &BillingHandler{
		billingUsecase: bu,
	}

	// Public Xendit callback endpoint (not protected by authMiddleware)
	r.POST("/invoices/xendit-callback", middleware.RateLimitMiddleware(60), handler.ProcessXenditCallback)

	invoiceGroup := r.Group("/invoices")
	invoiceGroup.Use(authMiddleware)
	{
		invoiceGroup.GET("", middleware.PermissionMiddleware("view_invoices"), handler.FetchInvoices)
		invoiceGroup.GET("/export", middleware.PermissionMiddleware("view_invoices"), handler.ExportInvoices)
		invoiceGroup.GET("/export-payment-links-excel", middleware.PermissionMiddleware("view_invoices"), handler.ExportPaymentLinksExcel)
		invoiceGroup.GET("/summary", middleware.PermissionMiddleware("view_invoices"), handler.GetInvoiceSummary)
		invoiceGroup.GET("/summary/", middleware.PermissionMiddleware("view_invoices"), handler.GetInvoiceSummary)
		invoiceGroup.GET("/:id", middleware.PermissionMiddleware("view_invoices"), handler.GetInvoice)
		invoiceGroup.POST("", middleware.PermissionMiddleware("create_invoices"), handler.CreateInvoice)
		invoiceGroup.POST("/generate", middleware.PermissionMiddleware("create_invoices"), handler.GenerateManualInvoice)
		invoiceGroup.PATCH("/:id/status", middleware.PermissionMiddleware("edit_invoices"), handler.UpdateInvoiceStatus)
		invoiceGroup.DELETE("/:id", middleware.PermissionMiddleware("delete_invoices"), handler.DeleteInvoice)
	}

	langgananGroup := r.Group("/langganan")
	langgananGroup.Use(authMiddleware)
	{
		langgananGroup.GET("", middleware.PermissionMiddleware("view_langganan"), handler.FetchLangganan)
		langgananGroup.GET("/new-users", middleware.PermissionMiddleware("view_langganan"), handler.GetNewUserLangganans)
		langgananGroup.GET("/:id", middleware.PermissionMiddleware("view_langganan"), handler.GetLangganan)
		langgananGroup.POST("", middleware.PermissionMiddleware("create_langganan"), handler.CreateLangganan)
		langgananGroup.PUT("/:id", middleware.PermissionMiddleware("edit_langganan"), handler.UpdateLangganan)
		langgananGroup.PATCH("/:id", middleware.PermissionMiddleware("edit_langganan"), handler.UpdateLangganan)
		langgananGroup.DELETE("/:id", middleware.PermissionMiddleware("delete_langganan"), handler.DeleteLangganan)
		langgananGroup.POST("/calculate-price", middleware.PermissionMiddleware("view_langganan"), handler.CalculatePrice)
		langgananGroup.POST("/calculate-prorate-plus-full", middleware.PermissionMiddleware("view_langganan"), handler.CalculateProratePlusFull)
		
		langgananGroup.GET("/export", middleware.PermissionMiddleware("view_langganan"), handler.ExportLangganan)
		langgananGroup.GET("/export/excel/multi-sheet", middleware.PermissionMiddleware("view_langganan"), handler.ExportLanggananMultiSheet)
		langgananGroup.POST("/import/csv", middleware.PermissionMiddleware("create_langganan"), handler.ImportLangganan)
		langgananGroup.GET("/template/csv", middleware.PermissionMiddleware("create_langganan"), handler.DownloadLanggananTemplate)
	}

	reportsGroup := r.Group("/reports")
	reportsGroup.Use(authMiddleware)
	{
		reportsGroup.GET("/revenue", handler.GetRevenueReport)
		reportsGroup.GET("/revenue/details", handler.GetRevenueReportDetails)
	}

	calculatorGroup := r.Group("/calculator")
	calculatorGroup.Use(authMiddleware)
	{
		calculatorGroup.POST("/prorate", handler.CalculateProrate)
		calculatorGroup.POST("/diskon", handler.CalculateDiskon)
	}
}

// Invoice Endpoints

func (h *BillingHandler) FetchInvoices(c *gin.Context) {
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	search := c.Query("search")
	status := c.Query("status_invoice")

	// Calculate page for usecase (internal logic still uses page/pageSize)
	// Or better: update usecase to use skip/limit too?
	// For now, let's just map skip to page
	page := (skip / limit) + 1

	invoices, total, err := h.billingUsecase.FetchInvoices(c.Request.Context(), page, limit, search, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  invoices,
		"total": total,
		"page":  page,
	})
}

func (h *BillingHandler) GetInvoiceSummary(c *gin.Context) {
	summary, err := h.billingUsecase.GetInvoiceSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *BillingHandler) GetInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	invoice, err := h.billingUsecase.GetInvoice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invoice})
}

func (h *BillingHandler) CreateInvoice(c *gin.Context) {
	var invoice domain.Invoice
	if err := c.ShouldBindJSON(&invoice); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.billingUsecase.CreateInvoice(c.Request.Context(), &invoice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": invoice})
}

func (h *BillingHandler) UpdateInvoiceStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.billingUsecase.UpdateInvoiceStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice status updated successfully"})
}

func (h *BillingHandler) GenerateManualInvoice(c *gin.Context) {
	var req struct {
		LanggananID uint64 `json:"langganan_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.billingUsecase.GenerateManualInvoice(c.Request.Context(), req.LanggananID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

func (h *BillingHandler) DeleteInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.billingUsecase.DeleteInvoice(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice berhasil dihapus"})
}

// Langganan Endpoints

func (h *BillingHandler) FetchLangganan(c *gin.Context) {
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	search := c.Query("search")
	status := c.Query("status")
	forInvoiceSelection := c.Query("for_invoice_selection") == "true"

	page := (skip / limit) + 1

	langganans, total, err := h.billingUsecase.FetchLangganan(c.Request.Context(), page, limit, search, status, forInvoiceSelection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]domain.LanggananResponse, len(langganans))
	for i, lang := range langganans {
		responses[i] = h.mapToLanggananResponse(c.Request.Context(), lang)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        responses,
		"total_count": total,
	})
}

func (h *BillingHandler) GetNewUserLangganans(c *gin.Context) {
	langganans, err := h.billingUsecase.GetNewUserLangganans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]domain.LanggananResponse, len(langganans))
	for i, lang := range langganans {
		responses[i] = h.mapToLanggananResponse(c.Request.Context(), lang)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        responses,
		"total_count": len(responses),
	})
}

func (h *BillingHandler) mapToLanggananResponse(ctx context.Context, lang domain.Langganan) domain.LanggananResponse {
	resp := domain.LanggananResponse{
		ID:                         lang.ID,
		PelangganID:                lang.PelangganID,
		PaketLayananID:             lang.PaketLayananID,
		Status:                     lang.Status,
		TglJatuhTempo:              lang.TglJatuhTempo,
		TglJatuhTempoPembayaran:    lang.TglJatuhTempoPembayaran,
		TglInvoiceTerakhir:         lang.TglInvoiceTerakhir,
		TglMulaiLangganan:          lang.TglMulaiLangganan,
		TglBerhenti:                lang.TglBerhenti,
		MetodePembayaran:           lang.MetodePembayaran,
		HargaAwal:                  lang.HargaAwal,
		AlasanBerhenti:             lang.AlasanBerhenti,
		StatusModem:                lang.StatusModem,
		WhatsappStatus:             lang.WhatsappStatus,
		LastWhatsappSent:           lang.LastWhatsappSent,
		CreatedAt:                  lang.CreatedAt,
		UpdatedAt:                  lang.UpdatedAt,
	}

	if lang.Pelanggan != nil {
		resp.NamaPelanggan = lang.Pelanggan.Nama
		resp.NoTelp = lang.Pelanggan.NoTelp
		resp.Alamat = lang.Pelanggan.Alamat
		if lang.Pelanggan.IDBrand != nil {
			resp.IDBrand = *lang.Pelanggan.IDBrand
		}
		resp.Pelanggan = &domain.LanggananPelangganResponse{
			ID:     lang.Pelanggan.ID,
			Nama:   lang.Pelanggan.Nama,
			NoTelp: lang.Pelanggan.NoTelp,
			Alamat: lang.Pelanggan.Alamat,
			Email:  lang.Pelanggan.Email,
		}
	}

	if lang.PaketLayanan != nil {
		resp.NamaPaket = lang.PaketLayanan.NamaPaket
		resp.Harga = lang.PaketLayanan.Harga

		if lang.HargaAwal != nil {
			resp.HargaFinal = h.billingUsecase.GetDiscountedPrice(ctx, resp.Alamat, *lang.HargaAwal)
		} else {
			basePrice := lang.PaketLayanan.Harga
			if lang.PaketLayanan.HargaLayanan != nil {
				basePrice = math.Round(lang.PaketLayanan.Harga * (1.0 + lang.PaketLayanan.HargaLayanan.Pajak/100.0))
			}
			resp.HargaFinal = h.billingUsecase.GetDiscountedPrice(ctx, resp.Alamat, basePrice)
		}

		if lang.PaketLayanan.HargaLayanan != nil {
			resp.Brand = lang.PaketLayanan.HargaLayanan.Brand
		}
	}
	return resp
}

func (h *BillingHandler) GetLangganan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	langganan, err := h.billingUsecase.GetLangganan(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	resp := h.mapToLanggananResponse(c.Request.Context(), *langganan)
	c.JSON(http.StatusOK, resp)
}

func (h *BillingHandler) CreateLangganan(c *gin.Context) {
	var langganan domain.Langganan
	if err := c.ShouldBindJSON(&langganan); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.billingUsecase.CreateLangganan(c.Request.Context(), &langganan); err != nil {
		if strings.Contains(err.Error(), "belum memiliki data teknis") {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "tidak ditemukan") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := h.billingUsecase.GetLangganan(c.Request.Context(), langganan.ID)
	if err != nil {
		c.JSON(http.StatusCreated, langganan)
		return
	}

	resp := h.mapToLanggananResponse(c.Request.Context(), *created)
	c.JSON(http.StatusCreated, resp)
}

func (h *BillingHandler) UpdateLangganan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var langganan domain.Langganan
	if err := c.ShouldBindJSON(&langganan); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.billingUsecase.UpdateLangganan(c.Request.Context(), id, &langganan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.billingUsecase.GetLangganan(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := h.mapToLanggananResponse(c.Request.Context(), *updated)
	c.JSON(http.StatusOK, resp)
}

func (h *BillingHandler) DeleteLangganan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.billingUsecase.DeleteLangganan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *BillingHandler) CalculatePrice(c *gin.Context) {
	var req domain.LanggananCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	res, err := h.billingUsecase.CalculatePrice(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *BillingHandler) CalculateProratePlusFull(c *gin.Context) {
	var req domain.LanggananCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	res, err := h.billingUsecase.CalculateProratePlusFull(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *BillingHandler) CalculateProrate(c *gin.Context) {
	var req domain.ProrateCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	res, err := h.billingUsecase.CalculateProrate(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *BillingHandler) CalculateDiskon(c *gin.Context) {
	var req domain.DiskonCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	res, err := h.billingUsecase.CalculateDiskon(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *BillingHandler) ProcessXenditCallback(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
		return
	}

	xCallbackToken := c.GetHeader("x-callback-token")
	idempotencyKey := c.GetHeader("x-idempotency-key")
	if idempotencyKey == "" {
		idempotencyKey = c.GetHeader("idempotency-key")
	}

	err := h.billingUsecase.ProcessXenditCallback(c.Request.Context(), xCallbackToken, payload, idempotencyKey)
	if err != nil {
		if strings.Contains(err.Error(), "invalid callback token") || strings.Contains(err.Error(), "invalid brand") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Callback processed successfully"})
}

func (h *BillingHandler) GetRevenueReport(c *gin.Context) {
	params := &domain.RevenueReportParams{
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Alamat:    c.Query("alamat"),
		IDBrand:   c.Query("id_brand"),
	}

	report, err := h.billingUsecase.GetRevenueReport(c.Request.Context(), params)
	if err != nil {
		// Log the actual error for debugging
		fmt.Printf("[REPORT ERROR] GetRevenueReport failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *BillingHandler) GetRevenueReportDetails(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))

	params := &domain.RevenueReportParams{
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Alamat:    c.Query("alamat"),
		IDBrand:   c.Query("id_brand"),
		Limit:     limit,
		Skip:      skip,
	}

	items, err := h.billingUsecase.GetRevenueReportDetails(c.Request.Context(), params)
	if err != nil {
		// Log the actual error for debugging
		fmt.Printf("[REPORT ERROR] GetRevenueReportDetails failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *BillingHandler) ExportLangganan(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	data, contentType, err := h.billingUsecase.ExportLangganan(c.Request.Context(), format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	filename := "export_langganan." + format
	if format == "excel" { filename = "export_langganan.xlsx" }
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

func (h *BillingHandler) ExportLanggananMultiSheet(c *gin.Context) {
	data, contentType, err := h.billingUsecase.ExportLanggananMultiSheet(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=export_langganan_komprehensif.xlsx")
	c.Data(http.StatusOK, contentType, data)
}

func (h *BillingHandler) ImportLangganan(c *gin.Context) {
	file, _ := c.FormFile("file")
	opened, _ := file.Open()
	defer opened.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, opened)
	count, err := h.billingUsecase.ImportLanggananFromCSV(c.Request.Context(), buf.String())
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Partial success", "error": err.Error(), "count": count})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Success", "count": count})
}

func (h *BillingHandler) DownloadLanggananTemplate(c *gin.Context) {
	buf := new(bytes.Buffer)
	buf.WriteString("Email Pelanggan;ID Paket\nbudi@example.com;1\n")
	c.Header("Content-Disposition", "attachment; filename=template_langganan.csv")
	c.Data(http.StatusOK, "text/csv", buf.Bytes())
}

func (h *BillingHandler) ExportInvoices(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	data, contentType, err := h.billingUsecase.ExportInvoices(c.Request.Context(), format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	filename := "export_invoices." + format
	if format == "excel" { filename = "export_invoices.xlsx" }
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

func (h *BillingHandler) ExportPaymentLinksExcel(c *gin.Context) {
	filters := map[string]string{
		"search":         c.Query("search"),
		"status_invoice": c.Query("status_invoice"),
		"start_date":     c.Query("start_date"),
		"end_date":       c.Query("end_date"),
	}

	data, err := h.billingUsecase.ExportPaymentLinksExcel(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=payment-links.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
