package word

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"vocabulary-service/internal/constants"
)

func u64(v uint64) *uint64 { return &v }

func TestAddWord_Personal(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	owner := u64(7)

	repo.On("ExistsForOwner", context.Background(), owner, "apple", uint64(0)).Return(false, nil)
	repo.On("Create", context.Background(), &Word{Word: "apple", Meaning: "a fruit", OwnerUserID: owner}).Return(nil)

	w, err := svc.AddWord(context.Background(), owner, "apple", "a fruit")

	assert.NoError(t, err)
	assert.Equal(t, "apple", w.Word)
	assert.False(t, w.IsGlobal())
	repo.AssertExpectations(t)
}

func TestAddWord_Global_Duplicate(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("ExistsForOwner", context.Background(), (*uint64)(nil), "apple", uint64(0)).Return(true, nil)

	_, err := svc.AddWord(context.Background(), nil, "apple", "a fruit")

	assert.ErrorIs(t, err, constants.ErrWordDuplicate)
	repo.AssertNotCalled(t, "Create")
}

func TestAddWord_EmptyAfterTrim(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	_, err := svc.AddWord(context.Background(), nil, "   ", "a fruit")

	assert.Error(t, err)
	repo.AssertNotCalled(t, "Create")
}

func TestGetByID_NotFound(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("GetByID", context.Background(), uint64(99)).Return(nil, constants.ErrNotFound)

	_, err := svc.GetByID(context.Background(), 1, 99)

	assert.ErrorIs(t, err, constants.ErrWordNotFound)
	repo.AssertExpectations(t)
}

func TestGetByID_ForbiddenForOtherUsersPersonalWord(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("GetByID", context.Background(), uint64(1)).Return(&Word{ID: 1, Word: "apple", OwnerUserID: u64(7)}, nil)

	_, err := svc.GetByID(context.Background(), 8, 1)

	assert.ErrorIs(t, err, constants.ErrWordNotFound)
	repo.AssertExpectations(t)
}

func TestGetByWord_Found(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	want := &Word{ID: 1, Word: "apple", Meaning: "a fruit"}
	repo.On("GetByWord", context.Background(), uint64(1), "apple").Return(want, nil)

	w, err := svc.GetByWord(context.Background(), 1, "apple")

	assert.NoError(t, err)
	assert.Equal(t, want, w)
	repo.AssertExpectations(t)
}

func TestGetByWord_NotFound(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("GetByWord", context.Background(), uint64(1), "banana").Return(nil, constants.ErrNotFound)

	_, err := svc.GetByWord(context.Background(), 1, "banana")

	assert.ErrorIs(t, err, constants.ErrWordNotFound)
	repo.AssertExpectations(t)
}

func TestGetByWord_EmptyAfterTrim(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	_, err := svc.GetByWord(context.Background(), 1, "   ")

	assert.Error(t, err)
	repo.AssertNotCalled(t, "GetByWord")
}

func TestGetAllWord_HasMore(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	want := []Word{{ID: 1, Word: "apple"}, {ID: 2, Word: "banana"}}
	repo.On("List", context.Background(), uint64(1), 2, 0).Return(want, true, nil)

	words, hasMore, err := svc.GetAllWord(context.Background(), 1, 2, 0)

	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, words, 2)
	repo.AssertExpectations(t)
}

func TestGetAllWord_NoMore(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	want := []Word{{ID: 1, Word: "apple"}}
	repo.On("List", context.Background(), uint64(1), 20, 0).Return(want, false, nil)

	words, hasMore, err := svc.GetAllWord(context.Background(), 1, 0, 0)

	assert.NoError(t, err)
	assert.False(t, hasMore)
	assert.Len(t, words, 1)
	repo.AssertExpectations(t)
}

func TestUpdateWord_PartialUpdate_SkipsEmptyField(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	w := &Word{ID: 1, Meaning: "ผลไม้"}
	repo.On("GetByID", context.Background(), uint64(1)).Return(&Word{ID: 1, OwnerUserID: u64(9)}, nil)
	repo.On("Update", context.Background(), w).Return(nil)

	err := svc.UpdateWord(context.Background(), 9, false, w)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateWord_WhitespaceOnly_Rejected(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	w := &Word{ID: 1, Word: "   "}

	err := svc.UpdateWord(context.Background(), 1, false, w)

	assert.Error(t, err)
	repo.AssertNotCalled(t, "Update")
}

func TestUpdateWord_NotFound(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	w := &Word{ID: 99, Word: "apple"}
	repo.On("GetByID", context.Background(), uint64(99)).Return(nil, constants.ErrNotFound)

	err := svc.UpdateWord(context.Background(), 1, false, w)

	assert.ErrorIs(t, err, constants.ErrWordNotFound)
	repo.AssertExpectations(t)
}

func TestUpdateWord_Forbidden_NotOwner(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	w := &Word{ID: 1, Word: "apple"}
	repo.On("GetByID", context.Background(), uint64(1)).Return(&Word{ID: 1, OwnerUserID: u64(7)}, nil)

	err := svc.UpdateWord(context.Background(), 8, false, w)

	assert.ErrorIs(t, err, constants.ErrWordForbidden)
	repo.AssertNotCalled(t, "Update")
}

func TestUpdateWord_GlobalWord_RequiresAdmin(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	w := &Word{ID: 1, Word: "apple"}
	repo.On("GetByID", context.Background(), uint64(1)).Return(&Word{ID: 1, OwnerUserID: nil}, nil)

	err := svc.UpdateWord(context.Background(), 8, false, w)

	assert.ErrorIs(t, err, constants.ErrWordForbidden)
	repo.AssertNotCalled(t, "Update")
}

func TestDelete_InUse(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("GetByID", context.Background(), uint64(1)).Return(&Word{ID: 1, OwnerUserID: u64(1)}, nil)
	repo.On("Delete", context.Background(), uint64(1)).Return(constants.ErrWordInUse)

	err := svc.Delete(context.Background(), 1, false, 1)

	assert.ErrorIs(t, err, constants.ErrWordInUse)
	repo.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("GetByID", context.Background(), uint64(99)).Return(nil, constants.ErrNotFound)

	err := svc.Delete(context.Background(), 1, false, 99)

	assert.ErrorIs(t, err, constants.ErrWordNotFound)
	repo.AssertExpectations(t)
}

func TestUpdateWord_RenameToExistingWord_Duplicate(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	owner := u64(9)
	repo.On("GetByID", context.Background(), uint64(1)).Return(&Word{ID: 1, Word: "aple", OwnerUserID: owner}, nil)
	repo.On("ExistsForOwner", context.Background(), owner, "apple", uint64(1)).Return(true, nil)

	err := svc.UpdateWord(context.Background(), 9, false, &Word{ID: 1, Word: "apple"})

	assert.ErrorIs(t, err, constants.ErrWordDuplicate)
	repo.AssertNotCalled(t, "Update")
}

func TestUpdateWord_Rename_ChecksOwnersWordsExcludingItself(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	w := &Word{ID: 1, Word: "Apple"}
	repo.On("GetByID", context.Background(), uint64(1)).Return(&Word{ID: 1, Word: "apple", OwnerUserID: nil}, nil)
	repo.On("ExistsForOwner", context.Background(), (*uint64)(nil), "Apple", uint64(1)).Return(false, nil)
	repo.On("Update", context.Background(), w).Return(nil)

	err := svc.UpdateWord(context.Background(), 8, true, w)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetAllWord_CapsLimit(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("List", context.Background(), uint64(1), constants.MaxListLimit, 0).Return([]Word{}, false, nil)

	_, _, err := svc.GetAllWord(context.Background(), 1, 1_000_000, 0)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
