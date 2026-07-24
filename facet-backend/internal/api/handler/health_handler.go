package handler

import (
	"context"
	"facet-backend/internal/api/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	dbPool  *pgxpool.Pool
	redisCL *redis.Client
}

func NewHealthHandler(dbPool *pgxpool.Pool, redisCL *redis.Client) *HealthHandler {
	return &HealthHandler{dbPool: dbPool, redisCL: redisCL}
}

func (h *HealthHandler) Check(c *gin.Context) {
	if err := h.dbPool.Ping(c.Request.Context()); err != nil {
		response.CustomError(c, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	if _, err := h.redisCL.Ping(context.Background()).Result(); err != nil {
		response.CustomError(c, http.StatusServiceUnavailable, "cache unavailable")
		return
	}

	response.Success(c, gin.H{"status": "ok"})
}