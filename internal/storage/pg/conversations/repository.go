package conversations

import (
	"log/slog"
	"mzhn/chats/pkg/sl"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:   pool,
		logger: slog.With(sl.Module("pg-conversations")),
	}
}
