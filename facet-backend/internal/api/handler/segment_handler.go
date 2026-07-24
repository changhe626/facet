package handler

import (
	"errors"
	"facet-backend/internal/api/middleware"
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/domain"
	"facet-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SegmentHandler struct {
	segService service.SegmentService
}

func NewSegmentHandler(segService service.SegmentService) *SegmentHandler {
	return &SegmentHandler{segService: segService}
}

func toSegmentResponse(seg *domain.Segment) response.SegmentResponse {
	resp := response.SegmentResponse{
		ID:          seg.ID,
		Name:        seg.Name,
		Description: seg.Description,
		MatchType:   string(seg.MatchType),
		CreatedAt:   seg.CreatedAt,
		UpdatedAt:   seg.UpdatedAt,
	}
	for _, c := range seg.Conditions {
		resp.Conditions = append(resp.Conditions, response.SegmentConditionDTO{
			ID:           c.ID,
			AttributeKey: c.AttributeKey,
			Operator:     c.Operator,
			Value:        c.Value,
		})
	}
	return resp
}

func (h *SegmentHandler) List(c *gin.Context) {
	tenantIDVal, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	tenantID, _ := tenantIDVal.(string)
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	segments, total, err := h.segService.List(c.Request.Context(), tenantID, keyword, page, pageSize)
	if err != nil {
		_ = c.Error(err)
		return
	}

	items := make([]response.SegmentResponse, len(segments))
	for i, seg := range segments {
		items[i] = toSegmentResponse(seg)
	}

	response.Success(c, response.ListSegmentsResponse{
		Page:       int64(page),
		PageSize:   int64(pageSize),
		Total:      total,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
		Items:      items,
	})
}

func (h *SegmentHandler) Create(c *gin.Context) {
	tenantIDVal, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	userIDVal, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("userID not found in context"))
		return
	}
	tenantID, _ := tenantIDVal.(string)
	userID, _ := userIDVal.(string)

	var req request.CreateSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(domain.ErrInvalidInput)
		return
	}

	seg, err := h.segService.Create(c.Request.Context(), tenantID, userID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, toSegmentResponse(seg))
}

func (h *SegmentHandler) GetByID(c *gin.Context) {
	tenantIDVal, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	tenantID, _ := tenantIDVal.(string)
	id := c.Param("segmentId")

	seg, err := h.segService.GetByID(c.Request.Context(), id, tenantID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, toSegmentResponse(seg))
}

func (h *SegmentHandler) Update(c *gin.Context) {
	tenantIDVal, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	userIDVal, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("userID not found in context"))
		return
	}
	tenantID, _ := tenantIDVal.(string)
	userID, _ := userIDVal.(string)
	id := c.Param("segmentId")

	var req request.CreateSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(domain.ErrInvalidInput)
		return
	}

	updateReq := request.UpdateSegmentRequest{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		MatchType:   req.MatchType,
		Conditions:  req.Conditions,
	}

	seg, err := h.segService.Update(c.Request.Context(), tenantID, userID, updateReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, toSegmentResponse(seg))
}

func (h *SegmentHandler) Delete(c *gin.Context) {
	tenantIDVal, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	tenantID, _ := tenantIDVal.(string)
	id := c.Param("segmentId")

	if err := h.segService.Delete(c.Request.Context(), id, tenantID); err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}
