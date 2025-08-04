package fb

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

type FBCategory struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	UserId string `json:"userId"`
}

func GetCategories(user *auth.UserRecord) ([]*FBCategory, error) {
	db, err := fb.Database(context.Background())
	if err != nil {
		return nil, err
	}

	ref := db.NewRef("users").Child(user.UID).Child("Category")
	docs, err := ref.OrderByKey().GetOrdered(context.Background())
	if err != nil {
		return nil, err
	}
	categories := []*FBCategory{}
	for _, doc := range docs {
		row := FBCategory{}
		err = doc.Unmarshal(&row)
		if err != nil {
			return nil, err
		}
		row.UserId = user.UID
		categories = append(categories, &row)
	}
	return categories, nil
}
