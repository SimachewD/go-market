package middleware

import (
    "net/http"
    "strings"

    "go-market/internal/auth"

    "github.com/gin-gonic/gin"
)

func JWTAuth(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
            c.Abort()
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
            c.Abort()
            return
        }

        claims, err := auth.ValidateJWT(parts[1], secret)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
            c.Abort()
            return
        }

        // Store user ID in context
        c.Set("userID", claims.UserID)
        c.Set("isAdmin", claims.IsAdmin)
        c.Next()
    }
}
