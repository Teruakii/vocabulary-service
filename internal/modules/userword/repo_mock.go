package userword

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, uw *UserWord) error {
	args := m.Called(ctx, uw)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uint64) (*UserWord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserWord), args.Error(1)
}

func (m *MockRepository) ListByUser(ctx context.Context, userID uint64, status Status, categoryID *uint64, limit, offset int) ([]UserWordDetail, bool, error) {
	args := m.Called(ctx, userID, status, categoryID, limit, offset)
	var items []UserWordDetail
	if args.Get(0) != nil {
		items = args.Get(0).([]UserWordDetail)
	}
	return items, args.Bool(1), args.Error(2)
}

func (m *MockRepository) ListDue(ctx context.Context, userID uint64, before time.Time, limit int) ([]UserWordDetail, error) {
	args := m.Called(ctx, userID, before, limit)
	var items []UserWordDetail
	if args.Get(0) != nil {
		items = args.Get(0).([]UserWordDetail)
	}
	return items, args.Error(1)
}

// ReviewByID and ReviewByWord apply the review to the row the test returns,
// as the repo does to the locked row.
func (m *MockRepository) ReviewByID(ctx context.Context, userID, id uint64, apply func(*UserWord)) (*UserWord, error) {
	return reviewed(m.Called(ctx, userID, id), apply)
}

func (m *MockRepository) ReviewByWord(ctx context.Context, userID, wordID uint64, apply func(*UserWord)) (*UserWord, error) {
	return reviewed(m.Called(ctx, userID, wordID), apply)
}

func reviewed(args mock.Arguments, apply func(*UserWord)) (*UserWord, error) {
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	uw := args.Get(0).(*UserWord)
	apply(uw)
	return uw, args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
