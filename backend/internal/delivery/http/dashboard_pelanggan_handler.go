package http

import (
	"net/http"
	"strconv"
	"strings"

	"billing-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type DashboardPelangganHandler struct {
	dashboardRepository domain.DashboardRepository
}

func NewDashboardPelangganHandler(r *gin.RouterGroup, repo domain.DashboardRepository, authMiddleware gin.HandlerFunc) {
	handler := &DashboardPelangganHandler{
		dashboardRepository: repo,
	}

	dashboard := r.Group("/dashboard-pelanggan")
	dashboard.Use(authMiddleware)
	{
		dashboard.GET("", handler.GetDashboardPelangganData)
		dashboard.GET("/", handler.GetDashboardPelangganData)
	}
}

func (h *DashboardPelangganHandler) GetDashboardPelangganData(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=300")

	timespan := c.Query("timespan")
	if timespan == "" {
		timespan = "6m"
	}

	months := h.parseTimespan(timespan)

	mainStats, err := h.dashboardRepository.GetMainStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch main stats"})
		return
	}

	growthChart, err := h.dashboardRepository.GetGrowthChartData(c.Request.Context(), months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch growth chart"})
		return
	}

	revenueChart, err := h.dashboardRepository.GetRevenueChartData(c.Request.Context(), months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revenue chart"})
		return
	}

	response := &domain.DashboardPelangganResponse{
		MainStats:    mainStats,
		GrowthChart:  growthChart,
		RevenueChart: revenueChart,
	}

	c.JSON(http.StatusOK, response)
}

func (h *DashboardPelangganHandler) parseTimespan(timespan string) int {
	timespan = strings.ToLower(strings.TrimSpace(timespan))

	switch timespan {
	case "3m":
		return 3
	case "6m":
		return 6
	case "1y":
		return 12
	default:
		n, err := strconv.Atoi(strings.TrimSuffix(timespan, "m"))
		if err != nil {
			return 6
		}
		return n
	}
}
