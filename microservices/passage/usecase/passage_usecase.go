package usecase

import (
	"context"
	"net/http"

	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/service_response"

	"github.com/go-playground/validator/v10"
)

const noLimit = -1

type FormPassageUseCase interface {
	FormPass(ctx context.Context, formPassage *model.FormPassage) (*resp.Response, error)
}

type formPasageUseCase struct {
	formRepository repository.FormRepository
	validate       *validator.Validate
}

func NewformPasageUseCase(formRepository repository.FormRepository, validate *validator.Validate) FormPassageUseCase {
	return &formPasageUseCase{
		formRepository: formRepository,
		validate:       validate,
	}
}

func (s *formPasageUseCase) FormPass(ctx context.Context, formPassage *model.FormPassage) (*resp.Response, error) {
	userID := model.AnonUserID

	if err := s.validate.Struct(formPassage); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), err
	}

	existingForm, err := s.formRepository.FindByID(ctx, *formPassage.FormID)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if existingForm == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	if !existingForm.Anonymous {
		value := ctx.Value(model.ContextCurrentUser)
		if value == nil {
			return resp.NewResponse(http.StatusUnauthorized, nil), nil
		}

		currentUser := value.(*model.UserGet)
		userID = int(currentUser.ID)

		totalPassages, err := s.formRepository.UserFormPassageCount(ctx, *existingForm.ID, currentUser.ID)
		if err != nil {
			return resp.NewResponse(http.StatusInternalServerError, nil), nil
		}

		if existingForm.PassageMax != noLimit && totalPassages >= int64(existingForm.PassageMax) {
			return resp.NewResponse(http.StatusBadRequest, nil), nil
		}
	}

	var formValidator passageValidator
	err = formValidator.validateFormPassage(formPassage, existingForm)
	if err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), err
	}

	err = s.formRepository.FormPassageSave(ctx, formPassage, uint64(userID))
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	return resp.NewResponse(http.StatusNoContent, nil), nil
}
