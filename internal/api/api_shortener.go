package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	resp "go-form-hub/internal/services/service_response"
	"go-form-hub/internal/services/shortener"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type ShortenerAPIController struct {
	shortenerService shortener.Service
	validator        *validator.Validate
	responseEncoder  ResponseEncoder
}

func NewShortenerAPIController(shortenerService shortener.Service, v *validator.Validate, responseEncoder ResponseEncoder) Router {
	return &ShortenerAPIController{
		shortenerService: shortenerService,
		validator:        v,
		responseEncoder:  responseEncoder,
	}
}

func (c *ShortenerAPIController) Routes() []Route {
	return []Route{
		{
			Name:         "ShortenURL",
			Method:       http.MethodPost,
			Path:         "/shorten",
			Handler:      c.ShortenURL,
			AuthRequired: false,
		},
		{
			Name:         "RedirectHandler",
			Method:       http.MethodGet,
			Path:         "/redirect/{shortURL}",
			Handler:      c.RedirectHandler,
			AuthRequired: false,
		},
		{
			Name:         "GetLongURL",
			Method:       http.MethodGet,
			Path:         "/long-url/{shortURL}",
			Handler:      c.GetLongURL,
			AuthRequired: false,
		},
	}
}

func (c *ShortenerAPIController) ShortenURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Println("checker")

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()
	if err != nil {
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var requestData struct {
		LongURL string `json:"longURL" validate:"required,url"`
	}

	if err = json.Unmarshal(requestJSON, &requestData); err != nil {
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	if err := c.validator.Struct(requestData); err != nil {
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	shortURL, err := c.shortenerService.ShortenURL(ctx, requestData.LongURL)
	if err != nil {
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	responseData := struct {
		ShortURL string `json:"shortURL"`
	}{ShortURL: shortURL}

	c.responseEncoder.EncodeJSONResponse(ctx, responseData, http.StatusOK, w)
}

func (c *ShortenerAPIController) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := chi.URLParam(r, "shortURL")
	if params == "" {
		c.responseEncoder.HandleError(ctx, w, nil, &resp.Response{Body: nil, StatusCode: http.StatusBadRequest})
		return
	}

	longURL, err := c.shortenerService.GetLongURL(ctx, params)
	if err != nil {
		c.responseEncoder.HandleError(ctx, w, err, &resp.Response{Body: nil, StatusCode: http.StatusNotFound})
		return
	}

	http.Redirect(w, r, longURL[21:], http.StatusFound)
}

func (c *ShortenerAPIController) GetLongURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := chi.URLParam(r, "shortURL")
	if params == "" {
		c.responseEncoder.HandleError(ctx, w, nil, &resp.Response{Body: nil, StatusCode: http.StatusBadRequest})
		return
	}

	longURL, err := c.shortenerService.GetLongURL(ctx, params)
	if err != nil {
		c.responseEncoder.HandleError(ctx, w, err, &resp.Response{Body: nil, StatusCode: http.StatusNotFound})
		return
	}

	responseData := struct {
		LongURL string `json:"longURL"`
	}{LongURL: longURL}

	c.responseEncoder.EncodeJSONResponse(ctx, responseData, http.StatusOK, w)
}
