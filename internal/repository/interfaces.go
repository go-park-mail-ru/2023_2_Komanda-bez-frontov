package repository

import (
	"context"
	"go-form-hub/internal/model"
)

type FormRepository interface {
	FindAll(ctx context.Context) ([]*Form, error)
	FindByTitle(ctx context.Context, title string) (*Form, error)
	Insert(ctx context.Context, form *Form) error
	Update(ctx context.Context, title string, form *Form) error
	Delete(ctx context.Context, title string) error
	ToModel(form *Form) *model.Form
	FromModel(form *model.Form) *Form
}

type UserRepository interface {
	FindAll(ctx context.Context) ([]*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Insert(ctx context.Context, user *User) error
	Update(ctx context.Context, id string, user *User) error
	Delete(ctx context.Context, id string) error
}

type SessionRepository interface {
	FindAll(ctx context.Context) ([]*Session, error)
	FindByID(ctx context.Context, sessionID string) (*Session, error)
	FindByUsername(ctx context.Context, username string) (*Session, error)
	FindByUserID(ctx context.Context, userID string) (*Session, error)
	Insert(ctx context.Context, session *Session) error
	Delete(ctx context.Context, sessionID string) error
}
