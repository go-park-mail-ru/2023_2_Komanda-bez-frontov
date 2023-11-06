package api

import (
	"encoding/json"
	"go-form-hub/internal/model"
	resp "go-form-hub/internal/services/service_response"
	"net/http"
)

type ResponseEncoder interface {
	EncodeJSONResponse(i interface{}, status int, w http.ResponseWriter)
	HandleError(w http.ResponseWriter, err error, result *resp.Response)
}

type responseEncoder struct{}

func NewResponseEncoder() ResponseEncoder {
	return &responseEncoder{}
}

func (r *responseEncoder) EncodeJSONResponse(i interface{}, status int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(i)
}

func (r *responseEncoder) HandleError(w http.ResponseWriter, err error, result *resp.Response) {
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

	r.EncodeJSONResponse(response, code, w)
}
