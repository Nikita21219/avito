package segment

type SegmentDto struct {
	Slug string `json:"slug"`
}

func (s *SegmentDto) Valid() bool {
	return s.Slug != ""
}
