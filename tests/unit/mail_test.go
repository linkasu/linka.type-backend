package unit

import (
	"os"
	"testing"

	"linka.type-backend/mail"
)

func TestMailMocking(t *testing.T) {
	// Test that mail functions work in test mode
	os.Setenv("TEST_MODE", "true")
	os.Setenv("MAIL_SERVER", "smtp.test.com")
	os.Setenv("MAIL_ADRESS", "test@example.com")
	os.Setenv("MAIL_PASSWORD", "test_password")

	// Test SendOTPEmail
	err := mail.SendOTPEmail("test@example.com", "123456", "registration")
	if err != nil {
		t.Errorf("SendOTPEmail should not return error in test mode: %v", err)
	}

	// Test SendWelcomeEmail
	err = mail.SendWelcomeEmail("test@example.com")
	if err != nil {
		t.Errorf("SendWelcomeEmail should not return error in test mode: %v", err)
	}

	// Test without test mode (should fail due to invalid SMTP config)
	os.Setenv("TEST_MODE", "false")
	err = mail.SendOTPEmail("test@example.com", "123456", "registration")
	if err == nil {
		t.Error("SendOTPEmail should return error without test mode")
	}
}
