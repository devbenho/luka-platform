package main

import (
	"log"

	config "github.com/devbenho/luka-platform/configs"
	httpServer "github.com/devbenho/luka-platform/internal/server/http"
	"github.com/devbenho/luka-platform/pkg/database"
	"github.com/devbenho/luka-platform/pkg/validation"
)

//	@title			GoShop Swagger API
//	@version		1.0
//	@description	Swagger API for GoShop.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Quang Dang
//	@contact.email	quangdangfit@gmail.com

//	@license.name	MIT
//	@license.url	https://github.com/MartinHeinz/go-project-blueprint/blob/master/LICENSE

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization

//	@BasePath	/api/v1

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
