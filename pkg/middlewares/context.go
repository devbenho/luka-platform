package middleware

import "github.com/gin-gonic/gin"

// UserID returns the authenticated user's ID set by the JWT middleware.
func UserID(c *gin.Context) (string, bool) {
	value, exists := c.Get("userId")
	if !exists {
		return "", false
	}
	id, ok := value.(string)
	return id, ok && id != ""
}

// Role returns the authenticated user's role set by the JWT middleware.
func Role(c *gin.Context) string {
	role, _ := c.Get("role")
	s, _ := role.(string)
	return s
}
