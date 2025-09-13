/package models

// Event represents an event in the system
type Event struct {
	ID        string `json:"id" db:"id"`
	UserID    string `json:"userId" db:"user_id"`
	Event     string `json:"event" db:"event"`
	Data      string `json:"data" db:"data"`
	CreatedAt string `json:"createdAt" db:"created_at"`
}
