package queue

import (
	"context"
	"encoding/json"
	"log/slog"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/services/chatservice"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConsumer struct {
	rdb          *redis.Client
	chatsService *chatservice.Service
}

func NewRedisConsumer(rdb *redis.Client, chatsService *chatservice.Service) *RedisConsumer {
	return &RedisConsumer{
		rdb:          rdb,
		chatsService: chatsService,
	}
}

func (c *RedisConsumer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			result, err := c.rdb.BLPop(ctx, 0*time.Second, "messages_queue").Result()
			if err != nil {
				slog.Error("error when receiving a message from the queue")
				continue
			}

			event := result[1] // 0 ключ, 1 значение

			var handledMessage domain.HandledMessage
			err = json.Unmarshal([]byte(event), &handledMessage)
			if err != nil {
				slog.Error("error when unmarshalling the event", slog.Any("error", err))
				continue
			}

			_, err = c.chatsService.SendMessage(ctx, &domain.NewMessage{
				ConversationId: handledMessage.ConversationId,
				Body:           handledMessage.Question.Message,
				UserId:         &handledMessage.UserId,
			})
			if err != nil {
				slog.Error("error when sending a message", slog.Any("error", err))
			}

			c.chatsService.SendMessage(ctx, &domain.NewMessage{
				ConversationId: handledMessage.ConversationId,
				Body:           handledMessage.Question.Message,
				UserId:         nil,
			})
			if err != nil {
				slog.Error("error when sending a message", slog.Any("error", err))
			}
		}

	}
}
