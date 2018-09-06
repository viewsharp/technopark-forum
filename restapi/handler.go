package main

import (
	"database/sql"
	"encoding/json"
	"github.com/valyala/fasthttp"
)

type Handler struct {
	DB *sql.DB
}

func (*Handler) CreateForum(ctx *fasthttp.RequestCtx) {
	if string(ctx.Path()) != "/forum/create" {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	var forum Forum
	err:=json.Unmarshal(ctx.PostBody(), &forum)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
}

func (*Handler) CreateThread(ctx *fasthttp.RequestCtx) {

}

func (*Handler) GetForum(ctx *fasthttp.RequestCtx) {

}

func (*Handler) GetForumThreads(ctx *fasthttp.RequestCtx) {

}

func (*Handler) GetForumUsers(ctx *fasthttp.RequestCtx) {

}

func (*Handler) GetPost(ctx *fasthttp.RequestCtx) {

}

func (*Handler) UpdatePost(ctx *fasthttp.RequestCtx) {

}

func (*Handler) ClearDB(ctx *fasthttp.RequestCtx) {

}

func (*Handler) GetDBStatus(ctx *fasthttp.RequestCtx) {

}

func (*Handler) CreatePost(ctx *fasthttp.RequestCtx) {

}

func (*Handler) GetThread(ctx *fasthttp.RequestCtx) {

}

func (*Handler) UpdateThread(ctx *fasthttp.RequestCtx) {

}

func (*Handler) GetThreadPosts(ctx *fasthttp.RequestCtx) {

}

func (*Handler) CreateThreadVote(ctx *fasthttp.RequestCtx) {

}

func (*Handler) CreateUser(ctx *fasthttp.RequestCtx) {

}

func (*Handler) GetUser(ctx *fasthttp.RequestCtx) {

}

func (*Handler) UpdateUser(ctx *fasthttp.RequestCtx) {

}
