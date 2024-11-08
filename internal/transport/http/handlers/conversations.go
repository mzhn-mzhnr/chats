package handlers

import (
	"mzhn/chats/internal/common"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/services/chatservice"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

type ConversationsResponse struct {
	Conversations []Conversation `json:"conversations"`
}

// @Summary	Получение диалогов пользователя
// @Tags		conversation
// @Success	200	{object}	ConversationsResponse
// @Router		/ [get]
// @Security Bearer
func GetConversations(cs *chatservice.Service) echo.HandlerFunc {
	return func(c echo.Context) error {

		userId := c.Get(string(common.UserCtxKey)).(string)

		ctx := c.Request().Context()

		conversations, err := cs.Conversations(ctx, userId)
		if err != nil {
			return err
		}

		conv := lo.Map(conversations, func(c *domain.Conversation, _ int) Conversation {
			return Conversation{
				Id:        c.Id,
				Name:      c.Name,
				CreatedAt: c.CreatedAt,
			}
		})

		return c.JSON(200, &ConversationsResponse{Conversations: conv})
	}
}
