package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Login(ctx context.Context, req request.LoginRequest) (*response.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.LoginResponse), args.Error(1)
}

type mockCaptchaService struct {
	mock.Mock
}

func (m *mockCaptchaService) Generate() (*service.CaptchaResponse, error) {
	args := m.Called()
	return args.Get(0).(*service.CaptchaResponse), args.Error(1)
}

func (m *mockCaptchaService) Verify(id, answer string) bool {
	args := m.Called(id, answer)
	return args.Bool(0)
}

func setupAuthHandler() (*mockAuthService, *mockCaptchaService, *AuthHandler) {
	mockAuth := new(mockAuthService)
	mockCaptcha := new(mockCaptchaService)
	return mockAuth, mockCaptcha, NewAuthHandler(mockAuth, mockCaptcha)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockAuth, mockCaptcha, handler := setupAuthHandler()

	reqBody := request.LoginRequest{
		Email:         "admin@test.com",
		Password:      "password123",
		CaptchaID:     "captcha-1",
		CaptchaAnswer: "1234",
	}

	mockCaptcha.On("Verify", "captcha-1", "1234").Return(true)
	mockAuth.On("Login", mock.Anything, reqBody).
		Return(&response.LoginResponse{
			Token: "jwt-token",
			User:  response.UserInfo{ID: "u-1", DisplayName: "Admin", Role: "system_admin"},
		}, nil)

	body, _ := json.Marshal(reqBody)
	c, w := ginTestContext("POST", "/api/v1/manage/auth/login")
	c.Request = httptest.NewRequest("POST", "/api/v1/manage/auth/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockCaptcha.AssertExpectations(t)
	mockAuth.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidFormat(t *testing.T) {
	_, _, handler := setupAuthHandler()

	c, w := ginTestContext("POST", "/api/v1/manage/auth/login")
	c.Request = httptest.NewRequest("POST", "/api/v1/manage/auth/login", bytes.NewReader([]byte(`not-json`)))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InvalidCaptcha(t *testing.T) {
	mockAuth, mockCaptcha, handler := setupAuthHandler()

	reqBody := request.LoginRequest{
		Email:         "admin@test.com",
		Password:      "password123",
		CaptchaID:     "captcha-1",
		CaptchaAnswer: "wrong",
	}

	mockCaptcha.On("Verify", "captcha-1", "wrong").Return(false)

	body, _ := json.Marshal(reqBody)
	c, w := ginTestContext("POST", "/api/v1/manage/auth/login")
	c.Request = httptest.NewRequest("POST", "/api/v1/manage/auth/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockCaptcha.AssertExpectations(t)
	mockAuth.AssertNotCalled(t, "Login")
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockAuth, mockCaptcha, handler := setupAuthHandler()

	reqBody := request.LoginRequest{
		Email:         "bad@test.com",
		Password:      "wrong",
		CaptchaID:     "captcha-1",
		CaptchaAnswer: "1234",
	}

	mockCaptcha.On("Verify", "captcha-1", "1234").Return(true)
	mockAuth.On("Login", mock.Anything, reqBody).Return(nil, service.ErrInvalidCredentials)

	body, _ := json.Marshal(reqBody)
	c, w := ginTestContext("POST", "/api/v1/manage/auth/login")
	c.Request = httptest.NewRequest("POST", "/api/v1/manage/auth/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockCaptcha.AssertExpectations(t)
	mockAuth.AssertExpectations(t)
}
