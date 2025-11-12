package storage

import (
	"context"
	"time"

	"api-gateway/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
}

func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{client}, nil
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}
