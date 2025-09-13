package repositories

import (
	"time"

	"linka.type-backend/models"
)

// OTPCRUD provides CRUD operations for OTP entity
// This is a wrapper around OTPRepository for backward compatibility
type OTPCRUD struct {
	repo *OTPRepository
}

// NewOTPCRUD creates a new OTPCRUD
func NewOTPCRUD() *OTPCRUD {
	return &OTPCRUD{
		repo: NewOTPRepository(),
	}
}

// CreateOTP creates a new OTP record
func (o *OTPCRUD) CreateOTP(otp *models.OTPCode) error {
	return o.repo.CreateOTP(otp)
}

// GetOTPByCode retrieves an OTP by code, email and type
func (o *OTPCRUD) GetOTPByCode(code, email, otpType string) (*models.OTPCode, error) {
	return o.repo.GetOTPByCode(code, email, otpType)
}

// GetOTPByCodeAnyStatus retrieves an OTP by code, email and type regardless of used status
func (o *OTPCRUD) GetOTPByCodeAnyStatus(code, email, otpType string) (*models.OTPCode, error) {
	return o.repo.GetOTPByCodeAnyStatus(code, email, otpType)
}

// MarkOTPAsUsed marks an OTP as used
func (o *OTPCRUD) MarkOTPAsUsed(id string) error {
	return o.repo.MarkOTPAsUsed(id)
}

// DeleteOTPByEmail deletes all OTP codes for an email
func (o *OTPCRUD) DeleteOTPByEmail(email string) error {
	return o.repo.DeleteOTPByEmail(email)
}

// DeleteExpiredOTPs deletes expired OTP codes
func (o *OTPCRUD) DeleteExpiredOTPs() error {
	return o.repo.DeleteExpiredOTPs()
}

// GetOTPByID retrieves an OTP by ID
func (o *OTPCRUD) GetOTPByID(id string) (*models.OTPCode, error) {
	return o.repo.GetOTPByID(id)
}

// GetOTPByEmailAndType получает активный OTP код по email и типу
func (o *OTPCRUD) GetOTPByEmailAndType(email, otpType string) (*models.OTPCode, error) {
	return o.repo.GetOTPByEmailAndType(email, otpType)
}

// UpdateOTPExpiration обновляет время истечения OTP кода
func (o *OTPCRUD) UpdateOTPExpiration(id string, newExpiration time.Time) error {
	return o.repo.UpdateOTPExpiration(id, newExpiration)
}
