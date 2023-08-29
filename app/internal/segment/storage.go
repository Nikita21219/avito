package segment

import "context"

//go:generate mockgen -source=storage.go -destination=mocks/mock.go
type Repository interface {
	Create(ctx context.Context, segment *Segment) error
	Delete(ctx context.Context, slug string) error
}
