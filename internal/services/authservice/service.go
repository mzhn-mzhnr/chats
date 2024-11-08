package authservice

import (
	"context"
	"log/slog"
	"mzhn/chats/internal/storage/models"
	"mzhn/chats/pkg/sl"
)

type AuthProvider interface {
	Authenticate(ctx context.Context, in *models.AuthenticateRequest) (*models.User, error)
}

type Service struct {
	logger       *slog.Logger
	authProvider AuthProvider
}

func New(
	authProvider AuthProvider,
) *Service {
	return &Service{
		logger:       slog.With(sl.Module("authservice")),
		authProvider: authProvider,
	}
}
