package controller

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/viewsharp/technopark-forum/internal/api"
	"github.com/viewsharp/technopark-forum/internal/domain"
)

// ThreadCreate implements api.ServerInterface
func (s *Server) ThreadCreate(c *fiber.Ctx, slug string) error {
	var thread api.Thread
	if err := c.BodyParser(&thread); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	ucThread := domain.Thread{
		Title:   thread.Title,
		Message: thread.Message,
		Author:  thread.Author,
		Created: thread.Created,
		Forum:   &slug,
		Slug:    thread.Slug,
	}

	err := s.sb.Thread.Add(c.Context(), &ucThread)
	if err != nil {
		if errors.Is(err, domain.ErrUniqueViolation) {
			result, err := s.sb.Thread.BySlug(c.Context(), *thread.Slug)
			if err == nil {
				return c.Status(fiber.StatusConflict).JSON(domainThreadToAPI(result))
			}
		}
		if errors.Is(err, domain.ErrThreadNotFoundUser) {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find thread author by nickname: " + thread.Author),
			})
		}
		if errors.Is(err, domain.ErrThreadNotFoundForum) {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find thread forum by slug: " + slug),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}
	return c.Status(fiber.StatusCreated).JSON(domainThreadToAPI(&ucThread))
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

	result, err := s.sb.Thread.ByForumSlug(c.Context(), slug, desc, since, limit)
	if err != nil {
		if errors.Is(err, domain.ErrThreadNotFoundForum) {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find forum by slug: " + slug),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	return c.JSON(domainThreadsToAPI(result))
}

// ThreadGetOne implements api.ServerInterface
func (s *Server) ThreadGetOne(c *fiber.Ctx, slugOrId string) error {
	var result *domain.Thread
	var err error

	threadId, threadIdParseErr := strconv.Atoi(slugOrId)
	if threadIdParseErr == nil {
		result, err = s.sb.Thread.ById(c.Context(), threadId)
	} else {
		result, err = s.sb.Thread.BySlug(c.Context(), slugOrId)
	}

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
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
		c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	return c.JSON(domainThreadToAPI(result))
}

// ThreadUpdate implements api.ServerInterface
func (s *Server) ThreadUpdate(c *fiber.Ctx, slugOrId string) error {
	var update api.ThreadUpdate
	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	ucUpdate := domain.ThreadUpdate{
		Message: update.Message,
		Title:   update.Title,
	}

	threadId, threadIdErr := strconv.Atoi(slugOrId)
	if threadIdErr == nil {
		err := s.sb.Thread.UpdateById(c.Context(), threadId, &ucUpdate)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
		}
	} else {
		err := s.sb.Thread.UpdateBySlug(c.Context(), slugOrId, &ucUpdate)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
		}
	}

	var result *domain.Thread
	var err error
	if threadIdErr == nil {
		result, err = s.sb.Thread.ById(c.Context(), threadId)
	} else {
		result, err = s.sb.Thread.BySlug(c.Context(), slugOrId)
	}

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			if threadIdErr == nil {
				return c.Status(fiber.StatusNotFound).JSON(api.Error{
					Message: ptrString("Can't find thread by id: " + strconv.Itoa(threadId)),
				})
			} else {
				return c.Status(fiber.StatusNotFound).JSON(api.Error{
					Message: ptrString("Can't find thread by slug: " + slugOrId),
				})
			}
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	return c.JSON(domainThreadToAPI(result))
}
