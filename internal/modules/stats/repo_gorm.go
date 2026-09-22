package stats

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"vocabulary-service/internal/constants"
)

type UserStatsRepoGorm struct {
	db *gorm.DB
}

func NewUserStatsRepoGorm(db *gorm.DB) *UserStatsRepoGorm {
	return &UserStatsRepoGorm{db: db}
}

func (r *UserStatsRepoGorm) Get(ctx context.Context, userID uint64) (*UserStats, error) {
	if err := r.ensureExists(ctx, userID); err != nil {
		return nil, err
	}
	var s UserStats
	if err := r.db.WithContext(ctx).First(&s, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get user_stats: %w", err)
	}
	return &s, nil
}

func (r *UserStatsRepoGorm) IncrementWordsLearned(ctx context.Context, userID uint64, now time.Time) error {
	return r.applyActivity(ctx, userID, now, "total_words_learned")
}

func (r *UserStatsRepoGorm) IncrementQuizzesTaken(ctx context.Context, userID uint64, now time.Time) error {
	return r.applyActivity(ctx, userID, now, "total_quiz_taken")
}

func (r *UserStatsRepoGorm) RecordActivity(ctx context.Context, userID uint64, now time.Time) error {
	return r.applyActivity(ctx, userID, now, "")
}

func (r *UserStatsRepoGorm) ensureExists(ctx context.Context, userID uint64) error {
	if err := r.db.WithContext(ctx).Exec(
		"INSERT IGNORE INTO user_stats (user_id) VALUES (?)", userID,
	).Error; err != nil {
		return fmt.Errorf("ensure user_stats row: %w", err)
	}
	return nil
}

const streakCaseSQL = `
	CASE
		WHEN last_activity_at IS NULL THEN 1
		WHEN DATE(last_activity_at) = DATE(?) THEN current_streak
		WHEN DATEDIFF(DATE(?), DATE(last_activity_at)) = 1 THEN current_streak + 1
		ELSE 1
	END
`

func (r *UserStatsRepoGorm) applyActivity(ctx context.Context, userID uint64, now time.Time, counterColumn string) error {
	if err := r.ensureExists(ctx, userID); err != nil {
		return err
	}

	counterSQL := ""
	if counterColumn != "" {
		counterSQL = fmt.Sprintf(", %s = %s + 1", counterColumn, counterColumn)
	}

	sql := fmt.Sprintf(`
		UPDATE user_stats
		SET
			current_streak = %s,
			longest_streak = GREATEST(longest_streak, %s)
			%s,
			last_activity_at = ?
		WHERE user_id = ?
	`, streakCaseSQL, streakCaseSQL, counterSQL)

	if err := r.db.WithContext(ctx).Exec(sql, now, now, now, now, now, userID).Error; err != nil {
		return fmt.Errorf("apply activity: %w", err)
	}
	return nil
}
