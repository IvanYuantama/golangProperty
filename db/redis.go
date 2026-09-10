package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis(ctx context.Context) (*redis.Client, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable is not set")
	}

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse REDIS_URL: %w", err)
	}
	options.DialTimeout = 2 * time.Second
	options.ReadTimeout = 750 * time.Millisecond
	options.WriteTimeout = 750 * time.Millisecond
	options.MaxRetries = 1

	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("unable to ping Redis: %w", err)
	}

	RedisClient = client
	return RedisClient, nil
}
