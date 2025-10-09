package domain

import "time"

type Thread struct {
	Title   string
	Message string
	Author  string
	Created *time.Time
	Forum   *string
	Id      *int32
	Slug    *string
	Votes   *int32
}

//easyjson:json
type Threads []*Thread

type ThreadUpdate struct {
	Message *string
	Title   *string
}
