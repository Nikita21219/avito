package user

import (
	"context"
	"fmt"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"main/internal/e"
	"main/internal/segment"
	"main/pkg"
	"strings"
)

type repository struct {
	client pkg.DBClient
}

func (r *repository) FindByUserId(ctx context.Context, userId int) (*Segments, error) {
	q := `
		SELECT us.segment_id, slug 
		FROM segments JOIN user_segments us ON segments.segment_id = us.segment_id 
		WHERE user_id = $1;
	`

	rows, err := r.client.Query(ctx, q, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	segments := make([]*segment.Segment, 0)

	for rows.Next() {
		var s segment.Segment
		err = rows.Scan(&s.Id, &s.Slug)
		if err != nil {
			return nil, err
		}
		segments = append(segments, &s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	us := &Segments{
		UserId:   userId,
		Segments: segments,
	}
	return us, nil
}

func getSegmentIdsBySlugs(ctx context.Context, tx pgx.Tx, slugs []string) ([]int, error) {
	q := `SELECT segment_id FROM segments WHERE slug = ANY($1);`
	slugsArr := pgtype.TextArray{}
	if err := slugsArr.Set(slugs); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, q, slugsArr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	segmentIds := make([]int, 0)
	for rows.Next() {
		var segmentId int
		err = rows.Scan(&segmentId)
		if err != nil {
			return nil, err
		}
		segmentIds = append(segmentIds, segmentId)
	}
	return segmentIds, nil
}

func addSegments(ctx context.Context, userId int, slugs []string, tx pgx.Tx) error {
	segmentIds, err := getSegmentIdsBySlugs(ctx, tx, slugs)
	if err != nil {
		return err
	}

	q := `INSERT INTO user_segments (user_id, segment_id) VALUES `

	values := make([]string, 0, len(segmentIds))
	for _, segmentId := range segmentIds {
		value := fmt.Sprintf("(%d, %d)", userId, segmentId)
		values = append(values, value)
	}

	resQuery := fmt.Sprintf("%s %s;", q, strings.Join(values, ", "))
	_, err = tx.Exec(ctx, resQuery)
	if e.IsDuplicateError(err) {
		return &e.DuplicateSegmentError{}
	}
	if err != nil {
		return err
	}
	return nil
}

func delSegments(ctx context.Context, userId int, slugs []string, tx pgx.Tx) error {
	segmentIds, err := getSegmentIdsBySlugs(ctx, tx, slugs)
	if err != nil {
		return err
	}

	var segmentIdsArray pgtype.Int4Array
	err = segmentIdsArray.Set(segmentIds)
	if err != nil {
		return err
	}

	q := `DELETE FROM user_segments WHERE user_id = $1 AND segment_id = ANY($2);`
	_, err = tx.Exec(ctx, q, userId, &segmentIdsArray)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) AddDelSegments(ctx context.Context, s *SegmentsAddDelDto) error {
	tx, err := r.client.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	if len(s.SegmentsAdd) > 0 {
		if err = addSegments(ctx, s.UserId, s.SegmentsAdd, tx); err != nil {
			return err
		}
	}
	if len(s.SegmentsDel) > 0 {
		if err = delSegments(ctx, s.UserId, s.SegmentsDel, tx); err != nil {
			return err
		}
	}

	return nil
}

func NewRepo(client pkg.DBClient) Repository {
	return &repository{
		client: client,
	}
}
