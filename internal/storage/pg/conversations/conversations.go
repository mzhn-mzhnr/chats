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

func (r *Repository) Conversations(ctx context.Context, f *domain.ConversationsFilter) ([]*domain.Conversation, error) {

	fn := "Conversations"
	log := r.logger.With(sl.Method(fn))

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer conn.Release()

	qb := sq.
		Select("id", "name", "created_at").
		From(pg.ConversationsTable).
		Where(sq.Eq{"owner_id": f.UserId}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Dollar)

	sql, args, err := qb.ToSql()
	if err != nil {
		log.Error("failed to build query", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	log.Debug("executing", slog.String("sql", sql), slog.Any("args", args))

	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		var pgErr pgx.PgError
		if errors.As(err, &pgErr) {
			log.Error("pg error on insert new message", sl.PgError(pgErr))
		} else {
			log.Error("unexpected error on sending message", sl.Err(err))
		}

		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	cc := make([]*domain.Conversation, 0)
	for rows.Next() {
		c := new(domain.Conversation)
		if err := rows.Scan(&c.Id, &c.Name, c.CreatedAt); err != nil {
			log.Error("failed to scan row", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		cc = append(cc, c)
	}

	return cc, nil
}
