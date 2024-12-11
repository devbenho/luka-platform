package inventories

import (
	configs "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/internal/inventory/repositories"
	"github.com/devbenho/luka-platform/internal/inventory/services"
	"github.com/devbenho/luka-platform/pkg/database"
	middleware "github.com/devbenho/luka-platform/pkg/middlewares"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.RouterGroup, mongoDb database.IDatabase, validator *validation.Validator, config configs.Config) {
	inventoryRepo := repositories.NewInventoryRepository(mongoDb)
	inventorySvc := services.NewInventoryService(inventoryRepo, validator)
	inventoryHandler := NewInventoryHandler(inventorySvc)

	inventoriesRoute := r.Group("/inventories")
	{
		inventoriesRoute.POST("/", middleware.JWTAuth(), inventoryHandler.Create)
		inventoriesRoute.PATCH("/:id", middleware.JWTAuth(), inventoryHandler.Update)
		inventoriesRoute.GET("/:id", middleware.JWTAuth(), inventoryHandler.GetById)
		inventoriesRoute.DELETE("/:id", middleware.JWTAuth(), inventoryHandler.Delete)
	}
}
