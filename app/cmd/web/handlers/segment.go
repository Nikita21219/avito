package handlers

import (
	"context"
	"errors"
	"log"
	"main/internal/e"
	"main/internal/segment"
	"net/http"
)

func createSegment(w http.ResponseWriter, r *http.Request, segmentRepo segment.Repository) {
	// TODO add idempotent key
	ctx := context.Background()
	s, err := unmarshalSegment(w, r)
	if err != nil {
		return
	}

	err = segmentRepo.Create(ctx, &segment.Segment{Slug: s.Slug})
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

func deleteSegment(w http.ResponseWriter, r *http.Request, segmentRepo segment.Repository) {
	// TODO add idempotent key
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

func Segments(segmentRepo segment.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			createSegment(w, r, segmentRepo)
		} else if r.Method == "DELETE" {
			deleteSegment(w, r, segmentRepo)
		}
	}
}
