package main

import (
	"context"
	"fmt"
	"go-form-hub/internal/api"
	repository "go-form-hub/internal/repository/mocks"
	"go-form-hub/internal/services/auth"
	"go-form-hub/internal/services/form"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
)

func StartServer(r http.Handler) (*http.Server, error) {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		return nil, fmt.Errorf("tcp listen failed, net listen error %s", err)
	}
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	go func() {
		err = server.Serve(ln)
		if err != nil {
			fmt.Printf("http server stopped: %e\n", err)
		}
	}()
	return server, nil
}

func main() {
	fmt.Printf("Starting server...\n\n")
	validate := validator.New()

	formRepository := repository.NewFormMockRepository()
	sessionRepository := repository.NewSessionMockRepository()
	userRepository := repository.NewUserMockRepository()

	formService := form.NewFormService(formRepository, validate)
	authService := auth.NewAuthService(userRepository, sessionRepository, validate)

	formRouter := api.NewFormAPIController(formService, validate)
	authRouter := api.NewAuthAPIController(authService, validate)

	authMiddleware := api.AuthMiddleware(sessionRepository, userRepository)
	r := api.NewRouter(authMiddleware, formRouter, authRouter)

	server, err := StartServer(r)
	if err != nil {
		fmt.Printf("Failed to start server: %e\n", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-interrupt

	fmt.Printf("\nreceived system signal: %s, application will be shutdown\n", sig)

	if err := server.Shutdown(context.Background()); err != nil {
		fmt.Printf("failed to shutdown http server: %e\n", err)
	}
}
