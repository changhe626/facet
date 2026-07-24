package middleware

import (
	"encoding/json"
	"facet-backend/internal/api/response"
	"facet-backend/pkg/token" // <-- import our own token package
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const testSecret = "test-jwt-secret-key-for-unit-tests" // Define a secret key dedicated for testing

// TestJWTMiddleware_Success tests the success scenario with a valid token
func TestJWTMiddleware_Success(t *testing.T) {
	// --- Arrange ---
	gin.SetMode(gin.TestMode)

	// 1. Prepare a valid Token
	expectedUserID := "user-123"
	expectedTenantID := "tenant-abc"
	expectedRole := "admin"
	validToken, err := token.GenerateToken(expectedUserID, expectedTenantID, expectedRole, testSecret, 1)
	assert.NoError(t, err)

	// 2. Create a response recorder and Gin engine
	w := httptest.NewRecorder()
	r := gin.Default()

	// 3. Define a mock next handler that checks context values
	var contextUserID, contextTenantID, contextRole string
	mockNext := func(c *gin.Context) {
		val, _ := c.Get(string(UserIDKey))
		contextUserID, _ = val.(string)

		val, _ = c.Get(string(TenantIDKey))
		contextTenantID, _ = val.(string)

		val, _ = c.Get(string(UserRoleKey))
		contextRole, _ = val.(string)

		c.Status(http.StatusOK) // Simulate successful execution of next handler
	}

	// 4. Register our middleware and the mock next handler
	r.GET("/test", JWTMiddleware(testSecret), mockNext)

	// 5. Create a mock HTTP request with the correct Authorization header
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	// --- Act ---
	r.ServeHTTP(w, req)

	// --- Assert ---
	// 1. Assert HTTP status code is 200 OK, indicating the request was successfully processed
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Assert the next handler got the correct values from context
	assert.Equal(t, expectedUserID, contextUserID)
	assert.Equal(t, expectedTenantID, contextTenantID)
	assert.Equal(t, expectedRole, contextRole)
}

// TestJWTMiddleware_NoHeader tests the failure scenario with no Authorization header
func TestJWTMiddleware_NoHeader(t *testing.T) {
	// --- Arrange ---
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.Default()
	r.GET("/test", JWTMiddleware(testSecret), func(c *gin.Context) {
		// This handler should not be called
		t.Fatal("next handler should not be called")
	})
	req, _ := http.NewRequest(http.MethodGet, "/test", nil) // No Authorization header

	// --- Act ---
	r.ServeHTTP(w, req)

	// --- Assert ---
	// 1. Assert HTTP status code is 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 2. Assert the response body content
	var apiResp response.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &apiResp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, apiResp.Code)
	assert.Equal(t, "authorization header is required", apiResp.Message)
}

// TestJWTMiddleware_InvalidFormat tests the failure scenario with invalid Authorization header format
func TestJWTMiddleware_InvalidFormat(t *testing.T) {
	// --- Arrange ---
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.Default()
	r.GET("/test", JWTMiddleware(testSecret), func(c *gin.Context) {
		t.Fatal("next handler should not be called")
	})
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "NotBearer token") // Invalid format

	// --- Act ---
	r.ServeHTTP(w, req)

	// --- Assert ---
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var apiResp response.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &apiResp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid authorization header format", apiResp.Message)
}

// TestJWTMiddleware_InvalidToken tests the failure scenario with an invalid token
func TestJWTMiddleware_InvalidToken(t *testing.T) {
	// --- Arrange ---
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.Default()
	r.GET("/test", JWTMiddleware(testSecret), func(c *gin.Context) {
		t.Fatal("next handler should not be called")
	})
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer this-is-not-a-valid-jwt") // Invalid token

	// --- Act ---
	r.ServeHTTP(w, req)

	// --- Assert ---
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var apiResp response.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &apiResp)
	assert.NoError(t, err)
	assert.Contains(t, apiResp.Message, "invalid or expired token") // Check for "invalid or expired token" message
}
