package repository

import (
	"context"
	"go-form-hub/internal/model"
)

type FormRepository interface {
	FindAll(ctx context.Context) ([]*Form, map[int64]*User, error)
	FindByID(ctx context.Context, id int64) (*Form, *User, error)
	Insert(ctx context.Context, form *Form) (*int64, error)
	Update(ctx context.Context, id int64, form *Form) (*Form, error)
	Delete(ctx context.Context, id int64) error
	ToModel(form *Form, author *User) *model.Form
	FromModel(form *model.Form) *Form
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
