package userword

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/modules/word"
)

// StatsRecorder is the subset of stats.Service this module needs. It's
// declared locally (rather than importing the stats package) to avoid a
// module-to-module dependency; stats.Service already satisfies it.
type StatsRecorder interface {
	RecordActivity(ctx context.Context, userID uint64) error
	RecordWordMastered(ctx context.Context, userID uint64) error
}

type service struct {
	repo  Repository
	word  word.Service
	stats StatsRecorder
	now   func() time.Time
}

// NewService takes stats as an optional trailing argument so existing
// callers/tests that only pass (repo, wordSvc) keep compiling. Pass a
// StatsRecorder to wire up streaks / "words mastered" tracking.
func NewService(repo Repository, wordSvc word.Service, stats ...StatsRecorder) Service {
	var sr StatsRecorder
	if len(stats) > 0 {
		sr = stats[0]
	}
	return &service{repo: repo, word: wordSvc, stats: sr, now: time.Now}
}

// recordActivity is best-effort: a stats-tracking failure must never fail
// the caller's actual request.
func (s *service) recordActivity(ctx context.Context, userID uint64) {
	if s.stats == nil {
		return
	}
	_ = s.stats.RecordActivity(ctx, userID)
}

func (s *service) recordIfNewlyMastered(ctx context.Context, userID uint64, wasStatus, isStatus Status) {
	if s.stats == nil {
		return
	}
	if isStatus == StatusMastered && wasStatus != StatusMastered {
		_ = s.stats.RecordWordMastered(ctx, userID)
	}
}

func (s *service) AddWord(ctx context.Context, userID, wordID uint64) (*UserWord, error) {
	now := s.now()
	uw := &UserWord{
		UserID:       userID,
		WordID:       wordID,
		Status:       StatusNew,
		EaseFactor:   2.5,
		IntervalDays: 0,
		Repetitions:  0,
		NextReviewAt: &now,
	}

	if err := s.repo.Create(ctx, uw); err != nil {
		if errors.Is(err, constants.ErrUserWordDuplicate) {
			return nil, constants.ErrUserWordDuplicate
		}
		return nil, fmt.Errorf("add word %d for user %d: %w", wordID, userID, err)
	}
	s.recordActivity(ctx, userID)
	return uw, nil
}

func (s *service) AddCustomWord(ctx context.Context, userID uint64, wordStr, meaning string) (*UserWord, error) {
	owner := userID
	w, err := s.word.AddWord(ctx, &owner, wordStr, meaning)
	if err != nil {
		return nil, err
	}
	return s.AddWord(ctx, userID, w.ID)
}

func (s *service) GetByID(ctx context.Context, userID, id uint64) (*UserWord, error) {
	uw, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrUserWordNotFound
		}
		return nil, fmt.Errorf("get user_word %d: %w", id, err)
	}
	if uw.UserID != userID {
		return nil, constants.ErrUserWordNotFound
	}
	return uw, nil
}

func (s *service) List(ctx context.Context, userID uint64, status Status, categoryID *uint64, limit, offset int) ([]UserWordDetail, bool, error) {
	if limit <= 0 {
		limit = constants.DefaultListLimit
	}
	if offset < 0 {
		offset = 0
	}
	items, hasMore, err := s.repo.ListByUser(ctx, userID, status, categoryID, limit, offset)
	if err != nil {
		return nil, false, fmt.Errorf("list user_word for user %d: %w", userID, err)
	}
	return items, hasMore, nil
}

func (s *service) Due(ctx context.Context, userID uint64, limit int) ([]UserWordDetail, error) {
	if limit <= 0 {
		limit = constants.DefaultListLimit
	}
	items, err := s.repo.ListDue(ctx, userID, s.now(), limit)
	if err != nil {
		return nil, fmt.Errorf("list due user_word for user %d: %w", userID, err)
	}
	return items, nil
}

func (s *service) Review(ctx context.Context, userID, id uint64, quality int) (*UserWord, error) {
	if quality < 0 || quality > 5 {
		return nil, constants.ErrInvalidQuality
	}

	uw, err := s.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	prevStatus := uw.Status
	applySM2(uw, quality, s.now())

	if err := s.repo.Update(ctx, uw); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrUserWordNotFound
		}
		return nil, fmt.Errorf("update user_word %d: %w", id, err)
	}
	s.recordActivity(ctx, userID)
	s.recordIfNewlyMastered(ctx, userID, prevStatus, uw.Status)
	return uw, nil
}
func (s *service) ReviewByWord(ctx context.Context, userID, wordID uint64, quality int) (*UserWord, error) {
	if quality < 0 || quality > 5 {
		return nil, constants.ErrInvalidQuality
	}

	uw, err := s.repo.GetByUserAndWord(ctx, userID, wordID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrUserWordNotFound
		}
		return nil, fmt.Errorf("get user_word for user %d word %d: %w", userID, wordID, err)
	}

	prevStatus := uw.Status
	applySM2(uw, quality, s.now())

	if err := s.repo.Update(ctx, uw); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrUserWordNotFound
		}
		return nil, fmt.Errorf("update user_word %d: %w", uw.ID, err)
	}
	s.recordActivity(ctx, userID)
	s.recordIfNewlyMastered(ctx, userID, prevStatus, uw.Status)
	return uw, nil
}
func (s *service) Remove(ctx context.Context, userID, id uint64) error {
	if _, err := s.GetByID(ctx, userID, id); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrUserWordNotFound
		}
		return fmt.Errorf("delete user_word %d: %w", id, err)
	}
	return nil
}

func applySM2(uw *UserWord, quality int, now time.Time) {
	if quality < 3 {
		uw.Repetitions = 0
		uw.IntervalDays = 1
	} else {
		switch uw.Repetitions {
		case 0:
			uw.IntervalDays = 1
		case 1:
			uw.IntervalDays = 6
		default:
			uw.IntervalDays = uint(math.Round(float64(uw.IntervalDays) * uw.EaseFactor))
		}
		uw.Repetitions++
	}

	delta := 0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02)
	uw.EaseFactor += delta
	if uw.EaseFactor < 1.3 {
		uw.EaseFactor = 1.3
	}

	uw.Status = deriveStatus(uw.Repetitions, uw.IntervalDays)

	reviewedAt := now
	uw.LastReviewedAt = &reviewedAt
	next := now.AddDate(0, 0, int(uw.IntervalDays))
	uw.NextReviewAt = &next
}

func deriveStatus(repetitions, intervalDays uint) Status {
	switch {
	case repetitions == 0:
		return StatusNew
	case intervalDays >= 21:
		return StatusMastered
	default:
		return StatusLearning
	}
}
