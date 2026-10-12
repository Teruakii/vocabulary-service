package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vocabulary-service/internal/constants"
)

// fakeResets is an in-memory PasswordResetRepository for tests.
type fakeResets struct {
	byHash map[string]*PasswordReset
}

func newFakeResets() *fakeResets {
	return &fakeResets{byHash: map[string]*PasswordReset{}}
}

func (f *fakeResets) CreatePasswordReset(_ context.Context, r *PasswordReset) error {
	f.byHash[r.TokenHash] = r
	return nil
}

func (f *fakeResets) GetPasswordReset(_ context.Context, tokenHash string) (*PasswordReset, error) {
	r, ok := f.byHash[tokenHash]
	if !ok {
		return nil, constants.ErrNotFound
	}
	return r, nil
}

func (f *fakeResets) DeletePasswordResets(_ context.Context, userID uint64) error {
	for h, r := range f.byHash {
		if r.UserID == userID {
			delete(f.byHash, h)
		}
	}
	return nil
}

type sentMail struct{ to, subject, body string }

type fakeMailer struct{ sent []sentMail }

func (m *fakeMailer) Send(_ context.Context, to, subject, body string) error {
	m.sent = append(m.sent, sentMail{to, subject, body})
	return nil
}

var resetLink = regexp.MustCompile(`https://app\.example/reset-password\?token=([A-Za-z0-9_-]+)`)

func TestPasswordReset_EmailedLinkWorksOnce(t *testing.T) {
	users := new(MockUserRepository)
	sessions := newFakeSessions()
	tokens := testTokens()
	mailer := &fakeMailer{}
	svc := NewService(users, new(MockIdentityRepository), sessions, tokens,
		WithAdminEmails([]string{"boss@example.com"}),
		WithPasswordReset(newFakeResets(), mailer, "https://app.example/"))

	boss := &User{ID: 1, Email: "boss@example.com", Role: constants.RoleUser}
	oldSession := sessions.issue(t, tokens, 1)
	users.On("GetByEmail", context.Background(), "boss@example.com").Return(boss, nil)
	users.On("GetByID", context.Background(), uint64(1)).Return(boss, nil)
	users.On("UpdatePassword", context.Background(), uint64(1), mock.AnythingOfType("string")).Return(nil).Once()
	users.On("Update", context.Background(), mock.MatchedBy(func(u *User) bool {
		return u.EmailVerifiedAt != nil && u.Role == constants.RoleAdmin
	})).Return(nil).Once()

	assert.NoError(t, svc.RequestPasswordReset(context.Background(), " Boss@Example.com "))

	assert.Len(t, mailer.sent, 1)
	assert.Equal(t, "boss@example.com", mailer.sent[0].to)
	m := resetLink.FindStringSubmatch(mailer.sent[0].body)
	if !assert.NotNil(t, m, "email carries the reset link: %q", mailer.sent[0].body) {
		return
	}
	token := m[1]

	assert.NoError(t, svc.ResetPassword(context.Background(), token, "brandNew123"))
	assert.NotNil(t, boss.EmailVerifiedAt)
	assert.Equal(t, constants.RoleAdmin, boss.Role)
	users.AssertExpectations(t)
	_, _, err := svc.Refresh(context.Background(), oldSession)
	assert.ErrorIs(t, err, constants.ErrInvalidToken, "old sessions end")

	assert.ErrorIs(t, svc.ResetPassword(context.Background(), token, "another123"), constants.ErrResetTokenInvalid)
}

func TestRequestPasswordReset_UnknownEmailSendsNothing(t *testing.T) {
	users := new(MockUserRepository)
	mailer := &fakeMailer{}
	resets := newFakeResets()
	svc := NewService(users, new(MockIdentityRepository), newFakeSessions(), testTokens(),
		WithPasswordReset(resets, mailer, "https://app.example"))
	users.On("GetByEmail", context.Background(), "ghost@example.com").Return(nil, constants.ErrNotFound)

	assert.NoError(t, svc.RequestPasswordReset(context.Background(), "ghost@example.com"))
	assert.Empty(t, mailer.sent)
	assert.Empty(t, resets.byHash)
}

func TestRequestPasswordReset_UnavailableWithoutMailer(t *testing.T) {
	users := new(MockUserRepository)
	svc := NewService(users, new(MockIdentityRepository), newFakeSessions(), testTokens(),
		WithPasswordReset(newFakeResets(), nil, "https://app.example"))

	err := svc.RequestPasswordReset(context.Background(), "boss@example.com")

	assert.ErrorIs(t, err, constants.ErrPasswordResetUnavailable)
	users.AssertNotCalled(t, "GetByEmail")
}

func TestResetPassword_ExpiredTokenRejected(t *testing.T) {
	users := new(MockUserRepository)
	resets := newFakeResets()
	svc := NewService(users, new(MockIdentityRepository), newFakeSessions(), testTokens(),
		WithPasswordReset(resets, &fakeMailer{}, "https://app.example"))
	resets.byHash[hashResetToken("old-token")] = &PasswordReset{
		TokenHash: hashResetToken("old-token"), UserID: 1, Email: "a@example.com",
		ExpiresAt: time.Now().Add(-time.Minute),
	}

	err := svc.ResetPassword(context.Background(), "old-token", "brandNew123")

	assert.ErrorIs(t, err, constants.ErrResetTokenInvalid)
	users.AssertNotCalled(t, "UpdatePassword")
}

// A link sent to the account's old address must not verify the new one.
func TestResetPassword_RejectedAfterEmailChange(t *testing.T) {
	users := new(MockUserRepository)
	resets := newFakeResets()
	svc := NewService(users, new(MockIdentityRepository), newFakeSessions(), testTokens(),
		WithAdminEmails([]string{"boss@example.com"}),
		WithPasswordReset(resets, &fakeMailer{}, "https://app.example"))
	resets.byHash[hashResetToken("tok")] = &PasswordReset{
		TokenHash: hashResetToken("tok"), UserID: 1, Email: "attacker@example.com",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	users.On("GetByID", context.Background(), uint64(1)).
		Return(&User{ID: 1, Email: "boss@example.com", Role: constants.RoleUser}, nil)

	err := svc.ResetPassword(context.Background(), "tok", "brandNew123")

	assert.ErrorIs(t, err, constants.ErrResetTokenInvalid)
	users.AssertNotCalled(t, "UpdatePassword")
	users.AssertNotCalled(t, "Update")
}

type forgotStub struct {
	Service
	err error
}

func (s forgotStub) RequestPasswordReset(context.Context, string) error { return s.err }

func TestForgotPasswordHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	post := func(h *Handler) *httptest.ResponseRecorder {
		router := gin.New()
		router.POST("/forgot-password", h.ForgotPassword)
		req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(`{"email":"a@example.com"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	unavailable := NewHandler(forgotStub{err: constants.ErrPasswordResetUnavailable}, 0, false, nil, "")
	assert.Equal(t, http.StatusServiceUnavailable, post(unavailable).Code)

	h := NewHandler(forgotStub{}, 0, false, nil, "")
	for i := 0; i < 5; i++ {
		assert.Equal(t, http.StatusNoContent, post(h).Code, "request %d", i+1)
	}
	w := post(h)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.NotEmpty(t, w.Header().Get("Retry-After"))
}
