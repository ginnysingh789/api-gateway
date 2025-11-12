package middleware

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"api-gateway/internal/config"
	"api-gateway/pkg/storage"

	"github.com/gin-gonic/gin"
)

func RateLimiter(redisClient *storage.RedisClient, cfg config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s", ip)

		ctx := context.Background()
		pipe := redisClient.TxPipeline()

		// Get current bucket state
		bucketState := pipe.HGetAll(ctx, key)
		pipe.Exec(ctx)

		bucketData, err := bucketState.Result()
		if err != nil || len(bucketData) == 0 {
			// New user - initialize bucket
			initialTokens := cfg.Requests - 1
			pipe = redisClient.TxPipeline()
			pipe.HSet(ctx, key, "tokens", initialTokens)
			pipe.HSet(ctx, key, "timestamp", time.Now().Unix())
			pipe.Expire(ctx, key, cfg.Window*2)
			_, err := pipe.Exec(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter error"})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// Parse bucket data
		tokensStr := bucketData["tokens"]
		timestampStr := bucketData["timestamp"]

		tokens, _ := strconv.ParseFloat(tokensStr, 64)
		lastTimestamp, _ := strconv.ParseInt(timestampStr, 10, 64)

		// Calculate token refill
		now := time.Now().Unix()
		elapsed := now - lastTimestamp
		refillRate := float64(cfg.Requests) / cfg.Window.Seconds()
		tokensToAdd := float64(elapsed) * refillRate
		tokens = math.Min(float64(cfg.Requests), tokens+tokensToAdd)

		if tokens < 1 {
			retryAfter := int(math.Ceil((1 - tokens) / refillRate))
			c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Requests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(now+int64(retryAfter), 10))
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		// Consume token
		tokens -= 1
		pipe = redisClient.TxPipeline()
		pipe.HSet(ctx, key, "tokens", tokens)
		pipe.HSet(ctx, key, "timestamp", now)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter error"})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(int(tokens)))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(now+int64(cfg.Window.Seconds()), 10))

		c.Next()
	}
}
