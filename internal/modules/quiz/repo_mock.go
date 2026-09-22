package quiz

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateWithQuestions(ctx context.Context, userID uint64, n int, source Source) (*Quiz, []questionWithWord, error) {
	args := m.Called(ctx, userID, n, source)
	var q *Quiz
	if args.Get(0) != nil {
		q = args.Get(0).(*Quiz)
	}
	var questions []questionWithWord
	if args.Get(1) != nil {
		questions = args.Get(1).([]questionWithWord)
	}
	return q, questions, args.Error(2)
}

func (m *MockRepository) GetQuiz(ctx context.Context, id uint64) (*Quiz, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Quiz), args.Error(1)
}

func (m *MockRepository) GetQuestion(ctx context.Context, id uint64) (*questionWithWord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*questionWithWord), args.Error(1)
}

func (m *MockRepository) ListQuestions(ctx context.Context, quizID uint64) ([]questionWithWord, error) {
	args := m.Called(ctx, quizID)
	var qs []questionWithWord
	if args.Get(0) != nil {
		qs = args.Get(0).([]questionWithWord)
	}
	return qs, args.Error(1)
}

func (m *MockRepository) ListQuizzesByUser(ctx context.Context, userID uint64, limit, offset int) ([]Quiz, bool, error) {
	args := m.Called(ctx, userID, limit, offset)
	var qs []Quiz
	if args.Get(0) != nil {
		qs = args.Get(0).([]Quiz)
	}
	return qs, args.Bool(1), args.Error(2)
}

func (m *MockRepository) SaveAnswer(ctx context.Context, questionID uint64, userAnswer string, isCorrect bool) error {
	args := m.Called(ctx, questionID, userAnswer, isCorrect)
	return args.Error(0)
}

func (m *MockRepository) IncrementScore(ctx context.Context, quizID uint64) error {
	args := m.Called(ctx, quizID)
	return args.Error(0)
}

func (m *MockRepository) CountUnanswered(ctx context.Context, quizID uint64) (int, error) {
	args := m.Called(ctx, quizID)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) CompleteQuiz(ctx context.Context, quizID uint64) error {
	args := m.Called(ctx, quizID)
	return args.Error(0)
}

func (m *MockRepository) DistractorMeanings(ctx context.Context, callerID, excludeWordID uint64, n int) ([]string, error) {
	args := m.Called(ctx, callerID, excludeWordID, n)
	var meanings []string
	if args.Get(0) != nil {
		meanings = args.Get(0).([]string)
	}
	return meanings, args.Error(1)
}
