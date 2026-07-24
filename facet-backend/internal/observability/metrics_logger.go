package observability

import (
	"log/slog"
)

// MetricsLogger is a simple, slog-based metrics logger.
// It prints metrics as structured logs for easy analysis and processing.
type MetricsLogger struct {
	logger *slog.Logger
}

// NewMetricsLogger creates a new MetricsLogger instance.
func NewMetricsLogger(logger *slog.Logger) *MetricsLogger {
	return &MetricsLogger{logger: logger.With("component", "MetricsLogger")}
}

// Incr records a counter-type metric. The counter value always increases by 1.
func (m *MetricsLogger) Incr(name string, labels map[string]string) {
	m.logger.Info(
		"metric",
		slog.String("metric_name", name),
		slog.Int("value", 1),
		slog.Any("labels", labels),
	)
}

// Observe records an observation-type metric, such as latency or size.
func (m *MetricsLogger) Observe(name string, value float64, labels map[string]string) {
	m.logger.Info(
		"metric",
		slog.String("metric_name", name),
		slog.Float64("value", value),
		slog.Any("labels", labels),
	)
}
