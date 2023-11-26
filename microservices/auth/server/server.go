package main

import (
	"fmt"
	"net"

	"go-form-hub/internal/config"
	"go-form-hub/internal/database"
	"go-form-hub/internal/repository"
	"go-form-hub/internal/services/auth"
	service "go-form-hub/microservices/auth/service"
	"go-form-hub/microservices/auth/session"

	"github.com/Masterminds/squirrel"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func main() {
	log.Info().Msg("Starting microservice...")
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

	validate := validator.New()

	sessionRepository := repository.NewSessionDatabaseRepository(db, builder)
	userRepository := repository.NewUserDatabaseRepository(db, builder)
	authService := auth.NewAuthService(userRepository, sessionRepository, cfg, validate)
	authManager := service.NewAuthManager(authService, validate)

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal().Msg("cant listen to port: " + err.Error())
	}

	server := grpc.NewServer()

	session.RegisterAuthCheckerServer(server, authManager)

	fmt.Println("starting server at :8081")
	server.Serve(lis)
}
