package models

// User represents a user in the system
type User struct {
	ID            string `json:"id" db:"id"`
	Email         string `json:"email" db:"email"`
	Password      string `json:"-" db:"password"`
	EmailVerified bool   `json:"emailVerified" db:"email_verified"`
	CreatedAt     string `json:"createdAt" db:"created_at"`
	UpdatedAt     string `json:"updatedAt" db:"updated_at"`
}