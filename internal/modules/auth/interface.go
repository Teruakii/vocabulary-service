package auth

import (
	"context"

	"vocabulary-service/internal/infra/jwt"
)

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uint64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
	UpdatePassword(ctx context.Context, id uint64, hashedPassword string) error
	Delete(ctx context.Context, id uint64) error
}

type IdentityRepository interface {
	CreateIdentity(ctx context.Context, i *Identity) error
	GetIdentityByID(ctx context.Context, id uint64) (*Identity, error)
	ListIdentitiesByUser(ctx context.Context, userID uint64) ([]Identity, error)
	CountIdentitiesByUser(ctx context.Context, userID uint64) (int, error)
	FindIdentityByProviderIdentifier(ctx context.Context, provider, identifier string) (*Identity, error)
	DeleteIdentity(ctx context.Context, id uint64) error
}

type Service interface {
	Register(ctx context.Context, email, password string) (*User, error)
	GetByID(ctx context.Context, id uint64) (*User, error)
	UpdateProfile(ctx context.Context, id uint64, email string) (*User, error)
	Delete(ctx context.Context, id uint64) error
	Login(ctx context.Context, username, password string) (accessToken, refreshToken string, u *User, err error)
	Refresh(ctx context.Context, refreshToken string) (accessToken string, u *User, err error)
	ChangePassword(ctx context.Context, userID uint64, currentPassword, newPassword string) error
	AdminResetPassword(ctx context.Context, targetUserID uint64, newPassword string) error
	ParseAccessToken(token string) (*jwt.Claims, error)
	LoginOrRegisterWithOAuth(ctx context.Context, provider, identifier, email, firstname, lastname string) (accessToken, refreshToken string, u *User, err error)
	LinkIdentity(ctx context.Context, userID uint64, provider, identifier string) (*Identity, error)
	ListIdentities(ctx context.Context, userID uint64) ([]Identity, error)
	UnlinkIdentity(ctx context.Context, userID, identityID uint64) error
}
