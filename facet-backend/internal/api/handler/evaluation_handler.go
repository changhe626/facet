package handler

import (
	"facet-backend/internal/api/response"
	"facet-backend/internal/domain"
	"facet-backend/internal/service"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
)

type EvaluationHandler struct {
	evalService   service.EvaluationService
	eventService  domain.EventService
	snowflakeNode *snowflake.Node
}

func NewEvaluationHandler(evalService service.EvaluationService, eventService domain.EventService, snowflakeNode *snowflake.Node) *EvaluationHandler {
	return &EvaluationHandler{
		evalService:   evalService,
		eventService:  eventService,
		snowflakeNode: snowflakeNode,
	}
}

func (h *EvaluationHandler) Evaluate(c *gin.Context) {
	startTime := time.Now()

	sdkKey := c.Query("sdk_key")
	featureName := c.Query("feature_name")
	userID := c.Query("user_id")
	accountID := c.Query("account_id")
	appName := c.Query("app_name")
	contextJSON := c.Query("context")

	matched, err := h.evalService.Evaluate(c.Request.Context(), sdkKey, featureName, userID, accountID, appName, contextJSON)
	evalDurationMs := time.Since(startTime).Milliseconds()

	event := &domain.EvaluationEvent{
		EventID:              "evt-" + h.snowflakeNode.Generate().String(),
		EventTime:            startTime,
		SDKKey:               sdkKey,
		FeatureName:          featureName,
		AccountID:            accountID,
		UserID:               userID,
		AppName:              appName,
		MatchedResult:        matched,
		EvaluationDurationMs: evalDurationMs,
		ClientIP:             c.ClientIP(),
		UserAgent:            c.Request.UserAgent(),
		Host:                 c.Request.Host,
		Proto:                c.Request.Proto,
		Referer:              c.Request.Referer(),
	}
	if err != nil {
		event.ErrorMessage = err.Error()
	}

	h.eventService.RecordEvaluationEvent(event)

	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, gin.H{
		"matched": matched,
	})
}
