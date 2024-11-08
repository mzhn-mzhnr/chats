package handlers

import (
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/services/chatservice"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

type ConversationRequest struct {
	Id string `param:"id"`
}

type Conversation struct {
	Id        string    `json:"id"`
	Name      *string   `json:"name,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Message struct {
	Id        int       `json:"id"`
	Body      string    `json:"body"`
	IsUser    bool      `json:"isUser"`
	CreatedAt time.Time `json:"createdAt"`
}

type ConversationResponse struct {
	Conversation Conversation `json:"conversation"`
	Messages     []Message    `json:"messages"`
}

func GetConversation(cs *chatservice.Service) echo.HandlerFunc {
	return func(c echo.Context) error {

		var request ConversationRequest

		if err := c.Bind(&request); err != nil {
			return err
		}
		ctx := c.Request().Context()

		conv, err := cs.Conversation(ctx, request.Id)
		if err != nil {
			return err
		}

		response := &ConversationResponse{
			Conversation: Conversation{
				Id:        conv.Id,
				Name:      conv.Name,
				CreatedAt: conv.CreatedAt,
			},
			Messages: lo.Map(conv.Messages, func(m *domain.Message, _ int) Message {
				return Message{
					Id:        m.Id,
					Body:      m.Body,
					IsUser:    m.IsUser,
					CreatedAt: m.CreatedAt,
				}
			}),
		}

		return c.JSON(200, response)
	}
}
