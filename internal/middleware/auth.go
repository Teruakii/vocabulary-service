package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"vocabulary-service/internal/constants"
	"vocabulary-service/internal/infra/http/response"
	"vocabulary-service/internal/infra/jwt"
)

const (
	ContextKeyUserID = "authUserID"
	ContextKeyRole   = "authRole"
)

type TokenVerifier interface {
	ParseAccessToken(token string) (*jwt.Claims, error)
}

var errMissingBearerToken = errors.New("missing bearer token")

func RequireAuth(verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := extractClaims(c, verifier)
		if err != nil {
			response.ErrorMsg(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

func OptionalAuth(verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		if claims, err := extractClaims(c, verifier); err == nil {
			c.Set(ContextKeyUserID, claims.UserID)
			c.Set(ContextKeyRole, claims.Role)
		}
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ContextKeyRole)

		if role != string(constants.RoleAdmin) {
			response.ErrorMsg(c, http.StatusForbidden, "admin only")
			c.Abort()
			return
		}
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(ContextKeyUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}

func extractClaims(c *gin.Context, verifier TokenVerifier) (*jwt.Claims, error) {
	header := c.GetHeader("Authorization")
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || token == "" {
		return nil, errMissingBearerToken
	}
	return verifier.ParseAccessToken(token)
}
