package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"sort"
)

type contextKey struct{}

type Logger struct {
	l *slog.Logger
}

func New() *Logger {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	return &Logger{l: slog.New(handler)}
}

func fromContext(ctx context.Context) map[string]any {
	fields, ok := ctx.Value(contextKey{}).(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return fields
}

func (l *Logger) WithFields(ctx context.Context, f map[string]any) context.Context {
	fields := fromContext(ctx)
	newFields := make(map[string]any, len(fields)+len(f))
	for k, v := range fields {
		newFields[k] = v
	}
	for k, v := range f {
		newFields[k] = v
	}
	return context.WithValue(ctx, contextKey{}, newFields)
}

func sortedAttrs(ctx context.Context) []any {
	fields := fromContext(ctx)
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	attrs := make([]any, 0, len(fields))
	for _, k := range keys {
		attrs = append(attrs, slog.Any(k, fields[k]))
	}
	return attrs
}

// ✅ Короткий Info (без runtime.Caller)
func (l *Logger) Info(ctx context.Context, msg string) {
	attrs := sortedAttrs(ctx)
	l.l.Info(msg, attrs...)
}

// ✅ Warn — с caller
func (l *Logger) Warn(ctx context.Context, msg string) {
	pc, file, line, _ := runtime.Caller(1)

	attrs := sortedAttrs(ctx)
	attrs = append(attrs,
		slog.String("file", file),
		slog.Int("line", line),
		slog.String("func", runtime.FuncForPC(pc).Name()),
	)
	l.l.Warn(msg, attrs...)
}

// ✅ Error — с caller
func (l *Logger) Error(ctx context.Context, msg string) {
	pc, file, line, _ := runtime.Caller(1)

	attrs := sortedAttrs(ctx)
	attrs = append(attrs,
		slog.String("file", file),
		slog.Int("line", line),
		slog.String("func", runtime.FuncForPC(pc).Name()),
	)
	l.l.Error(msg, attrs...)
}

// ✅ Fatal — как Error, но завершает программу
func (l *Logger) Fatal(ctx context.Context, msg string) {
	pc, file, line, _ := runtime.Caller(1)

	attrs := sortedAttrs(ctx)
	attrs = append(attrs,
		slog.String("file", file),
		slog.Int("line", line),
		slog.String("func", runtime.FuncForPC(pc).Name()),
	)
	l.l.Error(msg, attrs...)
	os.Exit(1)
}
