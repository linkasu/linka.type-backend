package db

type Statement struct {
	ID         string `json:"id"`
	Title      string `json:"text"`
	UserId     string `json:"userId"`
	CategoryId string `json:"categoryId"`
}

type Category struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	UserId string `json:"userId"`
}

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Password      string `json:"password"` //md5 hash
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
