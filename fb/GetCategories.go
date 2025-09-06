package fb

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/auth"
)

type Category struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	UserID string `json:"userId"`
}

func GetCategories(user *auth.UserRecord) ([]*Category, error) {
	if !IsFirebaseInitialized() {
		return nil, errors.New("Firebase is not initialized")
	}
	
	db, err := fb.Database(context.Background())
	if err != nil {
		return nil, err
	}

	ref := db.NewRef("users").Child(user.UID).Child("Category")
	docs, err := ref.OrderByKey().GetOrdered(context.Background())
	if err != nil {
		return nil, err
	}
	categories := []*Category{}
	for _, doc := range docs {
		row := Category{}
		err = doc.Unmarshal(&row)
		if err != nil {
			return nil, err
		}
		row.UserID = user.UID
		categories = append(categories, &row)
	}
	return categories, nil
}
