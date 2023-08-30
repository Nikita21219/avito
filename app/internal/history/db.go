package history

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"main/pkg"
	"strings"
	"time"
)

type repository struct {
	client pkg.DBClient
}

// Create is a method that adds a new segment to the segments table.
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

func (r *repository) GetFromDate(ctx context.Context, date time.Time) ([]HistoryDto, error) {
	q := `SELECT user_id, slug, operation, date
		  FROM history JOIN segments s on s.segment_id = history.segment_id
          WHERE date >= $1;`

	rows, err := r.client.Query(ctx, q, date.Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]HistoryDto, 0)
	for rows.Next() {
		var dto HistoryDto
		var d time.Time
		if err = rows.Scan(&dto.UserId, &dto.SegmentSlug, &dto.Operation, &d); err != nil {
			return nil, err
		}
		dto.Date = d.Format("2006-01-02 15:04:05")
		res = append(res, dto)
	}

	return res, nil
}

func NewRepo(client pkg.DBClient) Repository {
	return &repository{
		client: client,
	}
}
