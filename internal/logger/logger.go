package logger

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/olexsmir/onasty/internal/transport/http/reqid"
)

type CustomLogger struct {
	slog.Handler
}

//nolint:err113
func NewCustomLogger(lvl, format string, showLine bool) (*slog.Logger, error) {
	loggerLevels := map[string]slog.Level{
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
		"error": slog.LevelError,
		"warn":  slog.LevelWarn,
	}

	logLevel, ok := loggerLevels[lvl]
	if !ok {
		return nil, errors.New("unknown log level")
	}

	handlerOptions := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: showLine,
	}

	var slogHandler slog.Handler
	switch format {
	case "json":
		slogHandler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	case "text":
		slogHandler = slog.NewTextHandler(os.Stdout, handlerOptions)
	default:
		return nil, errors.New("unknown log format")
	}

	return slog.New(&CustomLogger{Handler: slogHandler}), nil
}

func (l *CustomLogger) Handle(ctx context.Context, r slog.Record) error {
	if requestID := reqid.GetFromContext(ctx); requestID != "" {
		r.AddAttrs(slog.String("request_id", requestID))
	}

	return l.Handler.Handle(ctx, r)
}
