package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type Config struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type Claims struct {
	UID   string `json:"uid"`
	Email string `json:"email,omitempty"`
	Type  string `json:"type"` // "access" or "refresh"
	JTI   string `json:"jti,omitempty"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type Manager struct {
	secret               []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewManager(cfg Config) *Manager {
	return &Manager{
		secret:               []byte(cfg.Secret),
		accessTokenDuration:  cfg.AccessTokenDuration,
		refreshTokenDuration: cfg.RefreshTokenDuration,
	}
}

func (m *Manager) GenerateTokenPair(uid, email string) (TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(m.accessTokenDuration)

	accessClaims := Claims{
		UID:   uid,
		Email: email,
		Type:  "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   uid,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(m.secret)
	if err != nil {
		return TokenPair{}, err
	}

	jti, err := generateJTI()
	if err != nil {
		return TokenPair{}, err
	}

	refreshClaims := Claims{
		UID:   uid,
		Email: email,
		Type:  "refresh",
		JTI:   jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   uid,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(m.secret)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry,
	}, nil
}

func (m *Manager) GenerateAccessToken(uid, email string) (string, time.Time, error) {
	now := time.Now()
	expiry := now.Add(m.accessTokenDuration)

	claims := Claims{
		UID:   uid,
		Email: email,
		Type:  "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   uid,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiry, nil
}

func (m *Manager) ValidateToken(tokenString string, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Type != expectedType {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.ValidateToken(tokenString, "access")
}

func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.ValidateToken(tokenString, "refresh")
}

func (m *Manager) RefreshTokenDuration() time.Duration {
	return m.refreshTokenDuration
}

func generateJTI() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
