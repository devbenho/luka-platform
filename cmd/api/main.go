package main

import (
	"log"

	config "github.com/devbenho/luka-platform/configs"
	httpServer "github.com/devbenho/luka-platform/internal/server/http"
	"github.com/devbenho/luka-platform/pkg/database"
	"github.com/devbenho/luka-platform/pkg/validation"
)

func main() {
	cfg, _ := config.LoadConfig()

	db, err := database.NewDatabase(cfg.Database.URI, cfg.Database.Name)
	if err != nil {
		log.Fatal("Cannot connect to database", err)
	}
	validator := validation.NewValidator()
	httpSvr := httpServer.NewServer(validator, db)
	log.Printf("Starting server on port %s", cfg.App.Port)
	if err = httpSvr.Run(); err != nil {
		log.Fatal(err)
	}

	//TODO: add gRPC server
}
