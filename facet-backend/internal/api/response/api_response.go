package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ApiResponse defines the standard JSON response format.
type ApiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // Use omitempty to hide data field when empty
}

// Success sends a standard success response (HTTP 200 OK).
// It wraps business data in the ApiResponse struct.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ApiResponse{
		Code:    http.StatusOK,
		Message: "OK",
		Data:    data,
	})
}

// Error sends a standard error response using an error object.
func Error(c *gin.Context, code int, err error) {
	c.JSON(code, ApiResponse{
		Code:    code,
		Message: err.Error(),
	})
}

// CustomError sends a standard error response with a custom string message.
func CustomError(c *gin.Context, code int, message string) {
	c.JSON(code, ApiResponse{
		Code:    code,
		Message: message,
	})
}
