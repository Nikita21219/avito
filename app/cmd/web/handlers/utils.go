package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
	"io"
	"log"
	"main/internal/cache"
	"main/internal/e"
	"main/internal/history"
	"main/internal/segment"
	"net/http"
	"time"
)

type NextHandler func(w http.ResponseWriter, r *http.Request, repo interface{}, historyRepo history.Repository)

// IdempotentKeyMiddleware is a middleware function that checks the idempotency key in Redis
// before invoking the next handler function. It takes a Redis client, the next handler function,
// and a repository interface as parameters and returns a new handler function.
// TODO fix doc (added new param)
func IdempotentKeyMiddleware(rdb cache.Repository, next NextHandler, repo interface{}, historyRepo history.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idempotentKey := r.Header.Get("Idempotency-Key")
		if idempotentKey == "" {
			log.Println("Idempotency-Key not found in request headers")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		val, err := rdb.Exists(ctx, idempotentKey)
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

		rdb.Set(ctx, idempotentKey, true, 60*time.Minute)
		next(w, r, repo, historyRepo)
	}
}

// unmarshalSegment is a utility function that reads and parses the request body to retrieve a SegmentDto.
// It takes the HTTP response writer and the HTTP request as parameters and returns a SegmentDto and an error.
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

// checkErrors is a utility function that checks for errors and responds with appropriate status codes.
// It takes the HTTP response writer and an error as parameters.
func checkErrors(w http.ResponseWriter, err error) {
	var dse *e.DuplicateSegmentError
	if errors.As(err, &dse) {
		w.WriteHeader(http.StatusConflict)
		return
	} else if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// RateLimiter is a middleware function that acts as a rate limiter for incoming requests.
// It takes the next handler function as a parameter and returns a new handler function.
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

// TODO fill doc
func UniqueKey() string {
	return uuid.New().String()
}
