package handlers

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"main/internal/segment"
	"net/http"
)

func createSegment(w http.ResponseWriter, r *http.Request, repo interface{}) {
	segmentRepo, ok := repo.(segment.Repository)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	s, err := unmarshalSegment(w, r)
	if err != nil {
		return
	}

	err = segmentRepo.Create(ctx, &segment.Segment{Slug: s.Slug})
	checkErrors(w, err)
}

func deleteSegment(w http.ResponseWriter, r *http.Request, repo interface{}) {
	segmentRepo, ok := repo.(segment.Repository)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	s, err := unmarshalSegment(w, r)
	if err != nil {
		return
	}

	if err = segmentRepo.Delete(ctx, s.Slug); err != nil {
		log.Println("error to delete segment:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func Segments(segmentRepo segment.Repository, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			IdempotentKeyMiddleware(rdb, createSegment, segmentRepo)(w, r)
		} else if r.Method == "DELETE" {
			IdempotentKeyMiddleware(rdb, deleteSegment, segmentRepo)(w, r)
		}
	}
}
