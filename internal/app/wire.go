//go:build wireinject

package app

import (
	"context"
	"fmt"
	"log/slog"
	"mzhn/chats/internal/services/authservice"
	"mzhn/chats/internal/services/chatservice"
	"mzhn/chats/internal/storage/api/auth"
	"mzhn/chats/internal/storage/api/rag"
	"mzhn/chats/internal/storage/pg/conversations"
	"mzhn/chats/internal/transport/http"
	"mzhn/chats/internal/transport/queue"
	"time"

	"mzhn/chats/internal/config"

	"github.com/google/wire"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

func New() (*App, func(), error) {
	panic(wire.Build(
		newApp,
		_servers,

		http.New,
		queue.NewRedisConsumer,

		authservice.New,
		wire.Bind(new(authservice.AuthProvider), new(*auth.Api)),

		chatservice.New,
		wire.Bind(new(chatservice.ConversationCreator), new(*conversations.Repository)),
		wire.Bind(new(chatservice.ConversationsProvider), new(*conversations.Repository)),
		wire.Bind(new(chatservice.MessageSaver), new(*conversations.Repository)),
		wire.Bind(new(chatservice.RagProvider), new(*rag.Api)),

		conversations.New,
		auth.New,
		rag.New,

		_redis,
		_pgxpool,
		config.New,
	))
}

func _pgxpool(cfg *config.Config) (*pgxpool.Pool, func(), error) {
	ctx := context.Background()
	cs := cfg.Pg.ConnectionString()
	db, err := pgxpool.Connect(ctx, cs)
	if err != nil {
		return nil, nil, err
	}

	slog.Debug("connecting to database", slog.String("cs", cs))
	t := time.Now()
	if err := db.Ping(ctx); err != nil {
		slog.Error("failed to connect to database", slog.String("err", err.Error()), slog.String("conn", cs))
		return nil, func() { db.Close() }, err
	}
	slog.Info("connected to database", slog.String("ping", fmt.Sprintf("%2.fs", time.Since(t).Seconds())))

	return db, func() { db.Close() }, nil
}

func _servers(cfg *config.Config, shttp *http.Server, _ *queue.RedisConsumer) []Server {
	servers := make([]Server, 0, 2)

	if cfg.Http.Enabled {
		servers = append(servers, shttp)
	}

	// servers = append(servers, rq)

	return servers
}

func _redis(cfg *config.Config) (*redis.Client, func(), error) {

	conf := cfg.Redis

	ctx := context.Background()

	slog.Debug("connecting to redis", slog.Any("config", conf))
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Password: conf.Pass,
		DB:       conf.DB,
	})

	slog.Debug("ping redis")
	start := time.Now()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, nil, err
	}
	slog.Debug("pinged redis", slog.Duration("took", time.Since(start)))

	return client, func() {
		client.Close()
	}, nil
}
