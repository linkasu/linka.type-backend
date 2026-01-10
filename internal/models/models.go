package models

import "encoding/json"

// Category matches client-visible fields.
type Category struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Created   int64  `json:"created"`
	Default   *bool  `json:"default,omitempty"`
	AIUse     bool   `json:"aiUse,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

// Statement matches client-visible fields.
type Statement struct {
	ID         string `json:"id"`
	CategoryID string `json:"categoryId"`
	Text       string `json:"text"`
	Created    int64  `json:"created"`
	UpdatedAt  int64  `json:"updated_at,omitempty"`
}

// DialogChat represents a chat session for dialog helper.
type DialogChat struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Created       int64  `json:"created"`
	UpdatedAt     int64  `json:"updated_at,omitempty"`
	LastMessageAt int64  `json:"last_message_at,omitempty"`
	MessageCount  int64  `json:"message_count,omitempty"`
}

// DialogMessage stores dialog history for a chat.
type DialogMessage struct {
	ID        string `json:"id"`
	ChatID    string `json:"chatId"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Source    string `json:"source,omitempty"`
	Created   int64  `json:"created"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

// DialogSuggestion is a suggested phrase for the user.
type DialogSuggestion struct {
	ID         string  `json:"id"`
	ChatID     string  `json:"chatId,omitempty"`
	MessageID  string  `json:"messageId,omitempty"`
	Text       string  `json:"text"`
	Status     string  `json:"status"`
	CategoryID *string `json:"categoryId,omitempty"`
	Created    int64   `json:"created"`
	UpdatedAt  int64   `json:"updated_at,omitempty"`
}

// DialogSuggestionJob is a background job for generating suggestions.
type DialogSuggestionJob struct {
	ID        string  `json:"id"`
	UserID    string  `json:"userId"`
	ChatID    string  `json:"chatId"`
	MessageID string  `json:"messageId"`
	Status    string  `json:"status"`
	Attempts  int     `json:"attempts"`
	LastError *string `json:"lastError,omitempty"`
	Created   int64   `json:"created"`
	UpdatedAt int64   `json:"updated_at,omitempty"`
}

// UserState combines onboarding and quick phrases.
type UserState struct {
	Inited      bool              `json:"inited"`
	Quickes     []string          `json:"quickes"`
	Preferences map[string]any    `json:"preferences,omitempty"`
}

// GlobalCategory mirrors global category definitions.
type GlobalCategory struct {
	ID         string      `json:"id"`
	Label      string      `json:"label"`
	Created    int64       `json:"created"`
	Default    *bool       `json:"default,omitempty"`
	UpdatedAt  int64       `json:"updated_at,omitempty"`
	Statements []Statement `json:"statements,omitempty"`
}

// FactoryQuestion defines onboarding templates.
type FactoryQuestion struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Phrases    []string `json:"phrases"`
	Category   string   `json:"category"`
	Type       string   `json:"type"`
	OrderIndex int      `json:"order_index"`
}

// ChangeEvent is emitted to realtime consumers.
type ChangeEvent struct {
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Op         string          `json:"op"`
	Payload    json.RawMessage `json:"payload"`
	UpdatedAt  int64           `json:"updated_at"`
	Cursor     string          `json:"cursor,omitempty"`
}
