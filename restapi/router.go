package main

import (
	"github.com/buaazp/fasthttprouter"
)

func NewRouter(handler *Handler) *fasthttprouter.Router {
	router := fasthttprouter.New()

	router.POST("/forum/:slug", handler.CreateForum) // "forum/create"
	router.POST("/forum/:slug/create", handler.CreateThread)
	router.GET("/forum/:slug/details", handler.GetForum)
	router.GET("/forum/:slug/threads", handler.GetForumThreads)
	router.GET("/forum/:slug/users", handler.GetForumUsers)
	router.GET("/post/:id/details", handler.GetPost)
	router.POST("/post/:id/details", handler.UpdatePost)
	router.POST("/service/clear", handler.ClearDB)
	router.GET("/service/status", handler.GetDBStatus)
	router.POST("/thread/:slug/create", handler.CreatePost)
	router.GET("/thread/:slug/details", handler.GetThread)
	router.POST("/thread/:slug/details", handler.UpdateThread)
	router.GET("/thread/:slug/posts", handler.GetThreadPosts)
	router.POST("/thread/:slug/vote", handler.CreateThreadVote)
	router.POST("/user/:nickname/create", handler.CreateUser)
	router.GET("/user/:nickname/profile", handler.GetUser)
	router.POST("/user/:nickname/profile", handler.UpdateUser)

	return router
}