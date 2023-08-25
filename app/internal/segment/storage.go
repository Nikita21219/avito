package segment

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, segment *Segment) error
	Delete(ctx context.Context, name string) error
}
