package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/infra/jwt"
	"vocabulary-service/internal/infra/mail"
)

const passwordResetTTL = time.Hour

type service struct {
	users       UserRepository
	identities  IdentityRepository
	sessions    SessionRepository
	tokens      *jwt.Signer
	adminEmails map[string]bool
	resets      PasswordResetRepository
	mailer      mail.Mailer
	frontendURL string
}

type Option func(*service)

// WithAdminEmails makes accounts with these emails admins once the email is
// verified.
func WithAdminEmails(emails []string) Option {
	return func(s *service) {
		for _, e := range emails {
			if e = normalizeEmail(e); e != "" {
				s.adminEmails[e] = true
			}
		}
	}
}

// WithPasswordReset enables password reset by email. Links point at
// frontendURL; a nil mailer or empty frontendURL leaves reset unavailable.
func WithPasswordReset(resets PasswordResetRepository, mailer mail.Mailer, frontendURL string) Option {
	return func(s *service) {
		s.resets = resets
		s.mailer = mailer
		s.frontendURL = strings.TrimRight(frontendURL, "/")
	}
}

func NewService(users UserRepository, identities IdentityRepository, sessions SessionRepository, tokens *jwt.Signer, opts ...Option) Service {
	s := &service{users: users, identities: identities, sessions: sessions, tokens: tokens, adminEmails: map[string]bool{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// roleFor is the role u should have: admin exactly when its email is listed
// and verified. Roles other than admin are left alone.
func (s *service) roleFor(u *User) constants.Role {
	if s.adminEmails[u.Email] && u.EmailVerifiedAt != nil {
		return constants.RoleAdmin
	}
	if u.Role == constants.RoleAdmin {
		return constants.RoleUser
	}
	return u.Role
}

// markVerified records that the user proved they own u.Email and applies
// the role that follows from it.
func (s *service) markVerified(ctx context.Context, u *User) error {
	verifiedAt, role := u.EmailVerifiedAt, u.Role
	if u.EmailVerifiedAt == nil {
		now := time.Now()
		u.EmailVerifiedAt = &now
	}
	u.Role = s.roleFor(u)
	if u.EmailVerifiedAt == verifiedAt && u.Role == role {
		return nil
	}
	if err := s.users.Update(ctx, u); err != nil {
		return fmt.Errorf("mark email verified for user %d: %w", u.ID, err)
	}
	return nil
}

func (s *service) Register(ctx context.Context, email, password string) (*User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return nil, constants.ErrEmailEmpty
	}
	if err := validateNewPassword(password); err != nil {
		return nil, err
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
	if email == "" {
		return s.GetByID(ctx, id)
	}
	e := normalizeEmail(email)
	if e == "" {
		return nil, constants.ErrEmailEmpty
	}
	u, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u.Email == e {
		return u, nil
	}
	u.Email = e
	u.EmailVerifiedAt = nil
	u.Role = s.roleFor(u)
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
	return s.issueTokenPair(ctx, u)
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (string, *User, error) {
	claims, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", nil, constants.ErrInvalidToken
	}

	session, err := s.sessions.GetSession(ctx, claims.ID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return "", nil, constants.ErrInvalidToken
		}
		return "", nil, fmt.Errorf("refresh lookup session: %w", err)
	}
	if session.UserID != claims.UserID || !session.Active(time.Now()) {
		return "", nil, constants.ErrInvalidToken
	}

	u, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return "", nil, constants.ErrInvalidToken
		}
		return "", nil, fmt.Errorf("refresh lookup user %d: %w", claims.UserID, err)
	}
	// Only a password change ends existing sessions. JWT iat has second
	// precision, so compare at that precision too: otherwise a token issued in
	// the same second as the change would look older than it is.
	if claims.IssuedAt != nil && u.PasswordChangedAt != nil &&
		u.PasswordChangedAt.Truncate(time.Second).After(claims.IssuedAt.Time) {
		return "", nil, constants.ErrInvalidToken
	}

	access, err := s.tokens.IssueAccessToken(u.ID, string(u.Role))
	if err != nil {
		return "", nil, fmt.Errorf("issue access token: %w", err)
	}
	return access, u, nil
}

// Logout revokes the session behind a refresh token. Tokens that are already
// invalid or expired have nothing left to revoke, so that is not an error.
func (s *service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil
	}
	if err := s.sessions.RevokeSession(ctx, claims.ID); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
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

	if err := validateNewPassword(newPassword); err != nil {
		return err
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
	if err := s.sessions.RevokeAllSessions(ctx, userID); err != nil {
		return fmt.Errorf("end sessions after password change %d: %w", userID, err)
	}
	return nil
}
func (s *service) AdminResetPassword(ctx context.Context, targetUserID uint64, newPassword string) error {
	if err := validateNewPassword(newPassword); err != nil {
		return err
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
	if err := s.sessions.RevokeAllSessions(ctx, targetUserID); err != nil {
		return fmt.Errorf("end sessions after password reset %d: %w", targetUserID, err)
	}
	return nil
}

func (s *service) RequestPasswordReset(ctx context.Context, email string) error {
	if s.mailer == nil || s.frontendURL == "" {
		return constants.ErrPasswordResetUnavailable
	}
	email = normalizeEmail(email)
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// Unknown emails get the same answer, so the endpoint does not reveal
		// which addresses have accounts.
		if errors.Is(err, constants.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("password reset lookup %q: %w", email, err)
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if err := s.resets.CreatePasswordReset(ctx, &PasswordReset{
		TokenHash: hashResetToken(token),
		UserID:    u.ID,
		Email:     u.Email,
		ExpiresAt: time.Now().Add(passwordResetTTL),
	}); err != nil {
		return fmt.Errorf("store reset token for user %d: %w", u.ID, err)
	}

	link := s.frontendURL + "/reset-password?token=" + token
	body := "Someone asked to reset the password for your account.\n\n" +
		"Open this link within 1 hour to choose a new password:\n" + link + "\n\n" +
		"If you did not ask for this, ignore this email. Your password stays the same.\n"
	if err := s.mailer.Send(ctx, u.Email, "Reset your password", body); err != nil {
		return fmt.Errorf("send reset email to user %d: %w", u.ID, err)
	}
	return nil
}

// ResetPassword sets a new password with an emailed token. Opening the link
// proves the user owns the email, so it also verifies it.
func (s *service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if err := validateNewPassword(newPassword); err != nil {
		return err
	}
	reset, err := s.resets.GetPasswordReset(ctx, hashResetToken(token))
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrResetTokenInvalid
		}
		return fmt.Errorf("lookup reset token: %w", err)
	}
	if !time.Now().Before(reset.ExpiresAt) {
		return constants.ErrResetTokenInvalid
	}
	u, err := s.users.GetByID(ctx, reset.UserID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrResetTokenInvalid
		}
		return fmt.Errorf("reset password lookup %d: %w", reset.UserID, err)
	}
	if u.Email != reset.Email {
		return constants.ErrResetTokenInvalid
	}

	hashed, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.users.UpdatePassword(ctx, u.ID, hashed); err != nil {
		return fmt.Errorf("reset password %d: %w", u.ID, err)
	}
	if err := s.markVerified(ctx, u); err != nil {
		return err
	}
	if err := s.resets.DeletePasswordResets(ctx, u.ID); err != nil {
		return fmt.Errorf("delete reset tokens for user %d: %w", u.ID, err)
	}
	if err := s.sessions.RevokeAllSessions(ctx, u.ID); err != nil {
		return fmt.Errorf("end sessions after password reset %d: %w", u.ID, err)
	}
	return nil
}

func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
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

	email = normalizeEmail(email)

	if existing, err := s.identities.FindIdentityByProviderIdentifier(ctx, string(prov), identifier); err == nil {
		u, err := s.users.GetByID(ctx, existing.UserID)
		if err != nil {
			return "", "", nil, fmt.Errorf("get user %d for oauth identity: %w", existing.UserID, err)
		}
		// Google vouches for its own email, which the account may have
		// changed away from with PUT /auth/me.
		if u.Email == email {
			if err := s.markVerified(ctx, u); err != nil {
				return "", "", nil, err
			}
		}
		return s.issueTokenPair(ctx, u)
	} else if !errors.Is(err, constants.ErrNotFound) {
		return "", "", nil, fmt.Errorf("find oauth identity: %w", err)
	}

	var u *User
	if email != "" {
		if existingUser, err := s.users.GetByEmail(ctx, email); err == nil {
			// A verified owner proved the same inbox Google just did, so the
			// account is theirs and only gains this identity.
			if existingUser.EmailVerifiedAt == nil {
				if err := s.takeOverUnverifiedAccount(ctx, existingUser); err != nil {
					return "", "", nil, err
				}
			}
			u = existingUser
		} else if !errors.Is(err, constants.ErrNotFound) {
			return "", "", nil, fmt.Errorf("lookup user by email %q: %w", email, err)
		}
	}
	if u == nil {
		now := time.Now()
		u = &User{Email: email, Firstname: firstname, Lastname: lastname, Role: constants.RoleUser, EmailVerifiedAt: &now}
		u.Role = s.roleFor(u)
		if err := s.users.Create(ctx, u); err != nil {
			if errors.Is(err, constants.ErrEmailDuplicate) {
				return "", "", nil, constants.ErrEmailDuplicate
			}
			return "", "", nil, fmt.Errorf("create user for oauth login: %w", err)
		}
	}
	if err := s.markVerified(ctx, u); err != nil {
		return "", "", nil, err
	}

	if err := s.identities.CreateIdentity(ctx, &Identity{UserID: u.ID, Provider: prov, Identifier: identifier}); err != nil {

		if !errors.Is(err, constants.ErrIdentityDuplicate) {
			return "", "", nil, fmt.Errorf("link oauth identity for user %d: %w", u.ID, err)
		}
	}

	return s.issueTokenPair(ctx, u)
}

// takeOverUnverifiedAccount runs when a Google login matches an account by an
// email that was never verified. Nothing proved the account's current way in
// belongs to the owner of that email: registration never verifies it, and
// PUT /auth/me lets a user put any unused email on an account that keeps its
// own Google identity. Google did verify it, so the Google user keeps the
// account and the password, the other linked identities and every open
// session are dropped. The owner can set a new password without the old one.
func (s *service) takeOverUnverifiedAccount(ctx context.Context, u *User) error {
	if u.HasPassword() {
		if err := s.users.UpdatePassword(ctx, u.ID, ""); err != nil {
			return fmt.Errorf("clear unverified password for user %d: %w", u.ID, err)
		}
		u.Password = nil
	}
	if err := s.identities.DeleteIdentitiesByUser(ctx, u.ID); err != nil {
		return fmt.Errorf("drop identities for user %d: %w", u.ID, err)
	}
	if err := s.sessions.RevokeAllSessions(ctx, u.ID); err != nil {
		return fmt.Errorf("end sessions for user %d: %w", u.ID, err)
	}
	return nil
}

func (s *service) issueTokenPair(ctx context.Context, u *User) (string, string, *User, error) {
	access, err := s.tokens.IssueAccessToken(u.ID, string(u.Role))
	if err != nil {
		return "", "", nil, fmt.Errorf("issue access token: %w", err)
	}
	refresh, claims, err := s.tokens.IssueRefreshToken(u.ID, string(u.Role))
	if err != nil {
		return "", "", nil, fmt.Errorf("issue refresh token: %w", err)
	}
	if err := s.sessions.CreateSession(ctx, &Session{
		ID:        claims.ID,
		UserID:    u.ID,
		ExpiresAt: claims.ExpiresAt.Time,
	}); err != nil {
		return "", "", nil, fmt.Errorf("store session: %w", err)
	}
	return access, refresh, u, nil
}
func (s *service) LinkIdentity(ctx context.Context, userID uint64, provider, identifier string) (*Identity, error) {
	provider = strings.TrimSpace(provider)
	identifier = strings.TrimSpace(identifier)
	if provider == "" || identifier == "" {
		return nil, constants.ErrIdentityInvalid
	}
	// A Google identity is only proven by the OAuth callback. Linking one by
	// hand would let a user claim someone else's Google sub.
	prov := constants.Provider(provider)
	if !prov.IsValid() || prov == constants.ProviderGoogle {
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

// validateNewPassword also caps the length: bcrypt rejects input over 72
// bytes, which would otherwise surface as a 500.
func validateNewPassword(password string) error {
	switch {
	case len(password) < 8:
		return constants.ErrPasswordTooShort
	case len(password) > 72:
		return constants.ErrPasswordTooLong
	}
	return nil
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
