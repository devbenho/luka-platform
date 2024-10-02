package main

import (
	"log"
	"net/http"

	"github.com/devbenho/bazar-user-service/ports/api/errors"
	"github.com/devbenho/bazar-user-service/ports/api/handlers"
	"github.com/devbenho/bazar-user-service/ports/api/middlewares"
	"github.com/gorilla/mux"

	configs "github.com/devbenho/bazar-user-service/configs"
	"github.com/devbenho/bazar-user-service/internal/database"
	"github.com/devbenho/bazar-user-service/internal/repositories"
	"github.com/devbenho/bazar-user-service/internal/services"
	"github.com/devbenho/bazar-user-service/pkg/hasher"
	"github.com/devbenho/bazar-user-service/pkg/tokens"
	"github.com/devbenho/bazar-user-service/pkg/validation"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Load configuration
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	initDB()

	// Initialize dependencies
	validator := validation.NewValidator()
	tokenService := tokens.NewTokenService(cfg.JWT.Secret)
	userRepo := repositories.NewUserRepository(database.Database)
	hasher := hasher.NewHasher()
	userService := services.NewUserService(validator, tokenService, userRepo, hasher)
	userHandler := handlers.NewUserHandler(userService)

	// Setup routes
	mux := mux.NewRouter()
	// add prefix to the routes /api/v1
	mux = mux.PathPrefix("/api/v1").Subrouter()
	mux.HandleFunc("/auth/register", errors.ErrorHandler(userHandler.Register)).Methods("POST")

	mux.HandleFunc("/auth/login", errors.ErrorHandler(userHandler.Login)).Methods("POST")
	mux.HandleFunc("/user/:id", errors.ErrorHandler(userHandler.GetUserByID)).Methods("GET")
	mux.HandleFunc("/user/:id", errors.ErrorHandler(userHandler.UpdateUser)).Methods("PUT")
	mux.HandleFunc("/user/:id", userHandler.DeleteUser).Methods("DELETE")

	// Swagger route
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Wrap the mux with the error middleware
	log.Printf("Starting server on port %s...", cfg.App.Port)
	if err := http.ListenAndServe(":"+cfg.App.Port, middlewares.ErrorMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
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
