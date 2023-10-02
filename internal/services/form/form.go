package form

import (
	"context"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/serviceresponse"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type Service interface {
	FormSave(ctx context.Context, form *model.Form) (*resp.Response, error)
	FormUpdate(ctx context.Context, title string, form *model.Form) (*resp.Response, error)
	FormList(ctx context.Context) (*resp.Response, error)
	FormDelete(ctx context.Context, title string) (*resp.Response, error)
	FormGet(ctx context.Context, title string) (*resp.Response, error)
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
	if err := s.validate.Struct(form); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), err
	}

	existing, err := s.formRepository.FindByTitle(ctx, form.Title)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if existing != nil {
		return resp.NewResponse(http.StatusConflict, nil), nil
	}

	err = s.formRepository.Insert(ctx, s.formRepository.FromModel(form))
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	return resp.NewResponse(http.StatusOK, form), nil
}

func (s *formService) FormUpdate(ctx context.Context, title string, form *model.Form) (*resp.Response, error) {
	if err := s.validate.Struct(form); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), err
	}

	existing, err := s.formRepository.FindByTitle(ctx, title)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if existing == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	err = s.formRepository.Update(ctx, s.formRepository.FromModel(form))
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	return resp.NewResponse(http.StatusOK, form), nil
}

func (s *formService) FormList(ctx context.Context) (*resp.Response, error) {
	var response model.FormList
	response.Forms = make([]*model.Form, 0)

	forms, err := s.formRepository.FindAll(ctx)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	for _, form := range forms {
		response.Forms = append(response.Forms, s.formRepository.ToModel(form))
	}

	response.Count = len(forms)
	return resp.NewResponse(http.StatusOK, response), nil
}

func (s *formService) FormDelete(ctx context.Context, title string) (*resp.Response, error) {
	if err := s.formRepository.Delete(ctx, title); err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}
	// TODO: check if current user is author

	return resp.NewResponse(http.StatusOK, nil), nil
}

func (s *formService) FormGet(ctx context.Context, title string) (*resp.Response, error) {
	form, err := s.formRepository.FindByTitle(ctx, title)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if form == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	return resp.NewResponse(http.StatusOK, s.formRepository.ToModel(form)), nil
}
