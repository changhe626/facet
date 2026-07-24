package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"facet-backend/internal/api/middleware"
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/domain"
	"facet-backend/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEnvironmentService struct {
	mock.Mock
}

func (m *mockEnvironmentService) List(ctx context.Context, tenantID string, req request.ListEnvironmentsRequest) (*response.ListEnvironmentsResponse, error) {
	args := m.Called(ctx, tenantID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ListEnvironmentsResponse), args.Error(1)
}

func (m *mockEnvironmentService) ListActive(ctx context.Context, tenantID string) ([]*response.ActiveEnvironmentListItem, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*response.ActiveEnvironmentListItem), args.Error(1)
}

func (m *mockEnvironmentService) Create(ctx context.Context, tenantID, creatorID string, req request.CreateEnvironmentRequest) (*response.CreateEnvironmentResponse, error) {
	args := m.Called(ctx, tenantID, creatorID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CreateEnvironmentResponse), args.Error(1)
}

func (m *mockEnvironmentService) Update(ctx context.Context, tenantID, envID, updaterID string, req request.UpdateEnvironmentRequest) (*response.UpdateEnvironmentResponse, error) {
	args := m.Called(ctx, tenantID, envID, updaterID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UpdateEnvironmentResponse), args.Error(1)
}

func (m *mockEnvironmentService) Delete(ctx context.Context, tenantID string, envID string) error {
	args := m.Called(ctx, tenantID, envID)
	return args.Error(0)
}

func setupEnvHandler() (*mockEnvironmentService, *EnvironmentHandler) {
	mockSvc := new(mockEnvironmentService)
	return mockSvc, NewEnvironmentHandler(mockSvc)
}

func TestEnvironmentHandler_ListActive_Success(t *testing.T) {
	mockSvc, handler := setupEnvHandler()

	mockSvc.On("ListActive", mock.Anything, "tenant-1").
		Return([]*response.ActiveEnvironmentListItem{
			{ID: "env-1", Name: "Production"},
		}, nil)

	c, w := ginTestContext("GET", "/api/v1/manage/environments/active")
	setContextAuth(c)
	handler.ListActive(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestEnvironmentHandler_List_Success(t *testing.T) {
	mockSvc, handler := setupEnvHandler()

	mockSvc.On("List", mock.Anything, "tenant-1", request.ListEnvironmentsRequest{Page: 1, PageSize: 10}).
		Return(&response.ListEnvironmentsResponse{
			Page: 1, PageSize: 10, Total: 1,
			Items: []response.EnvironmentListItem{{ID: "env-1", Name: "Prod", IsActive: true}},
		}, nil)

	c, w := ginTestContext("GET", "/api/v1/manage/environments")
	setContextAuth(c)
	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestEnvironmentHandler_Create_Success(t *testing.T) {
	mockSvc, handler := setupEnvHandler()

	reqBody := request.CreateEnvironmentRequest{Name: "Staging", Description: "staging env"}
	mockSvc.On("Create", mock.Anything, "tenant-1", "user-1", reqBody).
		Return(&response.CreateEnvironmentResponse{ID: "env-new", Name: "Staging", SDKKey: "sk_xxx", IsActive: true}, nil)

	body, _ := json.Marshal(reqBody)
	c, w := ginTestContext("POST", "/api/v1/manage/environments")
	c.Request = httptest.NewRequest("POST", "/api/v1/manage/environments", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setContextAuth(c)

	handler.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestEnvironmentHandler_Create_BindingError(t *testing.T) {
	_, handler := setupEnvHandler()

	c, w := ginTestContext("POST", "/api/v1/manage/environments")
	c.Request = httptest.NewRequest("POST", "/api/v1/manage/environments", bytes.NewReader([]byte(`{"name":""}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	setContextAuth(c)

	handler.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEnvironmentHandler_Create_QuotaExceeded(t *testing.T) {
	mockSvc, handler := setupEnvHandler()

	reqBody := request.CreateEnvironmentRequest{Name: "Extra"}
	mockSvc.On("Create", mock.Anything, "tenant-1", "user-1", reqBody).
		Return(nil, service.ErrQuotaExceededEnvironments)

	body, _ := json.Marshal(reqBody)
	c, w := ginTestContext("POST", "/api/v1/manage/environments")
	c.Request = httptest.NewRequest("POST", "/api/v1/manage/environments", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setContextAuth(c)

	handler.Create(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestEnvironmentHandler_Create_NameExists(t *testing.T) {
	mockSvc, handler := setupEnvHandler()

	reqBody := request.CreateEnvironmentRequest{Name: "Duplicate"}
	mockSvc.On("Create", mock.Anything, "tenant-1", "user-1", reqBody).
		Return(nil, domain.ErrEnvironmentNameExists)

	body, _ := json.Marshal(reqBody)
	c, w := ginTestContext("POST", "/api/v1/manage/environments")
	c.Request = httptest.NewRequest("POST", "/api/v1/manage/environments", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setContextAuth(c)

	handler.Create(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestEnvironmentHandler_Update_Success(t *testing.T) {
	mockSvc, handler := setupEnvHandler()

	active := true
	reqBody := request.UpdateEnvironmentRequest{Name: "Updated", Description: "desc", IsActive: &active}
	mockSvc.On("Update", mock.Anything, "tenant-1", "env-1", "user-1", reqBody).
		Return(&response.UpdateEnvironmentResponse{ID: "env-1", Name: "Updated"}, nil)

	body, _ := json.Marshal(reqBody)
	c, w := ginTestContext("PUT", "/api/v1/manage/environments/env-1")
	c.Params = []gin.Param{{Key: "envId", Value: "env-1"}}
	c.Request = httptest.NewRequest("PUT", "/api/v1/manage/environments/env-1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setContextAuth(c)

	handler.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestEnvironmentHandler_Delete_Success(t *testing.T) {
	mockSvc, handler := setupEnvHandler()

	mockSvc.On("Delete", mock.Anything, "tenant-1", "env-1").Return(nil)

	c, w := ginTestContext("DELETE", "/api/v1/manage/environments/env-1")
	c.Params = []gin.Param{{Key: "envId", Value: "env-1"}}
	setContextAuth(c)

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestEnvironmentHandler_Delete_NotFound(t *testing.T) {
	mockSvc, handler := setupEnvHandler()

	mockSvc.On("Delete", mock.Anything, "tenant-1", "env-missing").Return(domain.ErrEnvironmentNotFound)

	c, w := ginTestContext("DELETE", "/api/v1/manage/environments/env-missing")
	c.Params = []gin.Param{{Key: "envId", Value: "env-missing"}}
	setContextAuth(c)

	handler.Delete(c)

	errorHandler := middleware.ErrorHandler()
	errorHandler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}
