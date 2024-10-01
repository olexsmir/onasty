package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/olexsmir/onasty/internal/transport/http/reqid"
)

var _ slog.Handler = (*CustomLogger)(nil)

type CustomLogger struct{ slog.Handler }

type CustomLoggerOpts struct {
	// Level is the log level. Can be "info", "debug", "error", "warn".
	Level string

	// Format is the log format. Can be "json" or "text".
	Format string

	// ShowLine enables or disables the line number in the log output.
	ShowLine bool

	// Output is the writer to write logs to.
	// If not set, os.Stdout is used.
	Output io.Writer
}

//nolint:err113
func NewCustomLogger(opts CustomLoggerOpts) (*slog.Logger, error) {
	loggerLevels := map[string]slog.Level{
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
		"error": slog.LevelError,
		"warn":  slog.LevelWarn,
	}

	logLevel, ok := loggerLevels[opts.Level]
	if !ok {
		return nil, errors.New("unknown log level")
	}

	handlerOptions := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: opts.ShowLine,
	}

	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	var slogHandler slog.Handler
	switch opts.Format {
	case "json":
		slogHandler = slog.NewJSONHandler(opts.Output, handlerOptions)
	case "text":
		slogHandler = slog.NewTextHandler(opts.Output, handlerOptions)
	default:
		return nil, errors.New("unknown log format")
	}

	return slog.New(&CustomLogger{Handler: slogHandler}), nil
}

func (l *CustomLogger) Handle(ctx context.Context, r slog.Record) error {
	if requestID := reqid.GetContext(ctx); requestID != "" {
		r.AddAttrs(slog.String("request_id", requestID))
	}

	return l.Handler.Handle(ctx, r)
}
