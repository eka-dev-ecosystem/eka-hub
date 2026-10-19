// Package logging provides a structured logging layer built on top of
// log/slog. All output is JSON (console and optional file writer), including
// request-scoped access logs that carry a request ID through every line.
package logging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Level is the set of log levels supported by the application.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// ParseLevel converts a string (from LOG_LEVEL env) into a slog.Level.
// An empty or unknown value falls back to slog.LevelInfo.
func ParseLevel(level string) slog.Level {
	switch Level(strings.ToLower(strings.TrimSpace(level))) {
	case LevelDebug:
		return slog.LevelDebug
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Destroy closes any file opened for logging. Call it during shutdown.
func Destroy() {
	fileMu.Lock()
	defer fileMu.Unlock()
	if fileHandle != nil {
		_ = fileHandle.Close()
		fileHandle = nil
	}
}

// Setup initializes the package logger. Output is always JSON.
//
//   - level:    minimum level to output ("debug", "info", "warn", "error")
//   - filePath: when non-empty, every log line is also written to this file
//     in JSON format; parent directories are created as needed
//
// It returns the package-wide logger. Call Destroy() during shutdown to
// close the file handle.
func Setup(level string, filePath string) *slog.Logger {
	var handlers []slog.Handler

	console := newHandler(os.Stderr, ParseLevel(level))
	handlers = append(handlers, console)

	if filePath != "" {
		if err := openFile(filePath); err != nil {
			// Log the failure as JSON on stderr; don't crash startup.
			slog.New(newHandler(os.Stderr, slog.LevelWarn)).Warn(
				"failed to open log file",
				slog.String("path", filePath),
				slog.String("error", err.Error()),
			)
		} else if fileHandle != nil {
			handlers = append(handlers, slog.NewJSONHandler(fileHandle, &slog.HandlerOptions{Level: ParseLevel(level)}))
		}
	}

	var handler = console
	if len(handlers) > 1 {
		handler = newMultiHandler(handlers)
	}
	return Set(slog.New(handler))
}

func newHandler(w io.Writer, level slog.Level) slog.Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
}

// newMultiHandler fans out log records to several handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(hs []slog.Handler) slog.Handler {
	return &multiHandler{handlers: hs}
}

func (h *multiHandler) Enabled(_ context.Context, l slog.Level) bool {
	for _, hd := range h.handlers {
		if hd.Enabled(context.Background(), l) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs error
	for _, hd := range h.handlers {
		if hd.Enabled(ctx, r.Level) {
			if err := hd.Handle(ctx, r.Clone()); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	return errs
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hs := make([]slog.Handler, len(h.handlers))
	for i, hd := range h.handlers {
		hs[i] = hd.WithAttrs(attrs)
	}
	return newMultiHandler(hs)
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	hs := make([]slog.Handler, len(h.handlers))
	for i, hd := range h.handlers {
		hs[i] = hd.WithGroup(name)
	}
	return newMultiHandler(hs)
}

// openFile creates the log directory and opens/creates the file in append
// mode.
func openFile(filePath string) error {
	if dir := filepath.Dir(filePath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	fileMu.Lock()
	fileHandle = f
	fileMu.Unlock()
	return nil
}

var (
	logMu      sync.RWMutex
	logger     *slog.Logger
	fileMu     sync.Mutex
	fileHandle *os.File
)

// Set installs the provided logger as the package-wide logger and also as
// the slog default, so third-party code that relies on slog.Default() is
// covered as well.
func Set(l *slog.Logger) *slog.Logger {
	logMu.Lock()
	logger = l
	logMu.Unlock()

	slog.SetDefault(l)
	return l
}

// Logger returns the package-wide logger, creating a sensible default
// (INFO, JSON output, discarded) if Setup was never called.
func Logger() *slog.Logger {
	logMu.RLock()
	l := logger
	logMu.RUnlock()

	if l == nil {
		return Set(slog.New(newHandler(io.Discard, slog.LevelInfo)))
	}
	return l
}

// With returns a child logger with the given key-value attributes attached.
func With(args ...any) *slog.Logger {
	return Logger().With(args...)
}

// Group returns a child logger with the given group name.
func Group(name string) *slog.Logger {
	return Logger().WithGroup(name)
}

// Debug logs a message at debug level.
func Debug(msg string, args ...any) { Logger().Debug(msg, args...) }

// Info logs a message at info level.
func Info(msg string, args ...any) { Logger().Info(msg, args...) }

// Warn logs a message at warn level.
func Warn(msg string, args ...any) { Logger().Warn(msg, args...) }

// Error logs a message at error level.
func Error(msg string, args ...any) { Logger().Error(msg, args...) }

// Fatal logs a message at error level and exits the process with status 1.
func Fatal(msg string, args ...any) {
	Logger().Error(msg, args...)
	os.Exit(1)
}

// Attr helpers.

// Err builds an "error" attribute from an error value.
func Err(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "")
	}
	return slog.String("error", err.Error())
}

// contextKey is used for request-scoped values.
type contextKey string

const (
	ctxKeyRequestID contextKey = "logging:request_id"
	ctxKeyLogger    contextKey = "logging:logger"
)

// RequestID extracts the request ID from the context, or "".
func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyRequestID).(string); ok {
		return v
	}
	return ""
}

// ContextWithRequestID returns a context carrying the request ID.
func ContextWithRequestID(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, rid)
}

// ContextWithLogger returns a context carrying the given logger.
func ContextWithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, l)
}

// LoggerFromContext returns the logger stored in the context, or the
// package-wide logger when none is present.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKeyLogger).(*slog.Logger); ok && l != nil {
		return l
	}
	return Logger()
}
