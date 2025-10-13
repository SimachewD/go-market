package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     int           // allowed requests per window
	window   time.Duration // time window
}

type visitor struct {
	lastSeen time.Time
	tokens   int
}

func NewRateLimiter(rate int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// Background cleanup to remove old visitors
	go func() {
		for {
			time.Sleep(window)
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, v := range rl.visitors {
		if now.Sub(v.lastSeen) > rl.window*2 {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *rateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists {
			v = &visitor{lastSeen: time.Now(), tokens: rl.rate}
			rl.visitors[ip] = v
		}

		// refill tokens
		elapsed := time.Since(v.lastSeen)
		refill := int(elapsed / rl.window)
		if refill > 0 {
			v.tokens = rl.rate
			v.lastSeen = time.Now()
		}

		// consume a token
		if v.tokens > 0 {
			v.tokens--
			v.lastSeen = time.Now()
			rl.mu.Unlock()
			c.Next()
			return
		}

		rl.mu.Unlock()

		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": "too many requests, slow down",
		})
	}
}
