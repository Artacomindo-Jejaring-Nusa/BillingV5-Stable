package http

import (
	"log"
	"net/http"
	"strconv"

	"billing-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardUsecase domain.DashboardUsecase
	userUsecase      domain.UserUsecase
}

func NewDashboardHandler(r *gin.RouterGroup, du domain.DashboardUsecase, uu domain.UserUsecase, authMiddleware gin.HandlerFunc) {
	handler := &DashboardHandler{
		dashboardUsecase: du,
		userUsecase:      uu,
	}

	dashboard := r.Group("/dashboard")
	dashboard.Use(authMiddleware)
	{
		dashboard.GET("", handler.GetDashboardData)
		dashboard.GET("/", handler.GetDashboardData)
		dashboard.GET("/loyalitas-users-by-segment", handler.GetLoyaltyUsersBySegment)
		dashboard.GET("/sidebar-badges", handler.GetSidebarBadges)
		dashboard.GET("/paket-details", handler.GetPaketDetails)
		dashboard.GET("/invoice-generation-monitor", handler.GetInvoiceGenerationMonitor)
		dashboard.GET("/future-invoice-projection", handler.GetFutureInvoiceProjection)
	}
}

func (h *DashboardHandler) GetDashboardData(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=300")

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userUsecase.GetByID(c.Request.Context(), userID)
	if err != nil || user == nil || user.RoleID == nil {
		c.JSON(http.StatusOK, domain.DashboardData{})
		return
	}

	userPermissions := make(map[string]bool)
	for _, p := range user.Role.Permissions {
		userPermissions[p.Name] = true
	}

	data, err := h.dashboardUsecase.GetDashboardData(c.Request.Context(), userPermissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *DashboardHandler) GetLoyaltyUsersBySegment(c *gin.Context) {
	segmen := c.Query("segmen")
	if segmen == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'segmen' is required"})
		return
	}

	details, err := h.dashboardUsecase.GetLoyaltyUsersBySegment(c.Request.Context(), segmen)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, details)
}

func (h *DashboardHandler) GetSidebarBadges(c *gin.Context) {
	res, err := h.dashboardUsecase.GetSidebarBadges(c.Request.Context())
	if err != nil {
		log.Printf("[DASHBOARD ERROR] Failed to fetch sidebar badges: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *DashboardHandler) GetPaketDetails(c *gin.Context) {
	res, err := h.dashboardUsecase.GetPaketDetails(c.Request.Context())
	if err != nil {
		log.Printf("[DASHBOARD ERROR] Failed to fetch paket details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}


func (h *DashboardHandler) GetInvoiceGenerationMonitor(c *gin.Context) {
	targetDate := c.Query("target_date")
	userRoleVal, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userRole := userRoleVal.(string)

	log.Printf("[DASHBOARD] Fetching monitor for role: '%s', targetDate: '%s'", userRole, targetDate)

	res, err := h.dashboardUsecase.GetInvoiceGenerationMonitor(c.Request.Context(), targetDate, userRole)
	if err != nil {
		log.Printf("[DASHBOARD ERROR] Monitor access denied for role '%s': %v", userRole, err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *DashboardHandler) GetFutureInvoiceProjection(c *gin.Context) {
	targetDate := c.Query("target_date")
	userRoleVal, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userRole := userRoleVal.(string)

	res, err := h.dashboardUsecase.GetFutureInvoiceProjection(c.Request.Context(), targetDate, userRole)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
