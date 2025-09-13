package data

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/auth"
	fbauth "linka.type-backend/fb/auth"
)

// Category представляет категорию в Firebase
type Category struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	UserID string `json:"userId"`
}

// GetCategories получает категории пользователя из Firebase
func GetCategories(user *auth.UserRecord) ([]*Category, error) {
	if !fbauth.IsFirebaseInitialized() {
		return nil, errors.New("Firebase is not initialized")
	}
	
	fbApp := fbauth.GetFirebaseApp()
	db, err := fbApp.Database(context.Background())
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