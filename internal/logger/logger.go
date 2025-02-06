// Package logger provides logging capabilities.
package logger

import (
	"context"
	"log/slog"
	"os"
)

// Log levels supported.
const (
	LvlDebug = slog.LevelDebug
	LvlInfo  = slog.LevelInfo
	LvlWarn  = slog.LevelWarn
	LvlError = slog.LevelError
	LvlFatal = slog.LevelError
)

const (
	LogOutStderr = "stderr"
	LogOutStdout = "stdout"
)

const (
	LogFmtJSON = "json"
	LogFmtText = "text"
)

// Logger is the required logger interface for this service.
type Logger interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
}

// New returns a new logger.
func New(output, format string,
	level slog.Level,
) Logger {
	out := os.Stderr

	if output == LogOutStdout {
		out = os.Stdout
	}

	if format == LogFmtText {
		return slog.New(NewLogHandler(slog.NewTextHandler(out,
			&slog.HandlerOptions{Level: level})))
	}

	return slog.New(NewLogHandler(slog.NewJSONHandler(out,
		&slog.HandlerOptions{Level: level})))
}

// A LogHandler wraps an slog.Handler for use with this logger interface.
type LogHandler struct {
	handler slog.Handler
}

// NewLogHandler returns a new LogHandler for use as a log handler.
func NewLogHandler(h slog.Handler) *LogHandler {
	if lh, ok := h.(*LogHandler); ok {
		h = lh.Handler()
	}

	return &LogHandler{handler: h}
}

// Enabled implements Handler.Enabled.
func (h *LogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle implements Handler.Handle and adds the context data for this service.
func (h *LogHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.NumAttrs() > 0 {
		tID, ok := ctx.Value(5).(string)
		if !ok {
			tID = "none"
		}

		r.Add("source", "api", "trace_id", tID)
	}

	return h.handler.Handle(ctx, r)
}

// WithAttrs implements Handler.WithAttrs.
func (h *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewLogHandler(h.handler.WithAttrs(attrs))
}

// WithGroup implements Handler.WithGroup.
func (h *LogHandler) WithGroup(name string) slog.Handler {
	return NewLogHandler(h.handler.WithGroup(name))
}

// Handler returns the Handler wrapped by h.
func (h *LogHandler) Handler() slog.Handler {
	return h.handler
}

// NoOpLogger implements the Logger interface, but does nothing.
type NoOpLogger struct{}

// Log implements the interface, but intentionally does nothing.
func (nl NoOpLogger) Log(ctx context.Context,
	level slog.Level,
	msg string,
	args ...any,
) {
}

// NullLog is a singleton no-op logger.
var NullLog NoOpLogger
