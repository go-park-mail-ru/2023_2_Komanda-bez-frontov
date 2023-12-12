package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go-form-hub/internal/config"
	"go-form-hub/internal/database"
	"go-form-hub/internal/repository"
	"go-form-hub/microservices/user/controller"
	"go-form-hub/microservices/user/profile"
	"go-form-hub/microservices/user/usecase"

	"github.com/Masterminds/squirrel"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const defaultPort = ":8082"

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

	userRepository := repository.NewUserDatabaseRepository(db, builder)
	userService := usecase.NewUserUseCase(userRepository, cfg, validate)
	userController := controller.NewProfileController(userService, validate)

	lis, err := net.Listen("tcp", defaultPort) // #nosec G102
	if err != nil {
		log.Error().Msgf("failed to listen to port: %v", err)
		return
	}

	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*16),
		grpc.MaxSendMsgSize(1024*1024*16),
	)

	profile.RegisterProfileServer(server, userController)
	err = server.Serve(lis)
	if err != nil {
		log.Error().Msgf("failed to serve port: %v", err)
		return
	}

	log.Info().Msgf("Server started. Listening port %s", cfg.HTTPPort)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-interrupt

	log.Info().Msgf("Received system signal: %s, application will be shutdown", sig)
}
