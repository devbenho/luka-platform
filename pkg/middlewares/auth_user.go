package middleware

import (
	"log"
	"net/http"

	config "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/pkg/tokens"
	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return JWT(config.GetConfig().JWT.Type)
}

func JWT(tokenType string) gin.HandlerFunc {

	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}
		log.Print(token)
		payload, err := tokens.ValidateToken(token)

		if err != nil {
			c.JSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}
		c.Set("userId", payload["Id"])
		c.Set("role", payload["Role"])
		c.Set("username", payload["username"])
		c.Next()
	}
}
