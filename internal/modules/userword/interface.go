package userword

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, uw *UserWord) error
	GetByID(ctx context.Context, id uint64) (*UserWord, error)
	GetByUserAndWord(ctx context.Context, userID, wordID uint64) (*UserWord, error)
	ListByUser(ctx context.Context, userID uint64, status Status, categoryID *uint64, limit, offset int) ([]UserWordDetail, bool, error)
	ListDue(ctx context.Context, userID uint64, before time.Time, limit int) ([]UserWordDetail, error)
	Update(ctx context.Context, uw *UserWord) error
	Delete(ctx context.Context, id uint64) error
}

type Service interface {
	AddWord(ctx context.Context, userID, wordID uint64) (*UserWord, error)

	AddCustomWord(ctx context.Context, userID uint64, word, meaning string) (*UserWord, error)
	GetByID(ctx context.Context, userID, id uint64) (*UserWord, error)
	List(ctx context.Context, userID uint64, status Status, categoryID *uint64, limit, offset int) ([]UserWordDetail, bool, error)
	Due(ctx context.Context, userID uint64, limit int) ([]UserWordDetail, error)
	Review(ctx context.Context, userID, id uint64, quality int) (*UserWord, error)
	ReviewByWord(ctx context.Context, userID, wordID uint64, quality int) (*UserWord, error)
	Remove(ctx context.Context, userID, id uint64) error
}
