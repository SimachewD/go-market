package api

import (
	"go-market/internal/api/handlers"
	"go-market/internal/api/middleware"
	"go-market/internal/jobs"
	"go-market/internal/repo/cache"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(db *gorm.DB, redis *cache.RedisClient, jwtSecret string, queue *jobs.JobQueue) *gin.Engine {
    r := gin.Default()

    // Rate limiter: 10 requests per 10 seconds per IP
	rateLimiter := middleware.NewRateLimiter(10, 10*time.Second)
	r.Use(rateLimiter.Limit())

    // Public routes
    r.GET("/", handlers.SlowHandler)
    r.GET("/health", handlers.HealthCheck)

    authGroup := r.Group("/auth")
    {
        authGroup.POST("/register", handlers.RegisterUser(db))
        authGroup.POST("/login", handlers.LoginUser(db, jwtSecret))
    }

	// Protected routes
    adminGroup := r.Group("/admin")
    adminGroup.Use(middleware.JWTAuth(jwtSecret), middleware.RequireAdmin())
	{
		adminGroup.GET("/deadletters", handlers.ListDeadLetters(db))
		adminGroup.POST("/reprocess/:id", handlers.ReprocessOrder(db, queue))
	}

    userGroup := r.Group("/products")
    userGroup.Use(middleware.JWTAuth(jwtSecret))
    {
        userGroup.POST("/", handlers.CreateProduct(db))
        userGroup.GET("/", handlers.ListProducts(db))
        userGroup.GET("/:id", handlers.GetProduct(db))
        userGroup.PUT("/:id", handlers.UpdateProduct(db))
        userGroup.DELETE("/:id", handlers.DeleteProduct(db))
    }


    orderGroup := r.Group("/orders")
    orderGroup.Use(middleware.JWTAuth(jwtSecret))
    {
        orderGroup.POST("", handlers.CreateOrder(db, queue))
        orderGroup.GET("", handlers.ListOrders(db))
        orderGroup.GET("/:id", handlers.GetOrder(db))
    }

    metricsGroup := r.Group("/metrics")
    metricsGroup.Use(middleware.JWTAuth(jwtSecret), middleware.RequireAdmin())
    {
        metricsGroup.GET("", gin.WrapH(promhttp.Handler()))
    }

    // wsGroup := r.Group("/ws")
    // wsGroup.Use(middleware.JWTAuth(jwtSecret))
    // {
    //     wsGroup.GET("", handlers.WebSocketHandler(wsManager))
    // }

    return r
}
