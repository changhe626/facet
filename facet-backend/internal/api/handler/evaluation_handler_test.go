package handler

import (
	"context"
	"encoding/json"
	"errors"
	"facet-backend/internal/api/middleware"
	"facet-backend/internal/api/response"
	"facet-backend/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEvalService struct {
	mock.Mock
}

func (m *mockEvalService) Evaluate(ctx context.Context, sdkKey, featureName, userID, accountID, appName, evalContext string) (bool, error) {
	args := m.Called(ctx, sdkKey, featureName, userID, accountID, appName, evalContext)
	return args.Bool(0), args.Error(1)
}

type mockEventService struct {
	mock.Mock
}

func (m *mockEventService) RecordEvaluationEvent(event *domain.EvaluationEvent) {
	m.Called(event)
}

func (m *mockEventService) Close() error {
	return m.Called().Error(0)
}

func ginTestContext(method, url string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, url, nil)
	return c, w
}

func TestEvaluate_Success(t *testing.T) {
	mockEval := new(mockEvalService)
	mockEvent := new(mockEventService)
	sf, _ := snowflake.NewNode(1)

	handler := NewEvaluationHandler(mockEval, mockEvent, sf)

	mockEval.On("Evaluate", mock.Anything, "sk_test", "flag-a", "user1", "acct1", "myapp", "").
		Return(true, nil)
	mockEvent.On("RecordEvaluationEvent", mock.AnythingOfType("*domain.EvaluationEvent")).Return()

	c, w := ginTestContext("GET", "/evaluate?sdk_key=sk_test&feature_name=flag-a&user_id=user1&account_id=acct1&app_name=myapp")
	handler.Evaluate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 200, resp.Code)
	mockEval.AssertExpectations(t)
	mockEvent.AssertExpectations(t)
}

func TestEvaluate_NotFound(t *testing.T) {
	mockEval := new(mockEvalService)
	mockEvent := new(mockEventService)
	sf, _ := snowflake.NewNode(1)

	handler := NewEvaluationHandler(mockEval, mockEvent, sf)

	mockEval.On("Evaluate", mock.Anything, "sk_bad", "missing", "", "acct1", "myapp", "").
		Return(false, domain.ErrFeatureFlagNotFound)
	mockEvent.On("RecordEvaluationEvent", mock.AnythingOfType("*domain.EvaluationEvent")).Return()

	c, w := ginTestContext("GET", "/evaluate?sdk_key=sk_bad&feature_name=missing&user_id=&account_id=acct1&app_name=myapp")
	handler.Evaluate(c)

	errorHandler := middleware.ErrorHandler()
	errorHandler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var apiResp response.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &apiResp)
	assert.Equal(t, http.StatusNotFound, apiResp.Code)
	assert.Contains(t, apiResp.Message, "feature flag not found")
}

func TestEvaluate_ServiceError(t *testing.T) {
	mockEval := new(mockEvalService)
	mockEvent := new(mockEventService)
	sf, _ := snowflake.NewNode(1)

	handler := NewEvaluationHandler(mockEval, mockEvent, sf)

	mockEval.On("Evaluate", mock.Anything, "sk_err", "flag-x", "u1", "a1", "app", "").
		Return(false, errors.New("internal error"))
	mockEvent.On("RecordEvaluationEvent", mock.AnythingOfType("*domain.EvaluationEvent")).Return()

	c, w := ginTestContext("GET", "/evaluate?sdk_key=sk_err&feature_name=flag-x&user_id=u1&account_id=a1&app_name=app")
	handler.Evaluate(c)

	errorHandler := middleware.ErrorHandler()
	errorHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEvaluate_WithContext(t *testing.T) {
	mockEval := new(mockEvalService)
	mockEvent := new(mockEventService)
	sf, _ := snowflake.NewNode(1)

	handler := NewEvaluationHandler(mockEval, mockEvent, sf)

	ctxJSON := `{"region":"US","plan":"premium"}`
	mockEval.On("Evaluate", mock.Anything, "sk_test", "flag-b", "user2", "acct2", "myapp", ctxJSON).
		Return(true, nil)
	mockEvent.On("RecordEvaluationEvent", mock.AnythingOfType("*domain.EvaluationEvent")).Return()

	c, w := ginTestContext("GET", "/evaluate?sdk_key=sk_test&feature_name=flag-b&user_id=user2&account_id=acct2&app_name=myapp&context="+ctxJSON)
	handler.Evaluate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockEval.AssertExpectations(t)
}
