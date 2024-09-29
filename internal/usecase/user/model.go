package user

type User struct {
	About    *string `json:"about,omitempty"`
	Email    *string `json:"email"`
	FullName *string `json:"fullname"`
	Nickname *string `json:"nickname,omitempty"`
}

type Users []*User

type UserUpdate struct {
	About    *string `json:"about,omitempty"`
	Email    *string `json:"email,omitempty"`
	FullName *string `json:"fullname,omitempty"`
}
