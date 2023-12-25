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
	"go-form-hub/internal/services/form"
	"go-form-hub/internal/services/shortener"
	"go-form-hub/microservices/auth/session"
	passage "go-form-hub/microservices/passage/passage_client"
	"go-form-hub/microservices/user/profile"

	"github.com/Masterminds/squirrel"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	authGrpcConn, err := grpc.Dial(
		"127.0.0.1:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error().Msgf("cant connect to grpc: %s", err)
		return
	}
	defer authGrpcConn.Close()

	sessController := session.NewAuthCheckerClient(authGrpcConn)

	userGrpcConn, err := grpc.Dial(
		"127.0.0.1:8082",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error().Msgf("cant connect to grpc: %s", err)
		return
	}
	defer userGrpcConn.Close()

	userController := profile.NewProfileClient(userGrpcConn)

	passageGrpcConn, err := grpc.Dial(
		"127.0.0.1:8083",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error().Msgf("cant connect to grpc: %s", err)
		return
	}
	defer passageGrpcConn.Close()

	passageController := passage.NewFormPassageClient(passageGrpcConn)

	validate := validator.New()
	tokenParser := api.NewHMACHashToken(cfg.Secret)

	shortenerRepository := repository.NewShortenerRepository(db, builder)
	userRepository := repository.NewUserDatabaseRepository(db, builder)
	formRepository := repository.NewFormDatabaseRepository(db, builder)
	sessionRepository := repository.NewSessionDatabaseRepository(db, builder)
	questionRepository := repository.NewQuestionDatabaseRepository(db, builder)
	answerRepository := repository.NewAnswerDatabaseRepository(db, builder)

	formService := form.NewFormService(formRepository, questionRepository, answerRepository, validate)
	shortenerService := shortener.NewShortenerService(shortenerRepository, validate)

	responseEncoder := api.NewResponseEncoder()

	formRouter := api.NewFormAPIController(formService, passageController, validate, responseEncoder)
	authRouter := api.NewAuthAPIController(tokenParser, sessController, validate, cfg.CookieExpiration, responseEncoder)
	userRouter := api.NewUserAPIController(userController, validate, responseEncoder)
	shortenerRouter := api.NewShortenerAPIController(shortenerService, validate, responseEncoder)

	authMiddleware := api.AuthMiddleware(sessionRepository, userRepository, cfg.CookieExpiration, responseEncoder)
	currentUserMiddleware := api.CurrentUserMiddleware(sessionRepository, userRepository, cfg.CookieExpiration)
	csrfMiddleware := api.CSRFMiddleware(tokenParser, responseEncoder)

	r := api.NewRouter(cfg, authMiddleware, currentUserMiddleware, csrfMiddleware, formRouter, authRouter, userRouter, shortenerRouter)

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
