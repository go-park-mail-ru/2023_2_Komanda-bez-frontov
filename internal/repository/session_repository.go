package repository

import (
	"context"
	"go-form-hub/internal/database"

	"github.com/Masterminds/squirrel"
)

type Session struct {
	SessionID string
	UserID    int64
	Username  string
	CreatedAt int64
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

func (r *sessionRepository) FindAll(ctx context.Context) ([]*Session, error) {
	return nil, nil
}

func (r *sessionRepository) FindByID(ctx context.Context, sessionID string) (*Session, error) {
	return nil, nil
}

func (r *sessionRepository) FindByUsername(ctx context.Context, username string) (*Session, error) {
	return nil, nil
}

func (r *sessionRepository) FindByUserID(ctx context.Context, userID int64) (*Session, error) {
	return nil, nil
}

func (r *sessionRepository) Insert(ctx context.Context, session *Session) error {
	return nil
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	return nil
}
