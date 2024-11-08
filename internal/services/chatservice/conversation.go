package chatservice

import (
	"context"
	"log/slog"
	"mzhn/chats/internal/domain"
	"mzhn/chats/pkg/sl"
)

func (s *Service) Conversation(ctx context.Context, id string) (*domain.ConversationContent, error) {
	fn := "Conversation"
	log := s.logger.With(sl.Method(fn), slog.String("conversation_id", id))

	conv, err := s.convProvider.Conversation(ctx, id)
	if err != nil {
		log.Error("failed to get conversation", sl.Err(err))
		return nil, err
	}

	return conv, nil
}
