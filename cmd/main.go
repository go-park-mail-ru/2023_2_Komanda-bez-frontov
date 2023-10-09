package main

import (
	"context"
	"fmt"
	"go-form-hub/internal/api"
	"go-form-hub/internal/config"
	repository "go-form-hub/internal/repository/mocks"
	"go-form-hub/internal/services/auth"
	"go-form-hub/internal/services/form"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func StartServer(cfg *config.Config, r http.Handler) (*http.Server, error) {
	ln, err := net.Listen("tcp", ":"+cfg.HTTPPort)
	if err != nil {
		return nil, fmt.Errorf("tcp listen failed, net listen error %s", err)
	}
	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	}

	go func() {
		err = server.Serve(ln)
		if err != nil {
			log.Fatal().Err(err).Msg("http server stopped")
		}
	}()
	return server, nil
}

func main() {
	log.Info().Msg("Starting application...")
	cfg, err := config.NewConfig()
	if err != nil {
		log.Error().Msg(fmt.Sprintf("application failed to start: %s", err))
		return
	}
	zerolog.SetGlobalLevel(config.ZeroLogLevel(cfg.LogLevel))

	validate := validator.New()

	formRepository := repository.NewFormMockRepository()
	sessionRepository := repository.NewSessionMockRepository()
	userRepository := repository.NewUserMockRepository()

	formService := form.NewFormService(formRepository, validate)
	authService := auth.NewAuthService(userRepository, sessionRepository, cfg, validate)

	formRouter := api.NewFormAPIController(formService, validate)
	authRouter := api.NewAuthAPIController(authService, validate, cfg.CookieExpiration)

	authMiddleware := api.AuthMiddleware(sessionRepository, userRepository, cfg.CookieExpiration)
	r := api.NewRouter(authMiddleware, formRouter, authRouter)

	server, err := StartServer(cfg, r)
	if err != nil {
		log.Error().Msgf("Failed to start server: %e", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-interrupt

	log.Info().Msgf("Received system signal: %s, application will be shutdown", sig)

	if err := server.Shutdown(context.Background()); err != nil {
		log.Error().Msgf("failed to gracefully shutdown http server: %e", err)
	}
}
