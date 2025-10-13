package handlers

import (
    "net/http"
    "time"

    "go-market/internal/auth"
    "go-market/internal/models"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// POST /admin/create
func CreateAdmin(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get values from context
        isAdminVal, exists := c.Get("isAdmin")
        if !exists || !isAdminVal.(bool) {
            c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
            return
        }

        var req RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
            return
        }

        admin := models.User{
            Username: req.Username,
            Email:    req.Email,
            Password: string(hashedPassword),
            IsAdmin:  true,
        }

        if err := db.Create(&admin).Error; err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusCreated, gin.H{"message": "Admin created successfully"})
    }
}

func RegisterUser(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Hash password
        hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
            return
        }

        // Bootstrap logic: make the very first user an admin
        // var count int64
        // db.Model(&models.User{}).Count(&count)
        // isAdmin := count == 0

        user := models.User{
            Username: req.Username,
            Email:    req.Email,
            Password: string(hashed),
            // IsAdmin: isAdmin,
        }

        if err := db.Create(&user).Error; err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusCreated, gin.H{"message": "user created"})
    }
}

func LoginUser(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req LoginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var user models.User
        if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
            return
        }

        if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
            return
        }

        token, err := auth.GenerateJWT(user.ID, user.IsAdmin, jwtSecret, 24*time.Hour)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"token": token})
    }
}
