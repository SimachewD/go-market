package handlers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func SlowHandler(c *gin.Context) {
    fmt.Println("Started slow request")
    time.Sleep(20 * time.Second)
    c.String(200, "Done!")
    fmt.Println("Finished slow request")
}
