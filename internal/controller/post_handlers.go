package controller

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	oapitypes "github.com/oapi-codegen/runtime/types"

	"github.com/viewsharp/technopark-forum/internal/api"
	"github.com/viewsharp/technopark-forum/internal/domain"
)

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
			_, err = s.sb.Thread.ById(c.Context(), threadId)
		} else {
			_, err = s.sb.Thread.BySlug(c.Context(), slugOrId)
		}

		switch err {
		case nil:
			return c.Status(fiber.StatusCreated).JSON(posts)
		case domain.ErrNotFound:
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

	// Convert api.Post to domain.Post
	ucPosts := make([]domain.Post, len(posts))
	for i, p := range posts {
		ucPosts[i] = domain.Post{
			Message:  p.Message,
			Author:   p.Author,
			Created:  p.Created,
			Forum:    p.Forum,
			Id:       p.Id,
			IsEdited: p.IsEdited,
			Parent:   p.Parent,
			Thread:   p.Thread,
		}
	}

	var err error
	if threadIdParseErr == nil {
		err = s.sb.Post.AddByThreadId(c.Context(), ucPosts, int32(threadId))
	} else {
		err = s.sb.Post.AddByThreadSlug(c.Context(), ucPosts, slugOrId)
	}

	if err != nil {
		if errors.Is(err, domain.ErrPostInvalidParent) {
			return c.Status(fiber.StatusConflict).JSON(api.Error{
				Message: ptrString("Parent post was created in another thread"),
			})
		}
		if errors.Is(err, domain.ErrPostNotFoundThread) {
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

		var errNotFoundUser domain.ErrPostNotFoundUser
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
			Message:  p.Message,
			Author:   p.Author,
			Created:  p.Created,
			Forum:    p.Forum,
			Id:       p.Id,
			IsEdited: p.IsEdited,
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

	result, err := s.sb.Post.ById(c.Context(), id, related)

	switch err {
	case nil:
		response := api.PostFull{}
		if result.Post != nil {
			response.Post = &api.Post{
				Message:  result.Post.Message,
				Author:   result.Post.Author,
				Created:  result.Post.Created,
				Forum:    result.Post.Forum,
				Id:       result.Post.Id,
				IsEdited: result.Post.IsEdited,
				Parent:   result.Post.Parent,
				Thread:   result.Post.Thread,
			}
		}
		if result.Author != nil {
			response.Author = &api.User{
				About:    result.Author.About,
				Email:    oapitypes.Email(result.Author.Email),
				Fullname: result.Author.FullName,
				Nickname: &result.Author.Nickname,
			}
		}
		if result.Forum != nil {
			response.Forum = &api.Forum{
				Slug:    result.Forum.Slug,
				Title:   result.Forum.Title,
				User:    result.Forum.User,
				Posts:   result.Forum.Posts,
				Threads: result.Forum.Threads,
			}
		}
		if result.Thread != nil {
			response.Thread = &api.Thread{
				Author:  result.Thread.Author,
				Created: result.Thread.Created,
				Forum:   result.Thread.Forum,
				Id:      result.Thread.Id,
				Message: result.Thread.Message,
				Slug:    result.Thread.Slug,
				Title:   result.Thread.Title,
				Votes:   result.Thread.Votes,
			}
		}
		return c.JSON(response)
	case domain.ErrNotFound:
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
	var ucPosts []domain.Post
	sortType := ""
	if params.Sort != nil {
		sortType = string(*params.Sort)
	}

	switch sortType {
	case "tree":
		if threadIdParseErr == nil {
			ucPosts, err = s.sb.Post.TreeByThreadId(c.Context(), threadId, limit, desc, since)
		} else {
			ucPosts, err = s.sb.Post.TreeByThreadSlug(c.Context(), slugOrId, limit, desc, since)
		}
	case "parent_tree":
		if threadIdParseErr == nil {
			ucPosts, err = s.sb.Post.ParentTreeByThreadId(c.Context(), threadId, limit, desc, since)
		} else {
			ucPosts, err = s.sb.Post.ParentTreeByThreadSlug(c.Context(), slugOrId, limit, desc, since)
		}
	default:
		if threadIdParseErr == nil {
			ucPosts, err = s.sb.Post.FlatByThreadId(c.Context(), threadId, limit, desc, since)
		} else {
			ucPosts, err = s.sb.Post.FlatByThreadSlug(c.Context(), slugOrId, limit, desc, since)
		}
	}

	switch err {
	case nil:
		posts := make([]api.Post, len(ucPosts))
		for i, p := range ucPosts {
			posts[i] = api.Post{
				Message:  p.Message,
				Author:   p.Author,
				Created:  p.Created,
				Forum:    p.Forum,
				Id:       p.Id,
				IsEdited: p.IsEdited,
				Parent:   p.Parent,
				Thread:   p.Thread,
			}
		}
		return c.JSON(posts)
	case domain.ErrPostNotFoundThread:
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

	ucUpdate := domain.PostUpdate{
		Message: update.Message,
	}

	result, err := s.sb.Post.ById(c.Context(), id, nil)
	switch err {
	case nil:
		var updateErr error
		if update.Message != nil {
			if result.Post.Message != *update.Message {
				updateErr = s.sb.Post.UpdateById(c.Context(), id, ucUpdate)
				result.Post.IsEdited = ptrBool(true)
				result.Post.Message = *update.Message
			}
		}

		if updateErr == nil {
			return c.JSON(api.Post{
				Message:  result.Post.Message,
				Author:   result.Post.Author,
				Created:  result.Post.Created,
				Forum:    result.Post.Forum,
				Id:       result.Post.Id,
				IsEdited: result.Post.IsEdited,
				Parent:   result.Post.Parent,
				Thread:   result.Post.Thread,
			})
		}
	case domain.ErrNotFound:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString(fmt.Sprintf("Can't find post with id: %d", id)),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}
