package chatservice

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/storage/models"
	"mzhn/chats/pkg/sl"
	"time"
)

func (s *Service) SendMessage(ctx context.Context, in *domain.NewMessage) (*domain.SentMessage, error) {
	defer close(in.EventCh)

	fn := "SendMessage"
	log := s.logger.With(sl.Method(fn))

	if err := s.messageSaver.SaveQuestion(ctx, &models.Question{
		Message: models.Message{
			ConversationId: in.ConversationId,
			Body:           in.Body,
			CreatedAt:      in.CreatedAt,
		},
	}); err != nil {
		log.Error("failed to save message", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	events := make(chan []byte)
	meta := make(chan *models.AnswerMeta)
	done := make(chan error)
	sum := new(bytes.Buffer)

	go func() {
		m, err := s.rag.Stream(ctx, in.Body, events)
		done <- err
		if err != nil {
			log.Error("failed to stream rag response", sl.Err(err))
			close(meta)
		}
		meta <- m
		log.Info("rag stream ends", slog.Any("meta", m))
	}()

	for event := range events {
		log := log.With(slog.String("event", string(event)))
		log.Debug("writing event")

		if _, err := sum.Write(event); err != nil {
			log.Warn("failed to write event", sl.Err(err))
			return nil, err
		}

		in.EventCh <- event
		log.Debug("sent event")
	}

	log.Debug("not enough events")

	if err := <-done; err != nil {
		log.Error("failed to send message", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	log.Debug("waiting meta")
	var m *models.AnswerMeta
	if me, ok := <-meta; ok {
		log.Info("rag answer", slog.Any("answer", m))
		m = me
	}

	if err := s.messageSaver.SaveAnswer(ctx, &models.Answer{
		Message: models.Message{
			ConversationId: in.ConversationId,
			Body:           sum.String(),
			CreatedAt:      time.Now(),
		},
		AnswerMeta: *m,
	}); err != nil {
		log.Error("failed to save answer", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &domain.SentMessage{
		ConversationId: in.ConversationId,
	}, nil
}
