package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/viewsharp/technopark-forum/internal/api"
)

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
