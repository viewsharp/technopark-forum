package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
)

type Handler struct {
	DB *sql.DB
}

func (hlr *Handler) CreateForum(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	if string(ctx.Path()) != "/forum/create" {
		return nil, fasthttp.StatusNotFound
	}

	var forum Forum
	err := forum.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	var userId int
	err = hlr.DB.QueryRow(
		"SELECT id, nickname FROM users WHERE nickname = $1",
		forum.User,
	).Scan(&userId, &forum.User)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			message := "Can't find user with nickname: " + forum.User
			return Error{Message: &message}, fasthttp.StatusNotFound
		}

		return nil, fasthttp.StatusInternalServerError
	}

	_, err = hlr.DB.Exec(
		"INSERT INTO forums (slug, title, user_id) VALUES ($1, $2, $3)",
		forum.Slug, forum.Title, userId,
	)

	if err != nil {
		if err.(*pq.Error).Code.Name() == "unique_violation" {
			err := hlr.DB.QueryRow(
				"SELECT slug, title, users.nickname FROM forums JOIN users ON forums.user_id = users.id WHERE slug = $1",
				forum.Slug,
			).Scan(&forum.Slug, &forum.Title, &forum.User)
			if err != nil {
				return nil, fasthttp.StatusInternalServerError
			}

			return forum, fasthttp.StatusConflict
		}

		return nil, fasthttp.StatusInternalServerError
	}

	return forum, fasthttp.StatusCreated
}

func (hlr *Handler) CreateThread(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	var thread Thread
	err := thread.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	var userId int
	err = hlr.DB.QueryRow(
		"SELECT id, nickname FROM users WHERE nickname = $1",
		thread.Author,
	).Scan(&userId, &thread.Author)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			message := "Can't find user with nickname: " + *thread.Author
			return Error{Message: &message}, fasthttp.StatusNotFound
		}

		return nil, fasthttp.StatusInternalServerError
	}

	err = hlr.DB.QueryRow(
		"INSERT INTO threads (slug, created, title, message, user_id, forum_slug)	VALUES ($1, $2, $3, $4, $5, (SELECT slug FROM forums WHERE slug = $6)) RETURNING id, forum_slug, slug",
		thread.Slug, thread.Created, thread.Title, thread.Message, userId, ctx.UserValue("slug"),
	).Scan(&thread.Id, &thread.Forum, &thread.Slug)
	if err != nil {
		switch err.(*pq.Error).Code.Name() {
		case "unique_violation":
			err := hlr.DB.QueryRow(
				"SELECT threads.id, slug, created, title, message, users.nickname, forum_slug FROM threads JOIN users ON threads.user_id = users.id WHERE slug = $1",
				thread.Slug,
			).Scan(&thread.Id, &thread.Slug, &thread.Created, &thread.Title, &thread.Message, &thread.Author, &thread.Forum)
			if err != nil {
				return nil, fasthttp.StatusInternalServerError
			}

			return thread, fasthttp.StatusConflict
		case "not_null_violation":
			message := "Can't find thread forum by slug: " + *thread.Slug
			return Error{Message: &message}, fasthttp.StatusNotFound
		}

		return nil, fasthttp.StatusInternalServerError
	}

	return thread, fasthttp.StatusCreated
}

func (hlr *Handler) GetForum(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	slug := ctx.UserValue("slug").(string)
	var forum Forum

	err := hlr.DB.QueryRow("SELECT (SELECT count(*) FROM posts JOIN threads ON posts.thread_id = threads.id WHERE threads.forum_slug = $1) AS \"posts\", slug, (SELECT count(*) FROM threads WHERE threads.forum_slug = $1) AS \"threads\", title, users.nickname FROM forums JOIN users ON forums.user_id = users.id WHERE slug = $1",
		slug,
	).Scan(&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			message := "Can't find forum by slug: " + slug
			return Error{Message: &message}, fasthttp.StatusNotFound
		}

		return nil, fasthttp.StatusInternalServerError
	}

	return forum, fasthttp.StatusOK
}

func (hlr *Handler) GetForumThreads(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	slug := ctx.UserValue("slug").(string)

	limit := 1000
	limitParam := ctx.QueryArgs().Peek("limit")
	if limitParam != nil {
		var err error
		limit, err = strconv.Atoi(string(limitParam))
		if err != nil {
			return nil, fasthttp.StatusBadRequest
		}
	}

	desc := false
	descParam := ctx.QueryArgs().Peek("desc")
	if descParam != nil {
		desc = string(descParam) == "true"
	}

	since := string(ctx.QueryArgs().Peek("since"))

	forum, statusCode := hlr.GetForum(ctx)
	if statusCode != 200 {
		return forum, statusCode
	}

	var rows *sql.Rows
	var err error
	if desc {
		if since == "" {
			rows, err = hlr.DB.Query("SELECT threads.id, slug, created, title, message, 0 AS \"votes\", users.nickname, forum_slug FROM threads JOIN users ON threads.user_id = users.id WHERE forum_slug = $1 ORDER BY created DESC LIMIT $2", slug, limit)
		} else {
			rows, err = hlr.DB.Query("SELECT threads.id, slug, created, title, message, 0 AS \"votes\", users.nickname, forum_slug FROM threads JOIN users ON threads.user_id = users.id WHERE forum_slug = $1 AND created <= $3 ORDER BY created DESC LIMIT $2", slug, limit, since)
		}
	} else {
		if since == "" {
			rows, err = hlr.DB.Query("SELECT threads.id, slug, created, title, message, 0 AS \"votes\", users.nickname, forum_slug FROM threads JOIN users ON threads.user_id = users.id WHERE forum_slug = $1 ORDER BY created ASC LIMIT $2", slug, limit)
		} else {
			rows, err = hlr.DB.Query("SELECT threads.id, slug, created, title, message, 0 AS \"votes\", users.nickname, forum_slug FROM threads JOIN users ON threads.user_id = users.id WHERE forum_slug = $1 AND created >= $3 ORDER BY created ASC LIMIT $2", slug, limit, since)
		}
	}

	if err != nil {
		return nil, fasthttp.StatusInternalServerError
	}

	threads := make(Threads, 0, 1)
	for rows.Next() {
		var thread Thread
		err = rows.Scan(&thread.Id, &thread.Slug, &thread.Created, &thread.Title, &thread.Message, &thread.Votes, &thread.Author, &thread.Forum)
		if err != nil {
			return nil, fasthttp.StatusInternalServerError
		}
		threads = append(threads, &thread)
	}
	return threads, fasthttp.StatusOK
}

func (hlr *Handler) GetForumUsers(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	return nil, fasthttp.StatusOK
}

func (hlr *Handler) GetPost(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	return nil, fasthttp.StatusOK
}

func (hlr *Handler) UpdatePost(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	return nil, fasthttp.StatusOK
}

func (hlr *Handler) ClearDB(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	return nil, fasthttp.StatusOK
}

func (hlr *Handler) GetDBStatus(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	return nil, fasthttp.StatusOK
}

func (hlr *Handler) CreatePost(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	result, statusCode := hlr.GetThread(ctx)
	if statusCode != 200 {
		return result, statusCode
	}
	thread := result.(Thread)

	posts := make(Posts, 0, 1)
	err := posts.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	if len(posts) == 0 {
		return posts, fasthttp.StatusCreated
	}

	var queryBuilder strings.Builder
	queryParams := make([]interface{}, 0, 4*len(posts))
	queryBuilder.WriteString("INSERT INTO posts (user_id, message, parent_id, thread_id) VALUES ")
	for i, post := range posts {
		if i != 0 {
			queryBuilder.WriteString(",")
		}
		queryBuilder.WriteString(fmt.Sprintf(" ((SELECT id FROM users WHERE nickname = $%d), $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		queryParams = append(queryParams, post.Author, post.Message, post.Parent, thread.Id)
	}
	queryBuilder.WriteString(" RETURNING id, created, thread_id")

	rows, err := hlr.DB.Query(queryBuilder.String(), queryParams...)
	if err != nil {
		return nil, fasthttp.StatusInternalServerError
	}

	for i, _ := range posts {
		rows.Next()
		post := &posts[i]

		err = rows.Scan(&post.Id, &post.Created, &post.Thread)
		if err != nil {
			return nil, fasthttp.StatusInternalServerError
		}
		post.Forum = thread.Forum
	}

	return posts, fasthttp.StatusCreated
}

func (hlr *Handler) GetThread(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	var thread Thread
	slugOrId := ctx.UserValue("slug_or_id").(string)
	threadId, err := strconv.Atoi(slugOrId)
	if err != nil {
		err = hlr.DB.QueryRow("SELECT threads.id, slug, created, title, message, users.nickname, forum_slug FROM threads JOIN users ON threads.user_id = users.id WHERE slug = $1",
			slugOrId,
		).Scan(&thread.Id, &thread.Slug, &thread.Created, &thread.Title, &thread.Message, &thread.Author, &thread.Forum)
	} else {
		err = hlr.DB.QueryRow("SELECT threads.id, slug, created, title, message, users.nickname, forum_slug FROM threads JOIN users ON threads.user_id = users.id WHERE threads.id = $1",
			threadId,
		).Scan(&thread.Id, &thread.Slug, &thread.Created, &thread.Title, &thread.Message, &thread.Author, &thread.Forum)
	}
	if err != nil {
		return nil, fasthttp.StatusInternalServerError
	}

	return thread, fasthttp.StatusOK
}

func (hlr *Handler) UpdateThread(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	return nil, fasthttp.StatusOK
}

func (hlr *Handler) GetThreadPosts(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	return nil, fasthttp.StatusOK
}

func (hlr *Handler) CreateThreadVote(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	var vote Vote
	err := vote.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	result, statusCode := hlr.GetThread(ctx)
	if statusCode != 200 {
		return result, statusCode
	}
	thread := result.(Thread)

	_, err = hlr.DB.Exec(
		"INSERT INTO votes (thread_id, user_id, voice) VALUES ($1, (SELECT id FROM users WHERE nickname = $2), $3) ON CONFLICT ON CONSTRAINT votes_thread_user_unique DO UPDATE SET voice = $3 WHERE votes.thread_id = $1 AND votes.user_id = (SELECT id FROM users WHERE nickname = $2);",
		thread.Id, vote.Nickname, vote.Voice,
	)
	if err != nil {
		return nil, fasthttp.StatusInternalServerError
	}

	err = hlr.DB.QueryRow("SELECT sum(voice) FROM votes WHERE thread_id = $1",
		thread.Id,
	).Scan(&thread.Votes)
	if err != nil {
		return nil, fasthttp.StatusInternalServerError
	}

	return thread, fasthttp.StatusOK
}

func (hlr *Handler) CreateUser(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	var user User
	err := user.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	nickname := ctx.UserValue("nickname").(string)
	user.Nickname = &nickname
	_, err = hlr.DB.Exec(
		"INSERT INTO users (nickname, fullname, email, about) VALUES ($1, $2, $3, $4)",
		user.Nickname, user.FullName, user.Email, user.About,
	)

	if err != nil {
		if err.(*pq.Error).Code.Name() == "unique_violation" {
			rows, err := hlr.DB.Query(
				"SELECT nickname, fullname, email, about FROM users WHERE nickname = $1 UNION SELECT nickname, fullname, email, about FROM users WHERE email = $2",
				user.Nickname, user.Email,
			)
			if err != nil {
				return nil, fasthttp.StatusInternalServerError
			}

			var users Users
			for rows.Next() {
				var user User
				err = rows.Scan(&user.Nickname, &user.FullName, &user.Email, &user.About)
				if err != nil {
					return nil, fasthttp.StatusInternalServerError
				}
				users = append(users, &user)
			}
			return users, fasthttp.StatusConflict
		}

		return nil, fasthttp.StatusInternalServerError
	}

	return user, fasthttp.StatusCreated
}

func (hlr *Handler) GetUser(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	nickname := ctx.UserValue("nickname").(string)
	var user User

	err := hlr.DB.QueryRow("SELECT nickname, fullname, email, about FROM users WHERE nickname = $1",
		nickname,
	).Scan(&user.Nickname, &user.FullName, &user.Email, &user.About)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			message := "Can't find user by nickname: " + nickname
			return Error{Message: &message}, fasthttp.StatusNotFound
		}

		return nil, fasthttp.StatusInternalServerError
	}

	return user, fasthttp.StatusOK
}

func (hlr *Handler) UpdateUser(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	nickname := ctx.UserValue("nickname").(string)
	var user UserUpdate
	err := user.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	result, err := hlr.DB.Exec(
		"UPDATE users SET fullname = COALESCE($1, fullname), email = COALESCE($2, email), about = COALESCE($3, about) WHERE nickname = $4",
		user.FullName, user.Email, user.About, nickname,
	)

	if err != nil {
		if err.(*pq.Error).Code.Name() == "unique_violation" {
			message := "This email is already registered by user: " + nickname
			return Error{Message: &message}, fasthttp.StatusConflict
		}

		return nil, fasthttp.StatusInternalServerError
	}

	rowAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fasthttp.StatusInternalServerError
	}
	if rowAffected == 0 {
		message := "Can't find user by nickname: " + nickname
		return Error{Message: &message}, fasthttp.StatusNotFound
	}

	return hlr.GetUser(ctx)
}
