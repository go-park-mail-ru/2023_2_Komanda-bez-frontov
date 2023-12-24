package form

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/service_response"

	validator "github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

type Service interface {
	FormSave(ctx context.Context, form *model.Form) (*resp.Response, error)
	FormUpdate(ctx context.Context, id int64, form *model.FormUpdate) (*resp.Response, error)
	FormList(ctx context.Context) (*resp.Response, error)
	FormListByUser(ctx context.Context, username string, isArchived bool) (*resp.Response, error)
	FormDelete(ctx context.Context, id int64) (*resp.Response, error)
	FormArchive(ctx context.Context, id int64, archive bool) (*resp.Response, error)
	FormGet(ctx context.Context, id int64) (*resp.Response, error)
	FormSearch(ctx context.Context, title string, userID uint, isArchived bool) (*resp.Response, error)
	FormResults(ctx context.Context, id int64) (*resp.Response, error)
	FormResultsCsv(ctx context.Context, formID int64) ([]byte, error)
	FormResultsExel(ctx context.Context, formID int64) ([]byte, error)
	FormPassageList(ctx context.Context) (*resp.Response, error)
	FormPassageGet(ctx context.Context, id int64) (*resp.Response, error)
}

type formService struct {
	formRepository     repository.FormRepository
	questionRepository repository.QuestionRepository
	answerRepository   repository.AnswerRepository
	sanitizer          *bluemonday.Policy
	validate           *validator.Validate
}

func NewFormService(formRepository repository.FormRepository, questionRepository repository.QuestionRepository, answerRepository repository.AnswerRepository, validate *validator.Validate) Service {
	sanitizer := bluemonday.UGCPolicy()
	return &formService{
		formRepository:     formRepository,
		validate:           validate,
		questionRepository: questionRepository,
		sanitizer:          sanitizer,
		answerRepository:   answerRepository,
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
	formResults.Sanitize(s.sanitizer)

	return resp.NewResponse(http.StatusOK, formResults), nil
}

func (s *formService) FormResultsCsv(ctx context.Context, formID int64) ([]byte, error) {
	FormResultsExelCsv, err := s.formRepository.FormResultsCsv(ctx, formID)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_results_csv failed to run FormResults: %e", err)
	}

	if FormResultsExelCsv == nil {
		return nil, fmt.Errorf("form_repository form_results_csv returned nil result")
	}

	return FormResultsExelCsv, nil
}

func (s *formService) FormResultsExel(ctx context.Context, formID int64) ([]byte, error) {
	FormResultsExelCsv, err := s.formRepository.FormResultsExel(ctx, formID)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_results_exel failed to run FormResults: %e", err)
	}

	if FormResultsExelCsv == nil {
		return nil, fmt.Errorf("form_repository form_results_exel returned nil result")
	}

	return FormResultsExelCsv, nil
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

	result.Sanitize(s.sanitizer)

	return resp.NewResponse(http.StatusOK, result), nil
}

func (s *formService) FormUpdate(ctx context.Context, id int64, form *model.FormUpdate) (*resp.Response, error) {
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

	if len(form.RemovedAnswers) != 0 {
		err = s.answerRepository.DeleteAllByID(ctx, form.RemovedAnswers)
		if err != nil {
			return resp.NewResponse(http.StatusInternalServerError, nil), err
		}
	}

	if len(form.RemovedQuestions) != 0 {
		err = s.questionRepository.DeleteAllByID(ctx, form.RemovedQuestions)
		if err != nil {
			return resp.NewResponse(http.StatusInternalServerError, nil), err
		}
	}

	for _, question := range form.Questions {
		if *question.ID == 0 {
			err := s.questionRepository.Insert(ctx, question, id)
			if err != nil {
				return resp.NewResponse(http.StatusInternalServerError, nil), err
			}
		} else {
			if response, err := s.QuestionUpdate(ctx, question); err != nil {
				return response, err
			}
		}
	}

	formUpdate.Sanitize(s.sanitizer)

	return resp.NewResponse(http.StatusOK, formUpdate), nil
}

func (s *formService) QuestionUpdate(ctx context.Context, question *model.Question) (*resp.Response, error) {
	err := s.questionRepository.Update(ctx, *question.ID, question)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}
	if question.Type == model.InputAnswerType {
		err := s.answerRepository.DeleteByQuestionID(ctx, *question.ID)
		if err != nil {
			return resp.NewResponse(http.StatusInternalServerError, nil), err
		}
	}
	for _, answer := range question.Answers {
		if *answer.ID == 0 {
			err := s.answerRepository.Insert(ctx, *question.ID, answer)
			if err != nil {
				return resp.NewResponse(http.StatusInternalServerError, nil), err
			}
		} else {
			err := s.answerRepository.Update(ctx, *answer.ID, answer)
			if err != nil {
				return resp.NewResponse(http.StatusInternalServerError, nil), err
			}
		}
	}
	return resp.NewResponse(http.StatusOK, nil), nil
}

func (s *formService) FormList(ctx context.Context) (*resp.Response, error) {
	forms, err := s.formRepository.FindAll(ctx)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	formList := &model.FormList{
		Forms: forms,
	}
	formList.Count = len(forms)

	formList.Sanitize(s.sanitizer)
	return resp.NewResponse(http.StatusOK, formList), nil
}

func (s *formService) FormListByUser(ctx context.Context, username string, isArchived bool) (*resp.Response, error) {
	var forms []*model.FormTitle
	var err error

	if isArchived {
		forms, err = s.formRepository.FindAllByUserArchived(ctx, username)
		if err != nil {
			return resp.NewResponse(http.StatusInternalServerError, nil), err
		}
	} else {
		forms, err = s.formRepository.FindAllByUserActive(ctx, username)
		if err != nil {
			return resp.NewResponse(http.StatusInternalServerError, nil), err
		}
	}

	formList := &model.FormList{
		Forms: forms,
	}
	formList.Count = len(forms)

	formList.Sanitize(s.sanitizer)
	return resp.NewResponse(http.StatusOK, formList), nil
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

func (s *formService) FormArchive(ctx context.Context, id int64, archive bool) (*resp.Response, error) {
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

	if err := s.formRepository.Archive(ctx, id, archive); err != nil {
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

	if form.IsArchived && ctx.Value(model.ContextCurrentUser) == nil {
		return resp.NewResponse(http.StatusForbidden, nil), nil
	}
	form.Sanitize(s.sanitizer)

	if ctx.Value(model.ContextCurrentUser) != nil {
		currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

		if form.IsArchived && form.Author.ID != currentUser.ID {
			return resp.NewResponse(http.StatusForbidden, nil), nil
		}

		total, err := s.formRepository.UserFormPassageCount(ctx, *form.ID, currentUser.ID)
		if err != nil {
			return resp.NewResponse(http.StatusInternalServerError, nil), nil
		}

		form.CurrentPassageTotal = int(total)
	}
	form.Sanitize(s.sanitizer)

	return resp.NewResponse(http.StatusOK, form), nil
}

func (s *formService) FormSearch(ctx context.Context, title string, userID uint, isArchived bool) (*resp.Response, error) {
	var forms []*model.FormTitle
	var err error

	if isArchived {
		forms, err = s.formRepository.FormsSearchArchived(ctx, title, userID)
		if err != nil {
			return resp.NewResponse(http.StatusInternalServerError, nil), err
		}
	} else {
		forms, err = s.formRepository.FormsSearch(ctx, title, userID)
		if err != nil {
			return resp.NewResponse(http.StatusInternalServerError, nil), err
		}
	}

	formList := &model.FormList{
		Forms: forms,
	}
	formList.Count = len(forms)

	formList.Sanitize(s.sanitizer)

	return resp.NewResponse(http.StatusOK, formList), nil
}

func (s *formService) FormPassageList(ctx context.Context) (*resp.Response, error) {
	currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)
	
	forms, err := s.formRepository.FindPassagesAll(ctx, currentUser.ID)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	formList := &model.FormPassageList{
		Forms: forms,
	}
	formList.Count = len(forms)

	formList.Sanitize(s.sanitizer)
	return resp.NewResponse(http.StatusOK, formList), nil
}

func (s *formService) FormPassageGet(ctx context.Context, id int64) (*resp.Response, error) {
	form, err := s.formRepository.FindPassageByID(ctx, id)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if form == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	if ctx.Value(model.ContextCurrentUser) == nil {
		return resp.NewResponse(http.StatusForbidden, nil), nil
	}
	form.Sanitize(s.sanitizer)

	if ctx.Value(model.ContextCurrentUser) != nil {
		currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

		if form.Author.ID != currentUser.ID && form.UserID != currentUser.ID {
			return resp.NewResponse(http.StatusForbidden, nil), nil
		}
	}
	form.Sanitize(s.sanitizer)

	return resp.NewResponse(http.StatusOK, form), nil
}