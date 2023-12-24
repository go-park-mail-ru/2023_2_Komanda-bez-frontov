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
	passage "go-form-hub/microservices/passage/passage_client"

	"github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type FormAPIController struct {
	service         form.Service
	passageService  passage.FormPassageClient
	validator       *validator.Validate
	responseEncoder ResponseEncoder
}

func NewFormAPIController(service form.Service, passageService passage.FormPassageClient, v *validator.Validate, responseEncoder ResponseEncoder) Router {
	return &FormAPIController{
		service:         service,
		passageService:  passageService,
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
			Name:         "FormArchive",
			Method:       http.MethodPut,
			Path:         "/forms/{id}/archive",
			Handler:      c.FormArchive,
			AuthRequired: true,
		},
		{
			Name:         "FormUpdate",
			Method:       http.MethodPut,
			Path:         "/forms/{id}/update",
			Handler:      c.FormUpdate,
			AuthRequired: true,
		},
		{
			Name:         "FormSearch",
			Method:       http.MethodGet,
			Path:         "/forms/search",
			Handler:      c.FormSearch,
			AuthRequired: true,
		},
		{
			Name:         "FormResults",
			Method:       http.MethodGet,
			Path:         "/forms/{id}/results",
			Handler:      c.FormResults,
			AuthRequired: true,
		},
		{
			Name:         "FormPassage",
			Method:       http.MethodPost,
			Path:         "/forms/pass",
			Handler:      c.FormPass,
			AuthRequired: false,
		},
		{
			Name:         "FormPassageList",
			Method:       http.MethodGet,
			Path:         "/forms/pass/list",
			Handler:      c.FormPassageList,
			AuthRequired: true,
		},
		{
			Name:         "FormPassageGet",
			Method:       http.MethodGet,
			Path:         "/forms/pass/{id}",
			Handler:      c.FormPassageGet,
			AuthRequired: true,
		},
		{
			Name:         "FormResultsCsv",
			Method:       http.MethodGet,
			Path:         "/forms/{id}/results/csv",
			Handler:      c.FormResultsCsv,
			AuthRequired: true,
		},
		{
			Name:         "FormResultsExel",
			Method:       http.MethodGet,
			Path:         "/forms/{id}/results/excel",
			Handler:      c.FormResultsExel,
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
		log.Error().Msgf("form_api form_save body read error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var formSave model.Form
	if err = json.Unmarshal(requestJSON, &formSave); err != nil {
		log.Error().Msgf("form_api form_save unmarshal error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormSave(r.Context(), &formSave)
	if err != nil {
		log.Error().Msgf("form_api form_save error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormPass(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()
	if err != nil {
		log.Error().Msgf("form_api form_pass body read error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var formPassage model.FormPassage
	if err = json.Unmarshal(requestJSON, &formPassage); err != nil {
		log.Error().Msgf("form_api form_passage unmarshal error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	answersMsg := make([]*passage.PassageAnswer, 0)
	for _, passageAnswer := range formPassage.PassageAnswers {
		answersMsg = append(answersMsg, &passage.PassageAnswer{
			Text:       passageAnswer.Text,
			QuestionID: *passageAnswer.QuestionID,
		})
	}

	currentUser, ok := ctx.Value(model.ContextCurrentUser).(*model.UserGet)
	if !ok {
		currentUser = &model.UserGet{ID: model.AnonUserID}
	}

	passageMsg := &passage.Passage{
		UserID:  currentUser.ID,
		FormID:  *formPassage.FormID,
		Answers: answersMsg,
	}

	result, err := c.passageService.Pass(ctx, passageMsg)
	if err != nil {
		log.Error().Msgf("form_api form_pass error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, nil, int(result.Code), w)
}

func (c *FormAPIController) FormList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	author := r.URL.Query().Get("author")
	isArchived := r.URL.Query().Get("archive") == "true"

	if author != "" {
		result, err := c.service.FormListByUser(ctx, author, isArchived)
		if err != nil {
			log.Error().Msgf("form_api form_list error: %v", err)
			c.responseEncoder.HandleError(ctx, w, err, result)
			return
		}

		c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
	} else {
		result, err := c.service.FormList(ctx)
		if err != nil {
			log.Error().Msgf("form_api form_list error: %v", err)
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
		log.Error().Msgf("form_api form_delete unescape error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_delete parse_id error: %v", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormDelete(ctx, id)
	if err != nil {
		log.Error().Msgf("form_api form_delete error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

// nolint:dupl
func (c *FormAPIController) FormArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_archive unescape error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_archive parse_id error: %v", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	archive := r.URL.Query().Get("archive") == "true"

	result, err := c.service.FormArchive(ctx, id, archive)
	if err != nil {
		log.Error().Msgf("form_api form_archive error: %v", err)
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
		log.Error().Msgf("form_api form_get unescape error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_get parse_id error: %v", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormGet(ctx, id)
	if err != nil {
		log.Error().Msgf("form_api form_get error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)
	isArchived := r.URL.Query().Get("archive") == "true"

	title := r.URL.Query().Get("title")
	result, err := c.service.FormSearch(ctx, title, uint(currentUser.ID), isArchived)
	if err != nil {
		log.Error().Msgf("form_api form_search error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}
	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_result unescape error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_result parse_id error: %e", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormResults(ctx, id)
	if err != nil {
		log.Error().Msgf("form_api form_results error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormResultsCsv(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_result_exel unescape error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_result_exel parse_id error: %e", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormResultsCsv(ctx, id)
	if err != nil {
		log.Error().Msgf("form_api form_results_exel error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=export.xlsx")
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	_, err = w.Write(result)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func (c *FormAPIController) FormResultsExel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_result_exel unescape error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_result_exel parse_id error: %e", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormResultsExel(ctx, id)
	if err != nil {
		log.Error().Msgf("form_api form_results_exel error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=export.xlsx")
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	_, err = w.Write(result)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func (c *FormAPIController) FormUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_update unescape error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_update parse_id error: %v", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		log.Error().Msgf("form_api form_update read_body error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var updatedForm model.FormUpdate
	if err = json.Unmarshal(requestJSON, &updatedForm); err != nil {
		log.Error().Msgf("form_api form_update unmarshal error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormUpdate(ctx, id, &updatedForm)
	if err != nil {
		log.Error().Msgf("form_api form_update error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormPassageList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := c.service.FormPassageList(ctx)
	if err != nil {
		log.Error().Msgf("form_api form_passage_list error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *FormAPIController) FormPassageGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("form_api form_passage_get unescape error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("form_api form_passage_get parse_id error: %v", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.FormPassageGet(ctx, id)
	if err != nil {
		log.Error().Msgf("form_api form_passage_get error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}