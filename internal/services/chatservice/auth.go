package chatservice

import (
	"context"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/storage/models"
	"mzhn/chats/pkg/sl"
)

func (s *Service) auth(ctx context.Context, in *domain.AuthRequest) (*models.User, error) {
	fn := "auth"
	log := s.logger.With(sl.Method(fn))

	user, err := s.authProvider.Authenticate(ctx, &models.AuthenticateRequest{
		Token: in.Token,
		Roles: in.Roles,
	})
	if err != nil {
		log.Error("failed to auth", sl.Err(err))
		return nil, err
	}

	return user, nil
}
