package form

import (
	"fmt"
	"go-form-hub/internal/model"
)

type FormValidator struct {
	questionMap       map[int64]*model.Question
	foundAnswerMap    map[int64]bool
	foundQuestionsMap map[int64]bool
}

func (v *FormValidator) ValidateFormPassage(formPassage *model.FormPassage, form *model.Form) error {
	v.questionMap = questionMapFromArray(form.Questions)
	v.foundQuestionsMap = make(map[int64]bool)
	v.foundAnswerMap = make(map[int64]bool)

	for _, passageAnswer := range formPassage.PassageAnswers {
		err := v.validatePassageAnswer(passageAnswer)
		if err != nil {
			return fmt.Errorf("error validating answer: %e", err)
		}
	}

	for questionID, found := range v.foundQuestionsMap {
		if !found && v.questionMap[questionID].Required {
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

	if v.foundQuestionsMap[*passageAnswer.QuestionID] && question.Type != 3 {
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
