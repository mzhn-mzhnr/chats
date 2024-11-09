package chatservice

import (
	"bytes"
	"context"
	"fmt"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/storage/models"
	"mzhn/chats/pkg/sl"
	"time"
)

func (s *Service) SendMessage(ctx context.Context, in *domain.NewMessage) (*domain.SentMessage, error) {
	fn := "SendMessage"
	log := s.logger.With(sl.Method(fn))

	if err := s.messageSaver.SaveMessage(ctx, &models.NewMessage{
		ConversationId: in.ConversationId,
		IsUser:         in.IsUser,
		Body:           in.Body,
		CreatedAt:      in.CreatedAt,
	}); err != nil {
		log.Error("failed to save message", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	events := make(chan []byte)
	done := make(chan error)
	sum := new(bytes.Buffer)

	go func() {
		done <- s.rag.Stream(ctx, in.Body, events)
	}()

	for event := range events {
		if _, err := sum.Write(event); err != nil {
			log.Warn("failed to write event", sl.Err(err))
			return nil, err
		}
		in.EventCh <- event
	}

	if err := <-done; err != nil {
		log.Error("failed to send message", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	if err := s.messageSaver.SaveMessage(ctx, &models.NewMessage{
		ConversationId: in.ConversationId,
		IsUser:         false,
		Body:           sum.String(),
		CreatedAt:      time.Now(),
	}); err != nil {
		log.Error("failed to save answer", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &domain.SentMessage{
		ConversationId: in.ConversationId,
	}, nil
}
