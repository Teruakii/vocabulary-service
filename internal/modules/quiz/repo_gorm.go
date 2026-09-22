package quiz

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"

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
	if limit <= 0 {
		limit = constants.DefaultListLimit
	}
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

func (r *QuizRepoGorm) SaveAnswer(ctx context.Context, questionID uint64, userAnswer string, isCorrect bool) error {
	// WHERE is_correct IS NULL makes this an atomic "answer once" guard at the DB
	// level, so two concurrent requests for the same question can't both succeed
	// (which used to let the score get incremented twice).
	result := r.db.WithContext(ctx).Model(&Question{}).
		Where("quiz_question_id = ? AND is_correct IS NULL", questionID).
		Updates(map[string]interface{}{
			"user_answer": userAnswer,
			"is_correct":  isCorrect,
		})
	if result.Error != nil {
		return fmt.Errorf("save answer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		// Either the question doesn't exist, or it was already answered
		// (possibly by a concurrent request). The caller already checked
		// existence/ownership, so treat this as "already answered".
		return constants.ErrQuestionAlreadyAnswered
	}
	return nil
}

func (r *QuizRepoGorm) IncrementScore(ctx context.Context, quizID uint64) error {
	result := r.db.WithContext(ctx).Model(&Quiz{}).
		Where("quiz_id = ?", quizID).
		Update("score", gorm.Expr("score + 1"))
	if result.Error != nil {
		return fmt.Errorf("increment score: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}
	return nil
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

func (r *QuizRepoGorm) CompleteQuiz(ctx context.Context, quizID uint64) error {
	result := r.db.WithContext(ctx).Model(&Quiz{}).
		Where("quiz_id = ? AND completed_at IS NULL", quizID).
		Update("completed_at", time.Now())
	if result.Error != nil {
		return fmt.Errorf("complete quiz: %w", result.Error)
	}

	return nil
}

func (r *QuizRepoGorm) DistractorMeanings(ctx context.Context, callerID, excludeWordID uint64, n int) ([]string, error) {
	// Only pull distractors from words the caller can actually see (global
	// words + their own words) so another user's private word meaning can
	// never leak out as a multiple-choice option.
	var meanings []string
	err := r.db.WithContext(ctx).
		Table("word").
		Select("meaning").
		Where("word_id != ? AND (owner_user_id IS NULL OR owner_user_id = ?)", excludeWordID, callerID).
		Order("RAND()").
		Limit(n).
		Pluck("meaning", &meanings).Error
	if err != nil {
		return nil, fmt.Errorf("distractor meanings: %w", err)
	}
	return meanings, nil
}
