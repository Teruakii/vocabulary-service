package category

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"vocabulary-service/internal/constants"
)

func TestTag_Success(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	userID := uint64(1)

	repo.On("VocabBelongsToUser", context.Background(), uint64(10), userID).Return(true, nil)
	repo.On("GetByID", context.Background(), uint64(5)).Return(&Category{ID: 5, UserID: &userID}, nil)
	repo.On("CountTagged", context.Background(), uint64(5), userID).Return(3, nil)
	repo.On("Tag", context.Background(), uint64(10), uint64(5)).Return(nil)

	err := svc.Tag(context.Background(), userID, 10, 5)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestTag_DeckFull(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	userID := uint64(1)

	repo.On("VocabBelongsToUser", context.Background(), uint64(10), userID).Return(true, nil)
	repo.On("GetByID", context.Background(), uint64(5)).Return(&Category{ID: 5, UserID: &userID}, nil)
	repo.On("CountTagged", context.Background(), uint64(5), userID).Return(constants.MaxWordsPerCategory, nil)

	err := svc.Tag(context.Background(), userID, 10, 5)

	assert.ErrorIs(t, err, constants.ErrCategoryFull)
	repo.AssertNotCalled(t, "Tag")
}

func TestTag_VocabNotOwned(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("VocabBelongsToUser", context.Background(), uint64(10), uint64(1)).Return(false, nil)

	err := svc.Tag(context.Background(), 1, 10, 5)

	assert.ErrorIs(t, err, constants.ErrVocabNotFound)
	repo.AssertNotCalled(t, "CountTagged")
	repo.AssertNotCalled(t, "Tag")
}

func TestList_IncludesWordCount(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	userID := uint64(1)

	repo.On("ListVisibleToUser", context.Background(), userID).
		Return([]Category{{ID: 5, Name: "IELTS", UserID: &userID}}, nil)
	repo.On("CountTagged", context.Background(), uint64(5), userID).Return(12, nil)

	views, err := svc.List(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, 12, views[0].WordCount)
	assert.Equal(t, constants.MaxWordsPerCategory, views[0].MaxWords)
	repo.AssertExpectations(t)
}

func TestTag_GlobalCategoryLimitIsPerUser(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	userID := uint64(1)

	repo.On("VocabBelongsToUser", context.Background(), uint64(10), userID).Return(true, nil)
	repo.On("GetByID", context.Background(), uint64(5)).Return(&Category{ID: 5}, nil)
	repo.On("CountTagged", context.Background(), uint64(5), userID).Return(0, nil)
	repo.On("Tag", context.Background(), uint64(10), uint64(5)).Return(nil)

	err := svc.Tag(context.Background(), userID, 10, 5)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestList_GlobalCategoryCountsOnlyCallersWords(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	userID := uint64(1)

	repo.On("ListVisibleToUser", context.Background(), userID).Return([]Category{{ID: 5, Name: "Food"}}, nil)
	repo.On("CountTagged", context.Background(), uint64(5), userID).Return(2, nil)

	views, err := svc.List(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, 2, views[0].WordCount)
	repo.AssertExpectations(t)
}
