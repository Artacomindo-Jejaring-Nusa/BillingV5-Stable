package http

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"billing-backend/internal/domain"
	"billing-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type PelangganHandler struct {
	pelangganUsecase domain.PelangganUsecase
}

func NewPelangganHandler(r *gin.RouterGroup, pu domain.PelangganUsecase, authMiddleware gin.HandlerFunc) {
	handler := &PelangganHandler{
		pelangganUsecase: pu,
	}

	// Protect all pelanggan endpoints
	pelangganGroup := r.Group("/pelanggan")
	pelangganGroup.Use(authMiddleware)
	{
		pelangganGroup.GET("", middleware.PermissionMiddleware("view_pelanggan"), handler.FetchAll)
		pelangganGroup.GET("/lokasi/unik", middleware.PermissionMiddleware("view_pelanggan"), handler.GetUniqueLocations)
		pelangganGroup.GET("/export", middleware.PermissionMiddleware("view_pelanggan"), handler.Export)
		pelangganGroup.POST("/import", middleware.PermissionMiddleware("create_pelanggan"), handler.Import)
		pelangganGroup.GET("/template/csv", middleware.PermissionMiddleware("create_pelanggan"), handler.DownloadCSVTemplate)
		pelangganGroup.GET("/:id", middleware.PermissionMiddleware("view_pelanggan"), handler.GetByID)
		pelangganGroup.POST("", middleware.PermissionMiddleware("create_pelanggan"), handler.Store)
		pelangganGroup.PUT("/:id", middleware.PermissionMiddleware("edit_pelanggan"), handler.Update)
		pelangganGroup.DELETE("/:id", middleware.PermissionMiddleware("delete_pelanggan"), handler.Delete)
	}
}

func (h *PelangganHandler) Export(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	data, contentType, err := h.pelangganUsecase.Export(c.Request.Context(), format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := "export_pelanggan"
	if format == "excel" {
		filename += ".xlsx"
	} else {
		filename += ".csv"
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

func (h *PelangganHandler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer opened.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, opened); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	count, err := h.pelangganUsecase.ImportFromCSV(c.Request.Context(), buf.String())
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Import finished with errors",
			"errors":  strings.Split(err.Error(), "; "),
			"count":   count,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengimpor " + strconv.Itoa(count) + " data pelanggan",
		"count":   count,
	})
}

func (h *PelangganHandler) DownloadCSVTemplate(c *gin.Context) {
	headers := []string{"No KTP", "Nama", "Alamat", "Blok", "Unit", "No Telp", "Email", "Layanan", "ID Brand", "Tgl Instalasi"}
	buf := new(bytes.Buffer)
	buf.Write([]byte("\ufeff"))
	buf.WriteString(strings.Join(headers, ";") + "\n")
	sample := []string{"3201234567890001", "Budi Santoso", "Jl. Melati No. 123", "A", "101", "08123456789", "budi@example.com", "Internet 20 Mbps", "ajn-01", "2024-01-01"}
	buf.WriteString(strings.Join(sample, ";") + "\n")

	c.Header("Content-Disposition", "attachment; filename=template_import_pelanggan.csv")
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

func (h *PelangganHandler) GetUniqueLocations(c *gin.Context) {
	locations, err := h.pelangganUsecase.GetUniqueLocations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, locations)
}

func (h *PelangganHandler) FetchAll(c *gin.Context) {
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	search := c.Query("search")
	connectionStatus := c.Query("connection_status")

	pelanggans, total, err := h.pelangganUsecase.FetchAll(c.Request.Context(), skip, limit, search, connectionStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        pelanggans,
		"total_count": total,
	})
}

func (h *PelangganHandler) GetByID(c *gin.Context) {
	id, ok := h.parseID(c)
	if !ok {
		return
	}

	pelanggan, err := h.pelangganUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pelanggan)
}

func (h *PelangganHandler) Store(c *gin.Context) {
	var pelanggan domain.Pelanggan
	if err := c.ShouldBindJSON(&pelanggan); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.pelangganUsecase.Store(c.Request.Context(), &pelanggan); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, pelanggan)
}

func (h *PelangganHandler) Update(c *gin.Context) {
	id, ok := h.parseID(c)
	if !ok {
		return
	}

	var pelanggan domain.Pelanggan
	if err := c.ShouldBindJSON(&pelanggan); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.pelangganUsecase.Update(c.Request.Context(), id, &pelanggan); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pelanggan updated successfully"})
}

func (h *PelangganHandler) Delete(c *gin.Context) {
	id, ok := h.parseID(c)
	if !ok {
		return
	}

	if err := h.pelangganUsecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pelanggan deleted successfully"})
}

func (h *PelangganHandler) parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return 0, false
	}
	return id, true
}

func (h *PelangganHandler) handleError(c *gin.Context, err error) {
	if strings.Contains(err.Error(), "sudah terdaftar") || strings.Contains(err.Error(), "required") {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
