package handler

import (
	"context"
	"net/http"
	"time"

	"api-gateway/pkg/storage"
	"api-gateway/pkg/utils"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	redis *storage.RedisClient
	mongo *storage.MongoClient
}

func NewHealthHandler(redis *storage.RedisClient, mongo *storage.MongoClient) *HealthHandler {
	return &HealthHandler{
		redis: redis,
		mongo: mongo,
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	utils.SuccessResponse(c, http.StatusOK, "Service is healthy", gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "Redis unavailable")
		return
	}

	// Check MongoDB
	if err := h.mongo.Ping(ctx, nil); err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "MongoDB unavailable")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Service is ready", gin.H{
		"status": "ready",
	})
}
