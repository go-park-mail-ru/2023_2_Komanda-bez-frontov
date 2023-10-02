package main

import (
	"context"
	"fmt"
	"go-form-hub/internal/api"
	repository "go-form-hub/internal/repository/mocks"
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
	formService := form.NewFormService(formRepository, validate)
	formRouter := api.NewFormAPIController(formService, validate)

	r := api.NewRouter(formRouter)

	server, err := StartServer(r)
	if err != nil {
		fmt.Printf("Failed to start server: %e\n", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-interrupt

	fmt.Printf("received system signal: %s, application will be shutdown", sig)

	if err := server.Shutdown(context.Background()); err != nil {
		fmt.Printf("failed to shutdown http server: %e\n", err)
	}
}
