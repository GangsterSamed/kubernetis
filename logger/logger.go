package logger

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"os"
)

func SetupLogger(logLevel, logFormat string) *slog.Logger {
	level := parseLevel(logLevel)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}

	var handler slog.Handler
	if logFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	return logger
}

func parseLevel(logLevel string) slog.Level {
	switch logLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "fatal":
		return slog.LevelError
	default:
		// Log warning about invalid level and return default
		fmt.Fprintf(os.Stderr, "Warning: invalid log level '%s', using 'info' as default\n", logLevel)
		return slog.LevelInfo
	}

}

func LoggerFromContext(ctx *gin.Context, base *slog.Logger) *slog.Logger {
	if logger, ok := ctx.Get("logger"); ok {
		if l, ok := logger.(*slog.Logger); ok {
			return l
		}
	}
	return base
}
