package middleware

import (
	"encoding/json"
	"errors"
	"facet-backend/internal/api/response"
	"facet-backend/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestErrorHandler_NotFound tests that a 404 is returned when the error is resource not found
func TestErrorHandler_NotFound(t *testing.T) {
	// --- Arrange ---
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// ***** This is the key fix *****
	// Associate a valid http.Request with the context
	req, _ := http.NewRequest(http.MethodGet, "/test-url", nil)
	c.Request = req
	// *******************************

	c.Error(domain.ErrEnvironmentNotFound)
	handler := ErrorHandler()

	// --- Act ---
	handler(c)

	// --- Assert ---
	assert.Equal(t, http.StatusNotFound, w.Code)
	var apiResp response.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &apiResp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, apiResp.Code)
	assert.Equal(t, domain.ErrEnvironmentNotFound.Error(), apiResp.Message)
}

// TestErrorHandler_InternalServerError tests that a 500 is returned when the error is unknown
func TestErrorHandler_InternalServerError(t *testing.T) {
	// --- Arrange ---
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// ***** This is the key fix *****
	// Associate a valid http.Request with the context
	req, _ := http.NewRequest(http.MethodGet, "/test-url", nil)
	c.Request = req
	// *******************************

	c.Error(errors.New("some unexpected database explosion"))
	handler := ErrorHandler()

	// --- Act ---
	handler(c)

	// --- Assert ---
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var apiResp response.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &apiResp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, apiResp.Code)
	assert.Equal(t, "an internal error occurred", apiResp.Message)
}
