package form

import (
	"context"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/service_response"
	"net/http"
	"time"

	validator "github.com/go-playground/validator/v10"
)

type Service interface {
	FormSave(ctx context.Context, form *model.Form) (*resp.Response, error)
	FormUpdate(ctx context.Context, id int64, form *model.Form) (*resp.Response, error)
	FormList(ctx context.Context) (*resp.Response, error)
	FormDelete(ctx context.Context, id int64) (*resp.Response, error)
	FormGet(ctx context.Context, id int64) (*resp.Response, error)
}

type formService struct {
	formRepository repository.FormRepository
	validate       *validator.Validate
}

func NewFormService(formRepository repository.FormRepository, validate *validator.Validate) Service {
	return &formService{
		formRepository: formRepository,
		validate:       validate,
	}
}

func (s *formService) FormSave(ctx context.Context, form *model.Form) (*resp.Response, error) {
	currentUser := ctx.Value(model.CurrentUserInContext).(*model.UserGet)
	if err := s.validate.Struct(form); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), err
	}

	form.Author = currentUser
	form.CreatedAt = time.Now().UTC()

	id, err := s.formRepository.Insert(ctx, s.formRepository.FromModel(form))
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	form.ID = id
	return resp.NewResponse(http.StatusOK, form), nil
}

func (s *formService) FormUpdate(ctx context.Context, id int64, form *model.Form) (*resp.Response, error) {
	if err := s.validate.Struct(form); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), err
	}

	currentUser := ctx.Value(model.CurrentUserInContext).(*model.UserGet)

	existing, _, err := s.formRepository.FindByID(ctx, id)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if existing == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	if existing.AuthorID != currentUser.ID {
		return resp.NewResponse(http.StatusForbidden, nil), nil
	}

	form.Author = currentUser
	formUpdate, err := s.formRepository.Update(ctx, id, s.formRepository.FromModel(form))
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	res := s.formRepository.ToModel(formUpdate, &repository.User{
		ID:        currentUser.ID,
		Username:  currentUser.Username,
		FirstName: currentUser.FirstName,
		LastName:  currentUser.LastName,
		Email:     currentUser.Email,
	})
	return resp.NewResponse(http.StatusOK, res), nil
}

func (s *formService) FormList(ctx context.Context) (*resp.Response, error) {
	var response model.FormList
	response.Forms = make([]*model.Form, 0)

	forms, authors, err := s.formRepository.FindAll(ctx)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	for _, form := range forms {
		response.Forms = append(response.Forms, s.formRepository.ToModel(form, authors[form.AuthorID]))
	}

	response.Count = len(forms)
	return resp.NewResponse(http.StatusOK, response), nil
}

func (s *formService) FormDelete(ctx context.Context, id int64) (*resp.Response, error) {
	if err := s.formRepository.Delete(ctx, id); err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	return resp.NewResponse(http.StatusOK, nil), nil
}

func (s *formService) FormGet(ctx context.Context, id int64) (*resp.Response, error) {
	form, author, err := s.formRepository.FindByID(ctx, id)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if form == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	return resp.NewResponse(http.StatusOK, s.formRepository.ToModel(form, author)), nil
}
