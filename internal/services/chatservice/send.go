package chatservice

import (
	"context"
	"fmt"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/storage/models"
	"mzhn/chats/pkg/sl"
	"time"
)

func (s *Service) SendMessage(ctx context.Context, in *domain.NewMessageRequest) (*domain.SentMessage, error) {

	fn := "SendMessage"
	log := s.logger.With(sl.Method(fn))

	history, err := s.convProvider.Conversation(ctx, in.ConversationId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

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

	req := &models.RagRequest{
		Input:       in.Body,
		ChatHistory: make([]models.ChatHistoryEntry, 0, len(history.Messages)),
	}

	for _, m := range history.Messages {
		req.ChatHistory = append(req.ChatHistory, models.ChatHistoryEntry{
			IsUser: m.IsUser,
			Body:   m.Body,
		})
	}

	res, err := s.rag.Invoke(ctx, req)
	if err != nil {
		log.Error("failed to invoke rag", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	out := &domain.SentMessage{
		Answer:  res.Answer,
		Sources: make([]domain.AnswerMeta, len(res.Sources)),
	}
	ans := &models.Answer{
		Message: models.Message{
			ConversationId: in.ConversationId,
			Body:           res.Answer,
			CreatedAt:      time.Now(),
		},
		Sources: make([]models.AnswerMeta, len(res.Sources)),
	}

	for i, s := range res.Sources {
		out.Sources[i] = domain.AnswerMeta{
			FileId:   s.FileId,
			FileName: s.Filename,
			Slidenum: s.Slide,
		}
		ans.Sources[i] = models.AnswerMeta{
			Filename: s.Filename,
			Slide:    s.Slide,
			FileId:   s.FileId,
		}
	}

	if err := s.messageSaver.SaveAnswer(ctx, ans); err != nil {
		log.Error("failed to save answer", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return out, nil
}

// func (s *Service) StreamMessage(ctx context.Context, in *domain.StreamMessageRequest) (*domain.SentMessage, error) {
// 	defer close(in.EventCh)

// 	fn := "StreamMessage"
// 	log := s.logger.With(sl.Method(fn))

// 	history, err := s.convProvider.Conversation(ctx, in.ConversationId)
// 	if err != nil {
// 		return nil, fmt.Errorf("%s: %w", fn, err)
// 	}

// 	if err := s.messageSaver.SaveQuestion(ctx, &models.Question{
// 		Message: models.Message{
// 			ConversationId: in.ConversationId,
// 			Body:           in.Body,
// 			CreatedAt:      in.CreatedAt,
// 		},
// 	}); err != nil {
// 		log.Error("failed to save message", sl.Err(err))
// 		return nil, fmt.Errorf("%s: %w", fn, err)
// 	}

// 	events := make(chan []byte)
// 	meta := make(chan *models.AnswerMeta)
// 	done := make(chan error)
// 	sum := new(bytes.Buffer)

// 	req := &models.RagRequest{
// 		Input:       in.Body,
// 		ChatHistory: make([]models.ChatHistoryEntry, 0, len(history.Messages)),
// 	}

// 	for _, m := range history.Messages {
// 		req.ChatHistory = append(req.ChatHistory, models.ChatHistoryEntry{
// 			IsUser: m.IsUser,
// 			Body:   m.Body,
// 		})
// 	}

// 	go func() {
// 		m, err := s.rag.Stream(ctx, req, events)
// 		done <- err
// 		if err != nil {
// 			log.Error("failed to stream rag response", sl.Err(err))
// 			close(meta)
// 		} else {
// 			meta <- m
// 		}
// 		log.Info("rag stream ends", slog.Any("meta", m))
// 	}()

// 	for event := range events {
// 		log := log.With(slog.String("event", string(event)))
// 		log.Debug("writing event")

// 		if _, err := sum.Write(event); err != nil {
// 			log.Warn("failed to write event", sl.Err(err))
// 			return nil, err
// 		}

// 		data := struct {
// 			Response string `json:"response"`
// 		}{
// 			Response: string(event),
// 		}

// 		j, err := json.Marshal(data)
// 		if err != nil {
// 			log.Error("failed to marshal event", sl.Err(err))
// 			return nil, err
// 		}

// 		in.EventCh <- j
// 		log.Debug("sent event")
// 	}

// 	log.Debug("not enough events")

// 	if err := <-done; err != nil {
// 		log.Error("failed to send message", sl.Err(err))
// 		return nil, fmt.Errorf("%s: %w", fn, err)
// 	}

// 	log.Debug("waiting meta")
// 	var m *models.AnswerMeta
// 	if me, ok := <-meta; ok {
// 		log.Info("rag answer", slog.Any("answer", m))
// 		m = me
// 	}

// 	if err := s.messageSaver.SaveAnswer(ctx, &models.Answer{
// 		Message: models.Message{
// 			ConversationId: in.ConversationId,
// 			Body:           sum.String(),
// 			CreatedAt:      time.Now(),
// 		},
// 		AnswerMeta: *m,
// 	}); err != nil {
// 		log.Error("failed to save answer", sl.Err(err))
// 		return nil, fmt.Errorf("%s: %w", fn, err)
// 	}

// 	metadata := struct {
// 		FileId   string `json:"fileId"`
// 		FileName string `json:"filename"`
// 		Slidenum int    `json:"slidenum"`
// 	}{
// 		FileId:   m.FileId,
// 		FileName: m.Filename,
// 		Slidenum: m.Slide,
// 	}

// 	j, err := json.Marshal(metadata)
// 	if err != nil {
// 		log.Error("failed to marshal metadata", sl.Err(err))
// 		return nil, fmt.Errorf("%s: %w", fn, err)
// 	}

// 	in.EventCh <- j

// 	return &domain.SentMessage{}, nil
// }
