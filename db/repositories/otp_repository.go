package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"linka.type-backend/db"
	"linka.type-backend/models"
)

// OTPRepository provides CRUD operations for OTP entity
type OTPRepository struct{}

// NewOTPRepository creates a new OTPRepository
func NewOTPRepository() *OTPRepository {
	return &OTPRepository{}
}

// CreateOTP creates a new OTP record
func (o *OTPRepository) CreateOTP(otp *models.OTPCode) error {
	query := `
		INSERT INTO otp_codes (id, email, code, type, expires_at, used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := db.DB.Exec(query, otp.ID, otp.Email, otp.Code, otp.Type, otp.ExpiresAt, otp.Used, otp.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating OTP: %v", err)
	}

	return nil
}

// GetOTPByCode retrieves an OTP by code, email and type
func (o *OTPRepository) GetOTPByCode(code, email, otpType string) (*models.OTPCode, error) {
	query := `SELECT id, email, code, type, expires_at, used, created_at FROM otp_codes WHERE code = $1 AND email = $2 AND type = $3 AND used = false`

	var otp models.OTPCode
	err := db.DB.QueryRow(query, code, email, otpType).Scan(&otp.ID, &otp.Email, &otp.Code, &otp.Type, &otp.ExpiresAt, &otp.Used, &otp.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("OTP not found")
		}
		return nil, fmt.Errorf("error getting OTP: %v", err)
	}

	return &otp, nil
}

// GetOTPByCodeAnyStatus retrieves an OTP by code, email and type regardless of used status
func (o *OTPRepository) GetOTPByCodeAnyStatus(code, email, otpType string) (*models.OTPCode, error) {
	query := `SELECT id, email, code, type, expires_at, used, created_at FROM otp_codes WHERE code = $1 AND email = $2 AND type = $3`

	var otp models.OTPCode
	err := db.DB.QueryRow(query, code, email, otpType).Scan(&otp.ID, &otp.Email, &otp.Code, &otp.Type, &otp.ExpiresAt, &otp.Used, &otp.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("OTP not found")
		}
		return nil, fmt.Errorf("error getting OTP: %v", err)
	}

	return &otp, nil
}

// MarkOTPAsUsed marks an OTP as used
func (o *OTPRepository) MarkOTPAsUsed(id string) error {
	query := `UPDATE otp_codes SET used = true WHERE id = $1`

	result, err := db.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error marking OTP as used: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("OTP not found")
	}

	return nil
}

// DeleteOTPByEmail deletes all OTP codes for an email
func (o *OTPRepository) DeleteOTPByEmail(email string) error {
	query := `DELETE FROM otp_codes WHERE email = $1`

	result, err := db.DB.Exec(query, email)
	if err != nil {
		return fmt.Errorf("error deleting OTP codes: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no OTP codes found for email")
	}

	return nil
}

// DeleteExpiredOTPs deletes expired OTP codes
func (o *OTPRepository) DeleteExpiredOTPs() error {
	query := `DELETE FROM otp_codes WHERE expires_at < $1`

	now := time.Now().Format(time.RFC3339)
	result, err := db.DB.Exec(query, now)
	if err != nil {
		return fmt.Errorf("error deleting expired OTP codes: %v", err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	return nil
}

// GetOTPByID retrieves an OTP by ID
func (o *OTPRepository) GetOTPByID(id string) (*models.OTPCode, error) {
	query := `SELECT id, email, code, type, expires_at, used, created_at FROM otp_codes WHERE id = $1`

	var otp models.OTPCode
	err := db.DB.QueryRow(query, id).Scan(&otp.ID, &otp.Email, &otp.Code, &otp.Type, &otp.ExpiresAt, &otp.Used, &otp.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("OTP not found")
		}
		return nil, fmt.Errorf("error getting OTP: %v", err)
	}

	return &otp, nil
}

// GetOTPByEmailAndType получает активный OTP код по email и типу
func (o *OTPRepository) GetOTPByEmailAndType(email, otpType string) (*models.OTPCode, error) {
	query := `SELECT id, email, code, type, expires_at, used, created_at FROM otp_codes 
			  WHERE email = $1 AND type = $2 AND expires_at > NOW() 
			  ORDER BY created_at DESC LIMIT 1`

	var otp models.OTPCode
	err := db.DB.QueryRow(query, email, otpType).Scan(&otp.ID, &otp.Email, &otp.Code, &otp.Type, &otp.ExpiresAt, &otp.Used, &otp.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("OTP not found")
		}
		return nil, fmt.Errorf("error getting OTP: %v", err)
	}

	return &otp, nil
}

// UpdateOTPExpiration обновляет время истечения OTP кода
func (o *OTPRepository) UpdateOTPExpiration(id string, newExpiration time.Time) error {
	query := `UPDATE otp_codes SET expires_at = $1 WHERE id = $2`

	result, err := db.DB.Exec(query, newExpiration, id)
	if err != nil {
		return fmt.Errorf("error updating OTP expiration: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("OTP not found")
	}

	return nil
}
