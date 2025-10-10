package api

import (
	"go-market/internal/api/handlers"
	"go-market/internal/api/middleware"
	"go-market/internal/jobs"
	"go-market/internal/repo/cache"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, redis *cache.RedisClient, jwtSecret string, queue *jobs.JobQueue) *gin.Engine {
    r := gin.Default()

    // Public routes
    r.GET("/health", handlers.HealthCheck)

    authGroup := r.Group("/auth")
    {
        authGroup.POST("/register", handlers.RegisterUser(db))
        authGroup.POST("/login", handlers.LoginUser(db, jwtSecret))
    }

    // Protected routes
    productGroup := r.Group("/products")
    productGroup.Use(middleware.JWTAuth(jwtSecret))
    {
        productGroup.POST("", handlers.CreateProduct(db))
        productGroup.GET("", handlers.ListProducts(db))
        productGroup.GET("/:id", handlers.GetProduct(db))
        productGroup.PUT("/:id", handlers.UpdateProduct(db))
        productGroup.DELETE("/:id", handlers.DeleteProduct(db))
    }

    orderGroup := r.Group("/orders")
    orderGroup.Use(middleware.JWTAuth(jwtSecret))
    {
        orderGroup.POST("", handlers.CreateOrder(db, queue))
        orderGroup.GET("", handlers.ListOrders(db))
        orderGroup.GET("/:id", handlers.GetOrder(db))
    }

    // wsGroup := r.Group("/ws")
    // wsGroup.Use(middleware.JWTAuth(jwtSecret))
    // {
    //     wsGroup.GET("", handlers.WebSocketHandler(wsManager))
    // }

    return r
}
