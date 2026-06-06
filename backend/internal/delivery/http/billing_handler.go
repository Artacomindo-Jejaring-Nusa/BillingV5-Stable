package http

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"billing-backend/internal/domain"

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
	r.POST("/invoices/xendit-callback", handler.ProcessXenditCallback)

	invoiceGroup := r.Group("/invoices")
	invoiceGroup.Use(authMiddleware)
	{
		invoiceGroup.GET("", handler.FetchInvoices)
		invoiceGroup.GET("/summary", handler.GetInvoiceSummary)
		invoiceGroup.GET("/summary/", handler.GetInvoiceSummary)
		invoiceGroup.GET("/:id", handler.GetInvoice)
		invoiceGroup.POST("", handler.CreateInvoice)
		invoiceGroup.POST("/generate", handler.GenerateManualInvoice)
		invoiceGroup.PATCH("/:id/status", handler.UpdateInvoiceStatus)
	}

	langgananGroup := r.Group("/langganan")
	langgananGroup.Use(authMiddleware)
	{
		langgananGroup.GET("", handler.FetchLangganan)
		langgananGroup.GET("/new-users", handler.GetNewUserLangganans)
		langgananGroup.GET("/:id", handler.GetLangganan)
		langgananGroup.POST("", handler.CreateLangganan)
		langgananGroup.PUT("/:id", handler.UpdateLangganan)
		langgananGroup.PATCH("/:id", handler.UpdateLangganan)
		langgananGroup.DELETE("/:id", handler.DeleteLangganan)
		langgananGroup.POST("/calculate-price", handler.CalculatePrice)
		langgananGroup.POST("/calculate-prorate-plus-full", handler.CalculateProratePlusFull)
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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	invoices, total, err := h.billingUsecase.FetchInvoices(c.Request.Context(), page, pageSize)
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

// Langganan Endpoints

func (h *BillingHandler) FetchLangganan(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitVal, err := strconv.Atoi(limitStr); err == nil {
			pageSize = limitVal
		}
	}
	search := c.Query("search")
	status := c.Query("status")
	forInvoiceSelection := c.Query("for_invoice_selection") == "true"

	langganans, total, err := h.billingUsecase.FetchLangganan(c.Request.Context(), page, pageSize, search, status, forInvoiceSelection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert Langganan to LanggananResponse
	responses := make([]domain.LanggananResponse, len(langganans))
	for i, lang := range langganans {
		resp := domain.LanggananResponse{
			ID:                 lang.ID,
			PelangganID:        lang.PelangganID,
			PaketLayananID:     lang.PaketLayananID,
			Status:             lang.Status,
			TglJatuhTempo:      lang.TglJatuhTempo,
			TglInvoiceTerakhir: lang.TglInvoiceTerakhir,
			TglMulaiLangganan:  lang.TglMulaiLangganan,
			TglBerhenti:        lang.TglBerhenti,
			MetodePembayaran:   lang.MetodePembayaran,
			HargaAwal:          lang.HargaAwal,
			AlasanBerhenti:     lang.AlasanBerhenti,
			StatusModem:        lang.StatusModem,
			WhatsappStatus:     lang.WhatsappStatus,
			LastWhatsappSent:   lang.LastWhatsappSent,
			CreatedAt:          lang.CreatedAt,
			UpdatedAt:          lang.UpdatedAt,
		}

		// Add Pelanggan fields
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

		// Add PaketLayanan fields
		if lang.PaketLayanan != nil {
			resp.NamaPaket = lang.PaketLayanan.NamaPaket
			resp.Harga = lang.PaketLayanan.Harga
			resp.HargaFinal = lang.PaketLayanan.Harga

			// Add Brand from HargaLayanan
			if lang.PaketLayanan.HargaLayanan != nil {
				resp.Brand = lang.PaketLayanan.HargaLayanan.Brand
				resp.HargaFinal = math.Round(lang.PaketLayanan.Harga * (1.0 + lang.PaketLayanan.HargaLayanan.Pajak/100.0))
			}
		}

		responses[i] = resp
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

	// Convert Langganan to LanggananResponse
	responses := make([]domain.LanggananResponse, len(langganans))
	for i, lang := range langganans {
		resp := domain.LanggananResponse{
			ID:                 lang.ID,
			PelangganID:        lang.PelangganID,
			PaketLayananID:     lang.PaketLayananID,
			Status:             lang.Status,
			TglJatuhTempo:      lang.TglJatuhTempo,
			TglInvoiceTerakhir: lang.TglInvoiceTerakhir,
			TglMulaiLangganan:  lang.TglMulaiLangganan,
			TglBerhenti:        lang.TglBerhenti,
			MetodePembayaran:   lang.MetodePembayaran,
			HargaAwal:          lang.HargaAwal,
			AlasanBerhenti:     lang.AlasanBerhenti,
			StatusModem:        lang.StatusModem,
			WhatsappStatus:     lang.WhatsappStatus,
			LastWhatsappSent:   lang.LastWhatsappSent,
			CreatedAt:          lang.CreatedAt,
			UpdatedAt:          lang.UpdatedAt,
		}

		// Add Pelanggan fields
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

		// Add PaketLayanan fields
		if lang.PaketLayanan != nil {
			resp.NamaPaket = lang.PaketLayanan.NamaPaket
			resp.Harga = lang.PaketLayanan.Harga
			resp.HargaFinal = lang.PaketLayanan.Harga

			// Add Brand from HargaLayanan
			if lang.PaketLayanan.HargaLayanan != nil {
				resp.Brand = lang.PaketLayanan.HargaLayanan.Brand
				resp.HargaFinal = math.Round(lang.PaketLayanan.Harga * (1.0 + lang.PaketLayanan.HargaLayanan.Pajak/100.0))
			}
		}

		responses[i] = resp
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        responses,
		"total_count": len(responses),
	})
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

	c.JSON(http.StatusOK, langganan)
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

	c.JSON(http.StatusCreated, created)
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

	c.JSON(http.StatusOK, updated)
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
