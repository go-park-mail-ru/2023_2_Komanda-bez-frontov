package form

import (
	"context"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	"net/http"
	"time"

	resp "go-form-hub/internal/services/service_response"

	validator "github.com/go-playground/validator/v10"
)

type Service interface {
	FormSave(ctx context.Context, form *model.Form) (*resp.Response, error)
	FormUpdate(ctx context.Context, id int64, form *model.Form) (*resp.Response, error)
	FormList(ctx context.Context) (*resp.Response, error)
	FormListByUser(ctx context.Context, username string) (*resp.Response, error)
	FormDelete(ctx context.Context, id int64) (*resp.Response, error)
	FormGet(ctx context.Context, id int64) (*resp.Response, error)
	FormSearch(ctx context.Context, title string, userID uint) (*resp.Response, error)
	FormResults(ctx context.Context, id int64) (*resp.Response, error)
}

type formService struct {
	formRepository     repository.FormRepository
	questionRepository repository.QuestionRepository
	validate           *validator.Validate
}

func NewFormService(formRepository repository.FormRepository, questionRepository repository.QuestionRepository, validate *validator.Validate) Service {
	return &formService{
		formRepository:     formRepository,
		validate:           validate,
		questionRepository: questionRepository,
	}
}

func (s *formService) FormResults(ctx context.Context, formID int64) (*resp.Response, error) {
	formResults, err := s.formRepository.FormResults(ctx, formID)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if formResults == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	return resp.NewResponse(http.StatusOK, formResults), nil
}

func (s *formService) FormSave(ctx context.Context, form *model.Form) (*resp.Response, error) {
	currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)
	if err := s.validate.Struct(form); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), err
	}

	form.Author = currentUser
	form.CreatedAt = time.Now().UTC()

	result, err := s.formRepository.Insert(ctx, form, nil)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	return resp.NewResponse(http.StatusOK, result), nil
}

func (s *formService) FormUpdate(ctx context.Context, id int64, form *model.Form) (*resp.Response, error) {
	if err := s.validate.Struct(form); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), err
	}

	currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

	existing, err := s.formRepository.FindByID(ctx, id)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if existing == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	if existing.Author.ID != currentUser.ID {
		return resp.NewResponse(http.StatusForbidden, nil), nil
	}

	form.Author = currentUser
	formUpdate, err := s.formRepository.Update(ctx, id, form)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	formUpdate.ID = &id
	err = s.questionRepository.DeleteByFormID(ctx, id)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	newQuestions, err := s.questionRepository.BatchInsert(ctx, form.Questions, id)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	formUpdate.Questions = newQuestions

	return resp.NewResponse(http.StatusOK, formUpdate), nil
}

func (s *formService) FormList(ctx context.Context) (*resp.Response, error) {
	var response model.FormList
	response.Forms = make([]*model.Form, 0)

	forms, err := s.formRepository.FindAll(ctx)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	response.Count = len(forms)
	response.Forms = forms
	return resp.NewResponse(http.StatusOK, response), nil
}

func (s *formService) FormListByUser(ctx context.Context, username string) (*resp.Response, error) {
	var response model.FormList
	response.Forms = make([]*model.Form, 0)

	forms, err := s.formRepository.FindAllByUser(ctx, username)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	response.Count = len(forms)
	response.Forms = forms
	return resp.NewResponse(http.StatusOK, response), nil
}

func (s *formService) FormDelete(ctx context.Context, id int64) (*resp.Response, error) {
	currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

	existing, err := s.formRepository.FindByID(ctx, id)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if existing == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	if existing.Author.ID != currentUser.ID {
		return resp.NewResponse(http.StatusForbidden, nil), nil
	}

	if err := s.formRepository.Delete(ctx, id); err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	return resp.NewResponse(http.StatusOK, nil), nil
}

func (s *formService) FormGet(ctx context.Context, id int64) (*resp.Response, error) {
	form, err := s.formRepository.FindByID(ctx, id)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if form == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	return resp.NewResponse(http.StatusOK, form), nil
}

func (s *formService) FormSearch(ctx context.Context, title string, userID uint) (*resp.Response, error) {
	forms, err := s.formRepository.FormsSearch(ctx, title, userID)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	formTitleList := &model.FormTitleList{
		FormTitles: forms,
	}
	formTitleList.Count = len(forms)

	return resp.NewResponse(http.StatusOK, formTitleList), nil
}
