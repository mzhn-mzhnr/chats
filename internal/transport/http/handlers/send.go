package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/services/chatservice"
	"time"

	"github.com/labstack/echo/v4"
)

type SendMessageRequest struct {
	ConversationId string `json:"conversationId"`
	Input          string `json:"input"`
}

type Event struct {
	ID      []byte
	Data    []byte
	Event   []byte
	Retry   []byte
	Comment []byte
}

// MarshalTo marshals Event to given Writer
func (ev *Event) MarshalTo(w io.Writer) error {
	if len(ev.Data) == 0 && len(ev.Comment) == 0 {
		return nil
	}

	if len(ev.Data) > 0 {
		if _, err := fmt.Fprintf(w, "id: %s\n", ev.ID); err != nil {
			return err
		}

		sd := bytes.Split(ev.Data, []byte("\n"))
		for i := range sd {
			if _, err := fmt.Fprintf(w, "data: %s\n", sd[i]); err != nil {
				return err
			}
		}

		if len(ev.Event) > 0 {
			if _, err := fmt.Fprintf(w, "event: %s\n", ev.Event); err != nil {
				return err
			}
		}

		if len(ev.Retry) > 0 {
			if _, err := fmt.Fprintf(w, "retry: %s\n", ev.Retry); err != nil {
				return err
			}
		}
	}

	if len(ev.Comment) > 0 {
		if _, err := fmt.Fprintf(w, ": %s\n", ev.Comment); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return err
	}

	return nil
}

// @Summary	Send message
// @Tags		conversation
// @Param input body SendMessageRequest true "Данные для отправки"
// @Success	200	{object}	SendMessageRequest
// @Router		/send [post]
// @Security Bearer
func SendMessage(cs *chatservice.Service) echo.HandlerFunc {
	log := slog.With(slog.String("endpoint", "POST /send"))
	return func(c echo.Context) error {
		var req SendMessageRequest

		if err := c.Bind(&req); err != nil {
			return echo.ErrBadRequest
		}

		ctx := c.Request().Context()

		eventCh := make(chan []byte)
		done := make(chan error)

		go func() {
			_, err := cs.SendMessage(ctx, &domain.NewMessage{
				Body:           req.Input,
				ConversationId: req.ConversationId,
				IsUser:         true,
				CreatedAt:      time.Now(),
				EventCh:        eventCh,
			})

			done <- err
		}()

		w := c.Response()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		for {
			select {
			case <-c.Request().Context().Done():
				log.Info("SSE client disconnected, ip: %v", c.RealIP())
				return nil
			case event, ok := <-eventCh:
				if !ok {
					log.Info("streaming response ended")
					return nil
				}

				e := Event{
					Data: event,
				}
				log.Debug("marshaling event", slog.String("event", string(e.Data)))
				if err := e.MarshalTo(w); err != nil {
					return err
				}

				w.Flush()
			}
		}
	}
}
