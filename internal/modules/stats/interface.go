package stats

import (
	"context"
	"time"
)

type Repository interface {
	Get(ctx context.Context, userID uint64) (*UserStats, error)

	IncrementWordsLearned(ctx context.Context, userID uint64, now time.Time) error

	IncrementQuizzesTaken(ctx context.Context, userID uint64, now time.Time) error

	RecordActivity(ctx context.Context, userID uint64, now time.Time) error
}

type Service interface {
	Get(ctx context.Context, userID uint64) (*UserStats, error)

	RecordWordMastered(ctx context.Context, userID uint64) error

	RecordQuizCompleted(ctx context.Context, userID uint64) error

	RecordActivity(ctx context.Context, userID uint64) error
}
