package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/infra/http/response"
	"vocabulary-service/internal/infra/oauth"
	"vocabulary-service/internal/middleware"
)

const refreshCookieName = "refresh_token"
const oauthStateCookie = "oauth_state"

// Cookie paths must match the routes mounted under /api/v1/auth, otherwise
// the browser never sends the cookies back.
const authCookiePath = "/api/v1/auth"
const oauthCookiePath = authCookiePath + "/google"

type Handler struct {
	service       Service
	refreshTTL    time.Duration
	secureCookies bool
	google        *oauth.GoogleClient
	// frontendURL, when set, makes the Google callback redirect back to the web
	// app instead of answering with JSON.
	frontendURL string
	logins      *loginLimiter
	// resetRequests caps reset emails per IP and address, so the endpoint
	// cannot be used to flood an inbox.
	resetRequests *loginLimiter
}

func NewHandler(service Service, refreshTTL time.Duration, secureCookies bool, google *oauth.GoogleClient, frontendURL string) *Handler {
	return &Handler{service: service, refreshTTL: refreshTTL, secureCookies: secureCookies, google: google, frontendURL: strings.TrimRight(frontendURL, "/"),
		logins:        newLoginLimiter(5, 15*time.Minute, 15*time.Minute),
		resetRequests: newLoginLimiter(5, 15*time.Minute, 15*time.Minute)}
}

// oauthFail ends the Google callback with an error, either as JSON or as a
// redirect to the frontend login page carrying a short error code.
func (h *Handler) oauthFail(c *gin.Context, status int, code, msg string) {
	if h.frontendURL != "" {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?oauth_error="+url.QueryEscape(code))
		return
	}
	response.ErrorMsg(c, status, msg)
}

func (h *Handler) GoogleLogin(c *gin.Context) {
	if h.google == nil {
		response.ErrorMsg(c, http.StatusNotImplemented, "google oauth is not configured")
		return
	}
	state, err := randomState()
	if err != nil {
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, state, 300, oauthCookiePath, "", h.secureCookies, true)
	c.Redirect(http.StatusFound, h.google.AuthURL(state))
}

func (h *Handler) GoogleCallback(c *gin.Context) {
	if h.google == nil {
		response.ErrorMsg(c, http.StatusNotImplemented, "google oauth is not configured")
		return
	}

	wantState, err := c.Cookie(oauthStateCookie)
	c.SetCookie(oauthStateCookie, "", -1, oauthCookiePath, "", h.secureCookies, true)
	if err != nil || wantState == "" || c.Query("state") != wantState {
		h.oauthFail(c, http.StatusBadRequest, "state", "invalid oauth state")
		return
	}

	code := c.Query("code")
	if code == "" {
		h.oauthFail(c, http.StatusBadRequest, "cancelled", "missing code")
		return
	}

	info, err := h.google.Exchange(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, constants.ErrEmailNotVerified) {
			h.oauthFail(c, http.StatusUnauthorized, "email_not_verified", "google email is not verified")
			return
		}
		h.oauthFail(c, http.StatusUnauthorized, "exchange", "google oauth exchange failed")
		return
	}

	access, refresh, _, err := h.service.LoginOrRegisterWithOAuth(
		c.Request.Context(),
		string(constants.ProviderGoogle),
		info.Sub,
		info.Email,
		info.GivenName,
		info.FamilyName,
	)
	if err != nil {
		if errors.Is(err, constants.ErrEmailDuplicate) {
			h.oauthFail(c, http.StatusConflict, "email_exists", "an account with this email already exists")
			return
		}
		h.oauthFail(c, http.StatusInternalServerError, "server", "internal server error")
		return
	}

	h.setRefreshCookie(c, refresh)
	if h.frontendURL != "" {
		c.Redirect(http.StatusFound, h.frontendURL+"/")
		return
	}
	response.OK(c, http.StatusOK, TokenResponse{AccessToken: access})
}

func randomState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	u, err := h.service.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrEmailEmpty), errors.Is(err, constants.ErrPasswordTooShort), errors.Is(err, constants.ErrPasswordTooLong):
			response.Error(c, http.StatusBadRequest, err)
		case errors.Is(err, constants.ErrEmailDuplicate):
			response.ErrorMsg(c, http.StatusConflict, "email already registered")
		default:
			response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	response.OK(c, http.StatusCreated, toUserResponse(u))
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	limitKey := c.ClientIP() + "|" + strings.ToLower(strings.TrimSpace(req.Username))
	if wait := h.logins.retryAfter(limitKey); wait > 0 {
		c.Header("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
		response.ErrorMsg(c, http.StatusTooManyRequests, "too many login attempts")
		return
	}

	access, refresh, _, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, constants.ErrInvalidCredentials) {
			h.logins.fail(limitKey)
			response.ErrorMsg(c, http.StatusUnauthorized, "invalid username or password")
			return
		}
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logins.succeed(limitKey)
	h.setRefreshCookie(c, refresh)
	response.OK(c, http.StatusOK, TokenResponse{AccessToken: access})
}

func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshCookieName)
	if err != nil || refreshToken == "" {
		response.ErrorMsg(c, http.StatusUnauthorized, "missing refresh token")
		return
	}

	access, _, err := h.service.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		h.refreshFailed(c, err, "invalid or expired session")
		return
	}
	response.OK(c, http.StatusOK, TokenResponse{AccessToken: access})
}

func (h *Handler) Logout(c *gin.Context) {
	if refreshToken, err := c.Cookie(refreshCookieName); err == nil && refreshToken != "" {
		if err := h.service.Logout(c.Request.Context(), refreshToken); err != nil {
			response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
			return
		}
	}
	h.clearRefreshCookie(c)
	response.NoContent(c)
}

func (h *Handler) Me(c *gin.Context) {
	uid, ok := middleware.CurrentUserID(c)
	if !ok {
		h.meViaRefresh(c)
		return
	}

	u, err := h.service.GetByID(c.Request.Context(), uid)
	if err != nil {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	resp := toUserResponse(u)
	response.OK(c, http.StatusOK, MeResponse{User: &resp})
}

// refreshFailed answers a refresh that produced no token. Only a token that is
// really invalid clears the cookie; a database error must not log the user
// out, so it is a 500 and the cookie stays.
func (h *Handler) refreshFailed(c *gin.Context, err error, invalidMsg string) {
	if errors.Is(err, constants.ErrInvalidToken) {
		h.clearRefreshCookie(c)
		response.ErrorMsg(c, http.StatusUnauthorized, invalidMsg)
		return
	}
	response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
}

func (h *Handler) meViaRefresh(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshCookieName)
	if err != nil || refreshToken == "" {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	access, u, err := h.service.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		h.refreshFailed(c, err, "unauthorized")
		return
	}
	resp := toUserResponse(u)
	response.OK(c, http.StatusOK, MeResponse{User: &resp, AccessToken: access})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	uid, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}
	email := ""
	if req.Email != nil {
		email = *req.Email
	}

	u, err := h.service.UpdateProfile(c.Request.Context(), uid, email)
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrUserNotFound):
			response.ErrorMsg(c, http.StatusNotFound, "not found")
		case errors.Is(err, constants.ErrEmailEmpty):
			response.Error(c, http.StatusBadRequest, err)
		case errors.Is(err, constants.ErrEmailDuplicate):
			response.ErrorMsg(c, http.StatusConflict, "email already registered")
		default:
			response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	response.OK(c, http.StatusOK, toUserResponse(u))
}

func (h *Handler) ChangePassword(c *gin.Context) {
	uid, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), uid, req.CurrentPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, constants.ErrInvalidCredentials):
			response.ErrorMsg(c, http.StatusUnauthorized, "current password is incorrect")
		case errors.Is(err, constants.ErrPasswordTooShort), errors.Is(err, constants.ErrPasswordTooLong):
			response.Error(c, http.StatusBadRequest, err)
		case errors.Is(err, constants.ErrUserNotFound):
			response.ErrorMsg(c, http.StatusNotFound, "not found")
		default:
			response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	h.clearRefreshCookie(c)
	response.NoContent(c)
}

func (h *Handler) AdminResetPassword(c *gin.Context) {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid id")
		return
	}

	var req AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	if err := h.service.AdminResetPassword(c.Request.Context(), targetID, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, constants.ErrUserNotFound):
			response.ErrorMsg(c, http.StatusNotFound, "not found")
		case errors.Is(err, constants.ErrPasswordTooShort), errors.Is(err, constants.ErrPasswordTooLong):
			response.Error(c, http.StatusBadRequest, err)
		default:
			response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	response.NoContent(c)
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	limitKey := c.ClientIP() + "|" + normalizeEmail(req.Email)
	if wait := h.resetRequests.retryAfter(limitKey); wait > 0 {
		c.Header("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
		response.ErrorMsg(c, http.StatusTooManyRequests, "too many password reset requests")
		return
	}
	h.resetRequests.fail(limitKey)

	if err := h.service.RequestPasswordReset(c.Request.Context(), req.Email); err != nil {
		if errors.Is(err, constants.ErrPasswordResetUnavailable) {
			response.ErrorMsg(c, http.StatusServiceUnavailable, "password reset is not available")
			return
		}
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.NoContent(c)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, constants.ErrResetTokenInvalid):
			response.ErrorMsg(c, http.StatusBadRequest, "invalid or expired reset link")
		case errors.Is(err, constants.ErrPasswordTooShort), errors.Is(err, constants.ErrPasswordTooLong):
			response.Error(c, http.StatusBadRequest, err)
		default:
			response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	h.clearRefreshCookie(c)
	response.NoContent(c)
}

func (h *Handler) Delete(c *gin.Context) {
	uid, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.service.Delete(c.Request.Context(), uid); err != nil {
		if errors.Is(err, constants.ErrUserNotFound) {
			response.ErrorMsg(c, http.StatusNotFound, "not found")
			return
		}
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		return
	}
	h.clearRefreshCookie(c)
	response.NoContent(c)
}

func (h *Handler) LinkIdentity(c *gin.Context) {
	uid, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req LinkIdentityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	i, err := h.service.LinkIdentity(c.Request.Context(), uid, req.Provider, req.Identifier)
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrIdentityInvalid):
			response.Error(c, http.StatusBadRequest, err)
		case errors.Is(err, constants.ErrIdentityDuplicate):
			response.ErrorMsg(c, http.StatusConflict, "identity already linked")
		default:
			response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	response.OK(c, http.StatusCreated, i)
}

func (h *Handler) ListIdentities(c *gin.Context) {
	uid, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	ids, err := h.service.ListIdentities(c.Request.Context(), uid)
	if err != nil {
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, ids)
}

func (h *Handler) UnlinkIdentity(c *gin.Context) {
	uid, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.service.UnlinkIdentity(c.Request.Context(), uid, id); err != nil {
		switch {
		case errors.Is(err, constants.ErrIdentityNotFound):
			response.ErrorMsg(c, http.StatusNotFound, "not found")
		case errors.Is(err, constants.ErrCannotUnlinkLastIdentity):
			response.ErrorMsg(c, http.StatusConflict, "cannot unlink your only sign-in method")
		default:
			response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	response.NoContent(c)
}

func (h *Handler) setRefreshCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookieName, token, int(h.refreshTTL.Seconds()), authCookiePath, "", h.secureCookies, true)
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookieName, "", -1, authCookiePath, "", h.secureCookies, true)
}
