package handlers

import (
	"mzhn/chats/internal/services/chatservice"

	"github.com/labstack/echo/v4"
)

func Conversations(cs *chatservice.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(200, "conversations")
	}
}
