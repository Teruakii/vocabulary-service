package quiz

import "context"

type Repository interface {
	CreateWithQuestions(ctx context.Context, userID uint64, n int, source Source) (*Quiz, []questionWithWord, error)

	GetQuiz(ctx context.Context, id uint64) (*Quiz, error)
	GetQuestion(ctx context.Context, id uint64) (*questionWithWord, error)
	ListQuestions(ctx context.Context, quizID uint64) ([]questionWithWord, error)
	ListQuizzesByUser(ctx context.Context, userID uint64, limit, offset int) ([]Quiz, bool, error)
	SaveAnswer(ctx context.Context, questionID uint64, userAnswer string, isCorrect bool) error
	IncrementScore(ctx context.Context, quizID uint64) error
	CountUnanswered(ctx context.Context, quizID uint64) (int, error)
	CompleteQuiz(ctx context.Context, quizID uint64) error
	DistractorMeanings(ctx context.Context, callerID, excludeWordID uint64, n int) ([]string, error)
}

type Service interface {
	Create(ctx context.Context, userID uint64, totalQuestions int, source Source) (*QuizDetail, error)
	Get(ctx context.Context, userID, quizID uint64) (*QuizDetail, error)
	List(ctx context.Context, userID uint64, limit, offset int) ([]QuizSummary, bool, error)
	Answer(ctx context.Context, userID, quizID, questionID uint64, answer string) (*QuestionView, error)
	Complete(ctx context.Context, userID, quizID uint64) (*QuizDetail, error)
}
