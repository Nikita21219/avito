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
	TtlDays     *int     `json:"ttl_days"`
}

func (seg *SegmentsAddDelDto) Valid() bool {
	if seg.UserId <= 0 || seg.SegmentsAdd == nil || seg.SegmentsDel == nil {
		return false
	}
	if seg.TtlDays != nil && *seg.TtlDays <= 0 {
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
