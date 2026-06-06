package http

import (
	"net/http"
	"strconv"
	"time"

	"billing-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type LayananHandler struct {
	hargaUsecase domain.HargaLayananUsecase
	paketUsecase domain.PaketLayananUsecase
	diskonUsecase domain.DiskonUsecase
}

func NewLayananHandler(
	r *gin.RouterGroup,
	hu domain.HargaLayananUsecase,
	pu domain.PaketLayananUsecase,
	du domain.DiskonUsecase,
	authMiddleware gin.HandlerFunc,
) {
	handler := &LayananHandler{
		hargaUsecase:  hu,
		paketUsecase:  pu,
		diskonUsecase: du,
	}

	// HargaLayanan (Brand) Routes
	harga := r.Group("/harga_layanan")
	harga.Use(authMiddleware)
	{
		harga.POST("", handler.CreateBrand)
		harga.GET("", handler.GetAllBrands)
		harga.GET("/:id_brand", handler.GetBrandByID)
		harga.PATCH("/:id_brand", handler.UpdateBrand)
		harga.DELETE("/:id_brand", handler.DeleteBrand)
	}

	// PaketLayanan Routes
	paket := r.Group("/paket_layanan")
	paket.Use(authMiddleware)
	{
		paket.POST("", handler.CreatePaket)
		paket.GET("", handler.GetAllPakets)
		paket.GET("/:id", handler.GetPaketByID)
		paket.PATCH("/:id", handler.UpdatePaket)
		paket.DELETE("/:id", handler.DeletePaket)
	}

	// Diskon Routes
	diskon := r.Group("/diskon")
	diskon.Use(authMiddleware)
	{
		diskon.POST("", handler.CreateDiskon)
		diskon.GET("", handler.GetAllDiskons)
		diskon.GET("/clusters/list", handler.GetClusterList) // Positioned above detail route to prevent conflict
		diskon.GET("/:id", handler.GetDiskonByID)
		diskon.PUT("/:id", handler.UpdateDiskon)
		diskon.DELETE("/:id", handler.DeleteDiskon)
		diskon.POST("/:id/activate", handler.ActivateDiskon)
		diskon.GET("/cluster/:cluster", handler.GetDiskonByCluster)
	}
}

// --- HargaLayanan Structs & Handlers ---

type CreateBrandRequest struct {
	IDBrand       string  `json:"id_brand" binding:"required"`
	Brand         string  `json:"brand" binding:"required"`
	Pajak         float64 `json:"pajak" binding:"required,gte=0"`
	XenditKeyName string  `json:"xendit_key_name"`
}

type UpdateBrandRequest struct {
	Brand         *string  `json:"brand"`
	Pajak         *float64 `json:"pajak" binding:"omitempty,gte=0"`
	XenditKeyName *string  `json:"xendit_key_name"`
}

func (h *LayananHandler) CreateBrand(c *gin.Context) {
	var req CreateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	brand := &domain.HargaLayanan{
		IDBrand:       req.IDBrand,
		Brand:         req.Brand,
		Pajak:         req.Pajak,
		XenditKeyName: "JAKINET",
	}
	if req.XenditKeyName != "" {
		brand.XenditKeyName = req.XenditKeyName
	}

	created, err := h.hargaUsecase.Create(c.Request.Context(), brand)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *LayananHandler) GetAllBrands(c *gin.Context) {
	brands, err := h.hargaUsecase.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Client-side cache headers parity
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, brands)
}

func (h *LayananHandler) GetBrandByID(c *gin.Context) {
	idBrand := c.Param("id_brand")
	brand, err := h.hargaUsecase.GetByID(c.Request.Context(), idBrand)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, brand)
}

func (h *LayananHandler) UpdateBrand(c *gin.Context) {
	idBrand := c.Param("id_brand")
	var req UpdateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Brand != nil {
		updates["brand"] = *req.Brand
	}
	if req.Pajak != nil {
		updates["pajak"] = *req.Pajak
	}
	if req.XenditKeyName != nil {
		updates["xendit_key_name"] = *req.XenditKeyName
	}

	updated, err := h.hargaUsecase.Update(c.Request.Context(), idBrand, updates)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *LayananHandler) DeleteBrand(c *gin.Context) {
	idBrand := c.Param("id_brand")
	if err := h.hargaUsecase.Delete(c.Request.Context(), idBrand); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// --- PaketLayanan Structs & Handlers ---

type CreatePaketRequest struct {
	IDBrand   string  `json:"id_brand" binding:"required"`
	NamaPaket string  `json:"nama_paket" binding:"required"`
	Kecepatan int     `json:"kecepatan" binding:"required,gt=0"`
	Harga     float64 `json:"harga" binding:"required,gt=0"`
}

type UpdatePaketRequest struct {
	IDBrand   *string  `json:"id_brand"`
	NamaPaket *string  `json:"nama_paket"`
	Kecepatan *int     `json:"kecepatan" binding:"omitempty,gt=0"`
	Harga     *float64 `json:"harga" binding:"omitempty,gt=0"`
}

func (h *LayananHandler) CreatePaket(c *gin.Context) {
	var req CreatePaketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paket := &domain.PaketLayanan{
		IDBrand:   req.IDBrand,
		NamaPaket: req.NamaPaket,
		Kecepatan: req.Kecepatan,
		Harga:     req.Harga,
	}

	created, err := h.paketUsecase.Create(c.Request.Context(), paket)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *LayananHandler) GetAllPakets(c *gin.Context) {
	pakets, err := h.paketUsecase.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pakets)
}

func (h *LayananHandler) GetPaketByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	paket, err := h.paketUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, paket)
}

func (h *LayananHandler) UpdatePaket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req UpdatePaketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.IDBrand != nil {
		updates["id_brand"] = *req.IDBrand
	}
	if req.NamaPaket != nil {
		updates["nama_paket"] = *req.NamaPaket
	}
	if req.Kecepatan != nil {
		updates["kecepatan"] = *req.Kecepatan
	}
	if req.Harga != nil {
		updates["harga"] = *req.Harga
	}

	updated, err := h.paketUsecase.Update(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *LayananHandler) DeletePaket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.paketUsecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Paket layanan berhasil dihapus"})
}

// --- Diskon Structs & Handlers ---

type CreateDiskonRequest struct {
	NamaDiskon       string  `json:"nama_diskon" binding:"required,min=1"`
	PersentaseDiskon float64 `json:"persentase_diskon" binding:"required,gt=0,lte=100"`
	Cluster          string  `json:"cluster" binding:"required,min=1"`
	IsActive         *bool   `json:"is_active"`
	TglMulai         *string `json:"tgl_mulai"`
	TglSelesai       *string `json:"tgl_selesai"`
}

type UpdateDiskonRequest struct {
	NamaDiskon       *string  `json:"nama_diskon" binding:"omitempty,min=1"`
	PersentaseDiskon *float64 `json:"persentase_diskon" binding:"omitempty,gt=0,lte=100"`
	Cluster          *string  `json:"cluster" binding:"omitempty,min=1"`
	IsActive         *bool    `json:"is_active"`
	TglMulai         *string  `json:"tgl_mulai"`
	TglSelesai       *string  `json:"tgl_selesai"`
}

func (h *LayananHandler) CreateDiskon(c *gin.Context) {
	var req CreateDiskonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	diskon := &domain.Diskon{
		NamaDiskon:       req.NamaDiskon,
		PersentaseDiskon: req.PersentaseDiskon,
		Cluster:          req.Cluster,
		IsActive:         true,
	}

	if req.IsActive != nil {
		diskon.IsActive = *req.IsActive
	}

	if req.TglMulai != nil && *req.TglMulai != "" {
		if t, err := time.Parse("2006-01-02", *req.TglMulai); err == nil {
			diskon.TglMulai = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format tgl_mulai harus YYYY-MM-DD"})
			return
		}
	}

	if req.TglSelesai != nil && *req.TglSelesai != "" {
		if t, err := time.Parse("2006-01-02", *req.TglSelesai); err == nil {
			diskon.TglSelesai = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format tgl_selesai harus YYYY-MM-DD"})
			return
		}
	}

	created, err := h.diskonUsecase.Create(c.Request.Context(), diskon)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *LayananHandler) GetAllDiskons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	cluster := c.Query("cluster")

	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err == nil {
			isActive = &active
		}
	}

	diskons, total, err := h.diskonUsecase.GetAll(c.Request.Context(), page, pageSize, cluster, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      diskons,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *LayananHandler) GetDiskonByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	diskon, err := h.diskonUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, diskon)
}

func (h *LayananHandler) UpdateDiskon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req UpdateDiskonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.NamaDiskon != nil {
		updates["nama_diskon"] = *req.NamaDiskon
	}
	if req.PersentaseDiskon != nil {
		updates["persentase_diskon"] = *req.PersentaseDiskon
	}
	if req.Cluster != nil {
		updates["cluster"] = *req.Cluster
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	// Standardize date input checks
	if req.TglMulai != nil {
		if *req.TglMulai == "" {
			updates["tgl_mulai"] = nil
		} else {
			if t, err := time.Parse("2006-01-02", *req.TglMulai); err == nil {
				updates["tgl_mulai"] = t
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Format tgl_mulai harus YYYY-MM-DD"})
				return
			}
		}
	}

	if req.TglSelesai != nil {
		if *req.TglSelesai == "" {
			updates["tgl_selesai"] = nil
		} else {
			if t, err := time.Parse("2006-01-02", *req.TglSelesai); err == nil {
				updates["tgl_selesai"] = t
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Format tgl_selesai harus YYYY-MM-DD"})
				return
			}
		}
	}

	updated, err := h.diskonUsecase.Update(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *LayananHandler) DeleteDiskon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.diskonUsecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LayananHandler) ActivateDiskon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	activated, err := h.diskonUsecase.Activate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, activated)
}

func (h *LayananHandler) GetDiskonByCluster(c *gin.Context) {
	cluster := c.Param("cluster")
	tanggalStr := c.Query("tanggal")

	var checkDate *time.Time
	if tanggalStr != "" {
		if t, err := time.Parse("2006-01-02", tanggalStr); err == nil {
			checkDate = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal harus YYYY-MM-DD"})
			return
		}
	}

	diskon, err := h.diskonUsecase.GetActiveForCluster(c.Request.Context(), cluster, checkDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if diskon == nil {
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(http.StatusOK, diskon)
}

func (h *LayananHandler) GetClusterList(c *gin.Context) {
	clusters, err := h.diskonUsecase.GetClusterList(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, clusters)
}
