package repository

import (
	"context"
	"go-form-hub/internal/model"
)

type FormRepository interface {
	FindAll(ctx context.Context) ([]*Form, error)
	FindByTitle(ctx context.Context, title string) (*Form, error)
	Insert(ctx context.Context, form *Form) error
	Update(ctx context.Context, form *Form) error
	Delete(ctx context.Context, title string) error
	ToModel(form *Form) *model.Form
	FromModel(form *model.Form) *Form
}
