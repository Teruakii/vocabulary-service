package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	result := r.db.WithContext(ctx).Model(&User{}).
		Where("user_id = ?", u.ID).
		Select("Email", "EmailVerifiedAt", "Role").
		Updates(u)

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
		Updates(map[string]any{"password": hashedPassword, "password_changed_at": time.Now()})
	if result.Error != nil {
		return fmt.Errorf("update password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}
	return nil
}

// Delete removes the user and everything they own. The schema has no foreign
// keys, so the dependent rows are removed here, in one transaction.
func (r *AuthRepoGorm) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&User{}, id)
		if result.Error != nil {
			return fmt.Errorf("delete user: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return constants.ErrNotFound
		}

		cleanup := []struct {
			what string
			sql  string
		}{
			{"identities", "DELETE FROM identities WHERE user_id = ?"},
			{"sessions", "DELETE FROM refresh_sessions WHERE user_id = ?"},
			{"password resets", "DELETE FROM password_resets WHERE user_id = ?"},
			{"quiz questions", "DELETE qq FROM quiz_question qq JOIN quizzes q ON q.quiz_id = qq.quiz_id WHERE q.user_id = ?"},
			{"quizzes", "DELETE FROM quizzes WHERE user_id = ?"},
			{"word tags", "DELETE vt FROM vocab_tags vt JOIN user_word uw ON uw.user_word_id = vt.vocab_id WHERE uw.user_id = ?"},
			{"category tags", "DELETE vt FROM vocab_tags vt JOIN word_category c ON c.category_id = vt.category_id WHERE c.user_id = ?"},
			{"categories", "DELETE FROM word_category WHERE user_id = ?"},
			{"user words", "DELETE FROM user_word WHERE user_id = ?"},
			{"stats", "DELETE FROM user_stats WHERE user_id = ?"},
			{"daily activity", "DELETE FROM user_activity WHERE user_id = ?"},
			// Private words go too, unless another user still studies them.
			{"owned words", `DELETE w FROM word w
				WHERE w.owner_user_id = ?
				  AND NOT EXISTS (SELECT 1 FROM user_word uw WHERE uw.word_id = w.word_id)
				  AND NOT EXISTS (SELECT 1 FROM quiz_question qq WHERE qq.word_id = w.word_id)`},
		}
		for _, c := range cleanup {
			if err := tx.Exec(c.sql, id).Error; err != nil {
				return fmt.Errorf("delete user %s: %w", c.what, err)
			}
		}
		return nil
	})
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

func (r *AuthRepoGorm) DeleteIdentitiesByUser(ctx context.Context, userID uint64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&Identity{}).Error; err != nil {
		return fmt.Errorf("delete identities: %w", err)
	}
	return nil
}

func (r *AuthRepoGorm) CreateSession(ctx context.Context, s *Session) error {
	db := r.db.WithContext(ctx)
	// Drop this user's dead sessions so the table does not grow forever.
	if err := db.Where("user_id = ? AND (expires_at < ? OR revoked_at IS NOT NULL)", s.UserID, time.Now()).
		Delete(&Session{}).Error; err != nil {
		return fmt.Errorf("prune sessions: %w", err)
	}
	if err := db.Create(s).Error; err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *AuthRepoGorm) GetSession(ctx context.Context, id string) (*Session, error) {
	var s Session
	if err := r.db.WithContext(ctx).First(&s, "session_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &s, nil
}

func (r *AuthRepoGorm) RevokeSession(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Model(&Session{}).
		Where("session_id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", time.Now()).Error; err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (r *AuthRepoGorm) RevokeAllSessions(ctx context.Context, userID uint64) error {
	if err := r.db.WithContext(ctx).Model(&Session{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", time.Now()).Error; err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}
	return nil
}

// SyncAdmins makes the admin role match ADMIN_EMAILS: listed accounts with a
// verified email become admins and every other admin becomes a user.
func (r *AuthRepoGorm) SyncAdmins(ctx context.Context, emails []string) (promoted, demoted int64, err error) {
	normalized := make([]string, 0, len(emails))
	for _, e := range emails {
		if e = normalizeEmail(e); e != "" {
			normalized = append(normalized, e)
		}
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		demote := tx.Model(&User{}).Where("role = ?", constants.RoleAdmin)
		if len(normalized) > 0 {
			result := tx.Model(&User{}).
				Where("email IN ? AND email_verified_at IS NOT NULL AND role <> ?", normalized, constants.RoleAdmin).
				Update("role", constants.RoleAdmin)
			if result.Error != nil {
				return fmt.Errorf("promote admins: %w", result.Error)
			}
			promoted = result.RowsAffected
			demote = demote.Where("NOT (email IN ? AND email_verified_at IS NOT NULL)", normalized)
		}
		result := demote.Update("role", constants.RoleUser)
		if result.Error != nil {
			return fmt.Errorf("demote admins: %w", result.Error)
		}
		demoted = result.RowsAffected
		return nil
	})
	return promoted, demoted, err
}

func (r *AuthRepoGorm) CreatePasswordReset(ctx context.Context, pr *PasswordReset) error {
	db := r.db.WithContext(ctx)
	if err := db.Where("user_id = ? AND expires_at < ?", pr.UserID, time.Now()).
		Delete(&PasswordReset{}).Error; err != nil {
		return fmt.Errorf("prune password resets: %w", err)
	}
	if err := db.Create(pr).Error; err != nil {
		return fmt.Errorf("create password reset: %w", err)
	}
	return nil
}

func (r *AuthRepoGorm) GetPasswordReset(ctx context.Context, tokenHash string) (*PasswordReset, error) {
	var pr PasswordReset
	if err := r.db.WithContext(ctx).First(&pr, "token_hash = ?", tokenHash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("get password reset: %w", err)
	}
	return &pr, nil
}

func (r *AuthRepoGorm) DeletePasswordResets(ctx context.Context, userID uint64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&PasswordReset{}).Error; err != nil {
		return fmt.Errorf("delete password resets: %w", err)
	}
	return nil
}
