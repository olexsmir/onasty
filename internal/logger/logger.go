package logger

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/olexsmir/onasty/internal/transport/http/reqid"
)

var (
	ErrUnknownLevel  = errors.New("unknown log level")
	ErrUnknownFormat = errors.New("unknown log format")
)

// SetDefault configures and set default [slog.Logger]
func SetDefault(lvl, format string, showLine bool) error {
	loggerLevels := map[string]slog.Level{
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
		"error": slog.LevelError,
		"warn":  slog.LevelWarn,
	}

	logLevel, ok := loggerLevels[lvl]
	if !ok {
		return ErrUnknownLevel
	}

	handlerOptions := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: showLine,
	}

	var slogHandler slog.Handler
	switch format {
	case "json":
		slogHandler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	case "text", "txt":
		slogHandler = slog.NewTextHandler(os.Stdout, handlerOptions)
	default:
		return ErrUnknownFormat
	}

	slog.SetDefault(slog.New(&CustomLogger{Handler: slogHandler}))
	return nil
}

type CustomLogger struct{ slog.Handler }

func (l *CustomLogger) Handle(ctx context.Context, r slog.Record) error {
	if requestID := reqid.GetContext(ctx); requestID != "" {
		r.AddAttrs(slog.String("request_id", requestID))
	}

	return l.Handler.Handle(ctx, r)
}
