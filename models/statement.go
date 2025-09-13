package models

// Statement represents a statement in the system
type Statement struct {
	ID         string `json:"id" db:"id"`
	Title      string `json:"title" db:"title"`
	UserID     string `json:"userId" db:"user_id"`
	CategoryID string `json:"categoryId" db:"category_id"`
	CreatedAt  string `json:"createdAt" db:"created_at"`
	UpdatedAt  string `json:"updatedAt" db:"updated_at"`
}