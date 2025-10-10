package controller

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/viewsharp/technopark-forum/internal/api"
	"github.com/viewsharp/technopark-forum/internal/domain"
)

// UserCreate implements api.ServerInterface
func (s *Server) UserCreate(c *fiber.Ctx, nickname string) error {
	var user api.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.Error{Message: ptrString(err.Error())})
	}

	ucUser := domain.User{
		About:    user.About,
		Email:    string(user.Email),
		FullName: user.Fullname,
		Nickname: nickname,
	}

	err := s.sb.User.Add(c.Context(), &ucUser)
	if err != nil {
		if errors.Is(err, domain.ErrUniqueViolation) {
			var result []api.User

			userByEmail, err := s.sb.User.ByEmail(c.Context(), string(user.Email))
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
			}

			userByNickname, err := s.sb.User.ByNickname(c.Context(), nickname)
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
			}

			if userByEmail != nil {
				result = append(result, domainUserToAPI(userByEmail))
			}
			if userByNickname != nil && (userByEmail == nil || userByNickname.Nickname != userByEmail.Nickname) {
				result = append(result, domainUserToAPI(userByNickname))
			}

			return c.Status(fiber.StatusConflict).JSON(result)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	return c.Status(fiber.StatusCreated).JSON(api.User{
		About:    user.About,
		Email:    user.Email,
		Fullname: user.Fullname,
		Nickname: ptrString(nickname),
	})
}

// UserGetOne implements api.ServerInterface
func (s *Server) UserGetOne(c *fiber.Ctx, nickname string) error {
	result, err := s.sb.User.ByNickname(c.Context(), nickname)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user by nickname: " + nickname),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	return c.JSON(domainUserToAPI(result))
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

	ucUpdate := domain.UserUpdate{
		About:    update.About,
		Email:    emailStr,
		FullName: update.Fullname,
	}

	user, err := s.sb.User.UpdateByNickname(c.Context(), nickname, &ucUpdate)
	if err != nil {
		if errors.Is(err, domain.ErrUniqueViolation) {
			return c.Status(fiber.StatusConflict).JSON(api.Error{
				Message: ptrString("This email is already registered by user: " + string(*update.Email)),
			})
		}
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find user by nickname: " + nickname),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	return c.JSON(domainUserToAPI(user))
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

	result, err := s.sb.User.ByForumSlug(c.Context(), slug, desc, since, limit)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFoundForum) {
			return c.Status(fiber.StatusNotFound).JSON(api.Error{
				Message: ptrString("Can't find forum by slug: " + slug),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
	}

	return c.JSON(domainUsersToAPI(result))

}
