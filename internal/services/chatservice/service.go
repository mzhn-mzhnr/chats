package chatservice

import (
	"context"
	"log/slog"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/storage/models"
	"mzhn/chats/pkg/sl"
)

type ConversationsProvider interface {
	Conversation(ctx context.Context, id string) (*domain.ConversationContent, error)
	Conversations(ctx context.Context, f *domain.ConversationsFilter) ([]*domain.Conversation, error)
}

type ConversationCreator interface {
	CreateConversation(ctx context.Context, in *domain.NewConversation) (string, error)
}

type MessageSaver interface {
	SaveMessage(ctx context.Context, in *models.NewMessage) error
}

type Service struct {
	logger       *slog.Logger
	messageSaver MessageSaver
	convProvider ConversationsProvider
	convCreator  ConversationCreator
}

func New(
	msgsaver MessageSaver,
	convprovider ConversationsProvider,
	convcreator ConversationCreator,
) *Service {
	return &Service{
		logger:       slog.With(sl.Module("chatservice")),
		messageSaver: msgsaver,
		convProvider: convprovider,
		convCreator:  convcreator,
	}
}
