package segment

type SegmentDto struct {
	Name string `json:"name"`
}

func (s *SegmentDto) Valid() bool {
	return s.Name != ""
}
