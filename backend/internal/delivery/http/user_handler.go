package http

import (
	"log"
	"net/http"
	"strconv"

	"billing-backend/internal/domain"
	"billing-backend/internal/middleware"
	"billing-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type UserHandler struct {
	userUsecase domain.UserUsecase
}

// NewUserHandler sets up the routing for user and auth endpoints
func NewUserHandler(r *gin.RouterGroup, uu domain.UserUsecase, authMiddleware gin.HandlerFunc) {
	handler := &UserHandler{userUsecase: uu}

	// Auth routes (public)
	auth := r.Group("/auth")
	{
		auth.POST("/token", middleware.RateLimitMiddleware(15), handler.Login)
		auth.POST("/refresh", handler.RefreshToken)
		auth.POST("/logout", handler.Logout)
		auth.GET("/password-requirements", handler.GetPasswordRequirements)
	}

	// Auth routes (protected)
	authProtected := r.Group("/auth")
	authProtected.Use(authMiddleware)
	{
		authProtected.POST("/logout-all", handler.LogoutAll)
	}

	// User routes (protected)
	users := r.Group("/users")
	users.Use(authMiddleware)
	{
		users.GET("/me", handler.GetProfile)
		users.GET("", middleware.PermissionMiddleware("view_users"), handler.GetAll)
		users.GET("/:id", middleware.PermissionMiddleware("view_users"), handler.GetByID)
		users.POST("", middleware.PermissionMiddleware("create_users"), handler.CreateUser)
		users.PATCH("/:id", middleware.PermissionMiddleware("edit_users"), handler.UpdateUser)
		users.DELETE("/:id", middleware.PermissionMiddleware("delete_users"), handler.DeleteUser)
	}

	// Password management routes (public)
	r.POST("/users/forgot-password", handler.ForgotPassword)
	r.POST("/users/reset-password", handler.ResetPassword)
}

// --- Request/Response Structs ---

type LoginRequest struct {
	Email    string `json:"email" form:"email"`
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	NewPassword string `json:"new_password" binding:"required"`
	Token       string `json:"token" binding:"required"`
}

type CreateUserRequest struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required"`
	RoleID   *uint64 `json:"role_id"`
	IsActive *bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	RoleID   *uint64 `json:"role_id"`
	IsActive *bool   `json:"is_active"`
}

// --- Auth Handlers ---

// POST /auth/token - Login
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	contentType := c.GetHeader("Content-Type")
	log.Printf("[Login] Content-Type: %s", contentType)

	// Handle both JSON and form-urlencoded data
	if contentType == "application/x-www-form-urlencoded" {
		if err := c.ShouldBindWith(&req, binding.Form); err != nil {
			log.Printf("[Login Error] Form binding failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[Login Error] JSON binding failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	log.Printf("[Login Request] Email: '%s', Username: '%s'", req.Email, req.Username)

	email := req.Email
	if email == "" {
		email = req.Username
	}

	if email == "" {
		log.Printf("[Login Error] Both email and username are empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau username harus diisi"})
		return
	}

	user, accessToken, refreshToken, expiresIn, err := h.userUsecase.Authenticate(c.Request.Context(), email, req.Password)
	if err != nil {
		log.Printf("[Login Error] Authentication failed for %s: %v", email, err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":  err.Error(),
			"detail": "Email atau password salah",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "bearer",
		"expires_in":    expiresIn,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role.Name,
		},
	})
}

// POST /auth/refresh - Refresh access token
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, expiresIn, err := h.userUsecase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "bearer",
		"expires_in":    expiresIn,
	})
}

// POST /auth/logout - Logout current device
func (h *UserHandler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userUsecase.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout berhasil"})
}

// POST /auth/logout-all - Logout all devices
func (h *UserHandler) LogoutAll(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	if err := h.userUsecase.LogoutAll(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout dari semua perangkat berhasil"})
}

// GET /auth/password-requirements
func (h *UserHandler) GetPasswordRequirements(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   utils.DefaultPasswordRequirements(),
	})
}

// --- User CRUD Handlers ---

// GET /users/me - Get current user profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	user, err := h.userUsecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GET /users - List all users with search and filter
func (h *UserHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	search := c.Query("search")

	var roleID *uint64
	if rid := c.Query("role_id"); rid != "" {
		id, err := strconv.ParseUint(rid, 10, 64)
		if err == nil {
			roleID = &id
		}
	}

	users, total, err := h.userUsecase.GetAll(c.Request.Context(), page, pageSize, search, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        users,
		"total_count": total,
	})
}

// GET /users/:id - Get user by ID
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	user, err := h.userUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// POST /users - Create new user
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		IsActive: true,
		RoleID:   req.RoleID,
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	created, err := h.userUsecase.CreateUser(c.Request.Context(), user, req.Password)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": created})
}

// PATCH /users/:id - Update user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Password != nil {
		updates["password"] = *req.Password
	}
	if req.RoleID != nil {
		updates["role_id"] = *req.RoleID
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	updated, err := h.userUsecase.UpdateUser(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updated})
}

// DELETE /users/:id - Delete user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.userUsecase.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// --- Password Management ---

// POST /users/forgot-password
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userUsecase.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Silakan lanjutkan ke langkah reset password dengan token ini.",
		"token":   token,
	})
}

// POST /users/reset-password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userUsecase.ResetPassword(c.Request.Context(), req.Email, req.NewPassword, req.Token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diatur ulang. Silakan login dengan password baru."})
}
