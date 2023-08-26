package segment

import (
	"context"
	"main/internal/e"
	"main/pkg"
)

type repository struct {
	client pkg.DBClient
}

// Create is a method that adds a new segment to the segments table.
// It takes a context and a Segment pointer as parameters and returns an error.
func (r *repository) Create(ctx context.Context, segment *Segment) error {
	q := `INSERT INTO segments (slug) VALUES ($1) RETURNING segment_id`
	_, err := r.client.Exec(ctx, q, segment.Slug)
	if e.IsDuplicateError(err) {
		return &e.DuplicateSegmentError{SegmentName: segment.Slug}
	}
	return err
}

// Delete is a method that deletes a segment from the segments table based on the provided slug.
// It takes a context and a slug string as parameters and returns an error.
func (r *repository) Delete(ctx context.Context, slug string) error {
	q := `DELETE FROM segments WHERE slug = $1`
	_, err := r.client.Exec(ctx, q, slug)
	return err
}

func NewRepo(client pkg.DBClient) Repository {
	return &repository{
		client: client,
	}
}
