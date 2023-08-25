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
	q := `INSERT INTO segments (segment_name) VALUES ($1) RETURNING segment_id`
	_, err := r.client.Exec(ctx, q, segment.Name)
	if e.IsDuplicateError(err) {
		return &e.DuplicateSegmentError{SegmentName: segment.Name}
	}
	return err
}

func (r *repository) Delete(ctx context.Context, name string) error {
	q := `DELETE FROM segments WHERE segment_name = $1`
	_, err := r.client.Exec(ctx, q, name)
	return err
}

func NewRepo(client pkg.DBClient) Repository {
	return &repository{
		client: client,
	}
}
