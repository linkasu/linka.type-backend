package models

import "encoding/json"

// Category matches client-visible fields.
type Category struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Created   int64  `json:"created"`
	Default   *bool  `json:"default,omitempty"`
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
