package auth

import (
	"log/slog"
	"mzhn/chats/internal/config"
	"mzhn/chats/pkg/sl"
	"net/http"
	"time"
)

type Api struct {
	host   string
	client *http.Client
	logger *slog.Logger
}

func NewApi(cfg *config.Config) *Api {

	client := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   time.Minute,
	}

	return &Api{
		host:   cfg.AuthApi.Host,
		client: client,
		logger: slog.With(sl.Module("api-auth")),
	}
}
