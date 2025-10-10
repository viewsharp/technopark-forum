package controller

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/viewsharp/technopark-forum/internal/api"
	"github.com/viewsharp/technopark-forum/internal/domain"
)

// ThreadVote implements api.ServerInterface
func (s *Server) ThreadVote(c *fiber.Ctx, slugOrId string) error {
	var vote api.Vote
	if err := c.BodyParser(&vote); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	ucVote := domain.Vote{
		Nickname: &vote.Nickname,
		Voice:    &vote.Voice,
	}

	threadId, threadIdParseErr := strconv.Atoi(slugOrId)

	var err error
	if threadIdParseErr == nil {
		err = s.sb.Vote.AddByThreadId(c.Context(), &ucVote, threadId)
	} else {
		err = s.sb.Vote.AddByThreadSlug(c.Context(), &ucVote, slugOrId)
	}
	if err != nil {
		if errors.Is(err, domain.ErrVoteNotFoundThread) {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user by nickname: " + vote.Nickname),
			})
		}
		if errors.Is(err, domain.ErrVoteNotFoundUser) {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user by nickname: " + vote.Nickname),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	var result *domain.Thread
	if threadIdParseErr == nil {
		result, err = s.sb.Thread.ById(c.Context(), threadId)
	} else {
		result, err = s.sb.Thread.BySlug(c.Context(), slugOrId)
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	return c.JSON(domainThreadToAPI(result))
}
