package segment

import (
	"context"
	"main/internal/order_complete"
	"time"
)

type Repository interface {
	CreateAll(ctx context.Context, orders []*Order) error
	FindByLimitAndOffset(ctx context.Context, l, o int) (order []Order, err error)
	FindOne(ctx context.Context, id int) (Order, error)
	Update(ctx context.Context, o Order, oc *order_complete.OrderCompleteDto) error
	FindAllInTimeInterval(ctx context.Context, startDate, endDate time.Time, courierId int) ([]Order, error)
}
