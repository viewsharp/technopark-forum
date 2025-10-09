package domain

type Forum struct {
	Slug    string
	Title   string
	User    string
	Posts   *int64
	Threads *int32
}
