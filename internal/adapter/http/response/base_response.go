package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BaseResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

func RespondSuccess[T any](c *gin.Context, data T, message string) {

	switch v := any(data).(type) {
	case []any:
		if v == nil {
			data = any([]any{}).(T)
		}
	}

	c.JSON(200, BaseResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func RespondCreated[T any](c *gin.Context, data T, message string) {
	c.JSON(http.StatusCreated, BaseResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, BaseResponse[any]{
		Success: false,
		Message: message,
	})
}
