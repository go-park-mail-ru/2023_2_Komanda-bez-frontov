package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"go-form-hub/internal/model"
	"go-form-hub/internal/services/form"

	"github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type FormAPIController struct {
	service         form.Service
	validator       *validator.Validate
	responseEncoder ResponseEncoder
}

func NewFormAPIController(service form.Service, v *validator.Validate, responseEncoder ResponseEncoder) Router {
	return &FormAPIController{
		service:         service,
		validator:       v,
		responseEncoder: responseEncoder,
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
	ctx := r.Context()

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()
	if err != nil {
		log.Error().Msgf("form_api form_save body read error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var formSave model.Form
	if err = json.Unmarshal(requestJSON, &formSave); err != nil {
		log.Error().Msgf("form_api form_save unmarshal error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormSave(r.Context(), &formSave)
	if err != nil {
		log.Error().Msgf("form_api form_save error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	author := r.URL.Query().Get("author")

	if author != "" {
		result, err := c.service.FormListByUser(ctx, author)
		if err != nil {
			log.Error().Msgf("form_api form_list error: %e", err)
			c.responseEncoder.HandleError(ctx, w, err, result)
			return
		}

		c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
	} else {
		result, err := c.service.FormList(ctx)
		if err != nil {
			log.Error().Msgf("form_api form_list error: %e", err)
			c.responseEncoder.HandleError(ctx, w, err, result)
			return
		}

		c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
	}
}

// nolint:dupl
func (c *FormAPIController) FormDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_delete unescape error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_delete parse_id error: %e", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormDelete(ctx, id)
	if err != nil {
		log.Error().Msgf("form_api form_delete error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

// nolint:dupl
func (c *FormAPIController) FormGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_get unescape error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_get parse_id error: %e", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormGet(ctx, id)
	if err != nil {
		log.Error().Msgf("form_api form_get error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_update unescape error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_update parse_id error: %e", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		log.Error().Msgf("form_api form_update read_body error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var updatedForm model.Form
	if err = json.Unmarshal(requestJSON, &updatedForm); err != nil {
		log.Error().Msgf("form_api form_update unmarshal error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormUpdate(ctx, id, &updatedForm)
	if err != nil {
		log.Error().Msgf("form_api form_update error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}
