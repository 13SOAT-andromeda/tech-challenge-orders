package middlewares

import (
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/clients"
	"github.com/gin-gonic/gin"
)

// InjectUserContext propagates the authenticated user values set by AuthRequired
// into the request's context.Context so outbound HTTP clients can forward them.
func InjectUserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get(UserIDKey)
		userEmail, _ := c.Get(UserEmailKey)
		userRole, _ := c.Get(UserRoleKey)

		ctx := clients.WithUserHeaders(
			c.Request.Context(),
			userID.(string),
			userEmail.(string),
			userRole.(string),
		)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
