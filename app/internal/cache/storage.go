package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"main/internal/user"
	"time"
)

//go:generate mockgen -source=storage.go -destination=mocks/mock.go
type Repository interface {
	AddToCache(ctx context.Context, key string, data interface{}, exp time.Duration) error
	GetFromCache(ctx context.Context, key string, data interface{}) error
	UpdateCache(ctx context.Context, userRepo user.Repository)
	Exists(ctx context.Context, keys ...string) (int64, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}
