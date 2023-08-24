package user

import "context"

type Repository interface {
	CreateOne(ctx context.Context, c *Courier) error
	FindAll(ctx context.Context) (c []Courier, err error)
	CreateAll(ctx context.Context, couriers []*Courier) error
	FindByLimitAndOffset(ctx context.Context, l, o int) (c []Courier, err error)
	FindOne(ctx context.Context, id int) (Courier, error)
}
