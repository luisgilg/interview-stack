package logging

import (
	"log/slog"

	"github.com/example/interview-stack/go-service/internal/domain"
)

// StructuredLogger implements domain.Logger backed by slog.
type StructuredLogger struct {
	logger *slog.Logger
}

func NewStructuredLogger(logger *slog.Logger) *StructuredLogger {
	return &StructuredLogger{logger: logger}
}

func (l *StructuredLogger) Info(msg string, fields ...domain.Field) {
	l.logger.Info(msg, toArgs(fields)...)
}

func (l *StructuredLogger) Warn(msg string, fields ...domain.Field) {
	l.logger.Warn(msg, toArgs(fields)...)
}

func (l *StructuredLogger) Error(msg string, err error, fields ...domain.Field) {
	args := append(toArgs(fields), slog.Any("error", err))
	l.logger.Error(msg, args...)
}

func toArgs(fields []domain.Field) []any {
	args := make([]any, 0, len(fields)*2)
	for _, field := range fields {
		args = append(args, field.Key, field.Value)
	}
	return args
}
