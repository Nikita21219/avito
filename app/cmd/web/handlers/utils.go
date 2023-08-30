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
	"net/url"
	"time"
)

type NextHandler func(w http.ResponseWriter, r *http.Request, repo interface{}, historyRepo history.Repository)

// IdempotentKeyMiddleware is a middleware function that checks the idempotency key in Redis
// before invoking the next handler function.
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

// UniqueKey generates and returns a unique string using the UUID4.
func UniqueKey() string {
	return uuid.New().String()
}

// GetDateQuery extracts and parses a date from the provided URL query parameters.
// The function expects a single "date" parameter in the query, formatted as "2006-01-02 15:04".
func GetDateQuery(query url.Values) (time.Time, error) {
	date, ok := query["date"]
	if !ok || len(date) != 1 {
		return time.Time{}, fmt.Errorf("bad request")
	}

	t, err := time.Parse("2006-01-02 15:04", date[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("bad request")
	}
	return t, nil
}
