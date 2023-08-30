package history

import (
	"context"
	"github.com/jackc/pgx/v4"
	"time"
)

//go:generate mockgen -source=storage.go -destination=mocks/mock.go
type Repository interface {
	Create(ctx context.Context, history *History, tx pgx.Tx) error
	GetFromDate(ctx context.Context, date time.Time) ([]HistoryDto, error)
}
