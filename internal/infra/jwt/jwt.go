package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"vocabulary-service/internal/constants"
)

type Claims struct {
	UserID uint64 `json:"userId"`
	Role   string `json:"role"`
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
	return s.issue(userID, role, s.cfg.AccessSecret, s.cfg.AccessTTL)
}

func (s *Signer) IssueRefreshToken(userID uint64, role string) (string, error) {
	return s.issue(userID, role, s.cfg.RefreshSecret, s.cfg.RefreshTTL)
}

func (s *Signer) issue(userID uint64, role string, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (s *Signer) ParseAccessToken(tokenStr string) (*Claims, error) {
	return s.parse(tokenStr, s.cfg.AccessSecret)
}

func (s *Signer) ParseRefreshToken(tokenStr string) (*Claims, error) {
	return s.parse(tokenStr, s.cfg.RefreshSecret)
}

func (s *Signer) parse(tokenStr string, secret []byte) (*Claims, error) {
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
	if !token.Valid {
		return nil, constants.ErrInvalidToken
	}
	return claims, nil
}
