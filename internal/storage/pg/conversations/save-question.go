package conversations

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mzhn/chats/internal/storage/models"
	"mzhn/chats/internal/storage/pg"
	"mzhn/chats/pkg/sl"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
)

func (r *Repository) SaveQuestion(ctx context.Context, in *models.Question) error {

	fn := "SaveQuestion"
	log := r.logger.With(sl.Method(fn))

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer conn.Release()

	qb := sq.
		Insert(pg.MessagesTable).
		Columns("conversation_id", "is_user", "body", "created_at").
		Values(in.ConversationId, true, in.Body, in.CreatedAt).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := qb.ToSql()
	if err != nil {
		log.Error("failed to build query", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	log.Debug("executing", slog.String("sql", sql), slog.Any("args", args))
	if _, err := conn.Exec(ctx, sql, args...); err != nil {
		var pgErr pgx.PgError
		if errors.As(err, &pgErr) {
			log.Error("pg error on insert new message", sl.PgError(pgErr))
		} else {
			log.Error("unexpected error on sending message", sl.Err(err))
		}

		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
