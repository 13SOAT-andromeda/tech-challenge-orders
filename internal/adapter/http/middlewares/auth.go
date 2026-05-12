package middlewares

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	claimsKey    = "userClaims"
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
	UserRoleKey  = "user_role"
)

// AuthRequired returns a Gin middleware that extracts user information
// from X-User-Id, X-User-Email, and X-User-Role headers.
// Returns 401 if any header is missing, empty, or if ID is not a number.
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

		if _, err := strconv.Atoi(userID); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid user ID format"})
			c.Abort()
			return
		}

		c.Set(UserIDKey, userID)
		c.Set(UserEmailKey, userEmail)
		c.Set(UserRoleKey, userRole)

		c.Next()
	}
}
