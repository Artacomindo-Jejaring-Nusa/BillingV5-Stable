package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"billing-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type DataTeknisHandler struct {
	dataTeknisUsecase domain.DataTeknisUsecase
}

type IPCheckRequest struct {
	IPAddress string  `json:"ip_address" binding:"required"`
	CurrentID *uint64 `json:"current_id"`
}

// NewDataTeknisHandler registers all DataTeknis routes
func NewDataTeknisHandler(r *gin.RouterGroup, du domain.DataTeknisUsecase, authMiddleware gin.HandlerFunc) {
	handler := &DataTeknisHandler{
		dataTeknisUsecase: du,
	}

	g := r.Group("/data_teknis")
	g.Use(authMiddleware)
	{
		g.GET("", handler.FetchAll)
		g.GET("/available-olt", handler.GetAvailableOLT)
		g.GET("/available-profiles", handler.GetAvailableProfiles)
		g.GET("/available-vlans", handler.GetAvailableVlans)
		g.GET("/onu-power-ranges", handler.GetOnuPowerRanges)
		g.GET("/by-pelanggan/:pelanggan_id", handler.GetByPelangganID)
		g.GET("/:id", handler.GetByID)
		g.POST("", handler.Store)
		g.PUT("/:id", handler.Update)
		g.PATCH("/:id", handler.Update)
		g.DELETE("/:id", handler.Delete)
		g.POST("/check-ip", handler.CheckIPAddress)
		g.GET("/available-profiles/:paket_layanan_id/:pelanggan_id", handler.GetAvailableProfilesForPackage)
		g.GET("/last-ip/:mikrotik_server_id", handler.GetLastUsedIP)
		g.GET("/export", handler.Export)
		g.POST("/import/csv", handler.ImportFromCSV)
		g.GET("/template/csv", handler.DownloadCSVTemplate)
		g.POST("/upload-speedtest", handler.UploadSpeedtest)
	}
}

func (h *DataTeknisHandler) Export(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	data, contentType, err := h.dataTeknisUsecase.Export(c.Request.Context(), format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := "export_data_teknis"
	if format == "excel" {
		filename += ".xlsx"
	} else {
		filename += ".csv"
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

func (h *DataTeknisHandler) FetchAll(c *gin.Context) {
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	search := c.Query("search")
	olt := c.Query("olt")
	profile := c.Query("profile")
	vlan := c.Query("vlan")

	var onuPowerMin *int
	if minStr := c.Query("onu_power_min"); minStr != "" {
		if val, err := strconv.Atoi(minStr); err == nil {
			onuPowerMin = &val
		}
	}

	var onuPowerMax *int
	if maxStr := c.Query("onu_power_max"); maxStr != "" {
		if val, err := strconv.Atoi(maxStr); err == nil {
			onuPowerMax = &val
		}
	}

	data, total, err := h.dataTeknisUsecase.FetchAll(c.Request.Context(), skip, limit, search, olt, profile, vlan, onuPowerMin, onuPowerMax)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        data,
		"total_count": total,
	})
}

func (h *DataTeknisHandler) GetAvailableOLT(c *gin.Context) {
	list, err := h.dataTeknisUsecase.GetAvailableOLT(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res := []string{"Semua"}
	res = append(res, list...)
	c.JSON(http.StatusOK, res)
}

func (h *DataTeknisHandler) GetAvailableProfiles(c *gin.Context) {
	list, err := h.dataTeknisUsecase.GetAvailableProfiles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res := []string{"Semua"}
	res = append(res, list...)
	c.JSON(http.StatusOK, res)
}

func (h *DataTeknisHandler) GetAvailableVlans(c *gin.Context) {
	list, err := h.dataTeknisUsecase.GetAvailableVlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res := []string{"Semua"}
	res = append(res, list...)
	c.JSON(http.StatusOK, res)
}

func (h *DataTeknisHandler) GetOnuPowerRanges(c *gin.Context) {
	ranges, err := h.dataTeknisUsecase.GetOnuPowerRanges(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"min": ranges["min"],
		"max": ranges["max"],
		"ranges": []gin.H{
			{"label": "Sinyal Baik (>-24 dBm)", "min": -23, "max": 0},
			{"label": "Sinyal Sedang (-27 s/d -24 dBm)", "min": -27, "max": -24},
			{"label": "Sinyal Lemah (<-27 dBm)", "min": -50, "max": -28},
		},
	})
}

func (h *DataTeknisHandler) GetByPelangganID(c *gin.Context) {
	pelangganID, err := strconv.ParseUint(c.Param("pelanggan_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pelanggan ID"})
		return
	}

	dt, err := h.dataTeknisUsecase.GetByPelangganID(c.Request.Context(), pelangganID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if dt == nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	c.JSON(http.StatusOK, []interface{}{dt})
}

func (h *DataTeknisHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	dt, err := h.dataTeknisUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if dt == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "data teknis not found"})
		return
	}

	c.JSON(http.StatusOK, dt)
}

func (h *DataTeknisHandler) Store(c *gin.Context) {
	var dt domain.DataTeknis
	if err := c.ShouldBindJSON(&dt); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.dataTeknisUsecase.Store(c.Request.Context(), &dt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dt)
}

func (h *DataTeknisHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var dt domain.DataTeknis
	if err := c.ShouldBindJSON(&dt); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.dataTeknisUsecase.Update(c.Request.Context(), id, &dt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.dataTeknisUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *DataTeknisHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.dataTeknisUsecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *DataTeknisHandler) CheckIPAddress(c *gin.Context) {
	var req IPCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isTaken, err := h.dataTeknisUsecase.CheckIPAddress(c.Request.Context(), req.IPAddress, req.CurrentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	message := "IP tersedia"
	if isTaken {
		message = "IP sudah terpakai"
	}

	c.JSON(http.StatusOK, gin.H{
		"is_taken": isTaken,
		"message":  message,
		"owner_id": nil,
	})
}

func (h *DataTeknisHandler) GetAvailableProfilesForPackage(c *gin.Context) {
	paketLayananID, err := strconv.ParseUint(c.Param("paket_layanan_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid paket_layanan_id"})
		return
	}

	pelangganID, err := strconv.ParseUint(c.Param("pelanggan_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pelanggan_id"})
		return
	}

	var mikrotikServerID *uint64
	if serverIDStr := c.Query("mikrotik_server_id"); serverIDStr != "" {
		if val, err := strconv.ParseUint(serverIDStr, 10, 64); err == nil {
			mikrotikServerID = &val
		}
	}

	res, err := h.dataTeknisUsecase.GetAvailableProfilesForPackage(c.Request.Context(), paketLayananID, pelangganID, mikrotikServerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *DataTeknisHandler) GetLastUsedIP(c *gin.Context) {
	mikrotikServerID, err := strconv.ParseUint(c.Param("mikrotik_server_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mikrotik_server_id"})
		return
	}

	res, err := h.dataTeknisUsecase.GetLastUsedIP(c.Request.Context(), mikrotikServerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *DataTeknisHandler) ImportFromCSV(c *gin.Context) {
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

	csvContent := buf.String()
	count, err := h.dataTeknisUsecase.ImportFromCSV(c.Request.Context(), csvContent)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation errors occurred",
			"errors":  strings.Split(err.Error(), "; "),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengimpor data teknis baru",
		"count":   count,
	})
}

func (h *DataTeknisHandler) DownloadCSVTemplate(c *gin.Context) {
	headers := []string{
		"email_pelanggan",
		"olt",
		"kode_odp",
		"port_odp",
		"id_vlan",
		"id_pelanggan",
		"password_pppoe",
		"ip_pelanggan",
		"profile_pppoe",
		"olt_custom",
		"pon",
		"otb",
		"odc",
		"onu_power",
		"sn",
	}

	buf := new(bytes.Buffer)
	buf.Write([]byte("\ufeff"))
	buf.WriteString(strings.Join(headers, ";") + "\n")
	sample := []string{
		"budi.s@example.com",
		"Mikrotik-Pusat",
		"ODP-TMB-01",
		"1",
		"101",
		"budi-santoso",
		"change_on_first_login",
		"10.10.1.25",
		"50mbps-profile",
		"OLT-Tambun-Satu",
		"1",
		"101",
		"3",
		"-22",
		"ZTEG1A2B3C4D",
	}
	buf.WriteString(strings.Join(sample, ";") + "\n")

	c.Header("Content-Disposition", "attachment; filename=template_import_teknis.csv")
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

func (h *DataTeknisHandler) UploadSpeedtest(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File must have a valid extension"})
		return
	}

	uniqueFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dir := "./uploads/speedtest"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	filePath := filepath.Join(dir, uniqueFilename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
		return
	}

	fileInfo, err := os.Stat(filePath)
	var size int64
	if err == nil {
		size = fileInfo.Size()
	}

	contentType := file.Header.Get("Content-Type")
	fileURL := fmt.Sprintf("/static/uploads/speedtest/%s", uniqueFilename)

	c.JSON(http.StatusOK, gin.H{
		"file_url":     fileURL,
		"filename":     file.Filename,
		"content_type": contentType,
		"size":         size,
	})
}
