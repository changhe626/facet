package service

import (
	"context"
	"encoding/json"
	"facet-backend/internal/domain"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

const (
	EvaluationLogQueueKey = "evaluation_logs_queue"
)

type redisEventService struct {
	redisClient *redis.Client
	sem         chan struct{}
}

func NewEventService(redisClient *redis.Client) domain.EventService {
	return &redisEventService{
		redisClient: redisClient,
		sem:         make(chan struct{}, 100),
	}
}

// Close waits for all in-flight goroutines to finish, then returns.
// After Close, new events are silently dropped (semaphore is full).
func (s *redisEventService) Close() error {
	// Acquire all semaphore slots — blocks until every goroutine finishes
	for i := 0; i < cap(s.sem); i++ {
		s.sem <- struct{}{}
	}
	return nil
}

func (s *redisEventService) RecordEvaluationEvent(event *domain.EvaluationEvent) {
	select {
	case s.sem <- struct{}{}:
	default:
		slog.Warn("event service overloaded, dropping event")
		return
	}

	go func() {
		defer func() {
			<-s.sem
			if r := recover(); r != nil {
				slog.Error("panic in event worker", "recover", r)
			}
		}()

		ctx := context.Background()

		payload, err := json.Marshal(event)
		if err != nil {
			slog.Warn("serialize evaluation event failed", "error", err)
			return
		}

		if err := s.redisClient.LPush(ctx, EvaluationLogQueueKey, payload).Err(); err != nil {
			slog.Warn("push evaluation event to redis failed", "error", err)
		}
	}()
}
