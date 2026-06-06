package http

import (
	"net/http"
	"strconv"

	"billing-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

// --- Role Handler ---

type RoleHandler struct {
	roleUsecase domain.RoleUsecase
}

type CreateRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	PermissionIDs []uint64 `json:"permission_ids"`
}

type UpdateRoleRequest struct {
	Name          *string  `json:"name"`
	PermissionIDs []uint64 `json:"permission_ids"`
}

// NewRoleHandler sets up the routing for role endpoints
func NewRoleHandler(r *gin.RouterGroup, ru domain.RoleUsecase, authMiddleware gin.HandlerFunc) {
	handler := &RoleHandler{roleUsecase: ru}

	roles := r.Group("/roles")
	roles.Use(authMiddleware)
	{
		roles.GET("", handler.GetAll)
		roles.GET("/:id", handler.GetByID)
		roles.POST("", handler.Create)
		roles.PATCH("/:id", handler.Update)
		roles.DELETE("/:id", handler.Delete)
	}
}

func (h *RoleHandler) GetAll(c *gin.Context) {
	roles, err := h.roleUsecase.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	role, err := h.roleUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": role})
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.roleUsecase.Create(c.Request.Context(), req.Name, req.PermissionIDs)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": role})
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := ""
	if req.Name != nil {
		name = *req.Name
	}

	role, err := h.roleUsecase.Update(c.Request.Context(), id, name, req.PermissionIDs)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": role})
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.roleUsecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// --- Permission Handler ---

type PermissionHandler struct {
	permUsecase domain.PermissionUsecase
}

// NewPermissionHandler sets up the routing for permission endpoints
func NewPermissionHandler(r *gin.RouterGroup, pu domain.PermissionUsecase, authMiddleware gin.HandlerFunc) {
	handler := &PermissionHandler{permUsecase: pu}

	perms := r.Group("/permissions")
	perms.Use(authMiddleware)
	{
		perms.GET("", handler.GetAll)
	}
}

func (h *PermissionHandler) GetAll(c *gin.Context) {
	permissions, err := h.permUsecase.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": permissions})
}
