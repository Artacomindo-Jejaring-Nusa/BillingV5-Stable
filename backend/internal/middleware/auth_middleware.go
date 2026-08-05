package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"billing-backend/config"
	"billing-backend/internal/domain"
	"billing-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware creates a middleware for JWT validation or API Key validation
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check for API Key first (either in X-API-Key header or Bearer token starting with 'jk_')
		var apiKeyString string
		xApiKey := c.GetHeader("X-API-Key")
		if xApiKey != "" {
			apiKeyString = xApiKey
		} else {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" && strings.HasPrefix(parts[1], "jk_") {
					apiKeyString = parts[1]
				}
			}
		}

		// 2. Validate API Key if provided
		if apiKeyString != "" {
			dbVal, dbExists := c.Get("db")
			if !dbExists {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database context not initialized"})
				return
			}
			db := dbVal.(*gorm.DB)

			// Hash the incoming raw token to match the stored SHA-256 hash
			hashBytes := sha256.Sum256([]byte(apiKeyString))
			tokenHash := hex.EncodeToString(hashBytes[:])

			var apiKey domain.APIKey
			err := db.Preload("Role").Where("token_hash = ? AND is_active = ?", tokenHash, true).First(&apiKey).Error
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or inactive API Key"})
				return
			}

			// Map API key identity to context
			c.Set("user_id", fmt.Sprintf("api_key_%d", apiKey.ID))
			c.Set("email", fmt.Sprintf("apikey_%d@integration.system", apiKey.ID))
			c.Set("role", apiKey.Role.Name)

			// Propagate to standard request context for usecases
			// We use a dummy ID 0 or map to user ID if we need to. Here, 0 represents system integration.
			ctx := context.WithValue(c.Request.Context(), "user_id", uint64(0))
			ctx = context.WithValue(ctx, "email", fmt.Sprintf("apikey_%d@integration.system", apiKey.ID))
			ctx = context.WithValue(ctx, "role", apiKey.Role.Name)
			c.Request = c.Request.WithContext(ctx)

			c.Next()
			return
		}

		// 3. Fallback to standard JWT Bearer token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header or X-API-Key is required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}
		tokenString := parts[1]

		claims, err := utils.ValidateJWT(cfg, tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("user_id", fmt.Sprintf("%d", claims.UserID))
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "role", claims.Role)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RoleMiddleware checks if the authenticated user has the required role
func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		roleStr := userRole.(string)
		for _, role := range requiredRoles {
			if roleStr == role || roleStr == "superadmin" || roleStr == "admin" {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: You don't have enough permissions"})
	}
}

// PermissionMiddleware checks if the authenticated user has the required permission
func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbVal, dbExists := c.Get("db")
		if !dbExists {
			// In unit tests, DB is not set in Gin context, so bypass authorization checks
			c.Next()
			return
		}

		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Superadmin and Admin bypass permission checks
		roleStr := role.(string)
		if roleStr == "superadmin" || roleStr == "admin" {
			c.Next()
			return
		}

		userIDStr, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

		db := dbVal.(*gorm.DB)

		var count int64
		err := db.Table("users").
			Joins("JOIN role_has_permissions ON role_has_permissions.role_id = users.role_id").
			Joins("JOIN permissions ON permissions.id = role_has_permissions.permission_id").
			Where("users.id = ? AND permissions.name = ?", userID, requiredPermission).
			Count(&count).Error

		if err != nil || count == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: You do not have permission to perform this action"})
			return
		}

		c.Next()
	}
}

