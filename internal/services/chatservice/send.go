package chatservice

import (
	"context"
	"fmt"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/storage/models"
	"mzhn/chats/pkg/sl"
)

func (s *Service) SendMessage(ctx context.Context, in *domain.NewMessage) (*domain.SentMessage, error) {
	fn := "SendMessage"
	log := s.logger.With(sl.Method(fn))

	if err := s.messageSaver.SaveMessage(ctx, &models.NewMessage{
		ConversationId: in.ConversationId,
		IsUser:         in.UserId != nil,
		Body:           in.Body,
	}); err != nil {
		log.Error("failed to save message", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &domain.SentMessage{
		ConversationId: in.ConversationId,
	}, nil
}
