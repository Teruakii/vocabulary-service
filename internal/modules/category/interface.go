package category

import "context"

type Repository interface {
	Create(ctx context.Context, c *Category) error
	GetByID(ctx context.Context, id uint64) (*Category, error)
	ListVisibleToUser(ctx context.Context, userID uint64) ([]Category, error)
	Delete(ctx context.Context, id uint64) error
	Tag(ctx context.Context, vocabID, categoryID uint64) error
	Untag(ctx context.Context, vocabID, categoryID uint64) error
	ListCategoriesForVocab(ctx context.Context, vocabID uint64) ([]Category, error)
	VocabBelongsToUser(ctx context.Context, vocabID, userID uint64) (bool, error)
	CountTagged(ctx context.Context, categoryID uint64) (int, error)
}

type Service interface {
	Create(ctx context.Context, userID uint64, name string) (*CategoryView, error)
	List(ctx context.Context, userID uint64) ([]CategoryView, error)
	Delete(ctx context.Context, userID, id uint64) error
	Tag(ctx context.Context, userID, vocabID, categoryID uint64) error
	Untag(ctx context.Context, userID, vocabID, categoryID uint64) error
	ListForVocab(ctx context.Context, userID, vocabID uint64) ([]CategoryView, error)
}
