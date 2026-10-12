package category

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"vocabulary-service/internal/constants"
)

type RepoGorm struct {
	db *gorm.DB
}

func NewRepoGorm(db *gorm.DB) *RepoGorm {
	return &RepoGorm{db: db}
}

func (r *RepoGorm) Create(ctx context.Context, c *Category) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("create category: %w", err)
	}
	return nil
}

func (r *RepoGorm) GetByID(ctx context.Context, id uint64) (*Category, error) {
	var c Category
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get category: %w", err)
	}
	return &c, nil
}

func (r *RepoGorm) ListVisibleToUser(ctx context.Context, userID uint64) ([]Category, error) {
	var items []Category
	err := r.db.WithContext(ctx).
		Where("user_id IS NULL OR user_id = ?", userID).
		Order("user_id IS NULL DESC, category_name ASC").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return items, nil
}

// Delete removes the category and its tags in one transaction.
func (r *RepoGorm) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("category_id = ?", id).Delete(&Tag{}).Error; err != nil {
			return fmt.Errorf("delete category tags: %w", err)
		}
		result := tx.Delete(&Category{}, id)
		if result.Error != nil {
			return fmt.Errorf("delete category: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return constants.ErrNotFound
		}
		return nil
	})
}

func (r *RepoGorm) Tag(ctx context.Context, vocabID, categoryID uint64) error {
	tag := &Tag{VocabID: vocabID, CategoryID: categoryID}
	if err := r.db.WithContext(ctx).Create(tag).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return constants.ErrTagDuplicate
		}
		return fmt.Errorf("create tag: %w", err)
	}
	return nil
}

func (r *RepoGorm) Untag(ctx context.Context, vocabID, categoryID uint64) error {
	result := r.db.WithContext(ctx).
		Where("vocab_id = ? AND category_id = ?", vocabID, categoryID).
		Delete(&Tag{})
	if result.Error != nil {
		return fmt.Errorf("delete tag: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *RepoGorm) ListCategoriesForVocab(ctx context.Context, vocabID uint64) ([]Category, error) {
	var items []Category
	err := r.db.WithContext(ctx).
		Table("word_category").
		Joins("JOIN vocab_tags ON vocab_tags.category_id = word_category.category_id").
		Where("vocab_tags.vocab_id = ?", vocabID).
		Order("word_category.category_name ASC").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("list categories for vocab: %w", err)
	}
	return items, nil
}

func (r *RepoGorm) VocabBelongsToUser(ctx context.Context, vocabID, userID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("user_word").
		Where("user_word_id = ? AND user_id = ?", vocabID, userID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check vocab ownership: %w", err)
	}
	return count > 0, nil
}

func (r *RepoGorm) CountTagged(ctx context.Context, categoryID, userID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Tag{}).
		Joins("JOIN user_word ON user_word.user_word_id = vocab_tags.vocab_id").
		Where("vocab_tags.category_id = ? AND user_word.user_id = ?", categoryID, userID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count tagged words for category %d: %w", categoryID, err)
	}
	return int(count), nil
}
