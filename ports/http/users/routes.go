package users

import (
	configs "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/internal/user/repositories"
	"github.com/devbenho/luka-platform/internal/user/services"
	"github.com/devbenho/luka-platform/pkg/database"
	"github.com/devbenho/luka-platform/pkg/hasher"
	middleware "github.com/devbenho/luka-platform/pkg/middlewares"
	"github.com/devbenho/luka-platform/pkg/tokens"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.RouterGroup, mongoDb database.IDatabase, validator *validation.Validator, config configs.Config) {
	userRepo := repositories.NewUserRepository(mongoDb)
	userSvc := services.NewUserService(validator, tokens.NewTokenService(config.JWT.Secret), userRepo, hasher.NewHasher())
	userHandler := NewUserHandler(userSvc)

	authRoute := r.Group("/auth")
	{
		authRoute.POST("/register", userHandler.Register)
		authRoute.POST("/login", userHandler.Login)
	}
	userRoute := r.Group("/users")
	{
		userRoute.GET("/:id", middleware.JWTAuth(), userHandler.GetUserByID)
		userRoute.PUT("/:id", middleware.JWTAuth(), middleware.OwnerAuth(), userHandler.UpdateUser)
		userRoute.DELETE("/:id", middleware.JWTAuth(), middleware.OwnerAuth(), userHandler.DeleteUser)
	}
}
