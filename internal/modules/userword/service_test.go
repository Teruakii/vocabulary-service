package userword

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/modules/word"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAddWord_Success(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil)

	repo.On("Create", context.Background(), mock.MatchedBy(func(uw *UserWord) bool {
		return uw.UserID == 1 && uw.WordID == 42 && uw.Status == StatusNew && uw.EaseFactor == 2.5
	})).Return(nil)

	uw, err := svc.AddWord(context.Background(), 1, 42)

	assert.NoError(t, err)
	assert.NotNil(t, uw.NextReviewAt)
	repo.AssertExpectations(t)
}

func TestAddWord_Duplicate(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil)

	repo.On("Create", context.Background(), mock.Anything).Return(constants.ErrUserWordDuplicate)

	_, err := svc.AddWord(context.Background(), 1, 42)

	assert.ErrorIs(t, err, constants.ErrUserWordDuplicate)
}

func TestGetByID_WrongOwner_TreatedAsNotFound(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil)

	repo.On("GetByID", context.Background(), uint64(10)).
		Return(&UserWord{ID: 10, UserID: 2}, nil)

	_, err := svc.GetByID(context.Background(), 1, 10)

	assert.ErrorIs(t, err, constants.ErrUserWordNotFound)
}

func TestGetByID_ActuallyNotFound(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil)

	repo.On("GetByID", context.Background(), uint64(999)).Return(nil, constants.ErrNotFound)

	_, err := svc.GetByID(context.Background(), 1, 999)

	assert.ErrorIs(t, err, constants.ErrUserWordNotFound)
}

func TestList_DefaultsLimit(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil)

	repo.On("ListByUser", context.Background(), uint64(1), Status(""), (*uint64)(nil), constants.DefaultListLimit, 0).
		Return([]UserWordDetail{}, false, nil)

	_, _, err := svc.List(context.Background(), 1, "", nil, 0, -5)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestReview_InvalidQuality(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil)

	_, err := svc.Review(context.Background(), 1, 10, 6)

	assert.ErrorIs(t, err, constants.ErrInvalidQuality)
	repo.AssertNotCalled(t, "GetByID")
}

func TestReview_WrongOwner(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil)

	repo.On("GetByID", context.Background(), uint64(10)).
		Return(&UserWord{ID: 10, UserID: 2}, nil)

	_, err := svc.Review(context.Background(), 1, 10, 4)

	assert.ErrorIs(t, err, constants.ErrUserWordNotFound)
	repo.AssertNotCalled(t, "Update")
}

func TestReview_SM2Sequence(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil).(*service)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc.now = fixedNow(now)

	uw := &UserWord{ID: 10, UserID: 1, Status: StatusNew, EaseFactor: 2.5}

	repo.On("GetByID", context.Background(), uint64(10)).Return(uw, nil).Once()
	repo.On("Update", context.Background(), mock.MatchedBy(func(u *UserWord) bool {
		return u.IntervalDays == 1 && u.Repetitions == 1 && u.Status == StatusLearning
	})).Return(nil).Once()

	_, err := svc.Review(context.Background(), 1, 10, 4)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), uw.IntervalDays)
	assert.Equal(t, uint(1), uw.Repetitions)

	repo.On("GetByID", context.Background(), uint64(10)).Return(uw, nil).Once()
	repo.On("Update", context.Background(), mock.MatchedBy(func(u *UserWord) bool {
		return u.IntervalDays == 6 && u.Repetitions == 2
	})).Return(nil).Once()

	_, err = svc.Review(context.Background(), 1, 10, 4)
	assert.NoError(t, err)
	assert.Equal(t, uint(6), uw.IntervalDays)

	expectedEase := uw.EaseFactor
	repo.On("GetByID", context.Background(), uint64(10)).Return(uw, nil).Once()
	repo.On("Update", context.Background(), mock.Anything).Return(nil).Once()

	_, err = svc.Review(context.Background(), 1, 10, 4)
	assert.NoError(t, err)
	wantInterval := uint(6 * expectedEase)
	assert.InDelta(t, float64(wantInterval), float64(uw.IntervalDays), 1)

	repo.AssertExpectations(t)
}

func TestReview_Lapse_ResetsStreakButNotEaseFactor(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil).(*service)
	svc.now = fixedNow(time.Now())

	uw := &UserWord{ID: 10, UserID: 1, Repetitions: 3, IntervalDays: 15, EaseFactor: 2.6, Status: StatusLearning}
	easeBefore := uw.EaseFactor

	repo.On("GetByID", context.Background(), uint64(10)).Return(uw, nil).Once()
	repo.On("Update", context.Background(), mock.Anything).Return(nil).Once()

	_, err := svc.Review(context.Background(), 1, 10, 1)

	assert.NoError(t, err)
	assert.Equal(t, uint(0), uw.Repetitions)
	assert.Equal(t, uint(1), uw.IntervalDays)
	assert.Equal(t, StatusNew, uw.Status)

	assert.NotEqual(t, 2.5, uw.EaseFactor)
	assert.Less(t, uw.EaseFactor, easeBefore)
}

func TestReview_EaseFactorNeverBelowFloor(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil).(*service)
	svc.now = fixedNow(time.Now())

	uw := &UserWord{ID: 10, UserID: 1, EaseFactor: 1.3}

	repo.On("GetByID", context.Background(), uint64(10)).Return(uw, nil).Once()
	repo.On("Update", context.Background(), mock.Anything).Return(nil).Once()

	_, err := svc.Review(context.Background(), 1, 10, 0)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, uw.EaseFactor, 1.3)
}

func TestRemove_WrongOwner(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil)

	repo.On("GetByID", context.Background(), uint64(10)).
		Return(&UserWord{ID: 10, UserID: 2}, nil)

	err := svc.Remove(context.Background(), 1, 10)

	assert.ErrorIs(t, err, constants.ErrUserWordNotFound)
	repo.AssertNotCalled(t, "Delete")
}

func TestRemove_Success(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil)

	repo.On("GetByID", context.Background(), uint64(10)).
		Return(&UserWord{ID: 10, UserID: 1}, nil)
	repo.On("Delete", context.Background(), uint64(10)).Return(nil)

	err := svc.Remove(context.Background(), 1, 10)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

type fakeWordService struct {
	created *word.Word
	err     error
}

func (f *fakeWordService) AddWord(ctx context.Context, ownerUserID *uint64, w, meaning string) (*word.Word, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.created, nil
}
func (f *fakeWordService) GetByID(ctx context.Context, callerID, id uint64) (*word.Word, error) {
	return nil, nil
}
func (f *fakeWordService) GetByWord(ctx context.Context, callerID uint64, w string) (*word.Word, error) {
	return nil, nil
}
func (f *fakeWordService) GetAllWord(ctx context.Context, callerID uint64, limit, offset int) ([]word.Word, bool, error) {
	return nil, false, nil
}
func (f *fakeWordService) UpdateWord(ctx context.Context, callerID uint64, isAdmin bool, w *word.Word) error {
	return nil
}
func (f *fakeWordService) Delete(ctx context.Context, callerID uint64, isAdmin bool, id uint64) error {
	return nil
}

func TestAddCustomWord_CreatesWordThenLinksIt(t *testing.T) {
	repo := new(MockRepository)
	fw := &fakeWordService{created: &word.Word{ID: 55, Word: "serendipity", Meaning: "โชคดีโดยบังเอิญ"}}
	svc := NewService(repo, fw)

	repo.On("Create", context.Background(), mock.MatchedBy(func(uw *UserWord) bool {
		return uw.UserID == 1 && uw.WordID == 55 && uw.Status == StatusNew
	})).Return(nil)

	uw, err := svc.AddCustomWord(context.Background(), 1, "serendipity", "โชคดีโดยบังเอิญ")

	assert.NoError(t, err)
	assert.Equal(t, uint64(55), uw.WordID)
	repo.AssertExpectations(t)
}

func TestAddCustomWord_PropagatesWordCreationError(t *testing.T) {
	repo := new(MockRepository)
	fw := &fakeWordService{err: constants.ErrWordDuplicate}
	svc := NewService(repo, fw)

	_, err := svc.AddCustomWord(context.Background(), 1, "apple", "ผลไม้")

	assert.ErrorIs(t, err, constants.ErrWordDuplicate)
	repo.AssertNotCalled(t, "Create")
}
