package http

import (
	"net/http"
	"strconv"

	"billing-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type APIKeyHandler struct {
	apiKeyUsecase domain.APIKeyUsecase
}

func NewAPIKeyHandler(r *gin.RouterGroup, uu domain.APIKeyUsecase, authMiddleware gin.HandlerFunc) {
	handler := &APIKeyHandler{apiKeyUsecase: uu}

	apiKeyGroup := r.Group("/api-keys")
	apiKeyGroup.Use(authMiddleware)
	{
		apiKeyGroup.GET("", handler.FetchAPIKeys)
		apiKeyGroup.POST("", handler.CreateAPIKey)
		apiKeyGroup.PATCH("/:id/toggle", handler.ToggleAPIKey)
		apiKeyGroup.DELETE("/:id", handler.DeleteAPIKey)
	}
}

func (h *APIKeyHandler) FetchAPIKeys(c *gin.Context) {
	keys, err := h.apiKeyUsecase.FetchAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]domain.APIKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = domain.APIKeyResponse{
			ID:        key.ID,
			Name:      key.Name,
			Prefix:    key.Prefix,
			RoleID:    key.RoleID,
			RoleName:  key.Role.Name,
			IsActive:  key.IsActive,
			CreatedAt: key.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req domain.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, rawKey, err := h.apiKeyUsecase.CreateAPIKey(c.Request.Context(), req.Name, req.RoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := domain.CreateAPIKeyResponse{
		APIKeyResponse: domain.APIKeyResponse{
			ID:        apiKey.ID,
			Name:      apiKey.Name,
			Prefix:    apiKey.Prefix,
			RoleID:    apiKey.RoleID,
			RoleName:  apiKey.Role.Name,
			IsActive:  apiKey.IsActive,
			CreatedAt: apiKey.CreatedAt,
		},
		RawKey: rawKey,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "API Key created successfully. Copy the raw key now, it will not be shown again.",
		"data":    resp,
	})
}

func (h *APIKeyHandler) ToggleAPIKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API Key ID"})
		return
	}

	apiKey, err := h.apiKeyUsecase.ToggleAPIKey(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := domain.APIKeyResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Prefix:    apiKey.Prefix,
		RoleID:    apiKey.RoleID,
		RoleName:  apiKey.Role.Name,
		IsActive:  apiKey.IsActive,
		CreatedAt: apiKey.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API Key status updated successfully",
		"data":    resp,
	})
}

func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API Key ID"})
		return
	}

	err = h.apiKeyUsecase.DeleteAPIKey(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API Key deleted/revoked successfully"})
}
