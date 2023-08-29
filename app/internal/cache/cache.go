package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/redis/go-redis/v9"
	"log"
	"main/internal/user"
	"time"
)

type repository struct {
	client *redis.Client
}

// AddToCache adds data to the Redis cache with the specified key and expiration time.
// It converts the data to JSON format and stores it in the cache using the provided Redis client.
// The function returns an error if there's an issue with data marshaling or cache insertion.
func (r *repository) AddToCache(ctx context.Context, key string, data interface{}, exp time.Duration) error {
	value, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = r.client.Set(ctx, key, value, exp).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetFromCache retrieves data from the Redis cache using the specified key.
// It retrieves the data from the cache and unmarshals it from JSON format to the provided data structure.
// The function returns an error if there's an issue with cache retrieval or unmarshaling.
func (r *repository) GetFromCache(ctx context.Context, key string, data interface{}) error {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(value), data)
	if err != nil {
		return err
	}

	return nil
}

// UpdateCache periodically updates the Redis cache with user data.
// It uses the provided Redis client and user repository to refresh the cache.
// The cache is refreshed every minute.
func (r *repository) UpdateCache(ctx context.Context, userRepo user.Repository) {
	s := gocron.NewScheduler(time.UTC)

	_, err := s.Every(1).Minutes().Do(func() error {
		users, err := userRepo.FindAll(ctx)
		if err != nil {
			log.Println("error to get users:", err)
			return err
		}

		for _, u := range users {
			us, err := userRepo.FindByUserId(ctx, u.Id)
			if err != nil {
				log.Println("error to find active users segments:", err)
				continue
			}
			key := fmt.Sprintf("avito_user_%d", us.UserId)
			err = r.AddToCache(ctx, key, us, 5*time.Minute)
			if err != nil {
				log.Println("error to add cache in redis:", err)
				continue
			}
		}
		return nil
	})

	if err != nil {
		log.Println("error update cache:", err)
	}
	s.StartAsync()
}

func (r *repository) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

func (r *repository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.client.Set(ctx, key, value, expiration)
}

func NewRepo(client *redis.Client) Repository {
	return &repository{
		client: client,
	}
}
