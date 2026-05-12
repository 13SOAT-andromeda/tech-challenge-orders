package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/http/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setAuthHeaders(req *http.Request, id, email, role string) {
	req.Header.Set("X-User-Id", id)
	req.Header.Set("X-User-Email", email)
	req.Header.Set("X-User-Role", role)
}

func setupRoleRouter(requiredRoles ...string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.AuthRequired())
	r.GET("/test", middlewares.RoleRequired(requiredRoles...), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestRoleRequired_MatchingRole(t *testing.T) {
	r := setupRoleRouter("mechanic")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	setAuthHeaders(req, "1", "test@example.com", "mechanic")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleRequired_MultipleRoles(t *testing.T) {
	r := setupRoleRouter("mechanic", "attendant")
	
	t.Run("Matches first role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		setAuthHeaders(req, "1", "test@example.com", "mechanic")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Matches second role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		setAuthHeaders(req, "1", "test@example.com", "attendant")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRoleRequired_InsufficientRole(t *testing.T) {
	r := setupRoleRouter("administrator")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	setAuthHeaders(req, "1", "test@example.com", "mechanic")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRoleRequired_MissingClaimsInContext(t *testing.T) {
	// No AuthRequired middleware — context will have no claims.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", middlewares.RoleRequired("mechanic"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExtractClaims_ReturnsCorrectValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.AuthRequired())
	r.GET("/test", func(c *gin.Context) {
		claims, err := middlewares.ExtractClaims(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": claims.ID, "role": claims.Role, "email": claims.Email})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	setAuthHeaders(req, "99", "admin@example.com", "administrator")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"id":"99"`)
	assert.Contains(t, body, `"role":"administrator"`)
	assert.Contains(t, body, `"email":"admin@example.com"`)
}
