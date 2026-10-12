package stats

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	db := r.db.WithContext(ctx)
	var s UserStats
	if err := db.First(&s, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get user_stats: %w", err)
	}
	var mastered int64
	if err := db.Table("user_word").
		Where("user_id = ? AND status = ?", userID, "mastered").
		Count(&mastered).Error; err != nil {
		return nil, fmt.Errorf("count mastered words: %w", err)
	}
	s.TotalWordsLearned = uint(mastered)
	return &s, nil
}

func (r *UserStatsRepoGorm) IncrementQuizzesTaken(ctx context.Context, userID uint64, now time.Time) error {
	return r.applyActivity(ctx, userID, now, func(s *UserStats) { s.TotalQuizTaken++ })
}

func (r *UserStatsRepoGorm) RecordActivity(ctx context.Context, userID uint64, now time.Time) error {
	if err := r.applyActivity(ctx, userID, now, nil); err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Exec(
		"INSERT INTO user_activity (user_id, activity_date, count) VALUES (?, ?, 1) ON DUPLICATE KEY UPDATE count = count + 1",
		userID, now.Format(dateLayout),
	).Error; err != nil {
		return fmt.Errorf("bump daily activity: %w", err)
	}
	return nil
}

const dateLayout = "2006-01-02"

func (r *UserStatsRepoGorm) ActivitySince(ctx context.Context, userID uint64, from time.Time) (map[string]uint, error) {
	var rows []struct {
		Day   string
		Count uint
	}
	if err := r.db.WithContext(ctx).Raw(
		"SELECT DATE_FORMAT(activity_date, '%Y-%m-%d') AS day, count FROM user_activity WHERE user_id = ? AND activity_date >= ?",
		userID, from.Format(dateLayout),
	).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list daily activity: %w", err)
	}
	counts := make(map[string]uint, len(rows))
	for _, row := range rows {
		counts[row.Day] = row.Count
	}
	return counts, nil
}

func (r *UserStatsRepoGorm) ensureExists(ctx context.Context, userID uint64) error {
	if err := r.db.WithContext(ctx).Exec(
		"INSERT IGNORE INTO user_stats (user_id) VALUES (?)", userID,
	).Error; err != nil {
		return fmt.Errorf("ensure user_stats row: %w", err)
	}
	return nil
}

// applyActivity locks the user's stats row, advances the streak and bumps
// one counter, so concurrent activity cannot lose an update.
func (r *UserStatsRepoGorm) applyActivity(ctx context.Context, userID uint64, now time.Time, bump func(*UserStats)) error {
	if err := r.ensureExists(ctx, userID); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var s UserStats
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&s, "user_id = ?", userID).Error; err != nil {
			return fmt.Errorf("lock user_stats: %w", err)
		}
		s.recordActivity(now)
		if bump != nil {
			bump(&s)
		}
		if err := tx.Save(&s).Error; err != nil {
			return fmt.Errorf("apply activity: %w", err)
		}
		return nil
	})
}
