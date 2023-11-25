package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go-form-hub/internal/api"
	"go-form-hub/internal/config"
	"go-form-hub/internal/database"
	"go-form-hub/internal/repository"
	"go-form-hub/internal/services/auth"
	"go-form-hub/internal/services/form"
	"go-form-hub/internal/services/user"

	"github.com/Masterminds/squirrel"
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
	log.Info().Interface("config", cfg).Msgf("Server config")

	db, err := database.ConnectDatabaseWithRetry(cfg)
	if err != nil {
		log.Error().Msgf("failed to connect database: %s", err)
		return
	}
	defer db.Close()

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	_, err = database.Migrate(db, cfg, builder)
	if err != nil {
		log.Error().Msgf("failed to migrate database: %s", err)
		return
	}

	validate := validator.New()

	userRepository := repository.NewUserDatabaseRepository(db, builder)
	formRepository := repository.NewFormDatabaseRepository(db, builder)
	sessionRepository := repository.NewSessionDatabaseRepository(db, builder)
	questionRepository := repository.NewQuestionDatabaseRepository(db, builder)
	answerRepository := repository.NewAnswerDatabaseRepository(db, builder)

	formService := form.NewFormService(formRepository, questionRepository, answerRepository, validate)
	authService := auth.NewAuthService(userRepository, sessionRepository, cfg, validate)
	userService := user.NewUserService(userRepository, cfg, validate)

	responseEncoder := api.NewResponseEncoder()
	formRouter := api.NewFormAPIController(formService, validate, responseEncoder)
	authRouter := api.NewAuthAPIController(authService, validate, cfg.CookieExpiration, responseEncoder)
	userRouter := api.NewUserAPIController(userService, validate, responseEncoder)

	authMiddleware := api.AuthMiddleware(sessionRepository, userRepository, cfg.CookieExpiration, responseEncoder)
	currentUserMiddleware := api.CurrentUserMiddleware(sessionRepository, userRepository, cfg.CookieExpiration)

	r := api.NewRouter(cfg, authMiddleware, currentUserMiddleware, formRouter, authRouter, userRouter)

	server, err := StartServer(cfg, r)
	if err != nil {
		log.Error().Msgf("Failed to start server: %e", err)
	}

	log.Info().Msgf("Server started. Listening port %s", cfg.HTTPPort)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-interrupt

	log.Info().Msgf("Received system signal: %s, application will be shutdown", sig)

	if err := server.Shutdown(context.Background()); err != nil {
		log.Error().Msgf("failed to gracefully shutdown http server: %e", err)
	}
}
