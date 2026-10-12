package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"vocabulary-service/internal/constants"
)

// Token types. Access and refresh tokens may share a secret (they do when
// both come from JWT_SECRET_KEY), so the type claim is what stops one being
// accepted as the other.
const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

type Claims struct {
	UserID uint64 `json:"userId"`
	Role   string `json:"role"`
	Type   string `json:"typ"`
	jwt.RegisteredClaims
}

type Config struct {
	AccessSecret  []byte
	RefreshSecret []byte
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	Issuer        string
}

type Signer struct {
	cfg Config
}

func NewSigner(cfg Config) *Signer {
	return &Signer{cfg: cfg}
}

func (s *Signer) IssueAccessToken(userID uint64, role string) (string, error) {
	token, _, err := s.issue(userID, role, TypeAccess, "", s.cfg.AccessSecret, s.cfg.AccessTTL)
	return token, err
}

// IssueRefreshToken returns the token and its claims; claims.ID is a random
// session id the caller stores so the token can be revoked later.
func (s *Signer) IssueRefreshToken(userID uint64, role string) (string, *Claims, error) {
	id, err := randomID()
	if err != nil {
		return "", nil, err
	}
	return s.issue(userID, role, TypeRefresh, id, s.cfg.RefreshSecret, s.cfg.RefreshTTL)
}

func (s *Signer) issue(userID uint64, role, typ, id string, secret []byte, ttl time.Duration) (string, *Claims, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Role:   role,
		Type:   typ,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", nil, err
	}
	return token, claims, nil
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Signer) ParseAccessToken(tokenStr string) (*Claims, error) {
	return s.parse(tokenStr, s.cfg.AccessSecret, TypeAccess)
}

func (s *Signer) ParseRefreshToken(tokenStr string) (*Claims, error) {
	return s.parse(tokenStr, s.cfg.RefreshSecret, TypeRefresh)
}

func (s *Signer) parse(tokenStr string, secret []byte, wantType string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, constants.ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, constants.ErrTokenExpired
		}
		return nil, constants.ErrInvalidToken
	}
	if !token.Valid || claims.Type != wantType {
		return nil, constants.ErrInvalidToken
	}
	return claims, nil
}
