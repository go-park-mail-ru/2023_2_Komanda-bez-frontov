package shortener

import (
	"context"
	"fmt"
	"go-form-hub/internal/repository"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

type Service interface {
	ShortenURL(ctx context.Context, longURL string) (string, error)
	GetLongURL(ctx context.Context, shortURL string) (string, error)
	RedirectHandler(w http.ResponseWriter, req *http.Request)
}

type shortenerService struct {
	repository repository.ShortenerRepository
	sanitizer  *bluemonday.Policy
	validate   *validator.Validate
}

func NewShortenerService(urlRepository repository.ShortenerRepository, validate *validator.Validate) Service {
	sanitizer := bluemonday.UGCPolicy()
	return &shortenerService{
		repository: urlRepository,
		validate:   validate,
		sanitizer:  sanitizer,
	}
}

func (s *shortenerService) ShortenURL(ctx context.Context, longURL string) (string, error) {
	shortURL, err := s.repository.Insert(ctx, &repository.URLMapping{LongURL: longURL})
	if err != nil {
		return "", fmt.Errorf("failed to shorten URL: %v", err)
	}

	return shortURL, nil
}

func (s *shortenerService) GetLongURL(ctx context.Context, shortURL string) (string, error) {
	longURL, err := s.repository.GetLongURL(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("failed to get long URL: %v", err)
	}

	return longURL, nil
}

func (s *shortenerService) RedirectHandler(w http.ResponseWriter, req *http.Request) {
	shortURL := req.URL.Path[len("/redirect/"):]
	if shortURL == "" {
		http.Error(w, "Short URL not provided", http.StatusBadRequest)
		return
	}

	longURL, err := s.GetLongURL(req.Context(), shortURL)
	if err != nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, req, longURL, http.StatusFound)
}
