package logger

import (
	"facet-backend/internal/config"
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger initializes the global slog logger
func InitLogger(cfg config.LoggerConfig) { // Receives LoggerConfig
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Use MultiWriter: stdout always; file only when FileOutput is enabled
	var multiWriter io.Writer = os.Stdout
	if cfg.FileOutput {
		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		multiWriter = io.MultiWriter(os.Stdout, lumberjackLogger)
	}

	// Create a JSON handler
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: level,
	})

	// Set as the global default logger
	slog.SetDefault(slog.New(handler))
}
