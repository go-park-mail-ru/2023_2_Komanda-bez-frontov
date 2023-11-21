package repository

import (
	"context"

	"go-form-hub/internal/model"

	"github.com/jackc/pgx/v5"
)

type FormRepository interface {
	FindAll(ctx context.Context) ([]*model.Form, error)
	FindAllByUser(ctx context.Context, username string) ([]*model.Form, error)
	FindByID(ctx context.Context, id int64) (*model.Form, error)
	Insert(ctx context.Context, form *model.Form, tx pgx.Tx) (*model.Form, error)
	Update(ctx context.Context, id int64, form *model.Form) (*model.Form, error)
	Delete(ctx context.Context, id int64) error
	FormsSearch(ctx context.Context, title string) (forms []*model.FormTitle, err error)
	FormResults(ctx context.Context, id int64) (*model.FormResult, error)}

type UserRepository interface {
	FindAll(ctx context.Context) ([]*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	Insert(ctx context.Context, user *User) (int64, error)
	Update(ctx context.Context, id int64, user *User) error
	Delete(ctx context.Context, id int64) error
}

type SessionRepository interface {
	FindByID(ctx context.Context, sessionID string) (*Session, error)
	FindByUserID(ctx context.Context, userID int64) (*Session, error)
	Insert(ctx context.Context, session *Session) error
	Delete(ctx context.Context, sessionID string) error
}

type QuestionRepository interface {
	DeleteByFormID(ctx context.Context, formID int64) error
	BatchInsert(ctx context.Context, questions []*model.Question, formID int64) ([]*model.Question, error)
}
