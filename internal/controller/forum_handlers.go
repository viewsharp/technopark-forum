package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/viewsharp/technopark-forum/internal/api"
	"github.com/viewsharp/technopark-forum/internal/domain"
)

// ForumCreate implements api.ServerInterface
func (s *Server) ForumCreate(c *fiber.Ctx) error {
	var forum api.Forum
	if err := c.BodyParser(&forum); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	// Convert api.Forum to domain.Forum
	ucForum := domain.Forum{
		Slug:  forum.Slug,
		Title: forum.Title,
		User:  forum.User,
	}

	createdForum, err := s.sb.Forum.Add(c.Context(), ucForum)
	if err != nil {
		switch err {
		case domain.ErrUniqueViolation:
			result, err := s.sb.Forum.BySlug(c.Context(), forum.Slug)
			if err == nil {
				return c.Status(fiber.StatusConflict).JSON(api.Forum{
					Slug:    result.Slug,
					Title:   result.Title,
					User:    result.User,
					Posts:   result.Posts,
					Threads: result.Threads,
				})
			}
		case domain.ErrForumNotFoundUser:
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user with nickname: " + forum.User),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	return c.Status(fiber.StatusCreated).JSON(api.Forum{
		Slug:    createdForum.Slug,
		Title:   createdForum.Title,
		User:    createdForum.User,
		Posts:   createdForum.Posts,
		Threads: createdForum.Threads,
	})
}

// ForumGetOne implements api.ServerInterface
func (s *Server) ForumGetOne(c *fiber.Ctx, slug string) error {
	result, err := s.sb.Forum.FullBySlug(c.Context(), slug)

	switch err {
	case nil:
		return c.JSON(api.Forum{
			Slug:    result.Slug,
			Title:   result.Title,
			User:    result.User,
			Posts:   result.Posts,
			Threads: result.Threads,
		})
	case domain.ErrNotFound:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString("Can't find forum by slug: " + slug),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}
