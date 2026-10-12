package category

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, c *Category) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uint64) (*Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Category), args.Error(1)
}

func (m *MockRepository) ListVisibleToUser(ctx context.Context, userID uint64) ([]Category, error) {
	args := m.Called(ctx, userID)
	var items []Category
	if args.Get(0) != nil {
		items = args.Get(0).([]Category)
	}
	return items, args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) Tag(ctx context.Context, vocabID, categoryID uint64) error {
	args := m.Called(ctx, vocabID, categoryID)
	return args.Error(0)
}

func (m *MockRepository) Untag(ctx context.Context, vocabID, categoryID uint64) error {
	args := m.Called(ctx, vocabID, categoryID)
	return args.Error(0)
}

func (m *MockRepository) ListCategoriesForVocab(ctx context.Context, vocabID uint64) ([]Category, error) {
	args := m.Called(ctx, vocabID)
	var items []Category
	if args.Get(0) != nil {
		items = args.Get(0).([]Category)
	}
	return items, args.Error(1)
}

func (m *MockRepository) VocabBelongsToUser(ctx context.Context, vocabID, userID uint64) (bool, error) {
	args := m.Called(ctx, vocabID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) CountTagged(ctx context.Context, categoryID, userID uint64) (int, error) {
	args := m.Called(ctx, categoryID, userID)
	return args.Int(0), args.Error(1)
}
