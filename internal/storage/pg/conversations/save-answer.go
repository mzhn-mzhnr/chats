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

func (r *Repository) SaveAnswer(ctx context.Context, in *models.Answer) error {

	fn := "SaveAnswer"
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
		Values(in.ConversationId, false, in.Body, in.CreatedAt).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	sql, args, err := qb.ToSql()
	if err != nil {
		log.Error("failed to build query", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	var id int
	log.Debug("executing", slog.String("sql", sql), slog.Any("args", args))
	if err := conn.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		var pgErr pgx.PgError
		if errors.As(err, &pgErr) {
			log.Error("pg error on insert new message", sl.PgError(pgErr))
		} else {
			log.Error("unexpected error on sending message", sl.Err(err))
		}

		return fmt.Errorf("%s: %w", fn, err)
	}

	if err := r.saveAnswerMeta(ctx, &models.AnswerMetaSave{
		MessageId:  id,
		AnswerMeta: in.AnswerMeta,
	}); err != nil {
		log.Error("failed to save answer meta", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (r *Repository) saveAnswerMeta(ctx context.Context, in *models.AnswerMetaSave) error {

	fn := "saveAnswerMeta"
	log := r.logger.With(sl.Method(fn))

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer conn.Release()

	qb := sq.
		Insert(pg.AnswerMetaTable).
		Columns("message_id", "slide_num", "file_id", "file_name").
		Values(in.MessageId, in.AnswerMeta.Slide, in.FileId, in.AnswerMeta.Filename).
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
