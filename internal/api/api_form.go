package api

import (
	"encoding/json"
	"fmt"
	"go-form-hub/internal/model"
	"go-form-hub/internal/services/form"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type FormAPIController struct {
	service      form.Service
	errorHandler ErrorHandler
	validator    *validator.Validate
}

func NewFormAPIController(service form.Service, v *validator.Validate) Router {
	return &FormAPIController{
		service:      service,
		errorHandler: HandleError,
		validator:    v,
	}
}

func (c *FormAPIController) Routes() []Route {
	return []Route{
		{
			Name:         "FormSave",
			Method:       http.MethodPost,
			Path:         "/forms/save",
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
			Name:         "FormGet",
			Method:       http.MethodGet,
			Path:         "/forms/{id}",
			Handler:      c.FormGet,
			AuthRequired: false,
		},
		{
			Name:         "FormDelete",
			Method:       http.MethodDelete,
			Path:         "/forms/{id}/delete",
			Handler:      c.FormDelete,
			AuthRequired: true,
		},
		{
			Name:         "FormUpdate",
			Method:       http.MethodPut,
			Path:         "/forms/{id}/update",
			Handler:      c.FormUpdate,
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
		log.Error().Msgf("form_api form_save body read error: %e", err)
		c.errorHandler(w, err, nil)
		return
	}

	var formSave model.Form
	if err = json.Unmarshal(requestJSON, &formSave); err != nil {
		log.Error().Msgf("form_api form_save unmarshal error: %e", err)
		c.errorHandler(w, err, nil)
		return
	}

	result, err := c.service.FormSave(r.Context(), &formSave)
	if err != nil {
		log.Error().Msgf("form_api form_save error: %e", err)
		c.errorHandler(w, err, result)
		return
	}

	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormList(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.FormList(r.Context())
	if err != nil {
		log.Error().Msgf("form_api form_list error: %e", err)
		c.errorHandler(w, err, result)
		return
	}

	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

// nolint:dupl
func (c *FormAPIController) FormDelete(w http.ResponseWriter, r *http.Request) {
	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_delete unescape error: %e", err)
		c.errorHandler(w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_get parse_id error: %e", err)
		log.Error().Msg(err.Error())
		c.errorHandler(w, err, nil)
	}

	result, err := c.service.FormDelete(r.Context(), id)
	if err != nil {
		log.Error().Msgf("form_api form_delete error: %e", err)
		c.errorHandler(w, err, result)
		return
	}

	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

// nolint:dupl
func (c *FormAPIController) FormGet(w http.ResponseWriter, r *http.Request) {
	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_get unescape error: %e", err)
		c.errorHandler(w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_get parse_id error: %e", err)
		log.Error().Msg(err.Error())
		c.errorHandler(w, err, nil)
	}

	result, err := c.service.FormGet(r.Context(), id)
	if err != nil {
		log.Error().Msgf("form_api form_get error: %e", err)
		c.errorHandler(w, err, result)
		return
	}

	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormUpdate(w http.ResponseWriter, r *http.Request) {
	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_update unescape error: %e", err)
		c.errorHandler(w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_get parse_id error: %e", err)
		log.Error().Msg(err.Error())
		c.errorHandler(w, err, nil)
	}

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		log.Error().Msgf("form_api form_update read_body error: %e", err)
		c.errorHandler(w, err, nil)
		return
	}

	var updatedForm model.Form
	if err = json.Unmarshal(requestJSON, &updatedForm); err != nil {
		log.Error().Msgf("form_api form_update unmarshal error: %e", err)
		c.errorHandler(w, err, nil)
		return
	}

	result, err := c.service.FormUpdate(r.Context(), id, &updatedForm)
	if err != nil {
		log.Error().Msgf("form_api form_update error: %e", err)
		c.errorHandler(w, err, result)
		return
	}

	EncodeJSONResponse(result.Body, result.StatusCode, w)
}
