package api

import (
	"context"
	"encoding/json"
	"go-form-hub/internal/model"
	"net/http"

	resp "go-form-hub/internal/services/service_response"

	"github.com/microcosm-cc/bluemonday"
)

type ResponseEncoder interface {
	EncodeJSONResponse(ctx context.Context, i interface{}, status int, w http.ResponseWriter)
	HandleError(ctx context.Context, w http.ResponseWriter, err error, result *resp.Response)
}

type responseEncoder struct{}

func NewResponseEncoder() ResponseEncoder {
	return &responseEncoder{}
}

func (r *responseEncoder) EncodeJSONResponse(ctx context.Context, i interface{}, status int, w http.ResponseWriter) {
	if status == http.StatusNoContent {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)

	user := r.getCurrentUserFromCtx(ctx)
	if user != nil {
		user.Sanitize(bluemonday.UGCPolicy())
	}
	result := model.BasicResponse{
		Data:        i,
		CurrentUser: user,
	}
	_ = json.NewEncoder(w).Encode(result)
}

func (r *responseEncoder) HandleError(ctx context.Context, w http.ResponseWriter, err error, result *resp.Response) {
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

	r.EncodeJSONResponse(ctx, response, code, w)
}

func (r *responseEncoder) getCurrentUserFromCtx(ctx context.Context) *model.UserGet {
	if ctx.Value(model.ContextCurrentUser) == nil {
		return nil
	}

	return ctx.Value(model.ContextCurrentUser).(*model.UserGet)
}
