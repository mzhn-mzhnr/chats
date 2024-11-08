package chatservice

import (
	"context"
	"mzhn/chats/internal/domain"
	"mzhn/chats/pkg/sl"
)

func (s *Service) Conversations(ctx context.Context) ([]*domain.Conversation, error) {

	fn := "Conversations"
	log := s.logger.With(sl.Method(fn))

	cc, err := s.convProvider.Conversations(ctx, &domain.ConversationsFilter{
		UserId: userId,
	})
	if err != nil {
		log.Error("failed to get conversations", sl.Err(err))
		return nil, err
	}

	return cc, nil
}
