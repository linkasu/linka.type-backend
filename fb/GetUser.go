package fb

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

func GetUser(email string) (*auth.UserRecord, error) {
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
