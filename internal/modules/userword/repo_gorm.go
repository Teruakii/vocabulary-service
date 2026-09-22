package userword

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

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

func (r *UserWordRepoGorm) GetByUserAndWord(ctx context.Context, userID, wordID uint64) (*UserWord, error) {
	var uw UserWord
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND word_id = ?", userID, wordID).
		First(&uw).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get user_word by user+word: %w", err)
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

func (r *UserWordRepoGorm) Update(ctx context.Context, uw *UserWord) error {
	result := r.db.WithContext(ctx).Model(&UserWord{}).
		Where("user_word_id = ?", uw.ID).
		Select("Status", "EaseFactor", "IntervalDays", "Repetitions", "NextReviewAt", "LastReviewedAt").
		Updates(uw)
	if result.Error != nil {
		return fmt.Errorf("update user_word: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *UserWordRepoGorm) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&UserWord{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete user_word: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}
	return nil
}
