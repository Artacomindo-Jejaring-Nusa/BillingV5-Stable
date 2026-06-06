package http

import (
	"net/http"
	"strconv"

	"billing-backend/internal/domain"

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
		pelangganGroup.GET("", handler.FetchAll)
		pelangganGroup.GET("/:id", handler.GetByID)
		pelangganGroup.POST("", handler.Store)
		pelangganGroup.PUT("/:id", handler.Update)
		pelangganGroup.DELETE("/:id", handler.Delete)
	}
}

func (h *PelangganHandler) FetchAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// Map 'limit' query parameter to pageSize if present
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitVal, err := strconv.Atoi(limitStr); err == nil && limitVal > 0 {
			pageSize = limitVal
		}
	}

	connectionStatus := c.Query("connection_status")

	pelanggans, total, err := h.pelangganUsecase.FetchAll(c.Request.Context(), page, pageSize, connectionStatus)
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pelanggan)
}

func (h *PelangganHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var pelanggan domain.Pelanggan
	if err := c.ShouldBindJSON(&pelanggan); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.pelangganUsecase.Update(c.Request.Context(), id, &pelanggan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pelanggan updated successfully"})
}

func (h *PelangganHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.pelangganUsecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pelanggan deleted successfully"})
}
