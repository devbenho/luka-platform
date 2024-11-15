package users

import (
	configs "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/internal/user/repositories"
	"github.com/devbenho/luka-platform/internal/user/services"
	"github.com/devbenho/luka-platform/pkg/database"
	"github.com/devbenho/luka-platform/pkg/hasher"
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
}
