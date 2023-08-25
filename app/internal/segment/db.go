package segment

import (
	"context"
	"main/internal/e"
	"main/pkg"
)

type repository struct {
	client pkg.DBClient
}

func (r *repository) Create(ctx context.Context, segment *Segment) error {
	q := `INSERT INTO segments (slug) VALUES ($1) RETURNING segment_id`
	_, err := r.client.Exec(ctx, q, segment.Slug)
	if e.IsDuplicateError(err) {
		return &e.DuplicateSegmentError{SegmentName: segment.Slug}
	}
	return err
}

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
