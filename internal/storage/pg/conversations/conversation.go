package conversations

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mzhn/chats/internal/domain"
	"mzhn/chats/internal/storage/models"
	"mzhn/chats/internal/storage/pg"
	"mzhn/chats/pkg/sl"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
)

type message struct {
	Id        int
	IsUser    bool
	Body      string
	CreatedAt time.Time
	Slide     *int
	FileId    *string
	Filename  *string
}

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
		Select("m.id", "m.is_user", "m.body", "m.created_at").
		From(pg.MessagesTable + " m").
		Where(sq.Eq{"conversation_id": id}).
		OrderBy("m.created_at asc").
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

	mm := make([]domain.Message, 0)
	for rows.Next() {
		m := &message{}
		if err := rows.Scan(
			&m.Id,
			&m.IsUser,
			&m.Body,
			&m.CreatedAt,
		); err != nil {
			log.Error("failed to scan row", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		log.Debug("scanned message", slog.Any("message", m))

		message := domain.Message{
			ConversationId: id,
			Id:             m.Id,
			Body:           m.Body,
			CreatedAt:      m.CreatedAt,
			IsUser:         m.IsUser,
		}

		if !m.IsUser {
			sources, err := r.getSources(ctx, m.Id)
			if err != nil {
				log.Error("failed to get sources", sl.Err(err))
				return nil, fmt.Errorf("%s: %w", fn, err)
			}

			message.Sources = make([]domain.AnswerMeta, len(sources))
			for i, s := range sources {
				message.Sources[i] = domain.AnswerMeta{
					FileId:   s.FileId,
					FileName: s.Filename,
					Slidenum: s.Slide,
				}
			}
		}

		mm = append(mm, message)
	}

	res := &domain.ConversationContent{
		Conversation: domain.Conversation{
			Id: id,
		},
		Messages: mm,
	}

	return res, nil
}

func (r *Repository) getSources(ctx context.Context, messageId int) ([]models.AnswerMeta, error) {

	fn := "setSources"
	log := r.logger.With(sl.Method(fn))

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer conn.Release()

	qb := sq.
		Select("m.slide_num", "m.file_id", "m.file_name").
		From(pg.AnswerMetaTable + " m").
		Where(sq.Eq{"m.message_id": messageId}).
		OrderBy("m.id asc").
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

	out := make([]models.AnswerMeta, 0)
	for rows.Next() {
		m := models.AnswerMeta{}
		if err := rows.Scan(&m.Slide, &m.FileId, &m.Filename); err != nil {
			log.Error("failed to scan row", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		out = append(out, m)
	}

	return out, nil
}
