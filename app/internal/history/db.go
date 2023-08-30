package history

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"main/pkg"
	"strings"
)

type repository struct {
	client pkg.DBClient
}

// Create is a method that adds a new segment to the segments table.
// It takes a context and a Segment pointer as parameters and returns an error.
func (r *repository) Create(ctx context.Context, history *History, tx pgx.Tx) error {
	if len(history.SegmentIds) == 0 {
		return nil
	}

	q := `INSERT INTO history (user_id, segment_id, operation, date) VALUES `

	values := make([]interface{}, 0, len(history.SegmentIds))
	date := history.Date.Format("2006-01-02 15:04:05")
	for _, segmentId := range history.SegmentIds {
		values = append(values, history.UserId, segmentId, history.Operation, date)
	}
	placeholders := make([]string, 0, len(values)/4)
	for i := 0; i < len(values)/4; i++ {
		placeholder := fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4)
		placeholders = append(placeholders, placeholder)
	}

	resQuery := fmt.Sprintf("%s %s", q, strings.Join(placeholders, ", "))
	_, err := tx.Exec(ctx, resQuery, values...)
	if err != nil {
		return err
	}

	return nil
}

func NewRepo(client pkg.DBClient) Repository {
	return &repository{
		client: client,
	}
}
