package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/middleware"
)

type deleteStub struct {
	Service
	err error
}

func (s deleteStub) Delete(context.Context, uint64) error { return s.err }

func TestDeleteHandler_AlreadyDeletedUser_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(deleteStub{err: constants.ErrUserNotFound}, 0, false, nil, "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/auth/me", nil)
	c.Set(middleware.ContextKeyUserID, uint64(7))

	h.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// brokenSessions simulates the database failing under the session lookup.
type brokenSessions struct{}

var errDBDown = errors.New("db down")

func (brokenSessions) CreateSession(context.Context, *Session) error { return errDBDown }
func (brokenSessions) GetSession(context.Context, string) (*Session, error) {
	return nil, errDBDown
}
func (brokenSessions) RevokeSession(context.Context, string) error     { return errDBDown }
func (brokenSessions) RevokeAllSessions(context.Context, uint64) error { return errDBDown }

func TestRefresh_DatabaseErrorIsServerErrorAndKeepsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := testTokens()
	refresh, _, err := tokens.IssueRefreshToken(1, "user")
	assert.NoError(t, err)

	svc := NewService(new(MockUserRepository), new(MockIdentityRepository), brokenSessions{}, tokens)
	h := NewHandler(svc, time.Hour, false, nil, "")
	router := gin.New()
	router.POST("/refresh", h.Refresh)

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: refresh})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	for _, c := range w.Result().Cookies() {
		if c.Name == refreshCookieName {
			assert.GreaterOrEqual(t, c.MaxAge, 0, "refresh cookie was cleared by a transient error")
		}
	}
}
