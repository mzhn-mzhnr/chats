package middleware

import (
	"errors"
	"mzhn/chats/internal/common"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/services/authservice"
	"mzhn/chats/internal/storage"
	"strings"

	"github.com/labstack/echo/v4"
)

type AuthMiddlewareFunc func(roles ...string) echo.MiddlewareFunc

func AuthGuard(svc *authservice.Service) AuthMiddlewareFunc {
	return func(roles ...string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				ctx := c.Request().Context()

				authorization := c.Request().Header.Get("Authorization")

				spl := strings.Split(authorization, " ")

				if len(spl) < 2 {
					return echo.ErrUnauthorized
				}

				token := spl[1]

				user, err := svc.Auth(ctx, &domain.AuthRequest{
					Token: token,
					Roles: roles,
				})
				if err != nil {
					if errors.Is(err, storage.ErrUnauthorized) {
						return echo.ErrUnauthorized
					}
					return err
				}

				c.Set(string(common.UserCtxKey), user.Id)

				return next(c)
			}
		}
	}
}
