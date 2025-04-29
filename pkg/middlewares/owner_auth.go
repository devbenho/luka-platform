package middleware

import (
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/gin-gonic/gin"
)

// OwnerAuth checks if the authenticated user is the owner of the resource
func OwnerAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authenticated user's ID
		authUserID, exists := c.Get("userId")
		if !exists {
			c.JSON(401, utils.NewUnauthorizedResponse("User not authenticated"))
			c.Abort()
			return
		}

		// Get the resource ID from the URL parameter
		resourceID := c.Param("id")
		if resourceID == "" {
			c.JSON(400, utils.NewErrorResponse(400, "Bad Request", "Resource ID is required"))
			c.Abort()
			return
		}

		// Check if the authenticated user is the owner of the resource
		if authUserID != resourceID {
			c.JSON(403, utils.NewForbiddenResponse("You can only modify your own data"))
			c.Abort()
			return
		}

		c.Next()
	}
}
