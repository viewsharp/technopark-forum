package status

type Status struct {
	Forum  *int32 `json:"forum"`
	Post   *int32 `json:"post"`
	Thread *int32 `json:"thread"`
	User   *int32 `json:"user"`
}
