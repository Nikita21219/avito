package user

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]*User, error)
	FindByUserId(ctx context.Context, userId int) (*Segments, error)
	AddDelSegments(ctx context.Context, s *SegmentsAddDelDto) error
}
