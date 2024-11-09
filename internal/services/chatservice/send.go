package chatservice

import (
	"bytes"
	"context"
	"encoding/json"
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

		data := struct {
			Response string `json:"response"`
		}{
			Response: string(event),
		}

		j, err := json.Marshal(data)
		if err != nil {
			log.Error("failed to marshal event", sl.Err(err))
			return nil, err
		}

		in.EventCh <- j
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

	metadata := struct {
		FileId   string `json:"fileId"`
		FileName string `json:"filename"`
		Slidenum int    `json:"slidenum"`
	}{
		FileId:   m.FileId,
		FileName: m.Filename,
		Slidenum: m.Slide,
	}

	j, err := json.Marshal(metadata)
	if err != nil {
		log.Error("failed to marshal metadata", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	in.EventCh <- j

	return &domain.SentMessage{
		ConversationId: in.ConversationId,
		AnswerMeta: &domain.AnswerMeta{
			FileId:   m.FileId,
			FileName: m.Filename,
			Slidenum: m.Slide,
		},
	}, nil
}
