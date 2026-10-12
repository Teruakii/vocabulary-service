package word

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"vocabulary-service/internal/constants"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) AddWord(ctx context.Context, ownerUserID *uint64, word, meaning string) (*Word, error) {
	word = strings.TrimSpace(word)
	meaning = strings.TrimSpace(meaning)

	if word == "" || meaning == "" {
		return nil, constants.ErrWordEmpty
	}

	dup, err := s.repo.ExistsForOwner(ctx, ownerUserID, word, 0)
	if err != nil {
		return nil, fmt.Errorf("check duplicate word %q: %w", word, err)
	}
	if dup {
		return nil, constants.ErrWordDuplicate
	}

	w := &Word{
		Word:        word,
		Meaning:     meaning,
		OwnerUserID: ownerUserID,
	}
	if err := s.repo.Create(ctx, w); err != nil {
		if errors.Is(err, constants.ErrWordDuplicate) {
			return nil, constants.ErrWordDuplicate
		}
		return nil, fmt.Errorf("failed to create word %q: %w", word, err)
	}
	return w, nil
}

func (s *service) GetByID(ctx context.Context, callerID uint64, id uint64) (*Word, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrWordNotFound
		}
		return nil, fmt.Errorf("get word by id %d: %w", id, err)
	}
	if !w.IsGlobal() && *w.OwnerUserID != callerID {

		return nil, constants.ErrWordNotFound
	}
	return w, nil
}

func (s *service) GetByWord(ctx context.Context, callerID uint64, word string) (*Word, error) {
	words := strings.TrimSpace(word)

	if words == "" {
		return nil, constants.ErrWordEmpty
	}
	w, err := s.repo.GetByWord(ctx, callerID, words)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrWordNotFound
		}
		return nil, fmt.Errorf("failed to find word %q: %w", words, err)
	}
	return w, nil
}

func (s *service) GetAllWord(ctx context.Context, callerID uint64, limit, offset int) ([]Word, bool, error) {
	limit = constants.ListLimit(limit)
	if offset < 0 {
		offset = 0
	}

	words, hasMore, err := s.repo.List(ctx, callerID, limit, offset)
	if err != nil {
		return nil, false, fmt.Errorf("service GetAllWord failed for caller %d: %w", callerID, err)
	}

	return words, hasMore, nil
}

func (s *service) UpdateWord(ctx context.Context, callerID uint64, isAdmin bool, word *Word) error {
	if word.Word != "" {
		trimmed := strings.TrimSpace(word.Word)
		if trimmed == "" {
			return constants.ErrWordEmpty
		}
		word.Word = trimmed
	}
	if word.Meaning != "" {
		trimmed := strings.TrimSpace(word.Meaning)
		if trimmed == "" {
			return constants.ErrMeaningEmpty
		}
		word.Meaning = trimmed
	}

	current, err := s.getModifiable(ctx, callerID, isAdmin, word.ID)
	if err != nil {
		return err
	}
	if word.Word != "" {
		dup, err := s.repo.ExistsForOwner(ctx, current.OwnerUserID, word.Word, word.ID)
		if err != nil {
			return fmt.Errorf("check duplicate word %q: %w", word.Word, err)
		}
		if dup {
			return constants.ErrWordDuplicate
		}
	}

	if err := s.repo.Update(ctx, word); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrWordNotFound
		}
		return fmt.Errorf("update word %d: %w", word.ID, err)
	}
	return nil
}

func (s *service) Delete(ctx context.Context, callerID uint64, isAdmin bool, id uint64) error {
	if _, err := s.getModifiable(ctx, callerID, isAdmin, id); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrWordNotFound
		}
		return fmt.Errorf("delete word %d: %w", id, err)
	}
	return nil
}

func (s *service) getModifiable(ctx context.Context, callerID uint64, isAdmin bool, id uint64) (*Word, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrWordNotFound
		}
		return nil, fmt.Errorf("get word %d: %w", id, err)
	}
	if w.IsGlobal() {
		if !isAdmin {
			return nil, constants.ErrWordForbidden
		}
		return w, nil
	}
	if *w.OwnerUserID != callerID {
		return nil, constants.ErrWordForbidden
	}
	return w, nil
}
