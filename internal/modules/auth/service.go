package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/infra/jwt"
)

type service struct {
	users      UserRepository
	identities IdentityRepository
	tokens     *jwt.Signer
}

func NewService(users UserRepository, identities IdentityRepository, tokens *jwt.Signer) Service {
	return &service{users: users, identities: identities, tokens: tokens}
}

func (s *service) Register(ctx context.Context, email, password string) (*User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return nil, constants.ErrEmailEmpty
	}
	if len(password) < 8 {
		return nil, constants.ErrPasswordTooShort
	}

	hashed, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &User{Email: email, Password: &hashed, Role: constants.RoleUser}
	if err := s.users.Create(ctx, u); err != nil {
		if errors.Is(err, constants.ErrEmailDuplicate) {
			return nil, constants.ErrEmailDuplicate
		}
		return nil, fmt.Errorf("register user %q: %w", email, err)
	}
	return u, nil
}

func (s *service) GetByID(ctx context.Context, id uint64) (*User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user %d: %w", id, err)
	}
	return u, nil
}

func (s *service) UpdateProfile(ctx context.Context, id uint64, email string) (*User, error) {
	u := &User{ID: id}
	if email != "" {
		e := normalizeEmail(email)
		if e == "" {
			return nil, constants.ErrEmailEmpty
		}
		u.Email = e
	}
	if err := s.users.Update(ctx, u); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrUserNotFound
		}
		if errors.Is(err, constants.ErrEmailDuplicate) {
			return nil, constants.ErrEmailDuplicate
		}
		return nil, fmt.Errorf("update profile %d: %w", id, err)
	}
	return s.GetByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uint64) error {
	if err := s.users.Delete(ctx, id); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrUserNotFound
		}
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

func (s *service) Login(ctx context.Context, username, password string) (string, string, *User, error) {
	email := normalizeEmail(username)
	if email == "" || password == "" {
		return "", "", nil, constants.ErrInvalidCredentials
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {

			return "", "", nil, constants.ErrInvalidCredentials
		}
		return "", "", nil, fmt.Errorf("login lookup %q: %w", email, err)
	}
	if !u.HasPassword() {
		return "", "", nil, constants.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(password)); err != nil {
		return "", "", nil, constants.ErrInvalidCredentials
	}

	access, err := s.tokens.IssueAccessToken(u.ID, string(u.Role))
	if err != nil {
		return "", "", nil, fmt.Errorf("issue access token: %w", err)
	}
	refresh, err := s.tokens.IssueRefreshToken(u.ID, string(u.Role))
	if err != nil {
		return "", "", nil, fmt.Errorf("issue refresh token: %w", err)
	}
	return access, refresh, u, nil
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (string, *User, error) {
	claims, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", nil, constants.ErrInvalidToken
	}

	u, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return "", nil, constants.ErrInvalidToken
		}
		return "", nil, fmt.Errorf("refresh lookup user %d: %w", claims.UserID, err)
	}
	if claims.IssuedAt != nil && u.UpdatedAt.After(claims.IssuedAt.Time) {
		return "", nil, constants.ErrInvalidToken
	}

	access, err := s.tokens.IssueAccessToken(u.ID, string(u.Role))
	if err != nil {
		return "", nil, fmt.Errorf("issue access token: %w", err)
	}
	return access, u, nil
}

func (s *service) ChangePassword(ctx context.Context, userID uint64, currentPassword, newPassword string) error {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrUserNotFound
		}
		return fmt.Errorf("change password lookup %d: %w", userID, err)
	}
	if u.HasPassword() {
		if err := bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(currentPassword)); err != nil {
			return constants.ErrInvalidCredentials
		}
	}

	if len(newPassword) < 8 {
		return constants.ErrPasswordTooShort
	}
	hashed, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.users.UpdatePassword(ctx, userID, hashed); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrUserNotFound
		}
		return fmt.Errorf("update password %d: %w", userID, err)
	}
	return nil
}
func (s *service) AdminResetPassword(ctx context.Context, targetUserID uint64, newPassword string) error {
	if len(newPassword) < 8 {
		return constants.ErrPasswordTooShort
	}
	hashed, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.users.UpdatePassword(ctx, targetUserID, hashed); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrUserNotFound
		}
		return fmt.Errorf("admin reset password %d: %w", targetUserID, err)
	}
	return nil
}

func (s *service) ParseAccessToken(token string) (*jwt.Claims, error) {
	return s.tokens.ParseAccessToken(token)
}

func (s *service) LoginOrRegisterWithOAuth(ctx context.Context, provider, identifier, email, firstname, lastname string) (string, string, *User, error) {
	prov := constants.Provider(provider)
	if !prov.IsValid() {
		return "", "", nil, constants.ErrIdentityInvalid
	}
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return "", "", nil, constants.ErrIdentityInvalid
	}

	if existing, err := s.identities.FindIdentityByProviderIdentifier(ctx, string(prov), identifier); err == nil {
		u, err := s.users.GetByID(ctx, existing.UserID)
		if err != nil {
			return "", "", nil, fmt.Errorf("get user %d for oauth identity: %w", existing.UserID, err)
		}
		return s.issueTokenPair(u)
	} else if !errors.Is(err, constants.ErrNotFound) {
		return "", "", nil, fmt.Errorf("find oauth identity: %w", err)
	}

	email = normalizeEmail(email)
	var u *User
	if email != "" {
		if existingUser, err := s.users.GetByEmail(ctx, email); err == nil {
			u = existingUser
		} else if !errors.Is(err, constants.ErrNotFound) {
			return "", "", nil, fmt.Errorf("lookup user by email %q: %w", email, err)
		}
	}
	if u == nil {
		u = &User{Email: email, Firstname: firstname, Lastname: lastname, Role: constants.RoleUser}
		if err := s.users.Create(ctx, u); err != nil {
			if errors.Is(err, constants.ErrEmailDuplicate) {
				return "", "", nil, constants.ErrEmailDuplicate
			}
			return "", "", nil, fmt.Errorf("create user for oauth login: %w", err)
		}
	}

	if err := s.identities.CreateIdentity(ctx, &Identity{UserID: u.ID, Provider: prov, Identifier: identifier}); err != nil {

		if !errors.Is(err, constants.ErrIdentityDuplicate) {
			return "", "", nil, fmt.Errorf("link oauth identity for user %d: %w", u.ID, err)
		}
	}

	return s.issueTokenPair(u)
}

func (s *service) issueTokenPair(u *User) (string, string, *User, error) {
	access, err := s.tokens.IssueAccessToken(u.ID, string(u.Role))
	if err != nil {
		return "", "", nil, fmt.Errorf("issue access token: %w", err)
	}
	refresh, err := s.tokens.IssueRefreshToken(u.ID, string(u.Role))
	if err != nil {
		return "", "", nil, fmt.Errorf("issue refresh token: %w", err)
	}
	return access, refresh, u, nil
}
func (s *service) LinkIdentity(ctx context.Context, userID uint64, provider, identifier string) (*Identity, error) {
	provider = strings.TrimSpace(provider)
	identifier = strings.TrimSpace(identifier)
	if provider == "" || identifier == "" {
		return nil, constants.ErrIdentityInvalid
	}

	i := &Identity{UserID: userID, Provider: constants.Provider(provider), Identifier: identifier}
	if err := s.identities.CreateIdentity(ctx, i); err != nil {
		if errors.Is(err, constants.ErrIdentityDuplicate) {
			return nil, constants.ErrIdentityDuplicate
		}
		return nil, fmt.Errorf("link identity for user %d: %w", userID, err)
	}
	return i, nil
}

func (s *service) ListIdentities(ctx context.Context, userID uint64) ([]Identity, error) {
	ids, err := s.identities.ListIdentitiesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list identities for user %d: %w", userID, err)
	}
	return ids, nil
}
func (s *service) UnlinkIdentity(ctx context.Context, userID, identityID uint64) error {
	i, err := s.identities.GetIdentityByID(ctx, identityID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrIdentityNotFound
		}
		return fmt.Errorf("get identity %d: %w", identityID, err)
	}
	if i.UserID != userID {
		return constants.ErrIdentityNotFound
	}

	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user %d: %w", userID, err)
	}
	if !u.HasPassword() {
		count, err := s.identities.CountIdentitiesByUser(ctx, userID)
		if err != nil {
			return fmt.Errorf("count identities for user %d: %w", userID, err)
		}
		if count <= 1 {
			return constants.ErrCannotUnlinkLastIdentity
		}
	}

	if err := s.identities.DeleteIdentity(ctx, identityID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrIdentityNotFound
		}
		return fmt.Errorf("delete identity %d: %w", identityID, err)
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
