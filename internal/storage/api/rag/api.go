package rag

import (
	"log/slog"
	"mzhn/chats/internal/config"
	"mzhn/chats/pkg/sl"
	"net/http"
	"time"
)

type Api struct {
	client *http.Client
	host   string
	logger *slog.Logger
}

func New(cfg *config.Config) *Api {
	client := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   time.Minute,
	}

	return &Api{
		client: client,
		host:   cfg.RagApi.Host,
		logger: slog.With(sl.Module("rag-api")),
	}
}
