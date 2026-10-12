package word

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, word *Word) error {
	args := m.Called(ctx, word)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uint64) (*Word, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Word), args.Error(1)
}

func (m *MockRepository) GetByWord(ctx context.Context, callerID uint64, wordStr string) (*Word, error) {
	args := m.Called(ctx, callerID, wordStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Word), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, callerID uint64, limit, offset int) ([]Word, bool, error) {
	args := m.Called(ctx, callerID, limit, offset)
	var words []Word
	if args.Get(0) != nil {
		words = args.Get(0).([]Word)
	}
	return words, args.Bool(1), args.Error(2)
}

func (m *MockRepository) ExistsForOwner(ctx context.Context, ownerUserID *uint64, wordStr string, excludeID uint64) (bool, error) {
	args := m.Called(ctx, ownerUserID, wordStr, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, word *Word) error {
	args := m.Called(ctx, word)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
