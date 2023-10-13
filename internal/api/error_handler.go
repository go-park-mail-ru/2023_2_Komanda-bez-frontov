package api

import (
	"go-form-hub/internal/model"
	resp "go-form-hub/internal/services/service_response"
	"net/http"
)

type ErrorHandler func(w http.ResponseWriter, err error, result *resp.Response)

// HandleError defines the default logic on how to handle errors from the controller. Any errors from parsing
// request params will return a StatusBadRequest. Otherwise, the error code originating from the servicer will be used.
func HandleError(w http.ResponseWriter, err error, result *resp.Response) {
	errors := make([]model.Error, 0, 1)
	str := err.Error()
	errorItem := model.Error{
		Status: &str,
	}
	errors = append(errors, errorItem)
	response := model.ErrorResponse{
		Errors: &errors,
	}
	code := http.StatusBadRequest
	if result != nil {
		code = result.StatusCode
	}

	EncodeJSONResponse(response, code, w)
}
