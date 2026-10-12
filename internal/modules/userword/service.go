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

// ActivityRecorder is the part of the stats module that reviews feed.
type ActivityRecorder interface {
	RecordActivity(ctx context.Context, userID uint64) error
}

type service struct {
	repo  Repository
	word  word.Service
	stats ActivityRecorder
	now   func() time.Time
}

// NewService builds the user-word service; stats may be nil (e.g. in tests).
func NewService(repo Repository, wordSvc word.Service, stats ActivityRecorder) Service {
	return &service{repo: repo, word: wordSvc, stats: stats, now: time.Now}
}

// recordReview updates the learner's stats after a saved review. Stats are
// best-effort: a failure here must not undo or fail the review itself.
func (s *service) recordReview(ctx context.Context, userID uint64) {
	if s.stats == nil {
		return
	}
	_ = s.stats.RecordActivity(ctx, userID)
}

// AddWord adds a global word or one of the caller's own. Another user's
// private word reports not found, like a missing one.
func (s *service) AddWord(ctx context.Context, userID, wordID uint64) (*UserWord, error) {
	if _, err := s.word.GetByID(ctx, userID, wordID); err != nil {
		return nil, err
	}
	return s.link(ctx, userID, wordID)
}

func (s *service) link(ctx context.Context, userID, wordID uint64) (*UserWord, error) {
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
	return uw, nil
}

func (s *service) AddCustomWord(ctx context.Context, userID uint64, wordStr, meaning string) (*UserWord, error) {
	owner := userID
	w, err := s.word.AddWord(ctx, &owner, wordStr, meaning)
	if err != nil {
		return nil, err
	}
	return s.link(ctx, userID, w.ID)
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
	limit = constants.ListLimit(limit)
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
	limit = constants.ListLimit(limit)
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
	uw, err := s.repo.ReviewByID(ctx, userID, id, s.sm2(quality))
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrUserWordNotFound
		}
		return nil, fmt.Errorf("review user_word %d: %w", id, err)
	}
	s.recordReview(ctx, userID)
	return uw, nil
}

func (s *service) ReviewByWord(ctx context.Context, userID, wordID uint64, quality int) (*UserWord, error) {
	if quality < 0 || quality > 5 {
		return nil, constants.ErrInvalidQuality
	}
	uw, err := s.repo.ReviewByWord(ctx, userID, wordID, s.sm2(quality))
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrUserWordNotFound
		}
		return nil, fmt.Errorf("review user_word for user %d word %d: %w", userID, wordID, err)
	}
	s.recordReview(ctx, userID)
	return uw, nil
}

func (s *service) sm2(quality int) func(*UserWord) {
	return func(uw *UserWord) { applySM2(uw, quality, s.now()) }
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
