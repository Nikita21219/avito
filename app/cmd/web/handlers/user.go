package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"main/internal/e"
	"main/internal/segment"
	"main/internal/user"
	"net/http"
	"strconv"
)

// getActiveSegments is a handler function responsible for retrieving the active segments of a user.
// It takes the HTTP response writer, HTTP request, and a user repository as parameters.
func getActiveSegments(w http.ResponseWriter, r *http.Request, userRepo user.Repository) {
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
	us, err := userRepo.FindByUserId(ctx, id)
	if err != nil {
		var notFound *e.UserNotFoundError
		if errors.As(err, &notFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	usDto := user.SegmentsDto{
		UserId:   id,
		Segments: make([]segment.SegmentDto, 0, len(us.Segments)),
	}
	for _, dto := range us.Segments {
		usDto.Segments = append(usDto.Segments, segment.SegmentDto{Slug: dto.Slug})
	}

	data, e := json.Marshal(usDto)
	if e != nil {
		log.Println("Error marshal data:", e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, e = w.Write(data)
	if e != nil {
		log.Println("Error write data:", e)
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
	checkErrors(w, err)
}

// Users is a handler function that checks the request method and calls the appropriate handler.
// It takes the user repository and Redis client as parameters.
func Users(userRepo user.Repository, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getActiveSegments(w, r, userRepo)
		} else if r.Method == "POST" {
			IdempotentKeyMiddleware(rdb, addDelSegment, userRepo)(w, r)
		}
	}
}
