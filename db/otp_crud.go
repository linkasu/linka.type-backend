package db

import (
	"database/sql"
	"fmt"
	"time"
)

// OTPCRUD предоставляет методы для работы с OTP кодами
type OTPCRUD struct{}

// CreateOTP создает новый OTP код
func (otp *OTPCRUD) CreateOTP(otpCode *OTPCode) error {
	query := `
		INSERT INTO otp_codes (id, email, code, type, expires_at, used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := DB.Exec(query,
		otpCode.ID,
		otpCode.Email,
		otpCode.Code,
		otpCode.Type,
		otpCode.ExpiresAt,
		otpCode.Used,
		otpCode.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create OTP code: %v", err)
	}

	return nil
}

// GetOTPByEmailAndType получает активный OTP код по email и типу
func (otp *OTPCRUD) GetOTPByEmailAndType(email, otpType string) (*OTPCode, error) {
	query := `
		SELECT id, email, code, type, expires_at, used, created_at
		FROM otp_codes
		WHERE email = $1 AND type = $2 AND used = false AND expires_at > $3
		ORDER BY created_at DESC
		LIMIT 1
	`

	var otpCode OTPCode
	err := DB.QueryRow(query, email, otpType, time.Now().Format(time.RFC3339)).Scan(
		&otpCode.ID,
		&otpCode.Email,
		&otpCode.Code,
		&otpCode.Type,
		&otpCode.ExpiresAt,
		&otpCode.Used,
		&otpCode.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get OTP code: %v", err)
	}

	return &otpCode, nil
}

// MarkOTPAsUsed помечает OTP код как использованный
func (otp *OTPCRUD) MarkOTPAsUsed(id string) error {
	query := `UPDATE otp_codes SET used = true WHERE id = $1`

	_, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to mark OTP as used: %v", err)
	}

	return nil
}

// DeleteExpiredOTP удаляет истекшие OTP коды
func (otp *OTPCRUD) DeleteExpiredOTP() error {
	query := `DELETE FROM otp_codes WHERE expires_at < $1`

	_, err := DB.Exec(query, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to delete expired OTP codes: %v", err)
	}

	return nil
}

// DeleteOTPByEmail удаляет все OTP коды для указанного email
func (otp *OTPCRUD) DeleteOTPByEmail(email string) error {
	query := `DELETE FROM otp_codes WHERE email = $1`

	_, err := DB.Exec(query, email)
	if err != nil {
		return fmt.Errorf("failed to delete OTP codes for email: %v", err)
	}

	return nil
}

// GetOTPByCode получает OTP код по самому коду
func (otp *OTPCRUD) GetOTPByCode(code, email, otpType string) (*OTPCode, error) {
	query := `
		SELECT id, email, code, type, expires_at, used, created_at
		FROM otp_codes
		WHERE code = $1 AND email = $2 AND type = $3 AND used = false AND expires_at > $4
		ORDER BY created_at DESC
		LIMIT 1
	`

	var otpCode OTPCode
	err := DB.QueryRow(query, code, email, otpType, time.Now().Format(time.RFC3339)).Scan(
		&otpCode.ID,
		&otpCode.Email,
		&otpCode.Code,
		&otpCode.Type,
		&otpCode.ExpiresAt,
		&otpCode.Used,
		&otpCode.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get OTP by code: %v", err)
	}

	return &otpCode, nil
}

// UpdateOTPExpiration обновляет время истечения OTP кода
func (otp *OTPCRUD) UpdateOTPExpiration(id, expiresAt string) error {
	query := `UPDATE otp_codes SET expires_at = $1 WHERE id = $2`

	_, err := DB.Exec(query, expiresAt, id)
	if err != nil {
		return fmt.Errorf("failed to update OTP expiration: %v", err)
	}

	return nil
}
