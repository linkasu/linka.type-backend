package fb

import (
	"context"
	"time"
)

type FBStatement struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created"`
	Text       string    `json:"text"`
	UserId     string    `json:"userId"`
	CategoryId string    `json:"categoryId"`
}

type fbStatementRaw struct {
	FBStatement
	CreatedAt int64 `json:"created"`
}

func (c *FBCategory) GetStatements() ([]*FBStatement, error) {
	db, err := fb.Database(context.Background())
	if err != nil {
		return nil, err
	}
	ref := db.NewRef("users").Child(c.UserId).Child("Category").Child(c.ID).Child("statements")
	docs, err := ref.OrderByKey().GetOrdered(context.Background())
	if err != nil {
		return nil, err
	}
	statements := []*FBStatement{}

	for _, doc := range docs {
		row := fbStatementRaw{}
		err = doc.Unmarshal(&row)
		if err != nil {
			return nil, err
		}
		statements = append(statements, &FBStatement{
			ID:         row.ID,
			CreatedAt:  time.Unix(row.CreatedAt/1000, 0),
			Text:       row.Text,
			UserId:     row.UserId,
			CategoryId: row.CategoryId,
		})
	}
	return statements, nil
}
