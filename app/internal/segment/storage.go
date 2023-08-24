package segment

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, segment *Segment) error
	//FindOne(ctx context.Context, id int) (Order, error)
}
