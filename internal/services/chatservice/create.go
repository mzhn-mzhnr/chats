package chatservice

import (
	"context"
	"fmt"
	"log/slog"
	"mzhn/chats/internal/domain"
	"mzhn/chats/pkg/sl"
)

func (s *Service) CreateConversation(ctx context.Context, userId string) (string, error) {
	fn := "CreateConversation"
	log := s.logger.With(sl.Method(fn))

	id, err := s.convCreator.CreateConversation(ctx, &domain.NewConversation{
		UserId: userId,
	})
	if err != nil {
		log.Error("failed to create conversation", sl.Err(err))
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	log.Debug("created conversation", slog.String("conversation_id", id))

	return id, nil
}
