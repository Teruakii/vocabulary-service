package userword

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"vocabulary-service/internal/constants"
)

type UserWordRepoGorm struct {
	db *gorm.DB
}

func NewUserWordRepoGorm(db *gorm.DB) *UserWordRepoGorm {
	return &UserWordRepoGorm{db: db}
}

func (r *UserWordRepoGorm) Create(ctx context.Context, uw *UserWord) error {
	if err := r.db.WithContext(ctx).Create(uw).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return constants.ErrUserWordDuplicate
		}
		return fmt.Errorf("create user_word: %w", err)
	}
	return nil
}
func (r *UserWordRepoGorm) GetByID(ctx context.Context, id uint64) (*UserWord, error) {
	var uw UserWord
	if err := r.db.WithContext(ctx).First(&uw, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get user_word by id: %w", err)
	}
	return &uw, nil
}

const detailSelect = "user_word.*, word.word AS word, word.meaning AS meaning"

func (r *UserWordRepoGorm) ListByUser(ctx context.Context, userID uint64, status Status, categoryID *uint64, limit, offset int) ([]UserWordDetail, bool, error) {
	if limit <= 0 {
		limit = constants.DefaultListLimit
	}
	fetchLimit := limit + 1

	q := r.db.WithContext(ctx).
		Table("user_word").
		Select(detailSelect).
		Joins("JOIN word ON word.word_id = user_word.word_id").
		Where("user_word.user_id = ?", userID)
	if status != "" {
		q = q.Where("user_word.status = ?", status)
	}
	if categoryID != nil {
		q = q.Joins("JOIN vocab_tags ON vocab_tags.vocab_id = user_word.user_word_id AND vocab_tags.category_id = ?", *categoryID)
	}

	var items []UserWordDetail
	err := q.Order("user_word.user_word_id ASC").
		Limit(fetchLimit).
		Offset(offset).
		Find(&items).Error
	if err != nil {
		return nil, false, fmt.Errorf("list user_word: %w", err)
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	return items, hasMore, nil
}

func (r *UserWordRepoGorm) ListDue(ctx context.Context, userID uint64, before time.Time, limit int) ([]UserWordDetail, error) {
	if limit <= 0 {
		limit = constants.DefaultListLimit
	}

	var items []UserWordDetail
	err := r.db.WithContext(ctx).
		Table("user_word").
		Select(detailSelect).
		Joins("JOIN word ON word.word_id = user_word.word_id").
		Where("user_word.user_id = ? AND user_word.next_review_at <= ?", userID, before).
		Order("user_word.next_review_at ASC").
		Limit(limit).
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("list due user_word: %w", err)
	}
	return items, nil
}

func (r *UserWordRepoGorm) ReviewByID(ctx context.Context, userID, id uint64, apply func(*UserWord)) (*UserWord, error) {
	return r.review(ctx, "user_word_id = ? AND user_id = ?", id, userID, apply)
}

func (r *UserWordRepoGorm) ReviewByWord(ctx context.Context, userID, wordID uint64, apply func(*UserWord)) (*UserWord, error) {
	return r.review(ctx, "word_id = ? AND user_id = ?", wordID, userID, apply)
}

func (r *UserWordRepoGorm) review(ctx context.Context, where string, key, userID uint64, apply func(*UserWord)) (*UserWord, error) {
	var uw UserWord
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(where, key, userID).First(&uw).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return constants.ErrNotFound
			}
			return fmt.Errorf("lock user_word: %w", err)
		}
		apply(&uw)
		if err := tx.Model(&uw).
			Select("Status", "EaseFactor", "IntervalDays", "Repetitions", "NextReviewAt", "LastReviewedAt").
			Updates(&uw).Error; err != nil {
			return fmt.Errorf("save user_word review: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &uw, nil
}

// Delete also drops the word's category tags; left behind they would still
// count toward the category's word limit.
func (r *UserWordRepoGorm) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM vocab_tags WHERE vocab_id = ?", id).Error; err != nil {
			return fmt.Errorf("delete user_word tags: %w", err)
		}
		result := tx.Delete(&UserWord{}, id)
		if result.Error != nil {
			return fmt.Errorf("delete user_word: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return constants.ErrNotFound
		}
		return nil
	})
}
