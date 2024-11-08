package handlers

import (
	"mzhn/chats/internal/common"
	"mzhn/chats/internal/services/chatservice"

	"github.com/labstack/echo/v4"
)

type CreateConversationResponse struct {
	Id string `json:"id"`
}

func CreateConversation(cs *chatservice.Service) echo.HandlerFunc {
	return func(c echo.Context) error {

		userId := c.Get(string(common.UserCtxKey)).(string)
		ctx := c.Request().Context()

		id, err := cs.CreateConversation(ctx, userId)
		if err != nil {
			return err
		}

		return c.JSON(200, &CreateConversationResponse{
			Id: id,
		})
	}
}
