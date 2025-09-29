package handlers

import (
    "time"
    "github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "ok",
        "time":   time.Now(),
    })
}
