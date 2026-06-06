package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"billing-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type TroubleTicketHandler struct {
	usecase domain.TroubleTicketUsecase
}

func NewTroubleTicketHandler(r *gin.RouterGroup, tu domain.TroubleTicketUsecase, authMiddleware gin.HandlerFunc) {
	handler := &TroubleTicketHandler{
		usecase: tu,
	}

	ticketGroup := r.Group("/trouble-tickets")
	ticketGroup.Use(authMiddleware)
	{
		ticketGroup.GET("", handler.FetchAll)
		ticketGroup.GET("/statistics", handler.GetStatistics)
		ticketGroup.GET("/statistics/dashboard", handler.GetStatistics)
		ticketGroup.GET("/:id", handler.GetByID)
		ticketGroup.POST("", handler.Create)
		ticketGroup.PUT("/:id", handler.Update)
		ticketGroup.PATCH("/:id", handler.Update) // PATCH updates ticket
		ticketGroup.DELETE("/:id", handler.Delete)
		
		ticketGroup.PATCH("/:id/status", handler.UpdateStatus)
		ticketGroup.POST("/:id/status", handler.UpdateStatus) // POST updates status
		
		ticketGroup.PATCH("/:id/downtime", handler.UpdateDowntime)
		ticketGroup.POST("/:id/downtime", handler.UpdateDowntime) // POST updates downtime
		
		ticketGroup.PATCH("/:id/assign", handler.AssignTicket)
		ticketGroup.POST("/:id/assign", handler.AssignTicket) // POST assigns ticket
		
		ticketGroup.POST("/:id/actions", handler.AddAction)
		ticketGroup.POST("/:id/action", handler.AddAction) // POST action
		
		ticketGroup.GET("/:id/history", handler.GetHistory)
		ticketGroup.GET("/:id/actions", handler.GetActions)
		ticketGroup.GET("/:id/action", handler.GetActions) // GET action list
		
		// Upload evidence
		ticketGroup.POST("/upload-evidence", handler.UploadEvidence)

		// Reporting endpoints
		ticketGroup.GET("/reports/monthly-trends", handler.GetMonthlyTrends)
		ticketGroup.GET("/reports/category-performance", handler.GetCategoryPerformance)
		ticketGroup.GET("/reports/user-performance", handler.GetUserPerformance)
		ticketGroup.GET("/reports/downtime-analysis", handler.GetDowntimeAnalysis)
	}
}

func (h *TroubleTicketHandler) FetchAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "15"))

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if priority := c.Query("priority"); priority != "" {
		filters["priority"] = priority
	}
	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if idBrand := c.Query("id_brand"); idBrand != "" {
		filters["id_brand"] = idBrand
	}
	if pelangganIDStr := c.Query("pelanggan_id"); pelangganIDStr != "" {
		if pid, err := strconv.ParseUint(pelangganIDStr, 10, 64); err == nil {
			filters["pelanggan_id"] = pid
		}
	}
	if assignedToStr := c.Query("assigned_to"); assignedToStr != "" {
		if aid, err := strconv.ParseUint(assignedToStr, 10, 64); err == nil {
			filters["assigned_to"] = aid
		}
	}

	tickets, total, err := h.usecase.FetchAll(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tickets,
		"total": total,
		"page":  page,
	})
}

func (h *TroubleTicketHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	ticket, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ticket})
}

func (h *TroubleTicketHandler) Create(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	var ticket domain.TroubleTicket
	if err := c.ShouldBindJSON(&ticket); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.Create(c.Request.Context(), &ticket, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": ticket})
}

func (h *TroubleTicketHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	var ticket domain.TroubleTicket
	if err := c.ShouldBindJSON(&ticket); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.Update(c.Request.Context(), id, &ticket, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ticket})
}

type UpdateStatusRequest struct {
	Status            string `json:"status" binding:"required"`
	Notes             string `json:"notes"`
	ActionDescription string `json:"action_description"`
	SummaryProblem    string `json:"summary_problem"`
	SummaryAction     string `json:"summary_action"`
	Evidence          string `json:"evidence"`
}

func (h *TroubleTicketHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	err = h.usecase.UpdateStatus(
		c.Request.Context(),
		id,
		domain.TicketStatus(req.Status),
		req.Notes,
		req.ActionDescription,
		req.SummaryProblem,
		req.SummaryAction,
		req.Evidence,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket status updated successfully"})
}

type UpdateDowntimeRequest struct {
	DowntimeStart *time.Time `json:"downtime_start"`
	DowntimeEnd   *time.Time `json:"downtime_end"`
}

func (h *TroubleTicketHandler) UpdateDowntime(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	var req UpdateDowntimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	update := make(map[string]interface{})
	if req.DowntimeStart != nil {
		update["downtime_start"] = *req.DowntimeStart
	}
	if req.DowntimeEnd != nil {
		update["downtime_end"] = *req.DowntimeEnd
	}

	if err := h.usecase.UpdateDowntime(c.Request.Context(), id, update, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket downtime updated successfully"})
}

type AssignTicketRequest struct {
	AssignedTo uint64 `json:"assigned_to" binding:"required"`
	Notes      string `json:"notes"`
}

func (h *TroubleTicketHandler) AssignTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	var req AssignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.AssignTicket(c.Request.Context(), id, req.AssignedTo, req.Notes, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket assigned successfully"})
}

func (h *TroubleTicketHandler) AddAction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	var action domain.ActionTaken
	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.AddAction(c.Request.Context(), id, &action, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": action})
}

func (h *TroubleTicketHandler) GetHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	history, err := h.usecase.GetHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": history})
}

func (h *TroubleTicketHandler) GetActions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	actions, err := h.usecase.GetActions(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": actions})
}

func (h *TroubleTicketHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	if err := h.usecase.Delete(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket deleted successfully"})
}

func (h *TroubleTicketHandler) GetStatistics(c *gin.Context) {
	stats, err := h.usecase.GetStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func parseDateTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return &t, nil
	}
	t, err = time.Parse("2006-01-02", s)
	if err == nil {
		return &t, nil
	}
	t, err = time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		return &t, nil
	}
	return nil, err
}

func (h *TroubleTicketHandler) UploadEvidence(c *gin.Context) {
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
	dir := "./uploads/evidence"
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
	fileURL := fmt.Sprintf("/static/uploads/evidence/%s", uniqueFilename)

	c.JSON(http.StatusOK, gin.H{
		"file_url":     fileURL,
		"filename":     file.Filename,
		"content_type": contentType,
		"size":         size,
	})
}

func (h *TroubleTicketHandler) GetMonthlyTrends(c *gin.Context) {
	months, _ := strconv.Atoi(c.DefaultQuery("months", "12"))
	if months < 1 {
		months = 12
	}

	trends, err := h.usecase.GetMonthlyTrends(c.Request.Context(), months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trends": trends})
}

func (h *TroubleTicketHandler) GetCategoryPerformance(c *gin.Context) {
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")

	dateFrom, err := parseDateTime(dateFromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_from format: " + err.Error()})
		return
	}

	dateTo, err := parseDateTime(dateToStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_to format: " + err.Error()})
		return
	}

	performance, err := h.usecase.GetCategoryPerformance(c.Request.Context(), dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"performance": performance})
}

func (h *TroubleTicketHandler) GetUserPerformance(c *gin.Context) {
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")

	dateFrom, err := parseDateTime(dateFromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_from format: " + err.Error()})
		return
	}

	dateTo, err := parseDateTime(dateToStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_to format: " + err.Error()})
		return
	}

	performance, err := h.usecase.GetUserPerformance(c.Request.Context(), dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"performance": performance})
}

func (h *TroubleTicketHandler) GetDowntimeAnalysis(c *gin.Context) {
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")

	dateFrom, err := parseDateTime(dateFromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_from format: " + err.Error()})
		return
	}

	dateTo, err := parseDateTime(dateToStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_to format: " + err.Error()})
		return
	}

	analysis, err := h.usecase.GetDowntimeAnalysis(c.Request.Context(), dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}
