package services

import (
	"context"
	"time"

	"superhue/internal/domain"
)

type AppLogger struct {
	repo domain.LogRepository
}

func NewAppLogger(repo domain.LogRepository) *AppLogger { return &AppLogger{repo: repo} }

func (l *AppLogger) Info(ctx context.Context, source, message string) {
	_ = l.repo.Add(ctx, &domain.LogEntry{Level: domain.LogInfo, Source: source, Message: message, CreatedAt: time.Now().UTC()})
}

func (l *AppLogger) Error(ctx context.Context, source, message string) {
	_ = l.repo.Add(ctx, &domain.LogEntry{Level: domain.LogError, Source: source, Message: message, CreatedAt: time.Now().UTC()})
}
