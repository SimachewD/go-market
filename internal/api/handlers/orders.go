package handlers

import (
    "net/http"

    "go-market/internal/models"
    "go-market/internal/jobs"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type CreateOrderRequest struct {
    ProductID uint `json:"product_id" binding:"required"`
    Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

func CreateOrder(db *gorm.DB, queue *jobs.JobQueue) gin.HandlerFunc {
    return func(c *gin.Context) {
        userIDRaw, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
            return
        }
        userID := userIDRaw.(uint)

        var req CreateOrderRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        order := models.Order{
            UserID:    userID,
            ProductID: req.ProductID,
            Quantity:  req.Quantity,
            Status:    "pending",
        }

        if err := db.Create(&order).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
            return
        }

        // Enqueue job for background processing
        queue.Enqueue(order.ID)

        c.JSON(http.StatusCreated, gin.H{"order_id": order.ID, "status": order.Status})
    }
}

func ListOrders(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userIDRaw, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
            return
        }
        userID := userIDRaw.(uint)

        var orders []models.Order
        db.Where("user_id = ?", userID).Find(&orders)
        c.JSON(http.StatusOK, orders)
    }
}
