package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	// "strconv"

	"go-market/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateProductRequest struct {
    Name        string  `json:"name" binding:"required"`
    Description string  `json:"description"`
    Price       float64 `json:"price" binding:"gt=0"`
    Stock       int     `json:"stock" binding:"gte=0"`
}

func CreateProduct(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req CreateProductRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        product := models.Product{
            Name:        req.Name,
            Description: req.Description,
            Price:       req.Price,
            Stock:       req.Stock,
        }

        if err := db.Create(&product).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
            return
        }

        c.JSON(http.StatusCreated, product)
    }
}

func ListProducts(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Query params
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
        sort := c.DefaultQuery("sort", "id")        // e.g., price, name
        order := c.DefaultQuery("order", "asc")     // asc or desc
        minPrice, _ := strconv.ParseFloat(c.DefaultQuery("min_price", "0"), 64)
        maxPrice, _ := strconv.ParseFloat(c.DefaultQuery("max_price", "0"), 64)

        var products []models.Product
        query := db.Model(&models.Product{})

        // Filter by price range
        if minPrice > 0 {
            query = query.Where("price >= ?", minPrice)
        }
        if maxPrice > 0 {
            query = query.Where("price <= ?", maxPrice)
        }

        // Sorting
        query = query.Order(fmt.Sprintf("%s %s", sort, order))

        // Pagination
        offset := (page - 1) * limit
        query = query.Offset(offset).Limit(limit)

        if err := query.Find(&products).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
            return
        }

        c.JSON(http.StatusOK, products)
    }
}

func GetProduct(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var product models.Product
        if err := db.First(&product, id).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
            return
        }
        c.JSON(http.StatusOK, product)
    }
}

func UpdateProduct(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var product models.Product
        if err := db.First(&product, id).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
            return
        }

        var req CreateProductRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        product.Name = req.Name
        product.Description = req.Description
        product.Price = req.Price
        product.Stock = req.Stock

        if err := db.Save(&product).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
            return
        }

        c.JSON(http.StatusOK, product)
    }
}

func DeleteProduct(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        if err := db.Delete(&models.Product{}, id).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
    }
}
