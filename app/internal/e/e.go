package e

import (
	"fmt"
	"github.com/jackc/pgconn"
)

type DuplicateSegmentError struct {
	SegmentName string
}

func (e *DuplicateSegmentError) Error() string {
	return fmt.Sprintf("segment with name '%s' already exists", e.SegmentName)
}

func IsDuplicateError(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23505"
	}
	return false
}

type UserNotFoundError struct {
	UserId int
}

func (e *UserNotFoundError) Error() string {
	return fmt.Sprintf("user with id '%d' active segments not found", e.UserId)
}
