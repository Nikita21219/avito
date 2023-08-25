package user

import "context"

type Repository interface {
	FindByUserId(ctx context.Context, userId int) (*Segments, error)
	AddDelSegments(ctx context.Context, s *SegmentsAddDelDto) error
}
