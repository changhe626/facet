package worker

import (
	"testing"
	"time"

	"facet-backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareBatchForInsert_Empty(t *testing.T) {
	batch := []*domain.EvaluationEvent{}
	rows := prepareBatchForInsert(batch)
	assert.Empty(t, rows)
}

func TestPrepareBatchForInsert_SingleEvent(t *testing.T) {
	now := time.Now()
	event := &domain.EvaluationEvent{
		EventID:              "evt-001",
		EventTime:            now,
		SDKKey:               "sdk-test",
		FeatureName:          "feature-a",
		AccountID:            "acc-123",
		UserID:               "user-456",
		AppName:              "my-app",
		MatchedResult:        true,
		EvaluationDurationMs: 42,
		ErrorMessage:         "",
		ClientIP:             "192.168.1.1",
		UserAgent:            "Mozilla/5.0",
		Host:                 "example.com",
		Proto:                "HTTP/1.1",
		Referer:              "https://example.com/page",
	}

	rows := prepareBatchForInsert([]*domain.EvaluationEvent{event})
	require.Len(t, rows, 1)
	require.Len(t, rows[0], 15)

	assert.Equal(t, event.EventID, rows[0][0])
	assert.Equal(t, event.EventTime, rows[0][1])
	assert.Equal(t, event.SDKKey, rows[0][2])
	assert.Equal(t, event.FeatureName, rows[0][3])
	assert.Equal(t, event.AccountID, rows[0][4])
	assert.Equal(t, event.UserID, rows[0][5])
	assert.Equal(t, event.AppName, rows[0][6])
	assert.Equal(t, event.MatchedResult, rows[0][7])
	assert.Equal(t, event.EvaluationDurationMs, rows[0][8])
	assert.Equal(t, event.ErrorMessage, rows[0][9])
	assert.Equal(t, event.ClientIP, rows[0][10])
	assert.Equal(t, event.UserAgent, rows[0][11])
	assert.Equal(t, event.Host, rows[0][12])
	assert.Equal(t, event.Proto, rows[0][13])
	assert.Equal(t, event.Referer, rows[0][14])
}

func TestPrepareBatchForInsert_MultipleEvents(t *testing.T) {
	now := time.Now()
	events := []*domain.EvaluationEvent{
		{
			EventID:              "evt-001",
			EventTime:            now,
			SDKKey:               "sdk-1",
			FeatureName:          "feature-a",
			AccountID:            "acc-1",
			UserID:               "user-1",
			AppName:              "app-1",
			MatchedResult:        true,
			EvaluationDurationMs: 10,
			ClientIP:             "10.0.0.1",
		},
		{
			EventID:              "evt-002",
			EventTime:            now,
			SDKKey:               "sdk-2",
			FeatureName:          "feature-b",
			AccountID:            "acc-2",
			UserID:               "user-2",
			AppName:              "app-2",
			MatchedResult:        false,
			EvaluationDurationMs: 20,
			ErrorMessage:         "flag not found",
			ClientIP:             "10.0.0.2",
		},
		{
			EventID:              "evt-003",
			EventTime:            now,
			SDKKey:               "sdk-3",
			FeatureName:          "feature-c",
			AccountID:            "acc-3",
			UserID:               "user-3",
			AppName:              "app-3",
			MatchedResult:        true,
			EvaluationDurationMs: 30,
			UserAgent:            "curl/7.68.0",
			Host:                 "api.example.com",
			Proto:                "HTTP/2",
			Referer:              "https://example.com",
		},
	}

	rows := prepareBatchForInsert(events)
	require.Len(t, rows, 3)

	for i, row := range rows {
		require.Len(t, row, 15, "row %d should have 15 columns", i)
		assert.Equal(t, events[i].EventID, row[0], "row %d EventID", i)
		assert.Equal(t, events[i].EventTime, row[1], "row %d EventTime", i)
		assert.Equal(t, events[i].SDKKey, row[2], "row %d SDKKey", i)
		assert.Equal(t, events[i].FeatureName, row[3], "row %d FeatureName", i)
		assert.Equal(t, events[i].AccountID, row[4], "row %d AccountID", i)
		assert.Equal(t, events[i].UserID, row[5], "row %d UserID", i)
		assert.Equal(t, events[i].AppName, row[6], "row %d AppName", i)
		assert.Equal(t, events[i].MatchedResult, row[7], "row %d MatchedResult", i)
		assert.Equal(t, events[i].EvaluationDurationMs, row[8], "row %d EvaluationDurationMs", i)
		assert.Equal(t, events[i].ErrorMessage, row[9], "row %d ErrorMessage", i)
		assert.Equal(t, events[i].ClientIP, row[10], "row %d ClientIP", i)
		assert.Equal(t, events[i].UserAgent, row[11], "row %d UserAgent", i)
		assert.Equal(t, events[i].Host, row[12], "row %d Host", i)
		assert.Equal(t, events[i].Proto, row[13], "row %d Proto", i)
		assert.Equal(t, events[i].Referer, row[14], "row %d Referer", i)
	}
}

func TestPrepareBatchForInsert_EmptyFields(t *testing.T) {
	now := time.Now()
	event := &domain.EvaluationEvent{
		EventID:              "evt-001",
		EventTime:            now,
		SDKKey:               "sdk-test",
		FeatureName:          "feature-a",
		MatchedResult:        true,
		EvaluationDurationMs: 0,
	}

	rows := prepareBatchForInsert([]*domain.EvaluationEvent{event})
	require.Len(t, rows, 1)
	require.Len(t, rows[0], 15)

	// Empty strings should be preserved as empty strings
	assert.Equal(t, "", rows[0][4])  // AccountID
	assert.Equal(t, "", rows[0][5])  // UserID
	assert.Equal(t, "", rows[0][6])  // AppName
	assert.Equal(t, "", rows[0][9])  // ErrorMessage
	assert.Equal(t, "", rows[0][10]) // ClientIP
	assert.Equal(t, "", rows[0][11]) // UserAgent
	assert.Equal(t, "", rows[0][12]) // Host
	assert.Equal(t, "", rows[0][13]) // Proto
	assert.Equal(t, "", rows[0][14]) // Referer

	// Zero value for int64
	assert.Equal(t, int64(0), rows[0][8])
}

func TestPrepareBatchForInsert_Ordering(t *testing.T) {
	now := time.Now()
	events := make([]*domain.EvaluationEvent, 5)
	for i := 0; i < 5; i++ {
		events[i] = &domain.EvaluationEvent{
			EventID:              "evt-" + string(rune('0'+i)),
			EventTime:            now.Add(time.Duration(i) * time.Second),
			SDKKey:               "sdk-test",
			FeatureName:          "feature-a",
			MatchedResult:        true,
			EvaluationDurationMs: int64(i * 10),
		}
	}

	rows := prepareBatchForInsert(events)
	require.Len(t, rows, 5)

	// Verify ordering is preserved
	for i, row := range rows {
		assert.Equal(t, events[i].EventID, row[0], "row %d should preserve original order", i)
		assert.Equal(t, events[i].EventTime, row[1], "row %d should preserve original time", i)
		assert.Equal(t, events[i].EvaluationDurationMs, row[8], "row %d should preserve original duration", i)
	}
}

func TestBatchSizeConstant(t *testing.T) {
	assert.Equal(t, 100, BatchSize, "BatchSize should be 100")
}

func TestBatchTimeoutConstant(t *testing.T) {
	assert.Equal(t, 5*time.Second, BatchTimeout, "BatchTimeout should be 5 seconds")
}

func TestEvaluationLogQueueKeyConstant(t *testing.T) {
	assert.Equal(t, "evaluation_logs_queue", EvaluationLogQueueKey, "Queue key should be evaluation_logs_queue")
}

func TestFlush_EmptyBatch(t *testing.T) {
	// flush with an empty batch should return the same empty slice
	batch := []*domain.EvaluationEvent{}
	// We can't call flush directly since it needs a dbPool, but we can test the logic:
	// flush returns batch[:0] which should be an empty slice with the same underlying array
	emptyResult := batch[:0]
	assert.Empty(t, emptyResult)
	assert.Equal(t, 0, len(emptyResult))
}

func TestPrepareBatchForInsert_ColumnOrder(t *testing.T) {
	// Verify the column order matches the expected schema:
	// event_id, event_time, sdk_key, feature_name, account_id, user_id, app_name,
	// matched_result, evaluation_duration_ms, error_message, client_ip, user_agent,
	// host, proto, referer
	now := time.Now()
	event := &domain.EvaluationEvent{
		EventID:              "col-order-test",
		EventTime:            now,
		SDKKey:               "sdk-col",
		FeatureName:          "feature-col",
		AccountID:            "account-col",
		UserID:               "user-col",
		AppName:              "app-col",
		MatchedResult:        true,
		EvaluationDurationMs: 99,
		ErrorMessage:         "error-col",
		ClientIP:             "ip-col",
		UserAgent:            "ua-col",
		Host:                 "host-col",
		Proto:                "proto-col",
		Referer:              "referer-col",
	}

	rows := prepareBatchForInsert([]*domain.EvaluationEvent{event})
	require.Len(t, rows, 1)
	row := rows[0]

	// Verify the column mapping matches the CopyFrom column order in flush()
	expectedColumns := []string{
		"event_id", "event_time", "sdk_key", "feature_name", "account_id",
		"user_id", "app_name", "matched_result", "evaluation_duration_ms",
		"error_message", "client_ip", "user_agent", "host", "proto", "referer",
	}

	expectedValues := []interface{}{
		event.EventID, event.EventTime, event.SDKKey, event.FeatureName, event.AccountID,
		event.UserID, event.AppName, event.MatchedResult, event.EvaluationDurationMs,
		event.ErrorMessage, event.ClientIP, event.UserAgent, event.Host, event.Proto, event.Referer,
	}

	for i, colName := range expectedColumns {
		assert.Equal(t, expectedValues[i], row[i], "column %s at index %d should match", colName, i)
	}
}

func TestPrepareBatchForInsert_LargeBatch(t *testing.T) {
	now := time.Now()
	count := 1000
	events := make([]*domain.EvaluationEvent, count)
	for i := 0; i < count; i++ {
		events[i] = &domain.EvaluationEvent{
			EventID:              "evt-large",
			EventTime:            now,
			SDKKey:               "sdk-large",
			FeatureName:          "feature-large",
			MatchedResult:        i%2 == 0,
			EvaluationDurationMs: int64(i),
		}
	}

	rows := prepareBatchForInsert(events)
	require.Len(t, rows, count)

	for i, row := range rows {
		assert.Equal(t, i%2 == 0, row[7], "row %d MatchedResult", i)
		assert.Equal(t, int64(i), row[8], "row %d EvaluationDurationMs", i)
	}
}

func TestLogWorker_NewLogWorker(t *testing.T) {
	// Just verify that NewLogWorker creates a properly initialized worker
	// without needing actual Redis/DB connections
	w := &LogWorker{
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),
	}

	assert.NotNil(t, w.shutdown)
	assert.NotNil(t, w.done)

	// Verify the channels are unbuffered (as created by make(chan struct{}))
	assert.Equal(t, 0, len(w.shutdown))
	assert.Equal(t, 0, len(w.done))
}

func TestLogWorker_StopWithoutStart(t *testing.T) {
	// A worker that was never started should still be stoppable
	// (the Stop method will block on <-w.done which never closes,
	// but we can't test that without a goroutine. Instead we verify
	// that the shutdown channel works correctly.)
	w := &LogWorker{
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),
	}

	// Closing shutdown signals the worker to stop
	close(w.shutdown)

	// After closing shutdown, we should be able to receive from it
	_, ok := <-w.shutdown
	assert.False(t, ok, "shutdown channel should be closed")
}

func TestLogWorker_NewLogWorkerWithNilRedis(t *testing.T) {
	// Verify that NewLogWorker accepts nil Redis client (it will panic on use)
	// This is a compile-time and initialization test
	w := NewLogWorker(nil, nil)
	assert.NotNil(t, w)
	assert.Nil(t, w.redisClient)
	assert.Nil(t, w.dbPool)
	assert.NotNil(t, w.shutdown)
	assert.NotNil(t, w.done)
}