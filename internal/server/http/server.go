package http

import (
	"fmt"
	"log"
	"net/http"

	config "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/devbenho/luka-platform/pkg/database"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/devbenho/luka-platform/ports/http/categories"
	"github.com/devbenho/luka-platform/ports/http/inventories"
	"github.com/devbenho/luka-platform/ports/http/products"
	"github.com/devbenho/luka-platform/ports/http/stores"
	"github.com/devbenho/luka-platform/ports/http/users"
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine    *gin.Engine
	cfg       *config.Config
	validator *validation.Validator
	db        database.IDatabase
}

func NewServer(validator *validation.Validator, db database.IDatabase) Server {
	cfg, _ := config.LoadConfig()
	return Server{
		engine:    gin.Default(),
		cfg:       cfg,
		validator: validator,
		db:        db,
	}
}

func (s Server) Run() error {
	_ = s.engine.SetTrustedProxies(nil)
	if s.cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := s.MapRoutes(); err != nil {
		log.Fatalf("MapRoutes Error: %v", err)
	}
	s.engine.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(http.StatusOK, "pong", nil))
	})

	if err := s.engine.Run(fmt.Sprintf(":%s", s.cfg.App.Port)); err != nil {
		log.Fatalf("Running HTTP server: %v", err)
	}

	return nil
}

func (s Server) GetEngine() *gin.Engine {
	return s.engine
}

func (s Server) MapRoutes() error {
	v1 := s.engine.Group("/api/v1")
	users.Routes(v1, s.db, s.validator, *s.cfg)
	stores.Routes(v1, s.db, s.validator, *s.cfg)
	categories.Routes(v1, s.db, s.validator, *s.cfg)
	products.Routes(v1, s.db, s.validator, *s.cfg)
	inventories.Routes(v1, s.db, s.validator, *s.cfg)
	return nil
}
