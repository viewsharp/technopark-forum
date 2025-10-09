package domain

type User struct {
	Email    string
	FullName string
	Nickname string
	About    *string
}

type UserUpdate struct {
	About    *string
	Email    *string
	FullName *string
}
