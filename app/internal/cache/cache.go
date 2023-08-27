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

// AddToCache adds data to the Redis cache with the specified key and expiration time.
// It converts the data to JSON format and stores it in the cache using the provided Redis client.
// The function returns an error if there's an issue with data marshaling or cache insertion.
func AddToCache(ctx context.Context, rdb *redis.Client, key string, data interface{}, exp time.Duration) error {
	value, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = rdb.Set(ctx, key, value, exp).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetFromCache retrieves data from the Redis cache using the specified key.
// It retrieves the data from the cache and unmarshals it from JSON format to the provided data structure.
// The function returns an error if there's an issue with cache retrieval or unmarshaling.
func GetFromCache(ctx context.Context, rdb *redis.Client, key string, data interface{}) error {
	value, err := rdb.Get(ctx, key).Result()
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
func UpdateCache(ctx context.Context, rdb *redis.Client, userRepo user.Repository) {
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
			err = AddToCache(ctx, rdb, key, us, 5*time.Minute)
			if err != nil {
				log.Println("error to add cache in redis:", err)
				continue
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("error update cache:", err)
	}
	s.StartAsync()
}
