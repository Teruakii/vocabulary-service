package stats

import (
	"context"
	"time"
)

type Repository interface {
	Get(ctx context.Context, userID uint64) (*UserStats, error)

	IncrementQuizzesTaken(ctx context.Context, userID uint64, now time.Time) error

	RecordActivity(ctx context.Context, userID uint64, now time.Time) error

	// ActivitySince returns per-day counts keyed by YYYY-MM-DD, for days on or after from.
	ActivitySince(ctx context.Context, userID uint64, from time.Time) (map[string]uint, error)
}

type Service interface {
	Get(ctx context.Context, userID uint64) (*UserStats, error)

	RecordQuizCompleted(ctx context.Context, userID uint64) error

	RecordActivity(ctx context.Context, userID uint64) error

	Activity(ctx context.Context, userID uint64, days int) ([]ActivityPoint, error)
}
