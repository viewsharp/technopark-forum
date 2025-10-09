package user

type User struct {
	Email    string  `json:"email"`
	FullName string  `json:"fullname"`
	Nickname string  `json:"nickname,omitempty"`
	About    *string `json:"about,omitempty"`
}

type Users []*User

type UserUpdate struct {
	About    *string `json:"about,omitempty"`
	Email    *string `json:"email,omitempty"`
	FullName *string `json:"fullname,omitempty"`
}
