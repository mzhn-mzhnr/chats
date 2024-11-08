package middleware

import (
	"mzhn/chats/internal/common"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/services/chatservice"

	"github.com/labstack/echo/v4"
)

type AuthMiddleware func(roles ...string) echo.MiddlewareFunc

func AuthGuard(svc *chatservice.Service) AuthMiddleware {
	return func(roles ...string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {

				ctx := c.Request().Context()
				user, err := svc.Auth(ctx, &domain.AuthRequest{
					Token: c.Request().Header.Get("Authorization"),
					Roles: roles,
				})
				if err != nil {
					return err
				}

				c.Set(string(common.UserCtxKey), user)

				return next(c)
			}
		}
	}
}
