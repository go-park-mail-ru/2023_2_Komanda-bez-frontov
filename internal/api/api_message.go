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

type MessageAPIController struct {
	service         message.Service
	validator       *validator.Validate
	responseEncoder ResponseEncoder
}

func NewMessageAPIController(service message.Service, v *validator.Validate, responseEncoder ResponseEncoder) Router {
	return &FormAPIController{
		service:         service,
		validator:       v,
		responseEncoder: responseEncoder,
	}
}

func (c *MessageAPIController) Routes() []Route {
	return []Route{
		{
			Name:         "MessageSend",
			Method:       http.MethodPost,
			Path:         "/message/send",
			Handler:      c.MessageSave,
			AuthRequired: true,
		},
		{
			Name:         "MessageSend",
			Method:       http.MethodPost,
			Path:         "/message/send",
			Handler:      c.FormSave,
			AuthRequired: true,
		},
	}
}

func (c *MessageAPIController) MessageSave(w http.ResponseWriter, r *http.Request) {
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

// getMessageByID получает сообщение по его идентификатору
func (h *messageHandler) getMessageByID(w http.ResponseWriter, r *http.Request) {
    // Извлечение идентификатора из запроса и получение сообщения из сервиса
}

// markAsRead помечает сообщение как прочитанное
func (h *messageHandler) markAsRead(w http.ResponseWriter, r *http.Request) {
    // Извлечение идентификатора сообщения из запроса и вызов соответствующего метода сервиса
}

