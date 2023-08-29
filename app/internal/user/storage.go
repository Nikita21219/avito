package user

import "context"

//go:generate mockgen -source=storage.go -destination=mocks/mock.go
type Repository interface {
	FindAll(ctx context.Context) ([]*User, error)
	FindByUserId(ctx context.Context, userId int) (*Segments, error)
	AddDelSegments(ctx context.Context, s *SegmentsAddDelDto) error
	CreateUser(ctx context.Context) (int, error)
	DelUser(ctx context.Context, userId int) error
	GetMaxId(ctx context.Context) (int, error)
	DeleteSegmentsEveryDay(ctx context.Context)
}
