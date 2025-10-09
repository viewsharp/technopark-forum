package domain

type User struct {
	Email    string
	FullName string
	Nickname string
	About    *string
}

type Users []*User

type UserUpdate struct {
	About    *string
	Email    *string
	FullName *string
}
