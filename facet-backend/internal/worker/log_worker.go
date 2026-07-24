package worker

import (
	"context"
	"encoding/json"
	"facet-backend/internal/domain"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	EvaluationLogQueueKey = "evaluation_logs_queue"
	BatchSize             = 100
	BatchTimeout          = 5 * time.Second
)

type LogWorker struct {
	redisClient *redis.Client
	dbPool      *pgxpool.Pool
	shutdown    chan struct{}
	done        chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewLogWorker(redisClient *redis.Client, dbPool *pgxpool.Pool) *LogWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &LogWorker{
		redisClient: redisClient,
		dbPool:      dbPool,
		shutdown:    make(chan struct{}),
		done:        make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (w *LogWorker) Start() {
	defer w.cancel()
	slog.Info("log worker started")
	ticker := time.NewTicker(BatchTimeout)
	defer ticker.Stop()

	var batch []*domain.EvaluationEvent

	for {
		select {
		case <-w.shutdown:
			if len(batch) > 0 {
				w.flush(batch)
			}
			close(w.done)
			slog.Info("log worker stopped")
			return
		default:
		}

		select {
		case <-ticker.C:
			if len(batch) > 0 {
				batch = w.flush(batch)
			}
		default:
		}

		result, err := w.redisClient.BRPop(w.ctx, 1*time.Second, EvaluationLogQueueKey).Result()
		if err != nil {
			if err != redis.Nil && err != context.Canceled {
				slog.Error("failed to pop log from redis", "error", err)
			}
			continue
		}

		if len(result) >= 2 {
			var event domain.EvaluationEvent
			if err := json.Unmarshal([]byte(result[1]), &event); err != nil {
				slog.Warn("failed to unmarshal evaluation event", "error", err, "data", result[1])
				continue
			}
			batch = append(batch, &event)
		}

		if len(batch) >= BatchSize {
			batch = w.flush(batch)
		}
	}
}

func (w *LogWorker) Stop() {
	close(w.shutdown)
	<-w.done
}

func prepareBatchForInsert(batch []*domain.EvaluationEvent) [][]interface{} {
	rows := make([][]interface{}, len(batch))
	for i, event := range batch {
		rows[i] = []interface{}{
			event.EventID, event.EventTime, event.SDKKey, event.FeatureName, event.AccountID, event.UserID,
			event.AppName, event.MatchedResult, event.EvaluationDurationMs, event.ErrorMessage,
			event.ClientIP, event.UserAgent, event.Host, event.Proto, event.Referer,
		}
	}
	return rows
}

func (w *LogWorker) flush(batch []*domain.EvaluationEvent) []*domain.EvaluationEvent {
	if len(batch) == 0 {
		return batch
	}

	slog.Info("flushing logs to database", "count", len(batch))

	rows := prepareBatchForInsert(batch)

	copyCount, err := w.dbPool.CopyFrom(
		w.ctx,
		pgx.Identifier{"facet", "evaluation_logs"},
		[]string{"event_id", "event_time", "sdk_key", "feature_name", "account_id", "user_id", "app_name", "matched_result", "evaluation_duration_ms", "error_message", "client_ip", "user_agent", "host", "proto", "referer"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		slog.Error("failed to batch insert evaluation logs", "error", err)
	} else {
		slog.Info("successfully inserted logs", "count", copyCount)
	}

	return batch[:0]
}
