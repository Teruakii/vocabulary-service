package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"vocabulary-service/internal/constants"
)

type AuthRepoGorm struct {
	db *gorm.DB
}

func NewAuthRepoGorm(db *gorm.DB) *AuthRepoGorm {
	return &AuthRepoGorm{db: db}
}

func (r *AuthRepoGorm) Create(ctx context.Context, u *User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return constants.ErrEmailDuplicate
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *AuthRepoGorm) GetByID(ctx context.Context, id uint64) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *AuthRepoGorm) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (r *AuthRepoGorm) Update(ctx context.Context, u *User) error {
	updates := map[string]interface{}{}
	if u.Email != "" {
		updates["email"] = u.Email
	}
	if len(updates) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).Model(&User{}).
		Where("user_id = ?", u.ID).
		Updates(updates)

	var mysqlErr *mysql.MySQLError
	if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
		return constants.ErrEmailDuplicate
	}
	if result.Error != nil {
		return fmt.Errorf("update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *AuthRepoGorm) UpdatePassword(ctx context.Context, id uint64, hashedPassword string) error {
	result := r.db.WithContext(ctx).Model(&User{}).
		Where("user_id = ?", id).
		Update("password", hashedPassword)
	if result.Error != nil {
		return fmt.Errorf("update password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *AuthRepoGorm) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&User{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *AuthRepoGorm) CreateIdentity(ctx context.Context, i *Identity) error {
	if err := r.db.WithContext(ctx).Create(i).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return constants.ErrIdentityDuplicate
		}
		return fmt.Errorf("create identity: %w", err)
	}
	return nil
}

func (r *AuthRepoGorm) GetIdentityByID(ctx context.Context, id uint64) (*Identity, error) {
	var i Identity
	if err := r.db.WithContext(ctx).First(&i, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get identity by id: %w", err)
	}
	return &i, nil
}

func (r *AuthRepoGorm) ListIdentitiesByUser(ctx context.Context, userID uint64) ([]Identity, error) {
	var ids []Identity
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&ids).Error; err != nil {
		return nil, fmt.Errorf("list identities: %w", err)
	}
	return ids, nil
}

func (r *AuthRepoGorm) CountIdentitiesByUser(ctx context.Context, userID uint64) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&Identity{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count identities: %w", err)
	}
	return int(count), nil
}

func (r *AuthRepoGorm) FindIdentityByProviderIdentifier(ctx context.Context, provider, identifier string) (*Identity, error) {
	var i Identity
	err := r.db.WithContext(ctx).
		Where("provider = ? AND identifier = ?", provider, identifier).
		First(&i).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("find identity: %w", err)
	}
	return &i, nil
}

func (r *AuthRepoGorm) DeleteIdentity(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&Identity{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete identity: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}
	return nil
}
