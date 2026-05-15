package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
	UserRoleKey  = "user_role"
)

// AuthRequired reads X-User-Id, X-User-Email and X-User-Role injected by the
// upstream API Gateway + Lambda Authorizer. Returns 401 if any header is absent.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-Id")
		userEmail := c.GetHeader("X-User-Email")
		userRole := c.GetHeader("X-User-Role")

		if userID == "" || userEmail == "" || userRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing required user headers"})
			c.Abort()
			return
		}

		c.Set(UserIDKey, userID)
		c.Set(UserEmailKey, userEmail)
		c.Set(UserRoleKey, userRole)

		c.Next()
	}
}
