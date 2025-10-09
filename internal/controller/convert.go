package controller

import (
	oapitypes "github.com/oapi-codegen/runtime/types"

	"github.com/viewsharp/technopark-forum/internal/api"
	"github.com/viewsharp/technopark-forum/internal/domain"
)

func ptrString(s string) *string {
	return &s
}

func ptrBool(b bool) *bool {
	return &b
}

// domainForumToAPI converts domain.Forum to api.Forum
func domainForumToAPI(forum *domain.Forum) api.Forum {
	return api.Forum{
		Slug:    forum.Slug,
		Title:   forum.Title,
		User:    forum.User,
		Posts:   forum.Posts,
		Threads: forum.Threads,
	}
}

// domainThreadToAPI converts domain.Thread to api.Thread
func domainThreadToAPI(thread *domain.Thread) api.Thread {
	return api.Thread{
		Title:   thread.Title,
		Message: thread.Message,
		Author:  thread.Author,
		Created: thread.Created,
		Forum:   thread.Forum,
		Id:      thread.Id,
		Slug:    thread.Slug,
		Votes:   thread.Votes,
	}
}

// domainThreadsToAPI converts slice of domain.Thread to slice of api.Thread
func domainThreadsToAPI(threads []domain.Thread) []api.Thread {
	result := make([]api.Thread, len(threads))
	for i, t := range threads {
		result[i] = domainThreadToAPI(&t)
	}
	return result
}

// domainUserToAPI converts domain.User to api.User
func domainUserToAPI(user *domain.User) api.User {
	return api.User{
		About:    user.About,
		Email:    oapitypes.Email(user.Email),
		Fullname: user.FullName,
		Nickname: &user.Nickname,
	}
}

// domainUsersToAPI converts slice of domain.User to slice of api.User
func domainUsersToAPI(users []domain.User) []api.User {
	result := make([]api.User, len(users))
	for i, u := range users {
		result[i] = domainUserToAPI(&u)
	}
	return result
}

// domainPostToAPI converts domain.Post to api.Post
func domainPostToAPI(post *domain.Post) api.Post {
	return api.Post{
		Message:  post.Message,
		Author:   post.Author,
		Created:  post.Created,
		Forum:    post.Forum,
		Id:       post.Id,
		IsEdited: post.IsEdited,
		Parent:   post.Parent,
		Thread:   post.Thread,
	}
}

// domainPostsToAPI converts slice of domain.Post to slice of api.Post
func domainPostsToAPI(posts []domain.Post) []api.Post {
	result := make([]api.Post, len(posts))
	for i, p := range posts {
		result[i] = domainPostToAPI(&p)
	}
	return result
}

// domainPostFullToAPI converts domain.PostFull to api.PostFull
func domainPostFullToAPI(postFull *domain.PostFull) api.PostFull {
	response := api.PostFull{}

	if postFull.Post != nil {
		post := domainPostToAPI(postFull.Post)
		response.Post = &post
	}

	if postFull.Author != nil {
		author := domainUserToAPI(postFull.Author)
		response.Author = &author
	}

	if postFull.Forum != nil {
		forum := domainForumToAPI(postFull.Forum)
		response.Forum = &forum
	}

	if postFull.Thread != nil {
		thread := domainThreadToAPI(postFull.Thread)
		response.Thread = &thread
	}

	return response
}
