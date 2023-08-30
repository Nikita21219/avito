package user

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq"
	"log"
	"main/internal/e"
	"main/internal/history"
	"main/internal/segment"
	"main/pkg"
	"strings"
	"time"
)

type repository struct {
	client pkg.DBClient
}

// FindAll retrieves all user IDs from the user_segments table in the repository.
// It executes a query to select all user IDs and returns a slice of User pointers.
func (r *repository) FindAll(ctx context.Context) ([]*User, error) {
	q := `SELECT user_id FROM user_segments;`
	rows, err := r.client.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*User, 0)

	for rows.Next() {
		var u User
		err = rows.Scan(&u.Id)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// FindByUserId is a method that retrieves segments associated with a user based on the provided user ID.
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

	if len(segments) == 0 {
		return nil, &e.UserNotFoundError{UserId: userId}
	}

	us := &Segments{
		UserId:   userId,
		Segments: segments,
	}
	return us, nil
}

// getSegmentIdsBySlugs is a function that retrieves segment IDs based on the provided segment slugs.
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

// addSegments is a function that adds the user to the specified segments.
func addSegments(ctx context.Context, userId int, slugs []string, ttlDays *int, historyRepo history.Repository, tx pgx.Tx) error {
	segmentIds, err := getSegmentIdsBySlugs(ctx, tx, slugs)
	if err != nil {
		return err
	}
	if len(segmentIds) == 0 {
		return &e.SegmentsNotFoundError{Slugs: slugs}
	}

	var aliveUntil pq.NullTime
	if ttlDays != nil {
		aliveUntil.Time = time.Now().AddDate(0, 0, *ttlDays)
		aliveUntil.Valid = true
	}

	aliveUntilStr := "NULL"
	if aliveUntil.Valid {
		aliveUntilStr = fmt.Sprintf("'%s'", aliveUntil.Time.Format("2006-01-02"))
	}

	q := `INSERT INTO user_segments (user_id, segment_id, alive_until) VALUES `

	values := make([]string, 0, len(segmentIds))
	for _, segmentId := range segmentIds {
		value := fmt.Sprintf("(%d, %d, %s)", userId, segmentId, aliveUntilStr)
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

	h := history.History{
		UserId:     userId,
		SegmentIds: segmentIds,
		Operation:  "added",
		Date:       time.Now(),
	}

	if err = historyRepo.Create(ctx, &h, tx); err != nil {
		return err
	}

	return nil
}

// delSegments is a function that deletes the specified segments from the user.
func delSegments(ctx context.Context, userId int, slugs []string, historyRepo history.Repository, tx pgx.Tx) error {
	segmentIds, err := getSegmentIdsBySlugs(ctx, tx, slugs)
	if err != nil {
		return err
	}
	if len(segmentIds) == 0 {
		return nil
	}

	var segmentIdsArray pgtype.Int4Array
	err = segmentIdsArray.Set(segmentIds)
	if err != nil {
		return err
	}

	q := `DELETE FROM user_segments WHERE user_id = $1 AND segment_id = ANY($2) RETURNING segment_id;`
	rows, err := tx.Query(ctx, q, userId, &segmentIdsArray)
	if err != nil {
		return err
	}
	defer rows.Close()
	deletedSegmentIds := make([]int, 0)
	for rows.Next() {
		var deletedId int
		if err = rows.Scan(&deletedId); err != nil {
			return err
		}
		deletedSegmentIds = append(deletedSegmentIds, deletedId)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	h := history.History{
		UserId:     userId,
		SegmentIds: deletedSegmentIds,
		Operation:  "deleted",
		Date:       time.Now(),
	}

	if err = historyRepo.Create(ctx, &h, tx); err != nil {
		return err
	}

	return nil
}

// AddDelSegments is a method of the repository that adds and deletes segments for a user within a single transaction.
// This function calls the functions to add and delete segments for the user in a single transaction.
// If an error occurs during the process, a rollback will be triggered.
func (r *repository) AddDelSegments(ctx context.Context, s *SegmentsAddDelDto, historyRepo history.Repository) error {
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
		if err = addSegments(ctx, s.UserId, s.SegmentsAdd, s.TtlDays, historyRepo, tx); err != nil {
			return err
		}
	}
	if len(s.SegmentsDel) > 0 {
		if err = delSegments(ctx, s.UserId, s.SegmentsDel, historyRepo, tx); err != nil {
			return err
		}
	}
	return nil
}

// CreateUser creates a new user record in the "users" table within the repository and returns the ID of the newly created row.
func (r *repository) CreateUser(ctx context.Context) (int, error) {
	maxId, err := r.GetMaxId(ctx)
	if err != nil {
		return 0, err
	}

	var userId int
	q := `INSERT INTO users (user_id) VALUES ($1) RETURNING user_id;`
	err = r.client.QueryRow(ctx, q, maxId+1).Scan(&userId)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

// GetMaxId retrieves the maximum user ID from the "users" table in the repository.
func (r *repository) GetMaxId(ctx context.Context) (int, error) {
	var userId int
	q := `SELECT COALESCE(MAX(user_id), 0) FROM users;`
	err := r.client.QueryRow(ctx, q).Scan(&userId)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

// DelUser deletes a user with the specified userID from the repository.
// This function is primarily intended for testing purposes, providing the ability to remove a user by their unique identifier.
func (r *repository) DelUser(ctx context.Context, userId int) error {
	q := `DELETE FROM users WHERE user_id = ($1);`
	_, err := r.client.Query(ctx, q, userId)
	if err != nil {
		return err
	}
	return nil
}

// DeleteSegmentsEveryDay deletes users from segments when the current date
// become greater than or equal to the user's lifetime (alive_until column)
// within the segment.
// The function is triggered every day at 03:00 AM in the Moscow time zone
// when server load is at its lowest
func (r *repository) DeleteSegmentsEveryDay(ctx context.Context, historyRepo history.Repository) {
	s := gocron.NewScheduler(time.UTC)
	_, err := s.Every(1).Day().At("00:00").Do(func() error {
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

		// key: userId, value: slice of segment ids witch deleted
		h := make(map[int][]int)

		q := `DELETE FROM user_segments WHERE alive_until <= $1 RETURNING user_id, segment_id;`
		rows, err := tx.Query(ctx, q, time.Now())
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var userId, segmentId int
			if err = rows.Scan(&userId, &segmentId); err != nil {
				return err
			}
			h[userId] = append(h[userId], segmentId)
		}
		if err = rows.Err(); err != nil {
			return err
		}

		// Add rows to history
		now := time.Now()
		for userId, segmentIds := range h {
			h := history.History{
				UserId:     userId,
				SegmentIds: segmentIds,
				Operation:  "deleted",
				Date:       now,
			}
			if err = historyRepo.Create(ctx, &h, tx); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Println("error delete segments every day:", err)
	}

	s.StartAsync()
}

func NewRepo(client pkg.DBClient) Repository {
	return &repository{
		client: client,
	}
}
