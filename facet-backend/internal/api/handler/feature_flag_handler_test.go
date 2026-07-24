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

type mockFeatureFlagService struct {
	mock.Mock
}

func (m *mockFeatureFlagService) List(ctx context.Context, envID string, req request.ListFeatureFlagsRequest) (*response.ListFeatureFlagsResponse, error) {
	args := m.Called(ctx, envID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ListFeatureFlagsResponse), args.Error(1)
}

func (m *mockFeatureFlagService) Create(ctx context.Context, tenantID, envID, userID string, req request.CreateFeatureFlagRequest) (*response.CreateFeatureFlagResponse, error) {
	args := m.Called(ctx, tenantID, envID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CreateFeatureFlagResponse), args.Error(1)
}

func (m *mockFeatureFlagService) GetByID(ctx context.Context, envID, id string) (*response.FeatureFlagDetailResponse, error) {
	args := m.Called(ctx, envID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FeatureFlagDetailResponse), args.Error(1)
}

func (m *mockFeatureFlagService) Update(ctx context.Context, envID, userID, id string, req request.UpdateFeatureFlagRequest) (*response.UpdateFeatureFlagResponse, error) {
	args := m.Called(ctx, envID, userID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UpdateFeatureFlagResponse), args.Error(1)
}

func (m *mockFeatureFlagService) Delete(ctx context.Context, envID, id, userID string) error {
	args := m.Called(ctx, envID, id, userID)
	return args.Error(0)
}

func (m *mockFeatureFlagService) GetHistory(ctx context.Context, envID, flagID string) ([]*response.FeatureFlagSnapshotResponse, error) {
	args := m.Called(ctx, envID, flagID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*response.FeatureFlagSnapshotResponse), args.Error(1)
}

func (m *mockFeatureFlagService) Archive(ctx context.Context, envID, id, userID string) error {
	args := m.Called(ctx, envID, id, userID)
	return args.Error(0)
}

func (m *mockFeatureFlagService) Unarchive(ctx context.Context, envID, id, userID string) error {
	args := m.Called(ctx, envID, id, userID)
	return args.Error(0)
}

func setupFFHandler() (*mockFeatureFlagService, *FeatureFlagHandler) {
	mockSvc := new(mockFeatureFlagService)
	return mockSvc, NewFeatureFlagHandler(mockSvc)
}

func setContextAuth(c *gin.Context) {
	c.Set(string(middleware.TenantIDKey), "tenant-1")
	c.Set(string(middleware.UserIDKey), "user-1")
}

func TestFeatureFlagHandler_List(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	mockSvc.On("List", mock.Anything, "env-1", request.ListFeatureFlagsRequest{Page: 1, PageSize: 10}).
		Return(&response.ListFeatureFlagsResponse{
			Page: 1, PageSize: 10, Total: 0, Items: []response.FeatureFlagListItem{},
		}, nil)

	c, w := ginTestContext("GET", "/api/v1/manage/environments/env-1/features")
	c.Params = []gin.Param{{Key: "envId", Value: "env-1"}}
	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_List_WithKeyword(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	expectedReq := request.ListFeatureFlagsRequest{Page: 1, PageSize: 10, Keyword: "test-flag"}
	mockSvc.On("List", mock.Anything, "env-1", expectedReq).
		Return(&response.ListFeatureFlagsResponse{
			Page: 1, PageSize: 10, Total: 1,
			Items: []response.FeatureFlagListItem{
				{ID: "ff-1", Name: "test-flag", Enabled: true},
			},
		}, nil)

	c, w := ginTestContext("GET", "/api/v1/manage/environments/env-1/features?keyword=test-flag")
	c.Params = []gin.Param{{Key: "envId", Value: "env-1"}}
	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 200, resp.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_Create_Success(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	reqBody := request.CreateFeatureFlagRequest{
		Name:              "new-flag",
		FlagType:          "boolean",
		Enabled:           true,
		RolloutPercentage: 50,
	}

	mockSvc.On("Create", mock.Anything, "tenant-1", "env-1", "user-1", reqBody).
		Return(&response.CreateFeatureFlagResponse{
			ID: "ff-new", Name: "new-flag", Enabled: true,
		}, nil)

	body, _ := json.Marshal(reqBody)
	c, w := ginTestContext("POST", "/api/v1/manage/environments/env-1/features")
	c.Params = []gin.Param{{Key: "envId", Value: "env-1"}}
	c.Request = httptest.NewRequest("POST", "/api/v1/manage/environments/env-1/features", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setContextAuth(c)

	handler.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_Create_BindingError(t *testing.T) {
	_, handler := setupFFHandler()

	c, w := ginTestContext("POST", "/api/v1/manage/environments/env-1/features")
	c.Params = []gin.Param{{Key: "envId", Value: "env-1"}}
	c.Request = httptest.NewRequest("POST", "/api/v1/manage/environments/env-1/features", bytes.NewReader([]byte(`{"name":""}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	setContextAuth(c)

	handler.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFeatureFlagHandler_Create_QuotaExceeded(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	reqBody := request.CreateFeatureFlagRequest{
		Name:     "another-flag",
		FlagType: "boolean",
	}

	mockSvc.On("Create", mock.Anything, "tenant-1", "env-1", "user-1", reqBody).
		Return(nil, service.ErrQuotaExceededFeatureFlags)

	body, _ := json.Marshal(reqBody)
	c, w := ginTestContext("POST", "/api/v1/manage/environments/env-1/features")
	c.Params = []gin.Param{{Key: "envId", Value: "env-1"}}
	c.Request = httptest.NewRequest("POST", "/api/v1/manage/environments/env-1/features", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setContextAuth(c)

	handler.Create(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_GetByID_Success(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	mockSvc.On("GetByID", mock.Anything, "env-1", "ff-1").
		Return(&response.FeatureFlagDetailResponse{ID: "ff-1", Name: "test-flag"}, nil)

	c, w := ginTestContext("GET", "/api/v1/manage/environments/env-1/features/ff-1")
	c.Params = []gin.Param{
		{Key: "envId", Value: "env-1"},
		{Key: "featureId", Value: "ff-1"},
	}
	handler.GetByID(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_GetByID_NotFound(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	mockSvc.On("GetByID", mock.Anything, "env-1", "ff-missing").
		Return(nil, domain.ErrFeatureFlagNotFound)

	c, w := ginTestContext("GET", "/api/v1/manage/environments/env-1/features/ff-missing")
	c.Params = []gin.Param{
		{Key: "envId", Value: "env-1"},
		{Key: "featureId", Value: "ff-missing"},
	}
	handler.GetByID(c)

	errorHandler := middleware.ErrorHandler()
	errorHandler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_Update_Success(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	reqBody := request.UpdateFeatureFlagRequest{
		Enabled:           true,
		FlagType:          "boolean",
		RolloutPercentage: 80,
	}
	desc := "updated desc"
	reqBody.Description = &desc

	mockSvc.On("Update", mock.Anything, "env-1", "user-1", "ff-1", reqBody).
		Return(&response.UpdateFeatureFlagResponse{ID: "ff-1", Name: "test-flag", Enabled: true}, nil)

	body, _ := json.Marshal(reqBody)
	c, w := ginTestContext("PUT", "/api/v1/manage/environments/env-1/features/ff-1")
	c.Request = httptest.NewRequest("PUT", "/api/v1/manage/environments/env-1/features/ff-1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{
		{Key: "envId", Value: "env-1"},
		{Key: "featureId", Value: "ff-1"},
	}
	setContextAuth(c)

	handler.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_Delete_Success(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	mockSvc.On("Delete", mock.Anything, "env-1", "ff-1", "user-1").Return(nil)

	c, w := ginTestContext("DELETE", "/api/v1/manage/environments/env-1/features/ff-1")
	c.Params = []gin.Param{
		{Key: "envId", Value: "env-1"},
		{Key: "featureId", Value: "ff-1"},
	}
	setContextAuth(c)

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_Delete_NotFound(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	mockSvc.On("Delete", mock.Anything, "env-1", "ff-missing", "user-1").
		Return(domain.ErrFeatureFlagNotFound)

	c, w := ginTestContext("DELETE", "/api/v1/manage/environments/env-1/features/ff-missing")
	c.Params = []gin.Param{
		{Key: "envId", Value: "env-1"},
		{Key: "featureId", Value: "ff-missing"},
	}
	setContextAuth(c)

	handler.Delete(c)

	errorHandler := middleware.ErrorHandler()
	errorHandler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_Archive_Success(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	mockSvc.On("Archive", mock.Anything, "env-1", "ff-1", "user-1").Return(nil)

	c, w := ginTestContext("POST", "/api/v1/manage/environments/env-1/features/ff-1/archive")
	c.Params = []gin.Param{
		{Key: "envId", Value: "env-1"},
		{Key: "featureId", Value: "ff-1"},
	}
	c.Set("userID", "user-1")

	handler.Archive(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_Unarchive_Success(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	mockSvc.On("Unarchive", mock.Anything, "env-1", "ff-1", "user-1").Return(nil)

	c, w := ginTestContext("POST", "/api/v1/manage/environments/env-1/features/ff-1/unarchive")
	c.Params = []gin.Param{
		{Key: "envId", Value: "env-1"},
		{Key: "featureId", Value: "ff-1"},
	}
	c.Set("userID", "user-1")

	handler.Unarchive(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_GetHistory(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	mockSvc.On("GetHistory", mock.Anything, "env-1", "ff-1").
		Return([]*response.FeatureFlagSnapshotResponse{}, nil)

	c, w := ginTestContext("GET", "/api/v1/manage/environments/env-1/features/ff-1/history")
	c.Params = []gin.Param{
		{Key: "envId", Value: "env-1"},
		{Key: "featureId", Value: "ff-1"},
	}

	handler.GetHistory(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFeatureFlagHandler_GetHistory_NotFound(t *testing.T) {
	mockSvc, handler := setupFFHandler()

	mockSvc.On("GetHistory", mock.Anything, "env-1", "ff-missing").
		Return(nil, domain.ErrFeatureFlagNotFound)

	c, w := ginTestContext("GET", "/api/v1/manage/environments/env-1/features/ff-missing/history")
	c.Params = []gin.Param{
		{Key: "envId", Value: "env-1"},
		{Key: "featureId", Value: "ff-missing"},
	}

	handler.GetHistory(c)

	errorHandler := middleware.ErrorHandler()
	errorHandler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}
