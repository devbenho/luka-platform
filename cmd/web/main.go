package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	configs "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/internal/database"
	"github.com/devbenho/luka-platform/internal/repositories"
	"github.com/devbenho/luka-platform/internal/services"
	"github.com/devbenho/luka-platform/pkg/hasher"
	"github.com/devbenho/luka-platform/pkg/tokens"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/devbenho/luka-platform/ports/http/errors"
	"github.com/devbenho/luka-platform/ports/http/handlers"
)

func main() {
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	initDB()

	validator := validation.NewValidator()
	tokenService := tokens.NewTokenService(cfg.JWT.Secret)
	userRepo := repositories.NewUserRepository(database.Database)
	hasher := hasher.NewHasher()
	userService := services.NewUserService(validator, tokenService, userRepo, hasher)
	userHandler := handlers.NewUserHandler(userService)

	mux := setupRoutes(userHandler)

	server := &http.Server{
		Addr:     ":" + cfg.App.Port,
		Handler:  mux,
		ErrorLog: log.New(log.Writer(), "", 0),
	}

	go func() {
		startServer(server, cfg.App.Port, cfg.App.Environment)
	}()

	gracefulShutdown(server)
}

func initDB() {
	if err := database.Connect(); err != nil {
		if dbErr, ok := err.(*database.DBConnectionError); ok {
			log.Fatalf("Custom error occurred: %s\n", dbErr.Error())
		} else {
			log.Fatalf("An error occurred: %s\n", err.Error())
		}
	}
}

func setupRoutes(userHandler *handlers.UserHandler) *http.ServeMux {
	mux := http.NewServeMux()
	// health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/auth/register", errors.ErrorHandler(userHandler.Register))
	mux.HandleFunc("/auth/login", errors.ErrorHandler(userHandler.Login))
	mux.HandleFunc("/user/", errors.ErrorHandler(userHandler.GetUserByID))
	// mux.HandleFunc("/user/", errors.ErrorHandler(userHandler.UpdateUser))
	// mux.HandleFunc("/user/", userHandler.DeleteUser)
	// mux.Handle("/swagger/", httpSwagger.WrapHandler)
	return mux
}

func startServer(server *http.Server, port, environment string) {
	log.Printf("🚀 Starting server on port %s...", port)
	log.Printf("🥝 ENV %s...", environment)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting server: %s\n", err)
	}
}

func gracefulShutdown(server *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Println("Got signal:", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server exited properly")
}
