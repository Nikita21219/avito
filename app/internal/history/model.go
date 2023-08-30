package history

import "time"

type History struct {
	UserId     int
	SegmentIds []int
	Operation  string
	Date       time.Time
}
