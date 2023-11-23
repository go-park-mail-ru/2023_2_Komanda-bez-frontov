package form

import (
	"context"
	"errors"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	"net/http"
	"time"

	resp "go-form-hub/internal/services/service_response"

	validator "github.com/go-playground/validator/v10"
)

var (
	ErrMultipleAnswers            = errors.New("multiple answers given to single answer question")
	ErrQuestionDoesntExist        = errors.New("answer to non-existent question was given")
	ErrAnswerDoesntExist          = errors.New("non selectable answer was given")
	ErrRequiredQuestionUnanswered = errors.New("required question was not answered")
)

type Service interface {
	FormSave(ctx context.Context, form *model.Form) (*resp.Response, error)
	FormUpdate(ctx context.Context, id int64, form *model.Form) (*resp.Response, error)
	FormList(ctx context.Context) (*resp.Response, error)
	FormListByUser(ctx context.Context, username string) (*resp.Response, error)
	FormDelete(ctx context.Context, id int64) (*resp.Response, error)
	FormGet(ctx context.Context, id int64) (*resp.Response, error)
	FormSearch(ctx context.Context, title string, userID uint) (*resp.Response, error)
	FormPass(ctx context.Context, formPassage *model.FormPassage) (*resp.Response, error)
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

func (s *formService) FormPass(ctx context.Context, formPassage *model.FormPassage) (*resp.Response, error) {
	value := ctx.Value(model.ContextCurrentUser)
	userID := 0
	if value != nil {
		currentUser := value.(*model.UserGet)
		userID = int(currentUser.ID)
	}
	if err := s.validate.Struct(formPassage); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), err
	}

	existingForm, err := s.formRepository.FindByID(ctx, int64(*formPassage.FormID))
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if existingForm == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	err = s.formRepository.FormPassageSave(ctx, formPassage, uint64(userID))
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	return resp.NewResponse(http.StatusOK, nil), nil
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

func validatePassageAnswers(formPassage *model.FormPassage, form *model.Form) error {
	questionMap := questionMapFromArray(form.Questions)
	foundQuestionsMap := make(map[int64]bool)
	foundAnswerMap := make(map[int64]bool)

	for _, passageAnswer := range formPassage.PassageAnswers {
		question, found := questionMap[*passageAnswer.QuestionID]
		if !found {
			return ErrQuestionDoesntExist
		}

		if foundQuestionsMap[*passageAnswer.QuestionID] && question.Type != 3 {
			return ErrMultipleAnswers
		}
		foundQuestionsMap[*passageAnswer.QuestionID] = true

		if question.Type == 1 {
			continue
		}

		if question.Type == 2 {
			found := false
			for _, answer := range question.Answers {
				if answer.Text == passageAnswer.Text {
					found = true
					break
				}
			}

			if !found {
				return ErrAnswerDoesntExist
			}
			continue
		}

		if question.Type == 3 {
			found := false
			for _, answer := range question.Answers {
				if answer.Text == passageAnswer.Text {
					_, questionFound := foundAnswerMap[*answer.ID]
					if questionFound {
						return ErrMultipleAnswers
					}
					foundAnswerMap[*answer.ID] = true

					found = true
					break
				}
			}

			if !found {
				return ErrAnswerDoesntExist
			}
			continue
		}
	}

	for questionID, found := range foundQuestionsMap {
		if !found && questionMap[questionID].Required {
			return ErrRequiredQuestionUnanswered
		}
	}

	return nil
}

func questionMapFromArray(questions []*model.Question) map[int64]*model.Question {
	questionMap := make(map[int64]*model.Question)
	for _, question := range questions {
		questionMap[*question.ID] = question
	}
	return questionMap
}
