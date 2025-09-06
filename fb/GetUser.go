package fb

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/auth"
)

func GetUser(email string) (*auth.UserRecord, error) {
	if !IsFirebaseInitialized() {
		return nil, errors.New("Firebase is not initialized")
	}
	
	auth, err := fb.Auth(context.Background())
	if err != nil {
		return nil, err
	}

	user, err := auth.GetUserByEmail(context.Background(), email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
