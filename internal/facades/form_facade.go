package facades

import (
	"context"
	"fmt"
	"go-form-hub/internal/database"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
)

type FormFacade interface {
	Insert(ctx context.Context, form *model.Form) (*model.Form, error)
	Update(ctx context.Context, id int64, form *model.Form) (*model.Form, error)
}

type formFacade struct {
	formRepository     repository.FormRepository
	questionRepository repository.QuestionRepository
	answerRepository   repository.AnswerRepository
	db                 database.ConnPool
}

func NewFormFacade(formRepository repository.FormRepository, questionRepository repository.QuestionRepository, answerRepository repository.AnswerRepository, db database.ConnPool) FormFacade {
	return &formFacade{
		formRepository:     formRepository,
		questionRepository: questionRepository,
		answerRepository:   answerRepository,
		db:                 db,
	}
}

func (f *formFacade) Insert(ctx context.Context, form *model.Form) (res *model.Form, err error) {
	tx, err := f.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_facade insert failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	res = form

	formID, err := f.formRepository.Insert(ctx, &repository.Form{
		Title:     form.Title,
		AuthorID:  form.Author.ID,
		CreatedAt: form.CreatedAt,
	}, tx)
	if err != nil {
		err = fmt.Errorf("form_facade insert failed to insert form: %e", err)
		return
	}

	res.ID = formID

	var questions []*repository.Question
	for _, question := range form.Questions {
		questions = append(questions, &repository.Question{
			FormID:      *formID,
			Title:       question.Title,
			Description: question.Description,
			Type:        question.Type,
			Shuffle:     question.Shuffle,
		})
	}

	ids, err := f.questionRepository.BatchInsert(ctx, questions, tx)
	if err != nil {
		err = fmt.Errorf("form_facade insert failed to insert questions: %e", err)
		return
	}

	var answers []*repository.Answer
	for i, question := range form.Questions {

		for _, answer := range question.Answers {
			answers = append(answers, &repository.Answer{
				QuestionID: ids[i],
				Text:       answer.Text,
			})
		}

		res.Questions[i].ID = &ids[i]
	}

	ids, err = f.answerRepository.BatchInsert(ctx, answers, tx)
	if err != nil {
		err = fmt.Errorf("form_facade insert failed to insert answers: %e", err)
		return
	}

	return
}

func (f *formFacade) Update(ctx context.Context, id int64, form *model.Form) (*model.Form, error) {
	return nil, nil
}
