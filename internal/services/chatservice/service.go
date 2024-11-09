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
	SaveQuestion(ctx context.Context, in *models.Question) error
	SaveAnswer(ctx context.Context, in *models.Answer) error
}

type RagProvider interface {
	Stream(ctx context.Context, input string, eventCh chan<- []byte) (*models.AnswerMeta, error)
}

type Service struct {
	logger       *slog.Logger
	messageSaver MessageSaver
	convProvider ConversationsProvider
	convCreator  ConversationCreator
	rag          RagProvider
}

func New(
	msgsaver MessageSaver,
	convprovider ConversationsProvider,
	convcreator ConversationCreator,
	rag RagProvider,
) *Service {
	return &Service{
		logger:       slog.With(sl.Module("chatservice")),
		messageSaver: msgsaver,
		convProvider: convprovider,
		convCreator:  convcreator,
		rag:          rag,
	}
}
