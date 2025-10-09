package forum

type Forum struct {
	Posts   *int64  `json:"posts"`
	Slug    *string `json:"slug"`
	Threads *int32  `json:"threads"`
	Title   *string `json:"title"`
	User    *string `json:"user"`
}
