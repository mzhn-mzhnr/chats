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

func (r *Repository) Conversation(ctx context.Context, id string) (*domain.ConversationContent, error) {

	fn := "Conversations"
	log := r.logger.With(sl.Method(fn))

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer conn.Release()

	qb := sq.
		Select("id", "is_user", "body", "created_at").
		From(pg.MessagesTable).
		Where(sq.Eq{"conversation_id": id}).
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

	mm := make([]*domain.Message, 0)
	for rows.Next() {
		m := new(domain.Message)
		if err := rows.Scan(&m.Id, &m.IsUser, &m.Body, &m.CreatedAt); err != nil {
			log.Error("failed to scan row", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		mm = append(mm, m)
	}

	res := &domain.ConversationContent{
		Conversation: domain.Conversation{
			Id: id,
		},
		Messages: mm,
	}

	return res, nil
}
