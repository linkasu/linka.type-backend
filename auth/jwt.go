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
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
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

// SetJWTSecret устанавливает секретный ключ для JWT (для конфигурации)
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}
