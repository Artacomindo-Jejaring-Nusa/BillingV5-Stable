package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"billing-backend/config"
	"billing-backend/pkg/utils"

	"github.com/gin-gonic/gin"
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
