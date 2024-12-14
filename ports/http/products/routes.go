package products

import (
	configs "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/internal/product/repositories"
	"github.com/devbenho/luka-platform/internal/product/services"
	storeRepo "github.com/devbenho/luka-platform/internal/store/repositories"
	"github.com/devbenho/luka-platform/pkg/database"
	middleware "github.com/devbenho/luka-platform/pkg/middlewares"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.RouterGroup, mongoDb database.IDatabase, validator *validation.Validator, config configs.Config) {
	productRepo := repositories.NewProductRepository(mongoDb)
	storeRepo := storeRepo.NewStoreRepository(mongoDb)
	productSvc := services.NewProductService(productRepo, storeRepo, validator)
	productHandler := NewProductHandler(productSvc)

	productsRoute := r.Group("/products")
	{
		productsRoute.POST("/", middleware.JWTAuth(), productHandler.Create)
		productsRoute.PATCH("/:id", middleware.JWTAuth(), productHandler.Update)
		productsRoute.GET("/:id", middleware.JWTAuth(), productHandler.GetById)
		productsRoute.DELETE("/:id", middleware.JWTAuth(), productHandler.Delete)
	}
}
