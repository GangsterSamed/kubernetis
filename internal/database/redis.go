package database

import (
	"context"
	"log/slog"

	"github.com/polzovatel/todo-learning/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.Config, logger *slog.Logger) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})

	// Проверяем подключение
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Warn("redis ping err", slog.Any("error", err))
		return nil
	}

	logger.Info("redis connect success")
	return redisClient
}
