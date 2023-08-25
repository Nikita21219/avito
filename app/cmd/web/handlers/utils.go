package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
	"io"
	"log"
	"main/internal/e"
	"main/internal/segment"
	"net/http"
	"time"
)

type NextHandler func(w http.ResponseWriter, r *http.Request, repo interface{})

func IdempotentKeyMiddleware(rdb *redis.Client, next NextHandler, repo interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idempotentKey := r.Header.Get("Idempotency-Key")
		if idempotentKey == "" {
			log.Println("Idempotency-Key not found in request headers")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		val, err := rdb.Exists(ctx, idempotentKey).Result()
		if err != nil {
			log.Println("error check idempotent key:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if val > 0 {
			log.Println("Idempotency-Key already processed")
			w.WriteHeader(http.StatusConflict)
			return
		}

		status := rdb.Set(ctx, idempotentKey, true, 60*60*time.Second)
		log.Println("Redis set Idempotency-Key status:", status)
		next(w, r, repo)
	}
}

func unmarshalSegment(w http.ResponseWriter, r *http.Request) (*segment.SegmentDto, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}
	defer r.Body.Close()

	var s *segment.SegmentDto
	if err = json.Unmarshal(body, &s); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	if !s.Valid() {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("segment not valid")
	}

	return s, nil
}

func checkErrors(w http.ResponseWriter, err error) {
	var dse *e.DuplicateSegmentError
	if errors.As(err, &dse) {
		w.WriteHeader(http.StatusConflict)
		return
	} else if err != nil {
		log.Println("error to create segment:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func RateLimiter(next http.HandlerFunc) http.HandlerFunc {
	limiter := rate.NewLimiter(10, 10)
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		} else {
			next(w, r)
		}
	}
}
