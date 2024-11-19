package stores

import (
	configs "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/internal/store/repositories"
	"github.com/devbenho/luka-platform/internal/store/services"
	"github.com/devbenho/luka-platform/pkg/database"
	middleware "github.com/devbenho/luka-platform/pkg/middlewares"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.RouterGroup, mongoDb database.IDatabase, validator *validation.Validator, config configs.Config) {
	storeRepo := repositories.NewStoreRepository(mongoDb)
	storeSvc := services.NewStoreService(storeRepo, validator)
	storeHandler := NewStoreHandler(storeSvc)

	authRoute := r.Group("/stores")
	{
		authRoute.POST("/", middleware.JWTAuth(), storeHandler.Create)
	}
}
