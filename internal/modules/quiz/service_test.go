package quiz

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/modules/userword"
)

func newTestService() (*MockRepository, *mockWordsService, *mockStatsRecorder, Service) {
	repo := new(MockRepository)
	words := new(mockWordsService)
	stats := new(mockStatsRecorder)
	svc := NewService(repo, words, stats)
	return repo, words, stats, svc
}

func TestCreate_InvalidSize(t *testing.T) {
	repo, _, _, svc := newTestService()

	_, err := svc.Create(context.Background(), 1, 0, SourceDue)

	assert.ErrorIs(t, err, constants.ErrInvalidQuizSize)
	repo.AssertNotCalled(t, "CreateWithQuestions")
}

func TestCreate_InvalidSource(t *testing.T) {
	repo, _, _, svc := newTestService()

	_, err := svc.Create(context.Background(), 1, 5, Source("bogus"))

	assert.ErrorIs(t, err, constants.ErrInvalidQuizSource)
	repo.AssertNotCalled(t, "CreateWithQuestions")
}

func TestCreate_NoWordsAvailable(t *testing.T) {
	repo, _, _, svc := newTestService()

	repo.On("CreateWithQuestions", context.Background(), uint64(1), 5, SourceDue).
		Return(nil, nil, constants.ErrNoWordsAvailable)

	_, err := svc.Create(context.Background(), 1, 5, SourceDue)

	assert.ErrorIs(t, err, constants.ErrNoWordsAvailable)
}

func TestCreate_DefaultsToSourceDue(t *testing.T) {
	repo, _, _, svc := newTestService()

	q := &Quiz{ID: 1, TotalQuestions: 1}
	qs := []questionWithWord{{
		Question: Question{ID: 100, QuizID: 1, WordID: 42, QuestionType: QuestionTyped},
		Word:     "apple", Meaning: "a fruit",
	}}
	repo.On("CreateWithQuestions", context.Background(), uint64(1), 5, SourceDue).Return(q, qs, nil)

	detail, err := svc.Create(context.Background(), 1, 5, "")

	assert.NoError(t, err)
	assert.Len(t, detail.Questions, 1)
	assert.False(t, detail.Questions[0].Answered)
	assert.Nil(t, detail.Questions[0].CorrectAnswer)
	repo.AssertExpectations(t)
}

func TestGet_WrongOwner(t *testing.T) {
	repo, _, _, svc := newTestService()

	repo.On("GetQuiz", context.Background(), uint64(1)).Return(&Quiz{ID: 1, UserID: 2}, nil)

	_, err := svc.Get(context.Background(), 1, 1)

	assert.ErrorIs(t, err, constants.ErrQuizNotFound)
	repo.AssertNotCalled(t, "ListQuestions")
}

func TestAnswer_Correct_IncrementsScoreAndFeedsSRS_NoAutoComplete(t *testing.T) {
	repo, words, stats, svc := newTestService()

	repo.On("GetQuiz", context.Background(), uint64(1)).Return(&Quiz{ID: 1, UserID: 1}, nil)
	repo.On("GetQuestion", context.Background(), uint64(100)).Return(&questionWithWord{
		Question: Question{ID: 100, QuizID: 1, WordID: 42, QuestionType: QuestionTyped},
		Word:     "apple", Meaning: "a fruit",
	}, nil)
	repo.On("SaveAnswer", context.Background(), uint64(100), "a fruit", true).Return(nil)
	repo.On("IncrementScore", context.Background(), uint64(1)).Return(nil)
	repo.On("CountUnanswered", context.Background(), uint64(1)).Return(2, nil)
	words.On("ReviewByWord", context.Background(), uint64(1), uint64(42), 4).
		Return(&userword.UserWord{}, nil)

	view, err := svc.Answer(context.Background(), 1, 1, 100, "  A Fruit ")

	assert.NoError(t, err)
	assert.True(t, view.Answered)
	assert.True(t, *view.IsCorrect)
	repo.AssertNotCalled(t, "CompleteQuiz")
	stats.AssertNotCalled(t, "RecordQuizCompleted")
	repo.AssertExpectations(t)
	words.AssertExpectations(t)
}

func TestAnswer_Incorrect_AutoCompletesAndRecordsStats(t *testing.T) {
	repo, words, stats, svc := newTestService()

	repo.On("GetQuiz", context.Background(), uint64(1)).Return(&Quiz{ID: 1, UserID: 1}, nil)
	repo.On("GetQuestion", context.Background(), uint64(100)).Return(&questionWithWord{
		Question: Question{ID: 100, QuizID: 1, WordID: 42},
		Word:     "apple", Meaning: "a fruit",
	}, nil)
	repo.On("SaveAnswer", context.Background(), uint64(100), "a vegetable", false).Return(nil)
	repo.On("CountUnanswered", context.Background(), uint64(1)).Return(0, nil)
	repo.On("CompleteQuiz", context.Background(), uint64(1)).Return(nil)
	stats.On("RecordQuizCompleted", context.Background(), uint64(1)).Return(nil)
	words.On("ReviewByWord", context.Background(), uint64(1), uint64(42), 1).
		Return(&userword.UserWord{}, nil)

	view, err := svc.Answer(context.Background(), 1, 1, 100, "a vegetable")

	assert.NoError(t, err)
	assert.False(t, *view.IsCorrect)
	repo.AssertNotCalled(t, "IncrementScore")
	repo.AssertExpectations(t)
	stats.AssertExpectations(t)
}

func TestAnswer_QuizAlreadyCompleted(t *testing.T) {
	repo, _, _, svc := newTestService()

	now := time.Now()
	repo.On("GetQuiz", context.Background(), uint64(1)).Return(&Quiz{ID: 1, UserID: 1, CompletedAt: &now}, nil)

	_, err := svc.Answer(context.Background(), 1, 1, 100, "anything")

	assert.ErrorIs(t, err, constants.ErrQuizCompleted)
	repo.AssertNotCalled(t, "GetQuestion")
}

func TestAnswer_QuestionFromDifferentQuiz(t *testing.T) {
	repo, _, _, svc := newTestService()

	repo.On("GetQuiz", context.Background(), uint64(1)).Return(&Quiz{ID: 1, UserID: 1}, nil)
	repo.On("GetQuestion", context.Background(), uint64(100)).Return(&questionWithWord{
		Question: Question{ID: 100, QuizID: 999},
	}, nil)

	_, err := svc.Answer(context.Background(), 1, 1, 100, "anything")

	assert.ErrorIs(t, err, constants.ErrQuestionNotFound)
	repo.AssertNotCalled(t, "SaveAnswer")
}

func TestAnswer_AlreadyAnswered(t *testing.T) {
	repo, _, _, svc := newTestService()

	correct := true
	repo.On("GetQuiz", context.Background(), uint64(1)).Return(&Quiz{ID: 1, UserID: 1}, nil)
	repo.On("GetQuestion", context.Background(), uint64(100)).Return(&questionWithWord{
		Question: Question{ID: 100, QuizID: 1, IsCorrect: &correct},
	}, nil)

	_, err := svc.Answer(context.Background(), 1, 1, 100, "anything")

	assert.ErrorIs(t, err, constants.ErrQuestionAlreadyAnswered)
	repo.AssertNotCalled(t, "SaveAnswer")
}

func TestAnswer_SRSFailureDoesNotFailTheAnswer(t *testing.T) {
	repo, words, _, svc := newTestService()

	repo.On("GetQuiz", context.Background(), uint64(1)).Return(&Quiz{ID: 1, UserID: 1}, nil)
	repo.On("GetQuestion", context.Background(), uint64(100)).Return(&questionWithWord{
		Question: Question{ID: 100, QuizID: 1, WordID: 42},
		Meaning:  "a fruit",
	}, nil)
	repo.On("SaveAnswer", context.Background(), uint64(100), "a fruit", true).Return(nil)
	repo.On("IncrementScore", context.Background(), uint64(1)).Return(nil)
	repo.On("CountUnanswered", context.Background(), uint64(1)).Return(1, nil)
	words.On("ReviewByWord", context.Background(), uint64(1), uint64(42), mock.Anything).
		Return(nil, constants.ErrUserWordNotFound)

	view, err := svc.Answer(context.Background(), 1, 1, 100, "a fruit")

	assert.NoError(t, err)
	assert.True(t, view.Answered)
}

func TestComplete_AlreadyCompleted_NoOp(t *testing.T) {
	repo, _, stats, svc := newTestService()

	now := time.Now()
	repo.On("GetQuiz", context.Background(), uint64(1)).Return(&Quiz{ID: 1, UserID: 1, CompletedAt: &now}, nil)
	repo.On("ListQuestions", context.Background(), uint64(1)).Return([]questionWithWord{}, nil)

	_, err := svc.Complete(context.Background(), 1, 1)

	assert.NoError(t, err)
	repo.AssertNotCalled(t, "CompleteQuiz")
	stats.AssertNotCalled(t, "RecordQuizCompleted")
}

func TestComplete_TransitionsAndRecordsStats(t *testing.T) {
	repo, _, stats, svc := newTestService()

	repo.On("GetQuiz", context.Background(), uint64(1)).Return(&Quiz{ID: 1, UserID: 1}, nil)
	repo.On("CompleteQuiz", context.Background(), uint64(1)).Return(nil)
	stats.On("RecordQuizCompleted", context.Background(), uint64(1)).Return(nil)
	repo.On("ListQuestions", context.Background(), uint64(1)).Return([]questionWithWord{}, nil)

	_, err := svc.Complete(context.Background(), 1, 1)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	stats.AssertExpectations(t)
}
