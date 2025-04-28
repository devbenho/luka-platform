package main

import (
	"log"

	config "github.com/devbenho/luka-platform/configs"
	_ "github.com/devbenho/luka-platform/docs" // This is required for Swagger
	httpServer "github.com/devbenho/luka-platform/internal/server/http"
	"github.com/devbenho/luka-platform/pkg/database"
	"github.com/devbenho/luka-platform/pkg/validation"
	"go.uber.org/zap"
)

// @title Luka Platform API
// @version 1.0
// @description This is the API documentation for the Luka Platform e-commerce application
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@luka-platform.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:2707
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg, _ := config.LoadConfig()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Cannot initialize logger", err)
	}
	defer logger.Sync()

	db, err := database.NewDatabase(cfg.Database.URI, cfg.Database.Name)
	if err != nil {
		log.Fatal("Cannot connect to database", err)
	}
	validator := validation.NewValidator()
	httpSvr := httpServer.NewServer(validator, db, logger)
	log.Printf("Starting server on port %s", cfg.App.Port)
	if err = httpSvr.Run(); err != nil {
		log.Fatal(err)
	}

	//TODO: add gRPC server
}
