package categories

import (
	configs "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/internal/category/repositories"
	"github.com/devbenho/luka-platform/internal/category/services"
	"github.com/devbenho/luka-platform/pkg/database"
	middleware "github.com/devbenho/luka-platform/pkg/middlewares"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.RouterGroup, mongoDb database.IDatabase, validator *validation.Validator, config configs.Config) {
	categoryRepo := repositories.NewCategoryRepository(mongoDb)
	categorySvc := services.NewCategoryService(categoryRepo, validator)
	categoryHandler := NewCategoryHandler(categorySvc)

	categoriesRoute := r.Group("/categories")
	{
		categoriesRoute.POST("/", middleware.JWTAuth(), categoryHandler.Create)
		categoriesRoute.PATCH("/:id", middleware.JWTAuth(), categoryHandler.Update)
		categoriesRoute.GET("/:id", middleware.JWTAuth(), categoryHandler.GetById)
		categoriesRoute.DELETE("/:id", middleware.JWTAuth(), categoryHandler.Delete)
	}
}
