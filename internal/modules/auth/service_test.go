package auth

import (
	"context"
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
	svc := NewService(users, identities, testTokens())

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
	svc := NewService(users, identities, testTokens())

	_, err := svc.Register(context.Background(), "test@example.com", "short")

	assert.ErrorIs(t, err, constants.ErrPasswordTooShort)
	users.AssertNotCalled(t, "Create")
}

func TestLogin_Success(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, testTokens())

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
	svc := NewService(users, identities, testTokens())

	hashed := mustHash(t, "password123")
	stored := &User{ID: 1, Email: "test@example.com", Password: &hashed}
	users.On("GetByEmail", context.Background(), "test@example.com").Return(stored, nil)

	_, _, _, err := svc.Login(context.Background(), "test@example.com", "wrong")

	assert.ErrorIs(t, err, constants.ErrInvalidCredentials)
}

func TestLogin_EmailNotFound_SameErrorAsWrongPassword(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, testTokens())

	users.On("GetByEmail", context.Background(), "ghost@example.com").Return(nil, constants.ErrNotFound)

	_, _, _, err := svc.Login(context.Background(), "ghost@example.com", "whatever")

	assert.ErrorIs(t, err, constants.ErrInvalidCredentials)
}

func TestLogin_OAuthOnlyAccount_HasNoPassword(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, testTokens())

	stored := &User{ID: 1, Email: "test@example.com", Password: nil}
	users.On("GetByEmail", context.Background(), "test@example.com").Return(stored, nil)

	_, _, _, err := svc.Login(context.Background(), "test@example.com", "anything")

	assert.ErrorIs(t, err, constants.ErrInvalidCredentials)
}

func TestRefresh_CutOffAfterPasswordChange(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	tokens := testTokens()
	svc := NewService(users, identities, tokens)

	refreshToken, err := tokens.IssueRefreshToken(1, string(constants.RoleUser))
	assert.NoError(t, err)

	stored := &User{ID: 1, Role: constants.RoleUser, UpdatedAt: time.Now().Add(time.Hour)}
	users.On("GetByID", context.Background(), uint64(1)).Return(stored, nil)

	_, _, err = svc.Refresh(context.Background(), refreshToken)

	assert.ErrorIs(t, err, constants.ErrInvalidToken)
}

func TestRefresh_RejectsAccessTokenUsedAsRefresh(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	tokens := testTokens()
	svc := NewService(users, identities, tokens)

	accessToken, err := tokens.IssueAccessToken(1, string(constants.RoleUser))
	assert.NoError(t, err)

	_, _, err = svc.Refresh(context.Background(), accessToken)

	assert.ErrorIs(t, err, constants.ErrInvalidToken)
	users.AssertNotCalled(t, "GetByID")
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, testTokens())

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
	svc := NewService(users, identities, testTokens())

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
	svc := NewService(users, identities, testTokens())

	identities.On("GetIdentityByID", context.Background(), uint64(10)).
		Return(&Identity{ID: 10, UserID: 2}, nil)

	err := svc.UnlinkIdentity(context.Background(), 1, 10)

	assert.ErrorIs(t, err, constants.ErrIdentityNotFound)
	identities.AssertNotCalled(t, "DeleteIdentity")
}

func TestUnlinkIdentity_BlocksLastOneWhenNoPassword(t *testing.T) {
	users := new(MockUserRepository)
	identities := new(MockIdentityRepository)
	svc := NewService(users, identities, testTokens())

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
	svc := NewService(users, identities, testTokens())

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
	svc := NewService(users, identities, testTokens())

	identities.On("CreateIdentity", context.Background(), mock.Anything).
		Return(constants.ErrIdentityDuplicate)

	_, err := svc.LinkIdentity(context.Background(), 1, "google", "sub-123")

	assert.ErrorIs(t, err, constants.ErrIdentityDuplicate)
}
