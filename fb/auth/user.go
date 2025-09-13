package auth

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/auth"
)

// GetUser получает пользователя из Firebase по email
func GetUser(email string) (*auth.UserRecord, error) {
	if !IsFirebaseInitialized() {
		return nil, errors.New("Firebase is not initialized")
	}
	
	auth, err := GetFirebaseAuth()
	if err != nil {
		return nil, err
	}

	user, err := auth.GetUserByEmail(context.Background(), email)
	if err != nil {
		return nil, err
	}

	return user, nil
}