package handlers

import (
	"mzhn/chats/internal/common"
	"mzhn/chats/internal/services/chatservice"

	"github.com/labstack/echo/v4"
)

func Conversations(cs *chatservice.Service) echo.HandlerFunc {
	return func(c echo.Context) error {

		userId := c.Get(string(common.UserCtxKey)).(string)

		ctx := c.Request().Context()

		conversations, err := cs.Conversations(ctx, userId)
		if err != nil {
			return err
		}

		return c.JSON(200, &map[string]any{
			"conversations": conversations,
		})
	}
}
