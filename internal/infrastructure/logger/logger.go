package logger

import (
	"log/slog"
	"os"
)

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	With(args ...any) Logger
}

type slogLogger struct {
	logger *slog.Logger
}

func NewLogger() Logger {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return &slogLogger{logger: slog.New(handler)}
}

func (l *slogLogger) Info(msg string, args ...any)  { l.logger.Info(msg, args...) }
func (l *slogLogger) Error(msg string, args ...any) { l.logger.Error(msg, args...) }
func (l *slogLogger) Debug(msg string, args ...any) { l.logger.Debug(msg, args...) }
func (l *slogLogger) Warn(msg string, args ...any)  { l.logger.Warn(msg, args...) }
func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{logger: l.logger.With(args...)}
}