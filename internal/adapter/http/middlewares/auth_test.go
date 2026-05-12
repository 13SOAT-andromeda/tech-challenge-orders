package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/http/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRefactoredAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Note: This will not compile initially because AuthRequired still expects a string argument
	r.Use(middlewares.AuthRequired()) 
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get(middlewares.UserIDKey)
		email, _ := c.Get(middlewares.UserEmailKey)
		role, _ := c.Get(middlewares.UserRoleKey)
		c.JSON(http.StatusOK, gin.H{
			"id":    id,
			"email": email,
			"role":  role,
		})
	})
	return r
}

func TestAuthRequired_Refactor(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Success - All Valid Headers",
			headers: map[string]string{
				"X-User-Id":    "123",
				"X-User-Email": "test@example.com",
				"X-User-Role":  "admin",
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":    "123",
				"email": "test@example.com",
				"role":  "admin",
			},
		},
		{
			name: "Failure - Missing X-User-Id",
			headers: map[string]string{
				"X-User-Email": "test@example.com",
				"X-User-Role":  "admin",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Failure - Empty X-User-Id",
			headers: map[string]string{
				"X-User-Id":    "",
				"X-User-Email": "test@example.com",
				"X-User-Role":  "admin",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Failure - Non-Numeric X-User-Id",
			headers: map[string]string{
				"X-User-Id":    "abc",
				"X-User-Email": "test@example.com",
				"X-User-Role":  "admin",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Failure - Missing X-User-Email",
			headers: map[string]string{
				"X-User-Id":    "123",
				"X-User-Role":  "admin",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Failure - Missing X-User-Role",
			headers: map[string]string{
				"X-User-Id":    "123",
				"X-User-Email": "test@example.com",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupRefactoredAuthRouter()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				for k, v := range tt.expectedBody {
					assert.Contains(t, w.Body.String(), v.(string))
					assert.Contains(t, w.Body.String(), k)
				}
			}
		})
	}
}
