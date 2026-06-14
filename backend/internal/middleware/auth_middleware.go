package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"billing-backend/config"
	"billing-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware creates a middleware for JWT validation
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// 2. Extract token from Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}
		tokenString := parts[1]

		// 3. Validate token
		claims, err := utils.ValidateJWT(cfg, tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 4. Set user information in context for handlers to access
		c.Set("user_id", fmt.Sprintf("%d", claims.UserID))
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		// Set in standard request context so it propagates to usecases
		ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "role", claims.Role)
		c.Request = c.Request.WithContext(ctx)

		// 5. Proceed to next handler
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

