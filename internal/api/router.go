package api

import (
    "go-market/internal/api/handlers"
    "go-market/internal/repo/cache"
    // "go-market/internal/repo/postgres"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func NewRouter(db *gorm.DB, redis *cache.RedisClient, jwtSecret string) *gin.Engine {
    r := gin.Default()

    // Routes
    r.GET("/health", handlers.HealthCheck)
    authGroup := r.Group("/auth")
    {
        authGroup.POST("/register", handlers.RegisterUser(db))
        authGroup.POST("/login", handlers.LoginUser(db, jwtSecret))
    }

    return r
}

