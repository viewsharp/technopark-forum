package forum

type Forum struct {
	Posts   int32  `json:"posts,omitempty"`
	Slug    string `json:"slug"`
	Threads int32  `json:"threads,omitempty"`
	Title   string `json:"title"`
	User    string `json:"user"`
}