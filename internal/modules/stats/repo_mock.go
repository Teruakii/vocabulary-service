package stats

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Get(ctx context.Context, userID uint64) (*UserStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserStats), args.Error(1)
}

func (m *MockRepository) IncrementQuizzesTaken(ctx context.Context, userID uint64, now time.Time) error {
	args := m.Called(ctx, userID, now)
	return args.Error(0)
}

func (m *MockRepository) RecordActivity(ctx context.Context, userID uint64, now time.Time) error {
	args := m.Called(ctx, userID, now)
	return args.Error(0)
}

func (m *MockRepository) ActivitySince(ctx context.Context, userID uint64, from time.Time) (map[string]uint, error) {
	args := m.Called(ctx, userID, from)
	var counts map[string]uint
	if args.Get(0) != nil {
		counts = args.Get(0).(map[string]uint)
	}
	return counts, args.Error(1)
}
