package controller

import (
	"context"

	"go-form-hub/internal/model"
	passage "go-form-hub/microservices/passage/passage_client"
	"go-form-hub/microservices/passage/usecase"

	"github.com/go-playground/validator/v10"
)

type PassageController struct {
	passage.UnimplementedFormPassageServer

	passageUseCase usecase.FormPassageUseCase
	validator      *validator.Validate
}

func NewAuthController(passageUsecase usecase.FormPassageUseCase, v *validator.Validate) *PassageController {
	return &PassageController{
		passageUseCase: passageUsecase,
		validator:      v,
	}
}

func (controller *PassageController) Pass(ctx context.Context, passageMsg *passage.Passage) (*passage.ResultCode, error) {
	passageAnswers := make([]*model.PassageAnswer, 0)
	for _, answerMsg := range passageMsg.Answers {
		passageAnswers = append(passageAnswers, &model.PassageAnswer{
			QuestionID: &answerMsg.QuestionID,
			Text:       answerMsg.Text,
		})
	}

	passageModel := &model.FormPassage{
		FormID:         &passageMsg.FormID,
		PassageAnswers: passageAnswers,
	}
	ctx = context.WithValue(ctx, model.ContextCurrentUser, &model.UserGet{
		ID: passageMsg.UserID,
	})

	response, err := controller.passageUseCase.FormPass(ctx, passageModel)
	if err != nil {
		return &passage.ResultCode{Code: int64(response.StatusCode)}, err
	}

	return &passage.ResultCode{Code: int64(response.StatusCode)}, nil
}
