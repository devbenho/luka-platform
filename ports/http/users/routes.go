package users

import (
	configs "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/internal/user/repositories"
	"github.com/devbenho/luka-platform/internal/user/services"
	"github.com/devbenho/luka-platform/pkg/database"
	"github.com/devbenho/luka-platform/pkg/hasher"
	"github.com/devbenho/luka-platform/pkg/tokens"
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.RouterGroup, mongoDb database.IDatabase, config configs.Config) {
	userRepo := repositories.NewUserRepository(mongoDb)
	userSvc := services.NewUserService(tokens.NewTokenService(config.JWT.Secret), userRepo, hasher.NewHasher())
	userHandler := NewUserHandler(userSvc)

	authRoute := r.Group("/auth")
	{
		authRoute.POST("/register", func(c *gin.Context) {
			if err := userHandler.Register(c); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
			}
		})
		// authRoute.POST("/login", userHandler.Login)
	}
}
