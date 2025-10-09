package controller

import (
	"github.com/gofiber/fiber/v2"
	oapitypes "github.com/oapi-codegen/runtime/types"

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
	switch err {
	case nil:
		return c.Status(fiber.StatusCreated).JSON(api.User{
			About:    user.About,
			Email:    user.Email,
			Fullname: user.Fullname,
			Nickname: ptrString(nickname),
		})
	case domain.ErrUniqueViolation:
		var result []api.User

		userByEmail, err := s.sb.User.ByEmail(c.Context(), string(user.Email))
		if err == nil {
			result = append(result, api.User{
				About:    userByEmail.About,
				Email:    oapitypes.Email(userByEmail.Email),
				Fullname: userByEmail.FullName,
				Nickname: &userByEmail.Nickname,
			})
		}

		userByNickname, err := s.sb.User.ByNickname(c.Context(), nickname)
		if err == nil {
			if userByEmail == nil {
				result = append(result, api.User{
					About:    userByNickname.About,
					Email:    oapitypes.Email(userByNickname.Email),
					Fullname: userByNickname.FullName,
					Nickname: &userByNickname.Nickname,
				})
			} else if userByNickname.Nickname != userByEmail.Nickname {
				result = append(result, api.User{
					About:    userByNickname.About,
					Email:    oapitypes.Email(userByNickname.Email),
					Fullname: userByNickname.FullName,
					Nickname: &userByNickname.Nickname,
				})
			}
		}

		return c.Status(fiber.StatusConflict).JSON(result)
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}

// UserGetOne implements api.ServerInterface
func (s *Server) UserGetOne(c *fiber.Ctx, nickname string) error {
	result, err := s.sb.User.ByNickname(c.Context(), nickname)

	switch err {
	case nil:
		return c.JSON(api.User{
			About:    result.About,
			Email:    oapitypes.Email(result.Email),
			Fullname: result.FullName,
			Nickname: &result.Nickname,
		})
	case domain.ErrNotFound:
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

	ucUpdate := domain.UserUpdate{
		About:    update.About,
		Email:    emailStr,
		FullName: update.Fullname,
	}

	user, err := s.sb.User.UpdateByNickname(c.Context(), nickname, &ucUpdate)

	switch err {
	case nil:
		return c.JSON(api.User{
			About:    user.About,
			Email:    oapitypes.Email(user.Email),
			Fullname: user.FullName,
			Nickname: &user.Nickname,
		})
	case domain.ErrUniqueViolation:
		return c.Status(fiber.StatusConflict).JSON(api.Error{
			Message: ptrString("This email is already registered by user: " + string(*update.Email)),
		})
	case domain.ErrNotFound:
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

	result, err := s.sb.User.ByForumSlug(c.Context(), slug, desc, since, limit)

	switch err {
	case nil:
		users := make([]api.User, len(*result))
		for i, u := range *result {
			users[i] = api.User{
				About:    u.About,
				Email:    oapitypes.Email(u.Email),
				Fullname: u.FullName,
				Nickname: &u.Nickname,
			}
		}
		return c.JSON(users)
	case domain.ErrUserNotFoundForum:
		return c.Status(fiber.StatusNotFound).JSON(api.Error{
			Message: ptrString("Can't find forum by slug: " + slug),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Message: ptrString(err.Error())})
}
