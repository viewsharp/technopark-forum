package domain

import "time"

type Thread struct {
	Author  *string
	Created *time.Time
	Forum   *string
	Id      *int32
	Message *string
	Slug    *string
	Title   *string
	Votes   *int32
}

//easyjson:json
type Threads []*Thread

type ThreadUpdate struct {
	Message *string
	Title   *string
}
