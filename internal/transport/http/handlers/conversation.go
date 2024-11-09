package handlers

import (
	"mzhn/chats/internal/services/chatservice"
	"time"

	"github.com/labstack/echo/v4"
)

type ConversationRequest struct {
	Id string `param:"id"`
}

type Conversation struct {
	Id        string    `json:"id"`
	Name      *string   `json:"name,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Meta struct {
	FileId   string `json:"fileId"`
	Filename string `json:"fileName"`
	SlideNum int    `json:"slideNum"`
}
type Message struct {
	Id        int       `json:"id"`
	Body      string    `json:"body"`
	IsUser    bool      `json:"isUser"`
	CreatedAt time.Time `json:"createdAt"`
	Sources   []Meta    `json:"sources"`
}

type ConversationResponse struct {
	Conversation Conversation `json:"conversation"`
	Messages     []Message    `json:"messages"`
}

// @Summary	Получение диалога по ID
// @Param		id	path	int	true	"conversation ID"
// @Tags		conversation
// @Success	200	{object}	ConversationResponse
// @Router		/{id} [get]
// @Security Bearer
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
			Messages: make([]Message, len(conv.Messages)),
		}

		for i, m := range conv.Messages {
			response.Messages[i] = Message{
				Id:        m.Id,
				Body:      m.Body,
				IsUser:    m.IsUser,
				CreatedAt: m.CreatedAt,
			}

			if !m.IsUser {
				response.Messages[i].Sources = make([]Meta, len(m.Sources))
				for j, s := range m.Sources {
					response.Messages[i].Sources[j] = Meta{
						FileId:   s.FileId,
						Filename: s.FileName,
						SlideNum: s.Slidenum,
					}
				}
			}
		}

		return c.JSON(200, response)
	}
}
