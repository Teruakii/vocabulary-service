package word

import "context"

type Repository interface {
	Create(ctx context.Context, word *Word) error
	GetByID(ctx context.Context, id uint64) (*Word, error)

	GetByWord(ctx context.Context, callerID uint64, wordStr string) (*Word, error)
	List(ctx context.Context, callerID uint64, limit, offset int) ([]Word, bool, error)

	ExistsForOwner(ctx context.Context, ownerUserID *uint64, wordStr string) (bool, error)
	Update(ctx context.Context, word *Word) error
	Delete(ctx context.Context, id uint64) error
}

type Service interface {
	AddWord(ctx context.Context, ownerUserID *uint64, word, meaning string) (*Word, error)
	GetByID(ctx context.Context, callerID uint64, id uint64) (*Word, error)
	GetByWord(ctx context.Context, callerID uint64, word string) (*Word, error)
	GetAllWord(ctx context.Context, callerID uint64, limit, offset int) ([]Word, bool, error)

	UpdateWord(ctx context.Context, callerID uint64, isAdmin bool, word *Word) error
	Delete(ctx context.Context, callerID uint64, isAdmin bool, id uint64) error
}
