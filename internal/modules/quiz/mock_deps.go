package quiz

import (
	"context"

	"github.com/stretchr/testify/mock"

	"vocabulary-service/internal/modules/userword"
)

type mockWordsService struct {
	mock.Mock
}

func (m *mockWordsService) AddWord(ctx context.Context, userID, wordID uint64) (*userword.UserWord, error) {
	args := m.Called(ctx, userID, wordID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userword.UserWord), args.Error(1)
}

func (m *mockWordsService) GetByID(ctx context.Context, userID, id uint64) (*userword.UserWord, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userword.UserWord), args.Error(1)
}

func (m *mockWordsService) List(ctx context.Context, userID uint64, status userword.Status, limit, offset int) ([]userword.UserWordDetail, bool, error) {
	args := m.Called(ctx, userID, status, limit, offset)
	var items []userword.UserWordDetail
	if args.Get(0) != nil {
		items = args.Get(0).([]userword.UserWordDetail)
	}
	return items, args.Bool(1), args.Error(2)
}

func (m *mockWordsService) Due(ctx context.Context, userID uint64, limit int) ([]userword.UserWordDetail, error) {
	args := m.Called(ctx, userID, limit)
	var items []userword.UserWordDetail
	if args.Get(0) != nil {
		items = args.Get(0).([]userword.UserWordDetail)
	}
	return items, args.Error(1)
}

func (m *mockWordsService) Review(ctx context.Context, userID, id uint64, quality int) (*userword.UserWord, error) {
	args := m.Called(ctx, userID, id, quality)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userword.UserWord), args.Error(1)
}

func (m *mockWordsService) ReviewByWord(ctx context.Context, userID, wordID uint64, quality int) (*userword.UserWord, error) {
	args := m.Called(ctx, userID, wordID, quality)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userword.UserWord), args.Error(1)
}

func (m *mockWordsService) Remove(ctx context.Context, userID, id uint64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

type mockStatsRecorder struct {
	mock.Mock
}

func (m *mockStatsRecorder) RecordQuizCompleted(ctx context.Context, userID uint64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
