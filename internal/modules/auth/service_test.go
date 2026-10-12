package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/infra/jwt"
)

func testTokens() *jwt.Signer {
	return jwt.NewSigner(jwt.Config{
		AccessSecret:  []byte("test-access-secret"),
		RefreshSecret: []byte("test-refresh-secret"),
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
		Issuer:        "vocabulary-service-test",
	})
}

func mustHash(t *testing.T, plain string) string {
	t.Helper()
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	assert.NoError(t, err)
	return string(b)
}

func TestRegister_Success(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	users.On("Create", context.Background(), mock.MatchedBy(func(u *User) bool {
		return u.Email == "test@example.com" && u.HasPassword() && u.Role == constants.RoleUser
	})).Return(nil)

	u, err := svc.Register(context.Background(), "  Test@Example.com  ", "password123")

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", u.Email)
	assert.NotEqual(t, "password123", *u.Password)
	users.AssertExpectations(t)
}

func TestRegister_PasswordTooShort(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	_, err := svc.Register(context.Background(), "test@example.com", "short")

	assert.ErrorIs(t, err, constants.ErrPasswordTooShort)
	users.AssertNotCalled(t, "Create")
}

func TestLogin_Success(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	hashed := mustHash(t, "password123")
	stored := &User{ID: 1, Email: "test@example.com", Role: constants.RoleUser, Password: &hashed}
	users.On("GetByEmail", context.Background(), "test@example.com").Return(stored, nil)

	access, refresh, u, err := svc.Login(context.Background(), "Test@Example.com", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
	assert.Equal(t, uint64(1), u.ID)
	users.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	hashed := mustHash(t, "password123")
	stored := &User{ID: 1, Email: "test@example.com", Password: &hashed}
	users.On("GetByEmail", context.Background(), "test@example.com").Return(stored, nil)

	_, _, _, err := svc.Login(context.Background(), "test@example.com", "wrong")

	assert.ErrorIs(t, err, constants.ErrInvalidCredentials)
}

func TestLogin_EmailNotFound_SameErrorAsWrongPassword(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	users.On("GetByEmail", context.Background(), "ghost@example.com").Return(nil, constants.ErrNotFound)

	_, _, _, err := svc.Login(context.Background(), "ghost@example.com", "whatever")

	assert.ErrorIs(t, err, constants.ErrInvalidCredentials)
}

func TestLogin_OAuthOnlyAccount_HasNoPassword(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	stored := &User{ID: 1, Email: "test@example.com", Password: nil}
	users.On("GetByEmail", context.Background(), "test@example.com").Return(stored, nil)

	_, _, _, err := svc.Login(context.Background(), "test@example.com", "anything")

	assert.ErrorIs(t, err, constants.ErrInvalidCredentials)
}

func TestRefresh_CutOffAfterPasswordChange(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	tokens := testTokens()
	sessions := newFakeSessions()
	svc := NewService(users, identities, sessions, tokens)

	refreshToken := sessions.issue(t, tokens, 1)

	changedAt := time.Now().Add(time.Hour)
	stored := &User{ID: 1, Role: constants.RoleUser, PasswordChangedAt: &changedAt}
	users.On("GetByID", context.Background(), uint64(1)).Return(stored, nil)

	_, _, err := svc.Refresh(context.Background(), refreshToken)

	assert.ErrorIs(t, err, constants.ErrInvalidToken)
}

func TestRefresh_ProfileUpdateDoesNotEndSession(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	tokens := testTokens()
	sessions := newFakeSessions()
	svc := NewService(users, identities, sessions, tokens)

	refreshToken := sessions.issue(t, tokens, 1)

	// A just-created OAuth user or an email change bumps UpdatedAt past the
	// token's second-precision iat; that must not log the user out.
	stored := &User{ID: 1, Role: constants.RoleUser, UpdatedAt: time.Now().Add(time.Hour)}
	users.On("GetByID", context.Background(), uint64(1)).Return(stored, nil)

	access, _, err := svc.Refresh(context.Background(), refreshToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, access)
}

func TestRefresh_RejectsAccessTokenUsedAsRefresh(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	tokens := testTokens()
	sessions := newFakeSessions()
	svc := NewService(users, identities, sessions, tokens)

	accessToken, err := tokens.IssueAccessToken(1, string(constants.RoleUser))
	assert.NoError(t, err)

	_, _, err = svc.Refresh(context.Background(), accessToken)

	assert.ErrorIs(t, err, constants.ErrInvalidToken)
	users.AssertNotCalled(t, "GetByID")
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	hashed := mustHash(t, "correctOld123")
	stored := &User{ID: 1, Password: &hashed}
	users.On("GetByID", context.Background(), uint64(1)).Return(stored, nil)

	err := svc.ChangePassword(context.Background(), 1, "wrongOld", "brandNew123")

	assert.ErrorIs(t, err, constants.ErrInvalidCredentials)
	users.AssertNotCalled(t, "UpdatePassword")
}

func TestChangePassword_OAuthOnlyAccount_SkipsCurrentPasswordCheck(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	stored := &User{ID: 1, Password: nil}
	users.On("GetByID", context.Background(), uint64(1)).Return(stored, nil)
	users.On("UpdatePassword", context.Background(), uint64(1), mock.AnythingOfType("string")).Return(nil)

	err := svc.ChangePassword(context.Background(), 1, "", "brandNew123")

	assert.NoError(t, err)
	users.AssertExpectations(t)
}

func TestUnlinkIdentity_WrongOwner(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	identities.On("GetIdentityByID", context.Background(), uint64(10)).
		Return(&Identity{ID: 10, UserID: 2}, nil)

	err := svc.UnlinkIdentity(context.Background(), 1, 10)

	assert.ErrorIs(t, err, constants.ErrIdentityNotFound)
	identities.AssertNotCalled(t, "DeleteIdentity")
}

func TestUnlinkIdentity_BlocksLastOneWhenNoPassword(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	identities.On("GetIdentityByID", context.Background(), uint64(10)).
		Return(&Identity{ID: 10, UserID: 1}, nil)
	users.On("GetByID", context.Background(), uint64(1)).
		Return(&User{ID: 1, Password: nil}, nil)
	identities.On("CountIdentitiesByUser", context.Background(), uint64(1)).
		Return(1, nil)

	err := svc.UnlinkIdentity(context.Background(), 1, 10)

	assert.ErrorIs(t, err, constants.ErrCannotUnlinkLastIdentity)
	identities.AssertNotCalled(t, "DeleteIdentity")
}

func TestUnlinkIdentity_AllowedWhenPasswordExists(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	hashed := mustHash(t, "somepassword")
	identities.On("GetIdentityByID", context.Background(), uint64(10)).
		Return(&Identity{ID: 10, UserID: 1}, nil)
	users.On("GetByID", context.Background(), uint64(1)).
		Return(&User{ID: 1, Password: &hashed}, nil)
	identities.On("DeleteIdentity", context.Background(), uint64(10)).Return(nil)

	err := svc.UnlinkIdentity(context.Background(), 1, 10)

	assert.NoError(t, err)
	identities.AssertNotCalled(t, "CountIdentitiesByUser")
	identities.AssertExpectations(t)
}

func TestLinkIdentity_Duplicate(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens())

	identities.On("CreateIdentity", context.Background(), mock.Anything).
		Return(constants.ErrIdentityDuplicate)

	_, err := svc.LinkIdentity(context.Background(), 1, "local", "alice")

	assert.ErrorIs(t, err, constants.ErrIdentityDuplicate)
}

func TestLinkIdentity_GoogleRequiresOAuth(t *testing.T) {
	identities := new(MockIdentityRepository)
	svc := NewService(new(MockUserRepository), identities, newFakeSessions(), testTokens())

	_, err := svc.LinkIdentity(context.Background(), 1, "google", "someone-elses-sub")

	assert.ErrorIs(t, err, constants.ErrIdentityInvalid)
	identities.AssertNotCalled(t, "CreateIdentity")
}

// fakeSessions is an in-memory SessionRepository for tests.
type fakeSessions struct {
	byID map[string]*Session
}

func newFakeSessions() *fakeSessions {
	return &fakeSessions{byID: map[string]*Session{}}
}

func (f *fakeSessions) CreateSession(_ context.Context, s *Session) error {
	f.byID[s.ID] = s
	return nil
}

func (f *fakeSessions) GetSession(_ context.Context, id string) (*Session, error) {
	s, ok := f.byID[id]
	if !ok {
		return nil, constants.ErrNotFound
	}
	return s, nil
}

func (f *fakeSessions) RevokeSession(_ context.Context, id string) error {
	if s, ok := f.byID[id]; ok && s.RevokedAt == nil {
		now := time.Now()
		s.RevokedAt = &now
	}
	return nil
}

func (f *fakeSessions) RevokeAllSessions(ctx context.Context, userID uint64) error {
	for id, s := range f.byID {
		if s.UserID == userID {
			_ = f.RevokeSession(ctx, id)
		}
	}
	return nil
}

// issue creates a refresh token for userID with a stored, active session.
func (f *fakeSessions) issue(t *testing.T, tokens *jwt.Signer, userID uint64) string {
	t.Helper()
	token, claims, err := tokens.IssueRefreshToken(userID, string(constants.RoleUser))
	assert.NoError(t, err)
	f.byID[claims.ID] = &Session{ID: claims.ID, UserID: userID, ExpiresAt: claims.ExpiresAt.Time}
	return token
}

func TestRefresh_RejectsAccessTokenEvenWithSharedSecret(t *testing.T) {
	// Production signs both token kinds with JWT_SECRET_KEY.
	tokens := jwt.NewSigner(jwt.Config{
		AccessSecret:  []byte("same-secret"),
		RefreshSecret: []byte("same-secret"),
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    time.Hour,
	})
	users := new(MockUserRepository)
	svc := NewService(users, new(MockIdentityRepository), newFakeSessions(), tokens)

	accessToken, err := tokens.IssueAccessToken(1, string(constants.RoleUser))
	assert.NoError(t, err)

	_, _, err = svc.Refresh(context.Background(), accessToken)

	assert.ErrorIs(t, err, constants.ErrInvalidToken)
	users.AssertNotCalled(t, "GetByID")
}

func TestLogout_EndsSession(t *testing.T) {
	users := new(MockUserRepository)
	tokens := testTokens()
	sessions := newFakeSessions()
	svc := NewService(users, new(MockIdentityRepository), sessions, tokens)
	users.On("GetByID", context.Background(), uint64(1)).Return(&User{ID: 1, Role: constants.RoleUser}, nil)

	refreshToken := sessions.issue(t, tokens, 1)
	_, _, err := svc.Refresh(context.Background(), refreshToken)
	assert.NoError(t, err)

	assert.NoError(t, svc.Logout(context.Background(), refreshToken))

	_, _, err = svc.Refresh(context.Background(), refreshToken)
	assert.ErrorIs(t, err, constants.ErrInvalidToken)
}

func TestRefresh_RejectsUnknownSession(t *testing.T) {
	users := new(MockUserRepository)
	tokens := testTokens()
	svc := NewService(users, new(MockIdentityRepository), newFakeSessions(), tokens)

	// Signed correctly, but never stored (e.g. issued before sessions existed).
	refreshToken, _, err := tokens.IssueRefreshToken(1, string(constants.RoleUser))
	assert.NoError(t, err)

	_, _, err = svc.Refresh(context.Background(), refreshToken)

	assert.ErrorIs(t, err, constants.ErrInvalidToken)
}

func TestChangePassword_EndsAllSessions(t *testing.T) {
	users := new(MockUserRepository)
	tokens := testTokens()
	sessions := newFakeSessions()
	svc := NewService(users, new(MockIdentityRepository), sessions, tokens)

	hashed := mustHash(t, "oldpassword1")
	users.On("GetByID", context.Background(), uint64(1)).Return(&User{ID: 1, Password: &hashed}, nil)
	users.On("UpdatePassword", context.Background(), uint64(1), mock.AnythingOfType("string")).Return(nil)

	phone := sessions.issue(t, tokens, 1)
	laptop := sessions.issue(t, tokens, 1)

	assert.NoError(t, svc.ChangePassword(context.Background(), 1, "oldpassword1", "newpassword1"))

	for _, tok := range []string{phone, laptop} {
		_, _, err := svc.Refresh(context.Background(), tok)
		assert.ErrorIs(t, err, constants.ErrInvalidToken)
	}
}

// Registering proves nothing about the email, so a listed address must not
// get admin from sign-up alone.
func TestRegister_AdminEmailStaysUserUntilVerified(t *testing.T) {
	users := new(MockUserRepository)
	svc := NewService(users, new(MockIdentityRepository), newFakeSessions(), testTokens(),
		WithAdminEmails([]string{" Boss@Example.com "}))

	users.On("Create", context.Background(), mock.MatchedBy(func(u *User) bool {
		return u.Role == constants.RoleUser && u.EmailVerifiedAt == nil
	})).Return(nil).Once()

	boss, err := svc.Register(context.Background(), "boss@example.com", "password123")

	assert.NoError(t, err)
	assert.Equal(t, constants.RoleUser, boss.Role)
	users.AssertExpectations(t)
}

func TestRoleFor(t *testing.T) {
	svc := NewService(new(MockUserRepository), new(MockIdentityRepository), newFakeSessions(), testTokens(),
		WithAdminEmails([]string{"boss@example.com"})).(*service)
	verified := time.Now()

	cases := []struct {
		name string
		user User
		want constants.Role
	}{
		{"listed and verified", User{Email: "boss@example.com", Role: constants.RoleUser, EmailVerifiedAt: &verified}, constants.RoleAdmin},
		{"listed teacher, verified", User{Email: "boss@example.com", Role: constants.RoleTeacher, EmailVerifiedAt: &verified}, constants.RoleAdmin},
		{"listed admin, unverified", User{Email: "boss@example.com", Role: constants.RoleAdmin}, constants.RoleUser},
		{"unlisted admin, verified", User{Email: "x@example.com", Role: constants.RoleAdmin, EmailVerifiedAt: &verified}, constants.RoleUser},
		{"unlisted teacher", User{Email: "x@example.com", Role: constants.RoleTeacher, EmailVerifiedAt: &verified}, constants.RoleTeacher},
		{"unlisted user", User{Email: "x@example.com", Role: constants.RoleUser}, constants.RoleUser},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, svc.roleFor(&tc.user))
		})
	}
}

func TestUpdateProfile_EmailChangeClearsVerificationAndAdmin(t *testing.T) {
	users := new(MockUserRepository)
	svc := NewService(users, new(MockIdentityRepository), newFakeSessions(), testTokens(),
		WithAdminEmails([]string{"boss@example.com", "other-boss@example.com"}))

	verified := time.Now()
	stored := &User{ID: 1, Email: "boss@example.com", Role: constants.RoleAdmin, EmailVerifiedAt: &verified}
	users.On("GetByID", context.Background(), uint64(1)).Return(stored, nil)
	users.On("Update", context.Background(), mock.MatchedBy(func(u *User) bool {
		return u.ID == 1 && u.Email == "other-boss@example.com" && u.EmailVerifiedAt == nil && u.Role == constants.RoleUser
	})).Return(nil).Once()

	u, err := svc.UpdateProfile(context.Background(), 1, "Other-Boss@Example.com")

	assert.NoError(t, err)
	assert.Equal(t, constants.RoleUser, u.Role)
	users.AssertExpectations(t)
}

func TestUpdateProfile_SameEmailKeepsVerification(t *testing.T) {
	users := new(MockUserRepository)
	svc := NewService(users, new(MockIdentityRepository), newFakeSessions(), testTokens(),
		WithAdminEmails([]string{"boss@example.com"}))

	verified := time.Now()
	stored := &User{ID: 1, Email: "boss@example.com", Role: constants.RoleAdmin, EmailVerifiedAt: &verified}
	users.On("GetByID", context.Background(), uint64(1)).Return(stored, nil)

	u, err := svc.UpdateProfile(context.Background(), 1, " BOSS@example.com ")

	assert.NoError(t, err)
	assert.Equal(t, constants.RoleAdmin, u.Role)
	assert.NotNil(t, u.EmailVerifiedAt)
	users.AssertNotCalled(t, "Update")
}

func TestLoginLimiter_LocksAfterRepeatedFailures(t *testing.T) {
	l := newLoginLimiter(3, 15*time.Minute, 10*time.Minute)
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	l.now = func() time.Time { return now }

	key := "1.2.3.4|a@example.com"
	l.fail(key)
	l.fail(key)
	assert.Zero(t, l.retryAfter(key), "still allowed before the limit")

	l.fail(key)
	assert.Equal(t, 10*time.Minute, l.retryAfter(key))
	assert.Zero(t, l.retryAfter("1.2.3.4|b@example.com"), "other emails unaffected")

	now = now.Add(10*time.Minute + time.Second)
	assert.Zero(t, l.retryAfter(key), "lock expires")
}

func TestLoginLimiter_SuccessResetsCount(t *testing.T) {
	l := newLoginLimiter(3, 15*time.Minute, 10*time.Minute)
	key := "ip|a@example.com"
	l.fail(key)
	l.fail(key)
	l.succeed(key)
	l.fail(key)
	l.fail(key)
	assert.Zero(t, l.retryAfter(key))
}

func TestOAuthLogin_MatchingPasswordAccount_DropsPasswordAndSessions(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	sessions := newFakeSessions()
	tokens := testTokens()
	svc := NewService(users, identities, sessions, tokens)

	// Someone registered the victim's email with a password before the
	// victim ever signed in with Google.
	hash := mustHash(t, "attacker-pass")
	squatted := &User{ID: 1, Email: "victim@gmail.com", Password: &hash, Role: constants.RoleUser}
	squatterRefresh := sessions.issue(t, tokens, 1)

	identities.On("FindIdentityByProviderIdentifier", context.Background(), "google", "sub-1").Return(nil, constants.ErrNotFound)
	users.On("GetByEmail", context.Background(), "victim@gmail.com").Return(squatted, nil)
	users.On("UpdatePassword", context.Background(), uint64(1), "").Return(nil)
	identities.On("DeleteIdentitiesByUser", context.Background(), uint64(1)).Return(nil)
	identities.On("CreateIdentity", context.Background(), mock.Anything).Return(nil)
	users.On("Update", context.Background(), mock.Anything).Return(nil)

	_, _, u, err := svc.LoginOrRegisterWithOAuth(context.Background(), "google", "sub-1", "victim@gmail.com", "", "")

	assert.NoError(t, err)
	assert.False(t, u.HasPassword())
	assert.NotNil(t, u.EmailVerifiedAt)
	users.AssertCalled(t, "UpdatePassword", context.Background(), uint64(1), "")
	users.On("GetByID", context.Background(), uint64(1)).Return(squatted, nil)
	_, _, err = svc.Refresh(context.Background(), squatterRefresh)
	assert.ErrorIs(t, err, constants.ErrInvalidToken)
}

// An attacker signs up with their own Google account, then sets the victim's
// email on it with PUT /auth/me. When the victim first signs in with Google
// they must not end up in an account the attacker can still open.
func TestOAuthLogin_EmailSetByAnotherGoogleAccount_DropsItsAccess(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	sessions := newFakeSessions()
	tokens := testTokens()
	svc := NewService(users, identities, sessions, tokens)

	planted := &User{ID: 1, Email: "victim@gmail.com", Role: constants.RoleUser}
	attackerRefresh := sessions.issue(t, tokens, 1)

	identities.On("FindIdentityByProviderIdentifier", context.Background(), "google", "victim-sub").Return(nil, constants.ErrNotFound)
	users.On("GetByEmail", context.Background(), "victim@gmail.com").Return(planted, nil)
	identities.On("DeleteIdentitiesByUser", context.Background(), uint64(1)).Return(nil)
	identities.On("CreateIdentity", context.Background(), mock.Anything).Return(nil)
	users.On("Update", context.Background(), mock.Anything).Return(nil)

	_, _, _, err := svc.LoginOrRegisterWithOAuth(context.Background(), "google", "victim-sub", "victim@gmail.com", "", "")

	assert.NoError(t, err)
	identities.AssertCalled(t, "DeleteIdentitiesByUser", context.Background(), uint64(1))
	users.AssertNotCalled(t, "UpdatePassword")
	users.On("GetByID", context.Background(), uint64(1)).Return(planted, nil)
	_, _, err = svc.Refresh(context.Background(), attackerRefresh)
	assert.ErrorIs(t, err, constants.ErrInvalidToken)
}

// The owner already proved this inbox, so signing in with Google for the
// first time only adds the identity.
func TestOAuthLogin_MatchingVerifiedAccount_KeepsPasswordAndSessions(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	sessions := newFakeSessions()
	tokens := testTokens()
	svc := NewService(users, identities, sessions, tokens)

	hash := mustHash(t, "owner-pass")
	verified := time.Now()
	owner := &User{ID: 1, Email: "owner@gmail.com", Password: &hash, Role: constants.RoleUser, EmailVerifiedAt: &verified}
	ownerRefresh := sessions.issue(t, tokens, 1)

	identities.On("FindIdentityByProviderIdentifier", context.Background(), "google", "owner-sub").Return(nil, constants.ErrNotFound)
	users.On("GetByEmail", context.Background(), "owner@gmail.com").Return(owner, nil)
	users.On("GetByID", context.Background(), uint64(1)).Return(owner, nil)
	identities.On("CreateIdentity", context.Background(), mock.MatchedBy(func(i *Identity) bool {
		return i.UserID == 1 && i.Identifier == "owner-sub"
	})).Return(nil)

	_, _, u, err := svc.LoginOrRegisterWithOAuth(context.Background(), "google", "owner-sub", "owner@gmail.com", "", "")

	assert.NoError(t, err)
	assert.True(t, u.HasPassword())
	users.AssertNotCalled(t, "UpdatePassword")
	identities.AssertNotCalled(t, "DeleteIdentitiesByUser")
	identities.AssertExpectations(t)
	_, _, err = svc.Refresh(context.Background(), ownerRefresh)
	assert.NoError(t, err, "existing session survives")
}

// Google proves the email it reports, not whatever email the account was
// changed to afterwards.
func TestOAuthLogin_KnownIdentityWithChangedEmail_DoesNotVerifyIt(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens(),
		WithAdminEmails([]string{"boss@example.com"}))

	changed := &User{ID: 1, Email: "boss@example.com", Role: constants.RoleUser}
	identities.On("FindIdentityByProviderIdentifier", context.Background(), "google", "attacker-sub").
		Return(&Identity{ID: 5, UserID: 1}, nil)
	users.On("GetByID", context.Background(), uint64(1)).Return(changed, nil)

	_, _, u, err := svc.LoginOrRegisterWithOAuth(context.Background(), "google", "attacker-sub", "attacker@gmail.com", "", "")

	assert.NoError(t, err)
	assert.Nil(t, u.EmailVerifiedAt)
	assert.Equal(t, constants.RoleUser, u.Role)
	users.AssertNotCalled(t, "Update")
}

func TestOAuthLogin_KnownIdentity_VerifiesAndPromotesListedEmail(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, newFakeSessions(), testTokens(),
		WithAdminEmails([]string{"boss@example.com"}))

	boss := &User{ID: 1, Email: "boss@example.com", Role: constants.RoleUser}
	identities.On("FindIdentityByProviderIdentifier", context.Background(), "google", "boss-sub").
		Return(&Identity{ID: 5, UserID: 1}, nil)
	users.On("GetByID", context.Background(), uint64(1)).Return(boss, nil)
	users.On("Update", context.Background(), mock.MatchedBy(func(u *User) bool {
		return u.EmailVerifiedAt != nil && u.Role == constants.RoleAdmin
	})).Return(nil).Once()

	_, _, u, err := svc.LoginOrRegisterWithOAuth(context.Background(), "google", "boss-sub", "Boss@Example.com", "", "")

	assert.NoError(t, err)
	assert.Equal(t, constants.RoleAdmin, u.Role)
	users.AssertExpectations(t)
}

func TestPasswordTooLong_RejectedBeforeHashing(t *testing.T) {
	long := strings.Repeat("a", 73)

	users := new(MockUserRepository)
	svc := NewService(users, new(MockIdentityRepository), newFakeSessions(), testTokens())

	_, err := svc.Register(context.Background(), "test@example.com", long)
	assert.ErrorIs(t, err, constants.ErrPasswordTooLong)
	users.AssertNotCalled(t, "Create")

	assert.ErrorIs(t, svc.AdminResetPassword(context.Background(), 1, long), constants.ErrPasswordTooLong)

	users.On("GetByID", context.Background(), uint64(1)).Return(&User{ID: 1}, nil)
	assert.ErrorIs(t, svc.ChangePassword(context.Background(), 1, "", long), constants.ErrPasswordTooLong)
}
