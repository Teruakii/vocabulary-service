package stats

import (
	"context"
	"fmt"
	"time"
)

type service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) Service {
	return &service{repo: repo, now: time.Now}
}

func (s *service) Get(ctx context.Context, userID uint64) (*UserStats, error) {
	stats, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get stats for user %d: %w", userID, err)
	}
	stats.CurrentStreak = stats.streakAt(s.now())
	return stats, nil
}

func (s *service) RecordQuizCompleted(ctx context.Context, userID uint64) error {
	if err := s.repo.IncrementQuizzesTaken(ctx, userID, s.now()); err != nil {
		return fmt.Errorf("record quiz completed for user %d: %w", userID, err)
	}
	return nil
}

func (s *service) RecordActivity(ctx context.Context, userID uint64) error {
	if err := s.repo.RecordActivity(ctx, userID, s.now()); err != nil {
		return fmt.Errorf("record activity for user %d: %w", userID, err)
	}
	return nil
}

// Activity returns one point per day for the last `days` days, oldest first,
// with zero for days without activity so the chart has no gaps.
func (s *service) Activity(ctx context.Context, userID uint64, days int) ([]ActivityPoint, error) {
	if days <= 0 {
		days = DefaultActivityDays
	}
	if days > MaxActivityDays {
		days = MaxActivityDays
	}

	now := s.now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	from := today.AddDate(0, 0, -(days - 1))

	counts, err := s.repo.ActivitySince(ctx, userID, from)
	if err != nil {
		return nil, fmt.Errorf("activity for user %d: %w", userID, err)
	}

	points := make([]ActivityPoint, 0, days)
	for d := from; !d.After(today); d = d.AddDate(0, 0, 1) {
		key := d.Format(dateLayout)
		points = append(points, ActivityPoint{Date: key, Count: counts[key]})
	}
	return points, nil
}
