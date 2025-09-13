package models

// OTPCode represents an OTP code in the system
type OTPCode struct {
	ID        string `json:"id" db:"id"`
	Email     string `json:"email" db:"email"`
	Code      string `json:"code" db:"code"`
	Type      string `json:"type" db:"type"`
	ExpiresAt string `json:"expiresAt" db:"expires_at"`
	Used      bool   `json:"used" db:"used"`
	CreatedAt string `json:"createdAt" db:"created_at"`
}