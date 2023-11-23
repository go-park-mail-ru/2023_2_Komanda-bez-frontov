package form

import (
	"errors"
	"fmt"

	"go-form-hub/internal/model"
)

type FormValidator struct {
	questionMap       map[int64]*model.Question
	foundAnswerMap    map[int64]bool
	foundQuestionsMap map[int64]bool
}

var (
	ErrMultipleAnswers            = errors.New("multiple answers given to single answer question")
	ErrQuestionDoesntExist        = errors.New("answer to non-existent question was given")
	ErrAnswerDoesntExist          = errors.New("non selectable answer was given")
	ErrRequiredQuestionUnanswered = errors.New("required question was not answered")
)

func (v *FormValidator) ValidateFormPassage(formPassage *model.FormPassage, form *model.Form) error {
	v.questionMap = questionMapFromArray(form.Questions)
	v.foundQuestionsMap = make(map[int64]bool)
	v.foundAnswerMap = make(map[int64]bool)

	for _, passageAnswer := range formPassage.PassageAnswers {
		err := v.validatePassageAnswer(passageAnswer)
		if err != nil {
			return fmt.Errorf("error validating answer: %v", err)
		}
	}

	for questionID, question := range v.questionMap {
		_, found := v.foundQuestionsMap[questionID]
		if question.Required && !found {
			return ErrRequiredQuestionUnanswered
		}
	}

	return nil
}

func (v *FormValidator) validatePassageAnswer(passageAnswer *model.PassageAnswer) error {
	question, found := v.questionMap[*passageAnswer.QuestionID]
	if !found {
		return ErrQuestionDoesntExist
	}

	passed, found := v.foundQuestionsMap[*passageAnswer.QuestionID]
	if passed && question.Type != 3 {
		return ErrMultipleAnswers
	}
	v.foundQuestionsMap[*passageAnswer.QuestionID] = true

	switch question.Type {
	case 2:
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
	case 3:
		found := false
		for _, answer := range question.Answers {
			if answer.Text == passageAnswer.Text {
				_, questionFound := v.foundAnswerMap[*answer.ID]
				if questionFound {
					return ErrMultipleAnswers
				}
				v.foundAnswerMap[*answer.ID] = true

				found = true
				break
			}
		}

		if !found {
			return ErrAnswerDoesntExist
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
