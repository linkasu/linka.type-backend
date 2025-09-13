package data

import (
	"context"
	"errors"
	"time"

	fbauth "linka.type-backend/fb/auth"
)

// Statement представляет statement в Firebase
type Statement struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created"`
	Text       string    `json:"text"`
	UserID     string    `json:"userId"`
	CategoryID string    `json:"categoryId"`
}

// fbStatementRaw представляет raw statement из Firebase
type fbStatementRaw struct {
	Statement
	CreatedAt int64 `json:"created"`
}

// GetStatements получает statements для категории из Firebase
func (c *Category) GetStatements() ([]*Statement, error) {
	if !fbauth.IsFirebaseInitialized() {
		return nil, errors.New("Firebase is not initialized")
	}
	
	fbApp := fbauth.GetFirebaseApp()
	db, err := fbApp.Database(context.Background())
	if err != nil {
		return nil, err
	}
	ref := db.NewRef("users").Child(c.UserID).Child("Category").Child(c.ID).Child("statements")
	docs, err := ref.OrderByKey().GetOrdered(context.Background())
	if err != nil {
		return nil, err
	}
	statements := []*Statement{}

	for _, doc := range docs {
		row := fbStatementRaw{}
		err = doc.Unmarshal(&row)
		if err != nil {
			return nil, err
		}
		statements = append(statements, &Statement{
			ID:         row.ID,
			CreatedAt:  time.Unix(row.CreatedAt/1000, 0),
			Text:       row.Text,
			UserID:     row.UserID,
			CategoryID: row.CategoryID,
		})
	}
	return statements, nil
}