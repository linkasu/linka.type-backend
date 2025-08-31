package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var jwtSecret = []byte("your-secret-key-change-in-production")

func getJWTSecret() ([]byte, error) {
	if len(jwtSecret) != 0 && string(jwtSecret) != "your-secret-key-change-in-production" {
		return jwtSecret, nil
	}
	env := os.Getenv("JWT_SECRET")
	if env == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}
	return []byte(env), nil
}

// Claims структура для JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// RefreshClaims структура для refresh token claims
type RefreshClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair структура для пары токенов
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GenerateToken генерирует JWT токен для пользователя
func GenerateToken(userID, email string) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", fmt.Errorf("jwt secret error: %w", err)
	}

	issuer := os.Getenv("JWT_ISSUER")
	audience := os.Getenv("JWT_AUDIENCE")

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 часа
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken генерирует refresh token для пользователя
func GenerateRefreshToken(userID, email string) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", fmt.Errorf("jwt secret error: %w", err)
	}

	issuer := os.Getenv("JWT_ISSUER")
	audience := os.Getenv("JWT_AUDIENCE")

	claims := &RefreshClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(365 * 24 * time.Hour)), // 1 год
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("error signing refresh token: %v", err)
	}

	return tokenString, nil
}

// GenerateTokenPair генерирует пару токенов (access + refresh)
func GenerateTokenPair(userID, email string) (*TokenPair, error) {
	accessToken, err := GenerateToken(userID, email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken(userID, email)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateToken валидирует JWT токен и возвращает claims
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		secret, sErr := getJWTSecret()
		if sErr != nil {
			return nil, sErr
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		expectedIssuer := os.Getenv("JWT_ISSUER")
		if expectedIssuer != "" && !claims.VerifyIssuer(expectedIssuer, true) {
			return nil, errors.New("invalid token issuer")
		}
		expectedAudience := os.Getenv("JWT_AUDIENCE")
		if expectedAudience != "" && !claims.VerifyAudience(expectedAudience, true) {
			return nil, errors.New("invalid token audience")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken валидирует refresh token и возвращает claims
func ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		secret, sErr := getJWTSecret()
		if sErr != nil {
			return nil, sErr
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing refresh token: %v", err)
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		expectedIssuer := os.Getenv("JWT_ISSUER")
		if expectedIssuer != "" && !claims.VerifyIssuer(expectedIssuer, true) {
			return nil, errors.New("invalid refresh token issuer")
		}
		expectedAudience := os.Getenv("JWT_AUDIENCE")
		if expectedAudience != "" && !claims.VerifyAudience(expectedAudience, true) {
			return nil, errors.New("invalid refresh token audience")
		}
		return claims, nil
	}

	return nil, errors.New("invalid refresh token")
}

// SetJWTSecret устанавливает секретный ключ для JWT (для конфигурации)
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}
