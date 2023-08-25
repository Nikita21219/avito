package user

import "main/internal/segment"

type SegmentsDto struct {
	UserId   int                  `json:"user_id"`
	Segments []segment.SegmentDto `json:"segments"`
}

type SegmentsAddDelDto struct {
	UserId      int      `json:"user_id"`
	SegmentsAdd []string `json:"add"`
	SegmentsDel []string `json:"del"`
}

func (seg *SegmentsAddDelDto) ToModel() *Segments {
	return nil
}

func (seg *SegmentsAddDelDto) Valid() bool {
	if seg.UserId <= 0 {
		return false
	}
	for _, s := range seg.SegmentsAdd {
		if s == "" {
			return false
		}
	}
	for _, s := range seg.SegmentsDel {
		if s == "" {
			return false
		}
	}
	return true
}
