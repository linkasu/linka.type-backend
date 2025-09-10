package db

type Statement struct {
	ID         string `json:"id"`
	Title      string `json:"text"`
	UserID     string `json:"userId"`
	CategoryID string `json:"categoryId"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type Category struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	UserID    string `json:"userId"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Password      string `json:"password"` // md5 hash
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// OTPCode представляет OTP код в базе данных
type OTPCode struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Code      string `json:"code"`
	Type      string `json:"type"` // "registration" или "reset_password"
	ExpiresAt string `json:"expires_at"`
	Used      bool   `json:"used"`
	CreatedAt string `json:"created_at"`
}
