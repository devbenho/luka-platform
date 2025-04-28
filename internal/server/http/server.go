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
	"github.com/devbenho/luka-platform/ports/http/orders"
	"github.com/devbenho/luka-platform/ports/http/products"
	"github.com/devbenho/luka-platform/ports/http/stores"
	"github.com/devbenho/luka-platform/ports/http/users"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type Server struct {
	engine    *gin.Engine
	cfg       *config.Config
	validator *validation.Validator
	db        database.IDatabase
	logger    *zap.Logger
}

func NewServer(validator *validation.Validator, db database.IDatabase, logger *zap.Logger) Server {
	cfg, _ := config.LoadConfig()
	router := gin.Default()

	return Server{
		engine:    router,
		cfg:       cfg,
		validator: validator,
		db:        db,
		logger:    logger,
	}
}

func (s Server) Run() error {
	_ = s.engine.SetTrustedProxies(nil)
	if s.cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Add Swagger UI with custom configuration
	swaggerURL := fmt.Sprintf("http://localhost:%s/swagger/doc.json", s.cfg.App.Port)

	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, func(c *ginSwagger.Config) {
		c.URL = swaggerURL
	}))

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
	orders.Routes(v1, s.db, s.validator, *s.cfg)
	return nil
}
