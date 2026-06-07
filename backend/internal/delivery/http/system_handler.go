package http

import (
	"net/http"
	"strconv"
	"fmt"

	"billing-backend/internal/domain"
	"billing-backend/internal/scheduler"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	usecase      domain.SystemUsecase
	schedulerMgr *scheduler.SchedulerManager
}

func NewSystemHandler(r *gin.RouterGroup, su domain.SystemUsecase, sm *scheduler.SchedulerManager, authMiddleware gin.HandlerFunc) {
	handler := &SystemHandler{
		usecase:      su,
		schedulerMgr: sm,
	}

	systemGroup := r.Group("/system")
	systemGroup.Use(authMiddleware)
	{
		systemGroup.GET("/settings/:key", handler.GetSetting)
		systemGroup.POST("/settings", handler.SetSetting)
		
		// Scheduler
		systemGroup.GET("/scheduler/status", handler.GetSchedulerStatus)
		systemGroup.POST("/scheduler/toggle", handler.ToggleScheduler)
		systemGroup.GET("/scheduler/jobs", handler.GetSchedulerJobs)
		systemGroup.POST("/scheduler/jobs/:job_key", handler.UpdateSchedulerJob)
		systemGroup.POST("/scheduler/jobs/:job_key/run", handler.RunSchedulerJob)
		
		// Syarat & Ketentuan (SK)
		systemGroup.GET("/sk", handler.GetSKAll)
		systemGroup.GET("/sk/:id", handler.GetSKByID)
		systemGroup.POST("/sk", handler.CreateSK)
		systemGroup.PUT("/sk/:id", handler.UpdateSK)
		systemGroup.DELETE("/sk/:id", handler.DeleteSK)

		// Activity Logs
		systemGroup.GET("/activity-logs", handler.GetActivityLogs)
	}

	// Register top-level compatibility routes for frontend
	compatGroup := r.Group("")
	compatGroup.Use(authMiddleware)
	{
		compatGroup.GET("/activity-logs", handler.GetActivityLogs)
		compatGroup.GET("/activity-logs/", handler.GetActivityLogs)
		compatGroup.GET("/sk", handler.GetSKAll)
		compatGroup.GET("/sk/", handler.GetSKAll)
		compatGroup.GET("/sk/:id", handler.GetSKByID)
		compatGroup.POST("/sk", handler.CreateSK)
		compatGroup.POST("/sk/", handler.CreateSK)
		compatGroup.PUT("/sk/:id", handler.UpdateSK)
		compatGroup.PATCH("/sk/:id", handler.UpdateSK)
		compatGroup.DELETE("/sk/:id", handler.DeleteSK)
	}
}

func (h *SystemHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	val, err := h.usecase.GetSetting(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"key": key, "value": val}})
}

type SetSettingRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

func (h *SystemHandler) SetSetting(c *gin.Context) {
	var req SetSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	err := h.usecase.SetSetting(c.Request.Context(), req.Key, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setting saved successfully"})
}

func (h *SystemHandler) GetSKAll(c *gin.Context) {
	sk, err := h.usecase.FetchSKAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sk})
}

func (h *SystemHandler) GetSKByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	sk, err := h.usecase.GetSKByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sk})
}

func (h *SystemHandler) CreateSK(c *gin.Context) {
	var sk domain.SyaratKetentuan
	if err := c.ShouldBindJSON(&sk); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.CreateSK(c.Request.Context(), &sk); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": sk})
}

func (h *SystemHandler) UpdateSK(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var sk domain.SyaratKetentuan
	if err := c.ShouldBindJSON(&sk); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.UpdateSK(c.Request.Context(), id, &sk); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sk})
}

func (h *SystemHandler) DeleteSK(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.usecase.DeleteSK(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Terms and conditions deleted successfully"})
}

func (h *SystemHandler) GetActivityLogs(c *gin.Context) {
	// Parse skip & limit (compat with skip/limit pagination)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	
	// If page/page_size are supplied instead
	if pageStr := c.Query("page"); pageStr != "" {
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		if pageSize < 1 {
			pageSize = 10
		}
		offset = (page - 1) * pageSize
		limit = pageSize
	}

	search := c.Query("search")
	action := c.Query("action")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	var userID *uint64
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 64); err == nil {
			userID = &id
		}
	}

	filters := domain.ActivityLogFilters{
		Limit:    limit,
		Offset:   offset,
		Search:   search,
		UserID:   userID,
		Action:   action,
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}

	logs, total, err := h.usecase.FetchActivityLogs(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"items": logs,
		"total": total,
		"meta": gin.H{
			"total":     total,
			"offset":    offset,
			"limit":     limit,
		},
	})
}

func (h *SystemHandler) GetSchedulerStatus(c *gin.Context) {
	enabled := h.schedulerMgr.IsGlobalEnabled(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"enabled": enabled}})
}

type ToggleSchedulerRequest struct {
	Enabled bool `json:"enabled"`
}

func (h *SystemHandler) ToggleScheduler(c *gin.Context) {
	var req ToggleSchedulerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	err := h.schedulerMgr.ToggleGlobal(c.Request.Context(), req.Enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	status := "enabled"
	if !req.Enabled {
		status = "disabled"
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Scheduler globally %s", status)})
}

func (h *SystemHandler) GetSchedulerJobs(c *gin.Context) {
	jobs := h.schedulerMgr.GetJobsStatus(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"data": jobs})
}

type UpdateJobRequest struct {
	Schedule string `json:"schedule" binding:"required"`
	Enabled  bool   `json:"enabled"`
}

func (h *SystemHandler) UpdateSchedulerJob(c *gin.Context) {
	key := c.Param("job_key")
	var req UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	err := h.schedulerMgr.UpdateJob(c.Request.Context(), key, req.Schedule, req.Enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduler job updated successfully"})
}

func (h *SystemHandler) RunSchedulerJob(c *gin.Context) {
	key := c.Param("job_key")
	err := h.schedulerMgr.RunJobNow(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Job %s triggered successfully in the background", key)})
}
