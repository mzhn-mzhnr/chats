package queue

import (
	"context"
	"encoding/json"
	"log/slog"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/services/chatservice"
	"mzhn/chats/pkg/sl"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConsumer struct {
	rdb          *redis.Client
	chatsService *chatservice.Service
	logger       *slog.Logger
}

func NewRedisConsumer(rdb *redis.Client, chatsService *chatservice.Service) *RedisConsumer {
	return &RedisConsumer{
		rdb:          rdb,
		chatsService: chatsService,
		logger:       slog.With(sl.Module("RedisConsumer")),
	}
}

func (c *RedisConsumer) Run(ctx context.Context) error {
	c.logger.Debug("123")
	for {
		c.logger.Debug("GET GET")
		result, err := c.rdb.BLPop(ctx, 0*time.Second, "messages_queue").Result()
		if err != nil {
			c.logger.Error("error when receiving a message from the queue")
			continue
		}

		event := result[1] // 0 ключ, 1 значение

		var handledMessage domain.HandledMessage
		err = json.Unmarshal([]byte(event), &handledMessage)
		if err != nil || !handledMessage.Valid() {
			c.logger.Error("error when unmarshalling the event", slog.Any("error", err))
			continue
		}

		_, err = c.chatsService.SendMessage(ctx, &domain.NewMessage{
			ConversationId: handledMessage.ConversationId,
			Body:           handledMessage.Question.Message,
			CreatedAt:      handledMessage.Question.CreatedAt,
			IsUser:         true,
		})
		if err != nil {
			c.logger.Error("error when sending a message", slog.Any("error", err))
		}

		_, err = c.chatsService.SendMessage(ctx, &domain.NewMessage{
			ConversationId: handledMessage.ConversationId,
			Body:           handledMessage.Answer.Message,
			CreatedAt:      handledMessage.Answer.CreatedAt,
			IsUser:         false,
		})
		if err != nil {
			c.logger.Error("error when sending a message", slog.Any("error", err))
		}
	}
}
