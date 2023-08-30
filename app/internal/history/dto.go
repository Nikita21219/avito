package history

type HistoryDto struct {
	UserId      int    `json:"user_id"`
	SegmentSlug string `json:"segment_slug"`
	Operation   string `json:"operation"`
	Date        string `json:"date"`
}
