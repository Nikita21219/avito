package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"main/internal/e"
	"main/internal/segment"
	"main/internal/user"
	"net/http"
	"strconv"
)

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

func addDelSegment(w http.ResponseWriter, r *http.Request, userRepo user.Repository) {
	// TODO add idempotent key

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

func Users(userRepo user.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getActiveSegments(w, r, userRepo)
			return
		} else if r.Method == "POST" {
			addDelSegment(w, r, userRepo)
			return
		}
	}
}