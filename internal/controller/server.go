package controller

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	oapitypes "github.com/oapi-codegen/runtime/types"

	"github.com/viewsharp/technopark-forum/internal/api"
	forumUC "github.com/viewsharp/technopark-forum/internal/usecase/forum"
	post2 "github.com/viewsharp/technopark-forum/internal/usecase/post"
	thread2 "github.com/viewsharp/technopark-forum/internal/usecase/thread"
	user2 "github.com/viewsharp/technopark-forum/internal/usecase/user"
	vote2 "github.com/viewsharp/technopark-forum/internal/usecase/vote"
)

// Server implements api.ServerInterface
type Server struct {
	sb *UsecaseSet
}

func NewServer(usecaseSet *UsecaseSet) *Server {
	return &Server{sb: usecaseSet}
}

// Forum handlers

// ForumCreate implements api.ServerInterface
func (s *Server) ForumCreate(c *fiber.Ctx) error {
	var forum api.Forum
	if err := c.BodyParser(&forum); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	// Convert api.Forum to forumUC.Forum
	ucForum := forumUC.Forum{
		Slug:  &forum.Slug,
		Title: &forum.Title,
		User:  &forum.User,
	}

	createdForum, err := s.sb.forum.Add(c.Context(), ucForum)
	if err != nil {
		switch err {
		case forumUC.ErrUniqueViolation:
			result, err := s.sb.forum.BySlug(c.Context(), forum.Slug)
			if err == nil {
				return c.Status(fiber.StatusConflict).JSON(api.Forum{
					Slug:    *result.Slug,
					Title:   *result.Title,
					User:    *result.User,
					Posts:   result.Posts,
					Threads: result.Threads,
				})
			}
		case forumUC.ErrNotFoundUser:
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user with nickname: " + forum.User),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	return c.Status(fiber.StatusCreated).JSON(api.Forum{
		Slug:    *createdForum.Slug,
		Title:   *createdForum.Title,
		User:    *createdForum.User,
		Posts:   createdForum.Posts,
		Threads: createdForum.Threads,
	})
}

// ForumGetOne implements api.ServerInterface
func (s *Server) ForumGetOne(c *fiber.Ctx, slug string) error {
	result, err := s.sb.forum.FullBySlug(c.Context(), slug)

	switch err {
	case nil:
		return c.JSON(api.Forum{
			Slug:    *result.Slug,
			Title:   *result.Title,
			User:    *result.User,
			Posts:   result.Posts,
			Threads: result.Threads,
		})
	case forumUC.ErrNotFound:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString("Can't find forum by slug: " + slug),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// Thread handlers

// ThreadCreate implements api.ServerInterface
func (s *Server) ThreadCreate(c *fiber.Ctx, slug string) error {
	var thread api.Thread
	if err := c.BodyParser(&thread); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	ucThread := thread2.Thread{
		Author:  &thread.Author,
		Created: thread.Created,
		Forum:   &slug,
		Message: &thread.Message,
		Slug:    thread.Slug,
		Title:   &thread.Title,
	}

	err := s.sb.thread.Add(c.Context(), &ucThread)
	switch err {
	case nil:
		return c.Status(fiber.StatusCreated).JSON(api.Thread{
			Author:  *ucThread.Author,
			Created: ucThread.Created,
			Forum:   ucThread.Forum,
			Id:      ucThread.Id,
			Message: *ucThread.Message,
			Slug:    ucThread.Slug,
			Title:   *ucThread.Title,
			Votes:   ucThread.Votes,
		})
	case thread2.ErrUniqueViolation:
		result, err := s.sb.thread.BySlug(c.Context(), *thread.Slug)
		if err == nil {
			return c.Status(fiber.StatusConflict).JSON(api.Thread{
				Author:  *result.Author,
				Created: result.Created,
				Forum:   result.Forum,
				Id:      result.Id,
				Message: *result.Message,
				Slug:    result.Slug,
				Title:   *result.Title,
				Votes:   result.Votes,
			})
		}
	case thread2.ErrNotFoundUser:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString("Can't find thread author by nickname: " + thread.Author),
		})
	case thread2.ErrNotFoundForum:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString("Can't find thread forum by slug: " + slug),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// ForumGetThreads implements api.ServerInterface
func (s *Server) ForumGetThreads(c *fiber.Ctx, slug string, params api.ForumGetThreadsParams) error {
	var limit int32 = 100
	if params.Limit != nil {
		limit = *params.Limit
	}

	desc := false
	if params.Desc != nil {
		desc = *params.Desc
	}

	since := ""
	if params.Since != nil {
		since = params.Since.Format(time.RFC3339Nano)
	}

	result, err := s.sb.thread.ByForumSlug(c.Context(), slug, desc, since, limit)

	switch err {
	case nil:
		threads := make([]api.Thread, len(*result))
		for i, t := range *result {
			threads[i] = api.Thread{
				Author:  *t.Author,
				Created: t.Created,
				Forum:   t.Forum,
				Id:      t.Id,
				Message: *t.Message,
				Slug:    t.Slug,
				Title:   *t.Title,
				Votes:   t.Votes,
			}
		}
		return c.JSON(threads)
	case thread2.ErrNotFoundForum:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString("Can't find forum by slug: " + slug),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// ThreadGetOne implements api.ServerInterface
func (s *Server) ThreadGetOne(c *fiber.Ctx, slugOrId string) error {
	var result *thread2.Thread
	var err error

	threadId, threadIdParseErr := strconv.Atoi(slugOrId)
	if threadIdParseErr == nil {
		result, err = s.sb.thread.ById(c.Context(), threadId)
	} else {
		result, err = s.sb.thread.BySlug(c.Context(), slugOrId)
	}

	switch err {
	case nil:
		return c.JSON(api.Thread{
			Author:  *result.Author,
			Created: result.Created,
			Forum:   result.Forum,
			Id:      result.Id,
			Message: *result.Message,
			Slug:    result.Slug,
			Title:   *result.Title,
			Votes:   result.Votes,
		})
	case thread2.ErrNotFound:
		if threadIdParseErr == nil {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString(fmt.Sprintf("Can't find thread by id: %d", threadId)),
			})
		} else {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find thread by slug: " + slugOrId),
			})
		}
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// ThreadUpdate implements api.ServerInterface
func (s *Server) ThreadUpdate(c *fiber.Ctx, slugOrId string) error {
	var update api.ThreadUpdate
	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	ucUpdate := thread2.ThreadUpdate{
		Message: update.Message,
		Title:   update.Title,
	}

	threadId, threadIdErr := strconv.Atoi(slugOrId)
	if threadIdErr == nil {
		err := s.sb.thread.UpdateById(c.Context(), threadId, &ucUpdate)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
		}
	} else {
		err := s.sb.thread.UpdateBySlug(c.Context(), slugOrId, &ucUpdate)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
		}
	}

	var result *thread2.Thread
	var err error
	if threadIdErr == nil {
		result, err = s.sb.thread.ById(c.Context(), threadId)
	} else {
		result, err = s.sb.thread.BySlug(c.Context(), slugOrId)
	}

	switch err {
	case nil:
		return c.JSON(api.Thread{
			Author:  *result.Author,
			Created: result.Created,
			Forum:   result.Forum,
			Id:      result.Id,
			Message: *result.Message,
			Slug:    result.Slug,
			Title:   *result.Title,
			Votes:   result.Votes,
		})
	case thread2.ErrNotFound:
		if threadIdErr == nil {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString(fmt.Sprintf("Can't find thread by id: %d", threadId)),
			})
		} else {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find thread by slug: " + slugOrId),
			})
		}
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// User handlers

// UserCreate implements api.ServerInterface
func (s *Server) UserCreate(c *fiber.Ctx, nickname string) error {
	var user api.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	ucUser := user2.User{
		About:    user.About,
		Email:    ptrString(string(user.Email)),
		FullName: &user.Fullname,
		Nickname: &nickname,
	}

	err := s.sb.user.Add(c.Context(), &ucUser)
	switch err {
	case nil:
		return c.Status(fiber.StatusCreated).JSON(api.User{
			About:    user.About,
			Email:    user.Email,
			Fullname: user.Fullname,
			Nickname: ptrString(nickname),
		})
	case user2.ErrUniqueViolation:
		var result []api.User

		userByEmail, err := s.sb.user.ByEmail(c.Context(), string(user.Email))
		if err == nil {
			result = append(result, api.User{
				About:    userByEmail.About,
				Email:    oapitypes.Email(*userByEmail.Email),
				Fullname: *userByEmail.FullName,
				Nickname: userByEmail.Nickname,
			})
		}

		userByNickname, err := s.sb.user.ByNickname(c.Context(), nickname)
		if err == nil {
			if userByEmail == nil {
				result = append(result, api.User{
					About:    userByNickname.About,
					Email:    oapitypes.Email(*userByNickname.Email),
					Fullname: *userByNickname.FullName,
					Nickname: userByNickname.Nickname,
				})
			} else if *userByNickname.Nickname != *userByEmail.Nickname {
				result = append(result, api.User{
					About:    userByNickname.About,
					Email:    oapitypes.Email(*userByNickname.Email),
					Fullname: *userByNickname.FullName,
					Nickname: userByNickname.Nickname,
				})
			}
		}

		return c.Status(fiber.StatusConflict).JSON(result)
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// UserGetOne implements api.ServerInterface
func (s *Server) UserGetOne(c *fiber.Ctx, nickname string) error {
	result, err := s.sb.user.ByNickname(c.Context(), nickname)

	switch err {
	case nil:
		return c.JSON(api.User{
			About:    result.About,
			Email:    oapitypes.Email(*result.Email),
			Fullname: *result.FullName,
			Nickname: result.Nickname,
		})
	case user2.ErrNotFound:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString("Can't find user by nickname: " + nickname),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// UserUpdate implements api.ServerInterface
func (s *Server) UserUpdate(c *fiber.Ctx, nickname string) error {
	var update api.UserUpdate
	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	var emailStr *string
	if update.Email != nil {
		emailStr = ptrString(string(*update.Email))
	}

	ucUpdate := user2.UserUpdate{
		About:    update.About,
		Email:    emailStr,
		FullName: update.Fullname,
	}

	err := s.sb.user.UpdateByNickname(c.Context(), nickname, &ucUpdate)

	switch err {
	case nil:
		return c.JSON(api.User{
			About:    update.About,
			Email:    *update.Email,
			Fullname: *update.Fullname,
			Nickname: ptrString(nickname),
		})
	case user2.ErrUniqueViolation:
		return c.Status(fiber.StatusConflict).JSON(api.Error{
			Message: ptrString("This email is already registered by user: " + string(*update.Email)),
		})
	case user2.ErrNotFound:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString("Can't find user by nickname: " + nickname),
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// ForumGetUsers implements api.ServerInterface
func (s *Server) ForumGetUsers(c *fiber.Ctx, slug string, params api.ForumGetUsersParams) error {
	var limit int32 = 100
	if params.Limit != nil {
		limit = *params.Limit
	}

	desc := false
	if params.Desc != nil {
		desc = *params.Desc
	}

	since := ""
	if params.Since != nil {
		since = *params.Since
	}

	result, err := s.sb.user.ByForumSlug(c.Context(), slug, desc, since, limit)

	switch err {
	case nil:
		users := make([]api.User, len(*result))
		for i, u := range *result {
			users[i] = api.User{
				About:    u.About,
				Email:    oapitypes.Email(*u.Email),
				Fullname: *u.FullName,
				Nickname: u.Nickname,
			}
		}
		return c.JSON(users)
	case user2.ErrNotFoundForum:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString("Can't find forum by slug: " + slug),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// Post handlers

// PostsCreate implements api.ServerInterface
func (s *Server) PostsCreate(c *fiber.Ctx, slugOrId string) error {
	var posts []api.Post
	if err := c.BodyParser(&posts); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	threadId, threadIdParseErr := strconv.Atoi(slugOrId)

	if len(posts) == 0 {
		var err error
		if threadIdParseErr == nil {
			_, err = s.sb.thread.ById(c.Context(), threadId)
		} else {
			_, err = s.sb.thread.BySlug(c.Context(), slugOrId)
		}

		switch err {
		case nil:
			return c.Status(fiber.StatusCreated).JSON(posts)
		case thread2.ErrNotFound:
			if threadIdParseErr == nil {
				return c.Status(fiber.StatusNotFound).JSON(api.Error{
					Message: ptrString(fmt.Sprintf("Can't find post thread by id: %d", threadId)),
				})
			} else {
				return c.Status(fiber.StatusNotFound).JSON(api.Error{
					Message: ptrString("Can't find post thread by slug: " + slugOrId),
				})
			}
		}

		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	// Convert api.Post to post2.Post
	ucPosts := make([]post2.Post, len(posts))
	for i, p := range posts {
		ucPosts[i] = post2.Post{
			Author:   &p.Author,
			Created:  p.Created,
			Forum:    p.Forum,
			Id:       p.Id,
			IsEdited: p.IsEdited,
			Message:  &p.Message,
			Parent:   p.Parent,
			Thread:   p.Thread,
		}
	}

	var err error
	if threadIdParseErr == nil {
		err = s.sb.post.AddByThreadId(c.Context(), ucPosts, int32(threadId))
	} else {
		err = s.sb.post.AddByThreadSlug(c.Context(), ucPosts, slugOrId)
	}

	if err != nil {
		if errors.Is(err, post2.ErrInvalidParent) {
			return c.Status(fiber.StatusConflict).JSON(api.Error{
				Message: ptrString("Parent post was created in another thread"),
			})
		}
		if errors.Is(err, post2.ErrNotFoundThread) {
			if threadIdParseErr == nil {
				return c.Status(fiber.StatusNotFound).JSON(api.Error{
					Message: ptrString(fmt.Sprintf("Can't find post thread by id: %d", threadId)),
				})
			} else {
				return c.Status(fiber.StatusNotFound).JSON(api.Error{
					Message: ptrString("Can't find post thread by slug: " + slugOrId),
				})
			}
		}

		var errNotFoundUser post2.ErrNotFoundUser
		if errors.As(err, &errNotFoundUser) {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find post author by nickname: " + errNotFoundUser.Nickname),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	// Convert back to api.Post
	for i, p := range ucPosts {
		posts[i] = api.Post{
			Author:   *p.Author,
			Created:  p.Created,
			Forum:    p.Forum,
			Id:       p.Id,
			IsEdited: p.IsEdited,
			Message:  *p.Message,
			Parent:   p.Parent,
			Thread:   p.Thread,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(posts)
}

// PostGetOne implements api.ServerInterface
func (s *Server) PostGetOne(c *fiber.Ctx, id int64, params api.PostGetOneParams) error {
	var related []string
	if params.Related != nil {
		for _, r := range *params.Related {
			related = append(related, string(r))
		}
	}

	result, err := s.sb.post.ById(c.Context(), id, related)

	switch err {
	case nil:
		response := api.PostFull{}
		if result.Post != nil {
			response.Post = &api.Post{
				Author:   *result.Post.Author,
				Created:  result.Post.Created,
				Forum:    result.Post.Forum,
				Id:       result.Post.Id,
				IsEdited: result.Post.IsEdited,
				Message:  *result.Post.Message,
				Parent:   result.Post.Parent,
				Thread:   result.Post.Thread,
			}
		}
		if result.Author != nil {
			response.Author = &api.User{
				About:    result.Author.About,
				Email:    oapitypes.Email(*result.Author.Email),
				Fullname: *result.Author.FullName,
				Nickname: result.Author.Nickname,
			}
		}
		if result.Forum != nil {
			response.Forum = &api.Forum{
				Slug:    *result.Forum.Slug,
				Title:   *result.Forum.Title,
				User:    *result.Forum.User,
				Posts:   result.Forum.Posts,
				Threads: result.Forum.Threads,
			}
		}
		if result.Thread != nil {
			response.Thread = &api.Thread{
				Author:  *result.Thread.Author,
				Created: result.Thread.Created,
				Forum:   result.Thread.Forum,
				Id:      result.Thread.Id,
				Message: *result.Thread.Message,
				Slug:    result.Thread.Slug,
				Title:   *result.Thread.Title,
				Votes:   result.Thread.Votes,
			}
		}
		return c.JSON(response)
	case post2.ErrNotFound:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString("Can't find user by nickname: "),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// ThreadGetPosts implements api.ServerInterface
func (s *Server) ThreadGetPosts(c *fiber.Ctx, slugOrId string, params api.ThreadGetPostsParams) error {
	threadId, threadIdParseErr := strconv.Atoi(slugOrId)

	var limit int32 = 100
	if params.Limit != nil {
		limit = *params.Limit
	}

	desc := false
	if params.Desc != nil {
		desc = *params.Desc
	}

	var since int64
	if params.Since != nil {
		since = *params.Since
	}

	var err error
	var ucPosts []post2.Post
	sortType := ""
	if params.Sort != nil {
		sortType = string(*params.Sort)
	}

	switch sortType {
	case "tree":
		if threadIdParseErr == nil {
			ucPosts, err = s.sb.post.TreeByThreadId(c.Context(), threadId, limit, desc, since)
		} else {
			ucPosts, err = s.sb.post.TreeByThreadSlug(c.Context(), slugOrId, limit, desc, since)
		}
	case "parent_tree":
		if threadIdParseErr == nil {
			ucPosts, err = s.sb.post.ParentTreeByThreadId(c.Context(), threadId, limit, desc, since)
		} else {
			ucPosts, err = s.sb.post.ParentTreeByThreadSlug(c.Context(), slugOrId, limit, desc, since)
		}
	default:
		if threadIdParseErr == nil {
			ucPosts, err = s.sb.post.FlatByThreadId(c.Context(), threadId, limit, desc, since)
		} else {
			ucPosts, err = s.sb.post.FlatByThreadSlug(c.Context(), slugOrId, limit, desc, since)
		}
	}

	switch err {
	case nil:
		posts := make([]api.Post, len(ucPosts))
		for i, p := range ucPosts {
			posts[i] = api.Post{
				Author:   *p.Author,
				Created:  p.Created,
				Forum:    p.Forum,
				Id:       p.Id,
				IsEdited: p.IsEdited,
				Message:  *p.Message,
				Parent:   p.Parent,
				Thread:   p.Thread,
			}
		}
		return c.JSON(posts)
	case post2.ErrNotFoundThread:
		if threadIdParseErr == nil {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString(fmt.Sprintf("Can't find thread by slug: %d", threadId)),
			})
		} else {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find thread by slug: " + slugOrId),
			})
		}
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// PostUpdate implements api.ServerInterface
func (s *Server) PostUpdate(c *fiber.Ctx, id int64) error {
	var update api.PostUpdate
	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	ucUpdate := post2.PostUpdate{
		Message: update.Message,
	}

	result, err := s.sb.post.ById(c.Context(), id, nil)
	switch err {
	case nil:
		var updateErr error
		if update.Message != nil {
			if *result.Post.Message != *update.Message {
				updateErr = s.sb.post.UpdateById(c.Context(), id, ucUpdate)
				result.Post.IsEdited = ptrBool(true)
				result.Post.Message = update.Message
			}
		}

		if updateErr == nil {
			return c.JSON(api.Post{
				Author:   *result.Post.Author,
				Created:  result.Post.Created,
				Forum:    result.Post.Forum,
				Id:       result.Post.Id,
				IsEdited: result.Post.IsEdited,
				Message:  *result.Post.Message,
				Parent:   result.Post.Parent,
				Thread:   result.Post.Thread,
			})
		}
	case post2.ErrNotFound:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString(fmt.Sprintf("Can't find post with id: %d", id)),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// Vote handlers

// ThreadVote implements api.ServerInterface
func (s *Server) ThreadVote(c *fiber.Ctx, slugOrId string) error {
	var vote api.Vote
	if err := c.BodyParser(&vote); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	ucVote := vote2.Vote{
		Nickname: &vote.Nickname,
		Voice:    &vote.Voice,
	}

	var result *thread2.Thread
	threadId, err := strconv.Atoi(slugOrId)

	if err == nil {
		err = s.sb.vote.AddByThreadId(c.Context(), &ucVote, threadId)
		switch err {
		case nil:
			result, err = s.sb.thread.ById(c.Context(), threadId)
		case vote2.ErrNotFoundThread:
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString(fmt.Sprintf("Can't find thread by id: %d", threadId)),
			})
		case vote2.ErrNotFoundUser:
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user by nickname: " + vote.Nickname),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
		}
	} else {
		err = s.sb.vote.AddByThreadSlug(c.Context(), &ucVote, slugOrId)
		switch err {
		case nil:
			result, err = s.sb.thread.BySlug(c.Context(), slugOrId)
		case vote2.ErrNotFoundThread:
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user by nickname: " + vote.Nickname),
			})
		case vote2.ErrNotFoundUser:
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString(fmt.Sprintf("Can't find thread by id: %d", threadId)),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
		}
	}

	switch err {
	case nil:
		return c.JSON(api.Thread{
			Author:  *result.Author,
			Created: result.Created,
			Forum:   result.Forum,
			Id:      result.Id,
			Message: *result.Message,
			Slug:    result.Slug,
			Title:   *result.Title,
			Votes:   result.Votes,
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// Service handlers

// Status implements api.ServerInterface
func (s *Server) Status(c *fiber.Ctx) error {
	var result api.Status
	err := s.sb.DB().QueryRow(c.Context(), `	SELECT (SELECT COUNT(*) FROM forums),
											(SELECT COUNT(*) FROM posts),
											(SELECT COUNT(*) FROM threads),
											(SELECT COUNT(*) FROM users);`,
	).Scan(&result.Forum, &result.Post, &result.Thread, &result.User)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{
			Message: ptrString(err.Error()),
		})
	}
	return c.JSON(result)
}

// Clear implements api.ServerInterface
func (s *Server) Clear(c *fiber.Ctx) error {
	_, err := s.sb.DB().Exec(c.Context(), "TRUNCATE votes, posts, threads, forums, users, forum_user")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{
			Message: ptrString(err.Error()),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}
