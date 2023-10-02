package api

import (
	"encoding/json"
	"go-form-hub/internal/model"
	"go-form-hub/internal/services/form"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type FormAPIController struct {
	service   form.Service
	validator *validator.Validate
}

func NewFormAPIController(service form.Service, v *validator.Validate) Router {
	return &FormAPIController{
		service:   service,
		validator: v,
	}
}

func (c *FormAPIController) Routes() []Route {
	return []Route{
		{
			Name:         "FormSave",
			Method:       http.MethodPost,
			Path:         "/forms/new",
			Handler:      c.FormSave,
			AuthRequired: true,
		},
		{
			Name:         "FormList",
			Method:       http.MethodGet,
			Path:         "/forms",
			Handler:      c.FormList,
			AuthRequired: false,
		},
		{
			Name:         "FormDelete",
			Method:       http.MethodDelete,
			Path:         "/forms/{title}",
			Handler:      c.FormDelete,
			AuthRequired: true,
		},
		{
			Name:         "FormGet",
			Method:       http.MethodGet,
			Path:         "/forms/{title}",
			Handler:      c.FormGet,
			AuthRequired: false,
		},
		{
			Name:         "FormUpdate",
			Method:       http.MethodPut,
			Path:         "/forms/{title}",
			Handler:      c.FormSave,
			AuthRequired: true,
		},
	}
}

func (c *FormAPIController) FormSave(w http.ResponseWriter, r *http.Request) {
	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	var form model.Form
	if err = json.Unmarshal(requestJSON, &form); err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	result, err := c.service.FormSave(r.Context(), &form)
	if err != nil {
		EncodeJSONResponse(err, result.StatusCode, w)
		return
	}

	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormList(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.FormList(r.Context())
	if err != nil {
		EncodeJSONResponse(err, result.StatusCode, w)
		return
	}

	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormDelete(w http.ResponseWriter, r *http.Request) {
	title, err := url.PathUnescape(chi.URLParam(r, "title"))
	if err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	result, err := c.service.FormDelete(r.Context(), title)
	if err != nil {
		EncodeJSONResponse(err, result.StatusCode, w)
		return
	}

	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormGet(w http.ResponseWriter, r *http.Request) {
	title, err := url.PathUnescape(chi.URLParam(r, "title"))
	if err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	result, err := c.service.FormGet(r.Context(), title)
	if err != nil {
		EncodeJSONResponse(err, result.StatusCode, w)
		return
	}

	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormUpdate(w http.ResponseWriter, r *http.Request) {
	title, err := url.PathUnescape(chi.URLParam(r, "title"))
	if err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	var form model.Form
	if err = json.Unmarshal(requestJSON, &form); err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	result, err := c.service.FormUpdate(r.Context(), title, &form)
	if err != nil {
		EncodeJSONResponse(err, result.StatusCode, w)
		return
	}

	EncodeJSONResponse(result.Body, result.StatusCode, w)
}
