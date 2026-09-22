package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/infra/http/response"
	"vocabulary-service/internal/infra/oauth"
	"vocabulary-service/internal/middleware"
)

const refreshCookieName = "refresh_token"
const oauthStateCookie = "oauth_state"

type Handler struct {
	service       Service
	refreshTTL    time.Duration
	secureCookies bool
	google        *oauth.GoogleClient
}

func NewHandler(service Service, refreshTTL time.Duration, secureCookies bool, google *oauth.GoogleClient) *Handler {
	return &Handler{service: service, refreshTTL: refreshTTL, secureCookies: secureCookies, google: google}
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
	c.SetCookie(oauthStateCookie, state, 300, "/api/v1/auth/google", "", h.secureCookies, true)
	c.Redirect(http.StatusFound, h.google.AuthURL(state))
}

func (h *Handler) GoogleCallback(c *gin.Context) {
	if h.google == nil {
		response.ErrorMsg(c, http.StatusNotImplemented, "google oauth is not configured")
		return
	}

	wantState, err := c.Cookie(oauthStateCookie)
	c.SetCookie(oauthStateCookie, "", -1, "/api/v1/auth/google", "", h.secureCookies, true)
	if err != nil || wantState == "" || c.Query("state") != wantState {
		response.ErrorMsg(c, http.StatusBadRequest, "invalid oauth state")
		return
	}

	code := c.Query("code")
	if code == "" {
		response.ErrorMsg(c, http.StatusBadRequest, "missing code")
		return
	}

	info, err := h.google.Exchange(c.Request.Context(), code)
	if err != nil {
		response.ErrorMsg(c, http.StatusUnauthorized, "google oauth exchange failed")
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
			response.ErrorMsg(c, http.StatusConflict, "an account with this email already exists")
			return
		}
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		return
	}

	h.setRefreshCookie(c, refresh)
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
		case errors.Is(err, constants.ErrEmailEmpty), errors.Is(err, constants.ErrPasswordTooShort):
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

	access, refresh, _, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, constants.ErrInvalidCredentials) {
			response.ErrorMsg(c, http.StatusUnauthorized, "invalid username or password")
			return
		}
		response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		return
	}

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
		h.clearRefreshCookie(c)
		response.ErrorMsg(c, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	response.OK(c, http.StatusOK, TokenResponse{AccessToken: access})
}

func (h *Handler) Logout(c *gin.Context) {
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

func (h *Handler) meViaRefresh(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshCookieName)
	if err != nil || refreshToken == "" {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	access, u, err := h.service.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		h.clearRefreshCookie(c)
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
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
		case errors.Is(err, constants.ErrPasswordTooShort):
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
		case errors.Is(err, constants.ErrPasswordTooShort):
			response.Error(c, http.StatusBadRequest, err)
		default:
			response.ErrorMsg(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	response.NoContent(c)
}

func (h *Handler) Delete(c *gin.Context) {
	uid, ok := middleware.CurrentUserID(c)
	if !ok {
		response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.service.Delete(c.Request.Context(), uid); err != nil {
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
	c.SetCookie(refreshCookieName, token, int(h.refreshTTL.Seconds()), "/api/v1/auth", "", h.secureCookies, true)
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookieName, "", -1, "/api/v1/auth", "", h.secureCookies, true)
}
