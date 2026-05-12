package middlewares

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserClaims holds the identity information extracted from the headers.
type UserClaims struct {
	ID    string
	Email string
	Role  string
}

// ExtractClaims reads user identity from the Gin context (set by AuthRequired).
// Returns an error if the claims are not present or if types are invalid.
func ExtractClaims(c *gin.Context) (*UserClaims, error) {
	id, idExists := c.Get(UserIDKey)
	email, emailExists := c.Get(UserEmailKey)
	role, roleExists := c.Get(UserRoleKey)

	if !idExists || !emailExists || !roleExists {
		return nil, errors.New("missing user claims in context")
	}

	idStr, okID := id.(string)
	emailStr, okEmail := email.(string)
	roleStr, okRole := role.(string)

	if !okID || !okEmail || !okRole {
		return nil, errors.New("invalid user claims type in context")
	}

	return &UserClaims{
		ID:    idStr,
		Email: emailStr,
		Role:  roleStr,
	}, nil
}

// RoleRequired returns a middleware that grants access if the request carries
// one of the required roles.
// Returns 401 if claims are missing from context, 403 if the role is insufficient.
func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := ExtractClaims(c)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing identity claims"})
			c.Abort()
			return
		}

		hasRole := false

		for _, role := range roles {
			if claims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID reads the user ID from the Gin context.
func GetUserID(c *gin.Context) string {
	id, exists := c.Get(UserIDKey)
	if !exists {
		return ""
	}
	idStr, _ := id.(string)
	return idStr
}

// GetUserEmail reads the user email from the Gin context.
func GetUserEmail(c *gin.Context) string {
	email, exists := c.Get(UserEmailKey)
	if !exists {
		return ""
	}
	emailStr, _ := email.(string)
	return emailStr
}

// GetUserRole reads the user role from the Gin context.
func GetUserRole(c *gin.Context) string {
	role, exists := c.Get(UserRoleKey)
	if !exists {
		return ""
	}
	roleStr, _ := role.(string)
	return roleStr
}
