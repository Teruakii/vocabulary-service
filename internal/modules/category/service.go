package category

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

func (s *service) Create(ctx context.Context, userID uint64, name string) (*CategoryView, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, constants.ErrCategoryNameEmpty
	}
	c := &Category{UserID: &userID, Name: name}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("create category for user %d: %w", userID, err)
	}
	return toView(c, 0), nil
}

func (s *service) List(ctx context.Context, userID uint64) ([]CategoryView, error) {
	cats, err := s.repo.ListVisibleToUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list categories for user %d: %w", userID, err)
	}
	views := make([]CategoryView, len(cats))
	for i := range cats {
		count, err := s.repo.CountTagged(ctx, cats[i].ID, userID)
		if err != nil {
			return nil, fmt.Errorf("count words in category %d: %w", cats[i].ID, err)
		}
		views[i] = *toView(&cats[i], count)
	}
	return views, nil
}

func (s *service) Delete(ctx context.Context, userID, id uint64) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrCategoryNotFound
		}
		return fmt.Errorf("get category %d: %w", id, err)
	}
	if c.IsGlobal() || *c.UserID != userID {
		return constants.ErrCategoryForbidden
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrCategoryNotFound
		}
		return fmt.Errorf("delete category %d: %w", id, err)
	}
	return nil
}

func (s *service) Tag(ctx context.Context, userID, vocabID, categoryID uint64) error {
	if err := s.assertCanUse(ctx, userID, vocabID, categoryID); err != nil {
		return err
	}

	count, err := s.repo.CountTagged(ctx, categoryID, userID)
	if err != nil {
		return fmt.Errorf("count words in category %d: %w", categoryID, err)
	}
	if count >= constants.MaxWordsPerCategory {
		return constants.ErrCategoryFull
	}

	if err := s.repo.Tag(ctx, vocabID, categoryID); err != nil {
		if errors.Is(err, constants.ErrTagDuplicate) {
			return constants.ErrTagDuplicate
		}
		return fmt.Errorf("tag vocab %d with category %d: %w", vocabID, categoryID, err)
	}
	return nil
}

func (s *service) Untag(ctx context.Context, userID, vocabID, categoryID uint64) error {
	if err := s.assertCanUse(ctx, userID, vocabID, categoryID); err != nil {
		return err
	}
	if err := s.repo.Untag(ctx, vocabID, categoryID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrTagNotFound
		}
		return fmt.Errorf("untag vocab %d from category %d: %w", vocabID, categoryID, err)
	}
	return nil
}

func (s *service) ListForVocab(ctx context.Context, userID, vocabID uint64) ([]CategoryView, error) {
	owned, err := s.repo.VocabBelongsToUser(ctx, vocabID, userID)
	if err != nil {
		return nil, fmt.Errorf("check ownership of vocab %d: %w", vocabID, err)
	}
	if !owned {
		return nil, constants.ErrVocabNotFound
	}
	cats, err := s.repo.ListCategoriesForVocab(ctx, vocabID)
	if err != nil {
		return nil, fmt.Errorf("list categories for vocab %d: %w", vocabID, err)
	}
	views := make([]CategoryView, len(cats))
	for i := range cats {
		count, err := s.repo.CountTagged(ctx, cats[i].ID, userID)
		if err != nil {
			return nil, fmt.Errorf("count words in category %d: %w", cats[i].ID, err)
		}
		views[i] = *toView(&cats[i], count)
	}
	return views, nil
}

func (s *service) assertCanUse(ctx context.Context, userID, vocabID, categoryID uint64) error {
	owned, err := s.repo.VocabBelongsToUser(ctx, vocabID, userID)
	if err != nil {
		return fmt.Errorf("check ownership of vocab %d: %w", vocabID, err)
	}
	if !owned {
		return constants.ErrVocabNotFound
	}

	cat, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrCategoryNotFound
		}
		return fmt.Errorf("get category %d: %w", categoryID, err)
	}
	if !cat.IsGlobal() && *cat.UserID != userID {
		return constants.ErrCategoryNotFound
	}
	return nil
}

func toView(c *Category, wordCount int) *CategoryView {
	return &CategoryView{ID: c.ID, Name: c.Name, IsGlobal: c.IsGlobal(), WordCount: wordCount, MaxWords: constants.MaxWordsPerCategory}
}
