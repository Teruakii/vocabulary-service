package quiz

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"vocabulary-service/internal/constants"
)

type QuizRepoGorm struct {
	db *gorm.DB
}

func NewQuizRepoGorm(db *gorm.DB) *QuizRepoGorm {
	return &QuizRepoGorm{db: db}
}

func (r *QuizRepoGorm) CreateWithQuestions(ctx context.Context, userID uint64, n int, source Source) (*Quiz, []questionWithWord, error) {
	var quizOut *Quiz
	var questionsOut []questionWithWord

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wordIDs, err := selectWordIDs(tx, userID, n, source)
		if err != nil {
			return err
		}
		if len(wordIDs) == 0 {
			return constants.ErrNoWordsAvailable
		}

		q := &Quiz{UserID: userID, TotalQuestions: uint(len(wordIDs))}
		if err := tx.Create(q).Error; err != nil {
			return fmt.Errorf("create quiz: %w", err)
		}

		questions := make([]Question, len(wordIDs))
		for i, wordID := range wordIDs {
			qt := QuestionTyped
			if rand.Intn(2) == 1 {
				qt = QuestionMultipleChoice
			}
			questions[i] = Question{QuizID: q.ID, WordID: wordID, QuestionType: qt}
		}
		if err := tx.Create(&questions).Error; err != nil {
			return fmt.Errorf("create quiz questions: %w", err)
		}

		var joined []questionWithWord
		err = tx.Table("quiz_question").
			Select("quiz_question.*, word.word AS word, word.meaning AS meaning").
			Joins("JOIN word ON word.word_id = quiz_question.word_id").
			Where("quiz_question.quiz_id = ?", q.ID).
			Find(&joined).Error
		if err != nil {
			return fmt.Errorf("reload quiz questions: %w", err)
		}

		quizOut = q
		questionsOut = joined
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return quizOut, questionsOut, nil
}

func selectWordIDs(tx *gorm.DB, userID uint64, n int, source Source) ([]uint64, error) {
	q := tx.Table("user_word").
		Select("word_id").
		Where("user_id = ?", userID)
	if source == SourceDue {
		q = q.Where("next_review_at <= ?", time.Now())
	}

	var ids []uint64
	if err := q.Order("RAND()").Limit(n).Pluck("word_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("select candidate words: %w", err)
	}
	return ids, nil
}

func (r *QuizRepoGorm) GetQuiz(ctx context.Context, id uint64) (*Quiz, error) {
	var q Quiz
	if err := r.db.WithContext(ctx).First(&q, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get quiz by id: %w", err)
	}
	return &q, nil
}

func (r *QuizRepoGorm) GetQuestion(ctx context.Context, id uint64) (*questionWithWord, error) {
	var q questionWithWord
	err := r.db.WithContext(ctx).
		Table("quiz_question").
		Select("quiz_question.*, word.word AS word, word.meaning AS meaning").
		Joins("JOIN word ON word.word_id = quiz_question.word_id").
		Where("quiz_question.quiz_question_id = ?", id).
		First(&q).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get question by id: %w", err)
	}
	return &q, nil
}

func (r *QuizRepoGorm) ListQuestions(ctx context.Context, quizID uint64) ([]questionWithWord, error) {
	var qs []questionWithWord
	err := r.db.WithContext(ctx).
		Table("quiz_question").
		Select("quiz_question.*, word.word AS word, word.meaning AS meaning").
		Joins("JOIN word ON word.word_id = quiz_question.word_id").
		Where("quiz_question.quiz_id = ?", quizID).
		Order("quiz_question.quiz_question_id ASC").
		Find(&qs).Error
	if err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}
	return qs, nil
}

func (r *QuizRepoGorm) ListQuizzesByUser(ctx context.Context, userID uint64, limit, offset int) ([]Quiz, bool, error) {
	fetchLimit := limit + 1

	var quizzes []Quiz
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("quiz_id DESC").
		Limit(fetchLimit).
		Offset(offset).
		Find(&quizzes).Error
	if err != nil {
		return nil, false, fmt.Errorf("list quizzes: %w", err)
	}

	hasMore := len(quizzes) > limit
	if hasMore {
		quizzes = quizzes[:limit]
	}
	return quizzes, hasMore, nil
}

// SaveAnswer records the answer and its point in one transaction. The quiz
// row is locked first, so an answer cannot land after CompleteQuiz has
// finished the quiz, and only an unanswered question is written, so two
// concurrent answers cannot both score.
func (r *QuizRepoGorm) SaveAnswer(ctx context.Context, quizID, questionID uint64, userAnswer string, isCorrect bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var q Quiz
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&q, quizID).Error; err != nil {
			return fmt.Errorf("lock quiz: %w", err)
		}
		if q.IsCompleted() {
			return constants.ErrQuizCompleted
		}

		result := tx.Model(&Question{}).
			Where("quiz_question_id = ? AND quiz_id = ? AND is_correct IS NULL", questionID, quizID).
			Updates(map[string]interface{}{
				"user_answer": userAnswer,
				"is_correct":  isCorrect,
			})
		if result.Error != nil {
			return fmt.Errorf("save answer: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return constants.ErrQuestionAlreadyAnswered
		}

		if isCorrect {
			if err := tx.Model(&Quiz{}).Where("quiz_id = ?", quizID).
				Update("score", gorm.Expr("score + 1")).Error; err != nil {
				return fmt.Errorf("increment score: %w", err)
			}
		}
		return nil
	})
}

func (r *QuizRepoGorm) CountUnanswered(ctx context.Context, quizID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Question{}).
		Where("quiz_id = ? AND is_correct IS NULL", quizID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count unanswered: %w", err)
	}
	return int(count), nil
}

// CompleteQuiz reports whether this call completed the quiz; false means it
// was already completed.
func (r *QuizRepoGorm) CompleteQuiz(ctx context.Context, quizID uint64) (bool, error) {
	result := r.db.WithContext(ctx).Model(&Quiz{}).
		Where("quiz_id = ? AND completed_at IS NULL", quizID).
		Update("completed_at", time.Now())
	if result.Error != nil {
		return false, fmt.Errorf("complete quiz: %w", result.Error)
	}
	return result.RowsAffected > 0, nil
}

// DistractorMeanings picks wrong answers only from words the user can see
// (global ones and their own), never repeating the correct meaning.
func (r *QuizRepoGorm) DistractorMeanings(ctx context.Context, userID, excludeWordID uint64, correctMeaning string, n int) ([]string, error) {
	var meanings []string
	err := r.db.WithContext(ctx).
		Table("word").
		Distinct("meaning").
		Where("word_id != ? AND meaning != ?", excludeWordID, correctMeaning).
		Where("owner_user_id IS NULL OR owner_user_id = ?", userID).
		Order("RAND()").
		Limit(n).
		Pluck("meaning", &meanings).Error
	if err != nil {
		return nil, fmt.Errorf("distractor meanings: %w", err)
	}
	return meanings, nil
}
