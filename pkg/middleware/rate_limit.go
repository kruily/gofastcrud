package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kruily/gofastcrud/pkg/errors"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
	ips    map[string]*rate.Limiter
	mu     *sync.RWMutex
	rate   rate.Limit
	burst  int
	ttl    time.Duration
	lastGC time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(r rate.Limit, burst int, ttl time.Duration) *RateLimiter {
	return &RateLimiter{
		ips:    make(map[string]*rate.Limiter),
		mu:     &sync.RWMutex{},
		rate:   r,
		burst:  burst,
		ttl:    ttl,
		lastGC: time.Now(),
	}
}

// RateLimit 限流中间件
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取IP
		ip := c.ClientIP()

		// 清理过期限流器
		rl.cleanupStale()

		// 获取限流器
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			c.Error(errors.New(errors.ErrRateLimit, "Too many requests"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// getLimiter 获取IP对应的限流器
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.ips[ip] = limiter
	}

	return limiter
}

// cleanupStale 清理过期的限流器
func (rl *RateLimiter) cleanupStale() {
	if time.Since(rl.lastGC) < rl.ttl {
		return
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip := range rl.ips {
		if time.Since(rl.lastGC) > rl.ttl {
			delete(rl.ips, ip)
		}
	}
	rl.lastGC = now
}
