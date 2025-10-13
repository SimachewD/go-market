package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireAdmin ensures the user has admin rights
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdminRaw, exists := c.Get("isAdmin")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user info not found"})
			return
		}

		isAdmin := isAdminRaw.(bool)
		if !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		c.Next()
	}
}
