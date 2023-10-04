package repository

import (
	"context"
	"fmt"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	"sync"
)

type formMockRepository struct {
	mockDB *sync.Map
}

// NewFormMockRepository creates a new instance of the FormMockRepository struct.
//
// It returns a FormRepository interface.
func NewFormMockRepository() repository.FormRepository {
	return &formMockRepository{
		mockDB: &sync.Map{},
	}
}

// FindAll retrieves all forms from the form mock repository.
//
// It takes a context.Context as its parameter and returns a slice of forms and an error.
func (r *formMockRepository) FindAll(_ context.Context) ([]*repository.Form, error) {
	forms := []*repository.Form{}
	r.mockDB.Range(func(key, value interface{}) bool {
		forms = append(forms, value.(*repository.Form))
		return true
	})

	return forms, nil
}

// FindByTitle retrieves a form from the formMockRepository based on its title.
//
// _ - The context.Context object for managing the request lifecycle.
// title - The title of the form to search for.
// Returns a pointer to the form object if found, otherwise returns an error.
func (r *formMockRepository) FindByTitle(_ context.Context, title string) (*repository.Form, error) {
	if form, ok := r.mockDB.Load(title); ok {
		return form.(*repository.Form), nil
	}

	return nil, nil
}

func (r *formMockRepository) Delete(_ context.Context, title string) error {
	r.mockDB.Delete(title)
	return nil
}

// Insert inserts a form into the formMockRepository.
//
// It takes a context.Context and a form as parameters.
// It returns an error.
func (r *formMockRepository) Insert(_ context.Context, form *repository.Form) error {
	if form == nil {
		return fmt.Errorf("form_repository insert form is nil")
	}

	r.mockDB.Store(form.Title, form)
	return nil
}

// Update updates a form in the form repository.
//
// It takes a context and a form as parameters.
// It returns an error.
func (r *formMockRepository) Update(_ context.Context, form *repository.Form) error {
	if form == nil {
		return fmt.Errorf("form_repository update form is nil")
	}

	r.mockDB.Store(form.Title, form)
	return nil
}

// ToModel converts a repository.Form object to a model.Form object.
//
// It takes a pointer to a repository.Form object as a parameter and returns a pointer to a model.Form object.
func (r *formMockRepository) ToModel(form *repository.Form) *model.Form {
	return &model.Form{
		Title: form.Title,
	}
}

// FromModel converts a form model to a form repository object.
//
// It takes a pointer to a model.Form object as a parameter and returns a pointer to a repository.Form object.
func (r *formMockRepository) FromModel(form *model.Form) *repository.Form {
	return &repository.Form{
		Title: form.Title,
	}
}
