//go:build wireinject

package app

import (
	"context"
	"fmt"
	"log/slog"
	"mzhn/chats/internal/services/chatservice"
	"mzhn/chats/internal/storage/api/auth"
	"mzhn/chats/internal/storage/pg/conversations"
	"mzhn/chats/internal/transport/http"
	"time"

	"mzhn/chats/internal/config"

	"github.com/google/wire"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jackc/pgx/v4/pgxpool"
)

func New() (*App, func(), error) {
	panic(wire.Build(
		newApp,
		_servers,

		chatservice.New,
		wire.Bind(new(chatservice.AuthProvider), new(*auth.Api)),
		wire.Bind(new(chatservice.ConversationCreator), new(*conversations.Repository)),
		wire.Bind(new(chatservice.ConversationsProvider), new(*conversations.Repository)),
		wire.Bind(new(chatservice.MessageSaver), new(*conversations.Repository)),

		conversations.New,
		auth.New,

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

func _servers(cfg *config.Config, svc *chatservice.Service) []Server {
	servers := make([]Server, 0, 2)

	if cfg.Http.Enabled {
		servers = append(servers, http.New(cfg, svc))
	}

	return servers
}
