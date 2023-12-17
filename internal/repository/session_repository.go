package repository

import (
	"context"
	"fmt"
	"time"

	"go-form-hub/internal/database"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type Session struct {
	SessionID string    `db:"id"`
	UserID    int64     `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type sessionRepository struct {
	db      database.ConnPool
	builder squirrel.StatementBuilderType
}

func NewSessionDatabaseRepository(db database.ConnPool, builder squirrel.StatementBuilderType) SessionRepository {
	return &sessionRepository{
		db:      db,
		builder: builder,
	}
}

func (r *sessionRepository) getTableName() string {
	return fmt.Sprintf("%s.session", r.db.GetSchema())
}

func (r *sessionRepository) FindByID(ctx context.Context, sessionID string) (session *Session, err error) {
	query, args, err := r.builder.
		Select("id", "user_id", "created_at").
		From(r.getTableName()).
		Where(squirrel.Eq{"id": sessionID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("session_repository find_by_id failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("session_repository find_by_id failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	row := tx.QueryRow(ctx, query, args...)

	session, err = r.fromRow(row)
	return session, err
}

func (r *sessionRepository) FindByUserID(ctx context.Context, userID int64) (session *Session, err error) {
	query, args, err := r.builder.
		Select("id", "user_id", "created_at").
		From(r.getTableName()).
		Where(squirrel.Eq{"user_id": userID}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("session_repository find_by_user_id failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("session_repository find_by_user_id failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	row := tx.QueryRow(ctx, query, args...)

	session, err = r.fromRow(row)
	return session, err
}

func (r *sessionRepository) Insert(ctx context.Context, session *Session) error {
	query, args, err := r.builder.
		Insert(r.getTableName()).
		Columns("id", "user_id", "created_at").
		Values(session.SessionID, session.UserID, session.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("session_repository insert failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("session_repository insert failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, query, args...)
	return err
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	query, args, err := r.builder.
		Delete(r.getTableName()).
		Where(squirrel.Eq{"id": sessionID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("session_repository delete failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("session_repository delete failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, query, args...)
	return err
}

func (r *sessionRepository) fromRow(row pgx.Row) (*Session, error) {
	session := &Session{}
	err := row.Scan(
		&session.SessionID,
		&session.UserID,
		&session.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("session_repository failed to scan row: %e", err)
	}

	return session, nil
}
