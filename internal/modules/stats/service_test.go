package stats

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"vocabulary-service/internal/constants"
)

func TestGet_Success(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	want := &UserStats{UserID: 1, CurrentStreak: 3}
	repo.On("Get", context.Background(), uint64(1)).Return(want, nil)

	got, err := svc.Get(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGet_PropagatesError(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("Get", context.Background(), uint64(1)).Return(nil, constants.ErrNotFound)

	_, err := svc.Get(context.Background(), 1)

	assert.Error(t, err)
}

func TestRecordQuizCompleted_CallsIncrementQuizzesTaken(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo).(*service)
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixed }

	repo.On("IncrementQuizzesTaken", context.Background(), uint64(1), fixed).Return(nil)

	err := svc.RecordQuizCompleted(context.Background(), 1)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestRecordActivity_CallsRecordActivity(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo).(*service)
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixed }

	repo.On("RecordActivity", context.Background(), uint64(1), fixed).Return(nil)

	err := svc.RecordActivity(context.Background(), 1)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestActivity_FillsMissingDaysWithZero(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo).(*service)
	svc.now = func() time.Time { return time.Date(2026, 10, 10, 15, 0, 0, 0, time.UTC) }

	from := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	repo.On("ActivitySince", context.Background(), uint64(1), from).
		Return(map[string]uint{"2026-10-08": 3, "2026-10-10": 5}, nil)

	points, err := svc.Activity(context.Background(), 1, 3)

	assert.NoError(t, err)
	assert.Equal(t, []ActivityPoint{
		{Date: "2026-10-08", Count: 3},
		{Date: "2026-10-09", Count: 0},
		{Date: "2026-10-10", Count: 5},
	}, points)
}
