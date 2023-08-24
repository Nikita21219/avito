package pkg

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"main/internal/config"
)

type RedisClient interface {
}

func NewRedisClient(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisCfg.Host, cfg.RedisCfg.Port),
		Password: "",
		DB:       0,
	})

	status := rdb.Ping(ctx)
	if err := status.Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}
