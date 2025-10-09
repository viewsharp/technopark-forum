package controller

import (
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

	var result *domain.Thread
	threadId, err := strconv.Atoi(slugOrId)

	if err == nil {
		err = s.sb.Vote.AddByThreadId(c.Context(), &ucVote, threadId)
		switch err {
		case nil:
			result, err = s.sb.Thread.ById(c.Context(), threadId)
		case domain.ErrVoteNotFoundThread:
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user by nickname: " + vote.Nickname),
			})
		case domain.ErrVoteNotFoundUser:
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user by nickname: " + vote.Nickname),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
		}
	} else {
		err = s.sb.Vote.AddByThreadSlug(c.Context(), &ucVote, slugOrId)
		switch err {
		case nil:
			result, err = s.sb.Thread.BySlug(c.Context(), slugOrId)
		case domain.ErrVoteNotFoundThread:
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user by nickname: " + vote.Nickname),
			})
		case domain.ErrVoteNotFoundUser:
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user by nickname: " + vote.Nickname),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
		}
	}

	switch err {
	case nil:
		return c.JSON(domainThreadToAPI(result))
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}
