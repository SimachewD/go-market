package handlers

import (
	"net/http"
	"strconv"

	"go-market/internal/jobs"
	"go-market/internal/models"

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

        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
        status := c.Query("status") // optional filter by order status

        var orders []models.Order
        query := db.Where("user_id = ?", userID)
        if status != "" {
            query = query.Where("status = ?", status)
        }

        offset := (page - 1) * limit
        query = query.Offset(offset).Limit(limit).Order("created_at desc")

        if err := query.Find(&orders).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
            return
        }

        c.JSON(http.StatusOK, orders)
    }
}

func GetOrder(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userIDRaw, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
            return
        }
        userID := userIDRaw.(uint)

        id := c.Param("id")
        if id == "" {
            c.JSON(400, gin.H{"error": "missing id"})
            return
        }
        
        var order models.Order
        if err := db.Where("user_id = ? AND id = ?", userID, id).First(&order).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
            return
        }

        c.JSON(http.StatusOK, order)
    }
}

func ListDeadLetters(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context){
        var orders []models.Order
        if err := db.Where("status = ?", "dead_letter").Find(&orders).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        c.JSON(http.StatusOK, gin.H{"dead_letters": orders})
    }
}

func ReprocessOrder(db *gorm.DB, queue *jobs.JobQueue) gin.HandlerFunc{
    return func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
            return
        }

        var order models.Order
        if err := db.First(&order, id).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
            return
        }

        if order.Status != "dead_letter" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "order is not in dead_letter state"})
            return
        }

        order.Status = "pending"
        order.RetryCount = 0
        db.Save(&order)

        queue.Enqueue(order.ID)

        c.JSON(http.StatusOK, gin.H{
            "message": "order requeued for processing",
            "order_id": order.ID,
        })
    }
}

