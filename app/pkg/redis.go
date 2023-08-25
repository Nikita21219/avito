package pkg

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"main/internal/config"
	"main/pkg/utils"
	"time"
)

func NewRedisClient(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	var rdb *redis.Client

	err := utils.DoWithTries(func() error {
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.RedisCfg.Host, cfg.RedisCfg.Port),
			Password: "",
			DB:       0,
		})
		status := rdb.Ping(ctx)
		if err := status.Err(); err != nil {
			return err
		}
		return nil
	}, 3, 5*time.Second)

	return rdb, err
}
