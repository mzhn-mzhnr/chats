// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package app

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log/slog"
	"mzhn/chats/internal/config"
	"mzhn/chats/internal/services/chatservice"
	"mzhn/chats/internal/storage/api/auth"
	"mzhn/chats/internal/storage/pg/conversations"
	"mzhn/chats/internal/transport/http"
	"time"
)

import (
	_ "github.com/jackc/pgx/stdlib"
)

// Injectors from wire.go:

func New() (*App, func(), error) {
	configConfig := config.New()
	pool, cleanup, err := _pgxpool(configConfig)
	if err != nil {
		return nil, nil, err
	}
	repository := conversations.New(pool)
	api := auth.New(configConfig)
	service := chatservice.New(repository, repository, repository, api)
	v := _servers(configConfig, service)
	app := newApp(configConfig, v)
	return app, func() {
		cleanup()
	}, nil
}

// wire.go:

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
