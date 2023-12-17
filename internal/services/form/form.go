package form

import (
	"context"
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
	FormListByUser(ctx context.Context, username string) (*resp.Response, error)
	FormDelete(ctx context.Context, id int64) (*resp.Response, error)
	FormGet(ctx context.Context, id int64) (*resp.Response, error)
	FormSearch(ctx context.Context, title string, userID uint) (*resp.Response, error)
	FormResults(ctx context.Context, id int64) (*resp.Response, error)
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
	var response model.FormList
	response.Forms = make([]*model.Form, 0)

	forms, err := s.formRepository.FindAll(ctx)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if ctx.Value(model.ContextCurrentUser) != nil {
		currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

		for _, form := range forms {
			total, err := s.formRepository.UserFormPassageCount(ctx, *form.ID, currentUser.ID)
			if err != nil {
				return resp.NewResponse(http.StatusInternalServerError, nil), nil
			}

			form.CurrentPassageTotal = int(total)
		}
	}

	response.Count = len(forms)
	response.Forms = forms

	response.Sanitize(s.sanitizer)
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

	response.Sanitize(s.sanitizer)
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

	if ctx.Value(model.ContextCurrentUser) != nil {
		currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

		total, err := s.formRepository.UserFormPassageCount(ctx, *form.ID, currentUser.ID)
		if err != nil {
			return resp.NewResponse(http.StatusInternalServerError, nil), nil
		}

		form.CurrentPassageTotal = int(total)
	}
	form.Sanitize(s.sanitizer)

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

	formTitleList.Sanitize(s.sanitizer)

	return resp.NewResponse(http.StatusOK, formTitleList), nil
}
