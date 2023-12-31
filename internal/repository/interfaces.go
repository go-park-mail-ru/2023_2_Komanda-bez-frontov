package repository

import (
	"context"

	"go-form-hub/internal/model"

	"github.com/jackc/pgx/v5"
)

type FormRepository interface {
	FindAll(ctx context.Context) ([]*model.FormTitle, error)
	FindAllByUser(ctx context.Context, username string) ([]*model.FormTitle, error)
	FindByID(ctx context.Context, id int64) (*model.Form, error)
	Insert(ctx context.Context, form *model.Form, tx pgx.Tx) (*model.Form, error)
	Update(ctx context.Context, id int64, form *model.FormUpdate) (*model.FormUpdate, error)
	Delete(ctx context.Context, id int64) error
	FormsSearch(ctx context.Context, title string, userID uint) (forms []*model.FormTitle, err error)
	FormResults(ctx context.Context, id int64) (*model.FormResult, error)
	FormResultsCsv(ctx context.Context, id int64) ([]byte, error)
	FormResultsExel(ctx context.Context, id int64) ([]byte, error)
	FormPassageSave(ctx context.Context, formPassage *model.FormPassage, userID uint64) error
	FormPassageCount(ctx context.Context, formID int64) (int64, error)
	UserFormPassageCount(ctx context.Context, formID int64, userID int64) (int64, error)
}

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
	DeleteAllByID(ctx context.Context, ids []int64) error
	Update(ctx context.Context, id int64, question *model.Question) error
	Insert(ctx context.Context, questions *model.Question, formID int64) error
}

type AnswerRepository interface {
	DeleteAllByID(ctx context.Context, ids []int64) error
	Update(ctx context.Context, id int64, answer *model.Answer) error
	Insert(ctx context.Context, questionID int64, answer *model.Answer) error
	DeleteByQuestionID(ctx context.Context, questionID int64) error
}
