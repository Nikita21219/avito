package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"main/internal/cache"
	"main/internal/e"
	"main/internal/segment"
	"main/internal/user"
	"net/http"
	"strconv"
)

// getActiveSegments is a handler function responsible for retrieving the active segments of a user.
// The function checks the data in redis.
// If there is no necessary data or an error has occurred, then it makes a request to the database
// It takes the HTTP response writer, HTTP request, and a user repository as parameters.
func getActiveSegments(w http.ResponseWriter, r *http.Request, rdb cache.Repository, userRepo user.Repository) {
	userId, ok := r.URL.Query()["id"]
	if !ok || len(userId) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(userId[0])
	if err != nil || id <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	var us user.Segments
	redisKey := fmt.Sprintf("avito_user_%d", id)
	err = rdb.GetFromCache(ctx, redisKey, &us)
	if err != nil {
		log.Println("error to retrive cache:", err)
		u, err := userRepo.FindByUserId(ctx, id)
		if err != nil {
			var notFound *e.UserNotFoundError
			if errors.As(err, &notFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		us = *u
	}
	if us.UserId == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	usDto := user.SegmentsDto{
		UserId:   id,
		Segments: make([]segment.SegmentDto, 0, len(us.Segments)),
	}
	for _, dto := range us.Segments {
		usDto.Segments = append(usDto.Segments, segment.SegmentDto{Slug: dto.Slug})
	}

	data, err := json.Marshal(usDto)
	if err != nil {
		log.Println("Error marshal data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Println("Error write data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// addDelSegment is a handler function responsible for adding and deleting user segments.
// It takes the HTTP response writer, HTTP request, and a repository interface as parameters.
func addDelSegment(w http.ResponseWriter, r *http.Request, repo interface{}) {
	userRepo, ok := repo.(user.Repository)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx := context.Background()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var seg *user.SegmentsAddDelDto
	if err = json.Unmarshal(body, &seg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !seg.Valid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = userRepo.AddDelSegments(ctx, seg)
	var segmentsNotFoundError *e.SegmentsNotFoundError
	if errors.As(err, &segmentsNotFoundError) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	checkErrors(w, err)
}

// Users is a handler function that checks the request method and calls the appropriate handler.
// It takes the user repository and Redis client as parameters.
func Users(userRepo user.Repository, rdb cache.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getActiveSegments(w, r, rdb, userRepo)
		} else if r.Method == "POST" {
			IdempotentKeyMiddleware(rdb, addDelSegment, userRepo)(w, r)
		}
	}
}
