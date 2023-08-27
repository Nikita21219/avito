package user

import "main/internal/segment"

type Segments struct {
	UserId   int                `json:"user_id"`
	Segments []*segment.Segment `json:"segments"`
}

type User struct {
	Id int `json:"user_id"`
}

type SegmentsAddDel struct {
	UserId      int      `json:"user_id"`
	SegmentsAdd []string `json:"add"`
	SegmentsDel []string `json:"delete"`
}
