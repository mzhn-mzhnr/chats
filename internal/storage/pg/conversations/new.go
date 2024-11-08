package conversations

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/storage/pg"
	"mzhn/chats/pkg/sl"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
)

func (r *Repository) NewConversation(ctx context.Context, in *domain.NewConversation) (id string, err error) {
	fn := "NewConversation"
	log := r.logger.With(sl.Method(fn))

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return id, fmt.Errorf("%s: %w", fn, err)
	}
	defer conn.Release()

	qb := sq.
		Insert(pg.ConversationsTable).
		Columns("owner_id").
		Values(in.UserId).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	sql, args, err := qb.ToSql()
	if err != nil {
		log.Error("failed to build query", sl.Err(err))
		return id, fmt.Errorf("%s: %w", fn, err)
	}

	log.Debug("executing", slog.String("sql", sql), slog.Any("args", args))
	if err := conn.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		var pgErr pgx.PgError
		if errors.As(err, &pgErr) {
			log.Error("pg error on insert new message", sl.PgError(pgErr))
		} else {
			log.Error("unexpected error on sending message", sl.Err(err))
		}

		return id, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}
