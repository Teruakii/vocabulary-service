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
	return stats, nil
}

func (s *service) RecordWordMastered(ctx context.Context, userID uint64) error {
	if err := s.repo.IncrementWordsLearned(ctx, userID, s.now()); err != nil {
		return fmt.Errorf("record word mastered for user %d: %w", userID, err)
	}
	return nil
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
