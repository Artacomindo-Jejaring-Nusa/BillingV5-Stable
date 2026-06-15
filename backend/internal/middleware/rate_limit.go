package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"billing-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

type ipLimit struct {
	count     int
	resetTime time.Time
}

var (
	localLimits = make(map[string]*ipLimit)
	localMu     sync.Mutex
	once        sync.Once
)

// RateLimitMiddleware limits requests per IP per minute per endpoint path
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	// Clean up local limits map periodically
	once.Do(func() {
		go func() {
			for {
				time.Sleep(5 * time.Minute)
				localMu.Lock()
				now := time.Now()
				for key, limit := range localLimits {
					if now.After(limit.resetTime) {
						delete(localLimits, key)
					}
				}
				localMu.Unlock()
			}
		}()
	})

	return func(c *gin.Context) {
		ip := c.ClientIP()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		redisClient := websocket.GetRedisClient()
		if redisClient != nil {
			// Redis-based Rate Limiting (Distributed)
			ctx := c.Request.Context()
			minuteKey := time.Now().Format("200601021504")
			key := fmt.Sprintf("rate_limit:%s:%s:%s", path, ip, minuteKey)

			count, err := redisClient.Incr(ctx, key).Result()
			if err == nil {
				if count == 1 {
					redisClient.Expire(ctx, key, 2*time.Minute)
				}
				if int(count) > requestsPerMinute {
					c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
						"error": "Too many requests. Please try again later.",
					})
					return
				}
			}
		} else {
			// In-memory Fallback (Local)
			limitKey := ip + ":" + path
			localMu.Lock()
			now := time.Now()
			limit, exists := localLimits[limitKey]
			if !exists || now.After(limit.resetTime) {
				localLimits[limitKey] = &ipLimit{
					count:     1,
					resetTime: now.Add(time.Minute),
				}
			} else {
				limit.count++
				if limit.count > requestsPerMinute {
					localMu.Unlock()
					c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
						"error": "Too many requests. Please try again later.",
					})
					return
				}
			}
			localMu.Unlock()
		}

		c.Next()
	}
}
