package logger

import (
	"context"
	"log/slog"
)

type Logger interface {
	Info(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)

	Error(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)

	Warn(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
}

type StdLogger struct {
	slog *slog.Logger
}

func NewStdLogger(slog *slog.Logger) *StdLogger {
	return &StdLogger{slog: slog}
}

func (l *StdLogger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

func (l *StdLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.slog.InfoContext(ctx, msg, args...)
}

func (l *StdLogger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

func (l *StdLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.slog.ErrorContext(ctx, msg, args...)
}

func (l *StdLogger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

func (l *StdLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.slog.WarnContext(ctx, msg, args...)
}
