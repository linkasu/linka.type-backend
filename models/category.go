package models

// Category represents a category in the system
type Category struct {
	ID        string `json:"id" db:"id"`
	Title     string `json:"title" db:"title"`
	UserID    string `json:"userId" db:"user_id"`
	CreatedAt string `json:"createdAt" db:"created_at"`
	UpdatedAt string `json:"updatedAt" db:"updated_at"`
}