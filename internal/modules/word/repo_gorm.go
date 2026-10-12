package word

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"vocabulary-service/internal/constants"
)

type WordRepoGorm struct {
	db *gorm.DB
}

func NewWordRepoGorm(db *gorm.DB) *WordRepoGorm {
	return &WordRepoGorm{db: db}
}

func (r *WordRepoGorm) Create(ctx context.Context, word *Word) error {
	if err := r.db.WithContext(ctx).Create(word).Error; err != nil {

		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return constants.ErrWordDuplicate
		}
		return fmt.Errorf("create word failed %w", err)
	}
	return nil
}

func (r *WordRepoGorm) GetByID(ctx context.Context, id uint64) (*Word, error) {
	var word Word
	if err := r.db.WithContext(ctx).First(&word, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get word by id: %w", err)
	}
	return &word, nil
}

func (r *WordRepoGorm) GetByWord(ctx context.Context, callerID uint64, wordStr string) (*Word, error) {
	var word Word
	if err := r.db.WithContext(ctx).
		Where("word = ? AND (owner_user_id IS NULL OR owner_user_id = ?)", wordStr, callerID).
		First(&word).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get word by word: %w", err)
	}
	return &word, nil
}

func (r *WordRepoGorm) ExistsForOwner(ctx context.Context, ownerUserID *uint64, wordStr string, excludeID uint64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&Word{}).Where("word = ? AND word_id <> ?", wordStr, excludeID)
	if ownerUserID == nil {
		q = q.Where("owner_user_id IS NULL")
	} else {
		q = q.Where("owner_user_id = ?", *ownerUserID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, fmt.Errorf("check existing word: %w", err)
	}
	return count > 0, nil
}

func (r *WordRepoGorm) List(ctx context.Context, callerID uint64, limit, offset int) ([]Word, bool, error) {
	if limit <= 0 {
		limit = constants.DefaultListLimit
	}
	fetchLimit := limit + 1

	var words []Word
	err := r.db.WithContext(ctx).
		Where("owner_user_id IS NULL OR owner_user_id = ?", callerID).
		Order("word_id ASC").
		Limit(fetchLimit).
		Offset(offset).
		Find(&words).Error
	if err != nil {
		return nil, false, fmt.Errorf("list word: %w", err)
	}

	hasMore := len(words) > limit
	if hasMore {
		words = words[:limit]
	}
	return words, hasMore, nil
}

func (r *WordRepoGorm) Update(ctx context.Context, word *Word) error {
	updates := map[string]interface{}{}
	if word.Word != "" {
		updates["word"] = word.Word
	}
	if word.Meaning != "" {
		updates["meaning"] = word.Meaning
	}
	if len(updates) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).Model(&Word{}).
		Where("word_id = ?", word.ID).
		Updates(updates)

	var mysqlErr *mysql.MySQLError
	if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
		return constants.ErrWordDuplicate
	}
	if result.Error != nil {
		return fmt.Errorf("update word: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}
	return nil
}

// Delete refuses words still referenced by a user's list or a quiz. The
// schema has no foreign keys, so the check is done here.
func (r *WordRepoGorm) Delete(ctx context.Context, id uint64) error {
	db := r.db.WithContext(ctx)
	var inUse int64
	if err := db.Raw(`SELECT
		EXISTS (SELECT 1 FROM user_word WHERE word_id = ?) +
		EXISTS (SELECT 1 FROM quiz_question WHERE word_id = ?)`, id, id).
		Scan(&inUse).Error; err != nil {
		return fmt.Errorf("check word usage: %w", err)
	}
	if inUse > 0 {
		return constants.ErrWordInUse
	}

	result := db.Delete(&Word{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete word: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}
	return nil
}
