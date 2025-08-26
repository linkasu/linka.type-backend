package mail

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

// MailConfig содержит конфигурацию SMTP
type Config struct {
	Server   string
	Port     int
	Address  string
	Password string
}

// GetMailConfig получает конфигурацию из переменных окружения
func GetMailConfig() (*Config, error) {
	server := os.Getenv("MAIL_SERVER")
	if server == "" {
		log.Printf("[SMTP ERROR] MAIL_SERVER environment variable is required")
		return nil, fmt.Errorf("MAIL_SERVER environment variable is required")
	}

	address := os.Getenv("MAIL_ADDRESS")
	if address == "" {
		log.Printf("[SMTP ERROR] MAIL_ADDRESS environment variable is required")
		return nil, fmt.Errorf("MAIL_ADDRESS environment variable is required")
	}

	password := os.Getenv("MAIL_PASSWORD")
	if password == "" {
		log.Printf("[SMTP ERROR] MAIL_PASSWORD environment variable is required")
		return nil, fmt.Errorf("MAIL_PASSWORD environment variable is required")
	}

	// Порт по умолчанию для Yandex SMTP
	port := 587
	if portStr := os.Getenv("MAIL_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	log.Printf("[SMTP INFO] Mail config loaded: Server=%s, Port=%d, Address=%s", server, port, address)
	return &Config{
		Server:   server,
		Port:     port,
		Address:  address,
		Password: password,
	}, nil
}

// SendOTPEmail отправляет email с OTP кодом
func SendOTPEmail(to, otpCode, otpType string) error {
	log.Printf("[SMTP INFO] Attempting to send OTP email to: %s, type: %s", to, otpType)

	// Check if we're in test mode
	if os.Getenv("TEST_MODE") == "true" {
		log.Printf("[SMTP TEST] Mocking OTP email to: %s, code: %s, type: %s", to, otpCode, otpType)
		return nil
	}

	config, err := GetMailConfig()
	if err != nil {
		log.Printf("[SMTP ERROR] Failed to get mail config: %v", err)
		return fmt.Errorf("failed to get mail config: %v", err)
	}

	// Определяем тему и содержимое в зависимости от типа OTP
	var subject, body string
	switch otpType {
	case "registration":
		subject = "Подтверждение регистрации - Linka Type"
		body = fmt.Sprintf(`
			<h2>Добро пожаловать в Linka Type!</h2>
			<p>Для завершения регистрации введите следующий код подтверждения:</p>
			<h1 style="color: #3498db; font-size: 32px; text-align: center; padding: 20px; background: #f8f9fa; border-radius: 8px;">%s</h1>
			<p><strong>Код действителен в течение 15 минут.</strong></p>
			<p>Если вы не регистрировались в Linka Type, проигнорируйте это письмо.</p>
		`, otpCode)
	case "reset_password":
		subject = "Сброс пароля - Linka Type"
		body = fmt.Sprintf(`
			<h2>Сброс пароля</h2>
			<p>Для сброса пароля введите следующий код подтверждения:</p>
			<h1 style="color: #e74c3c; font-size: 32px; text-align: center; padding: 20px; background: #f8f9fa; border-radius: 8px;">%s</h1>
			<p><strong>Код действителен в течение 15 минут.</strong></p>
			<p>Если вы не запрашивали сброс пароля, проигнорируйте это письмо.</p>
		`, otpCode)
	default:
		log.Printf("[SMTP ERROR] Unknown OTP type: %s", otpType)
		return fmt.Errorf("unknown OTP type: %s", otpType)
	}

	// Создаем сообщение
	m := gomail.NewMessage()
	m.SetHeader("From", config.Address)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Настраиваем SMTP
	d := gomail.NewDialer(config.Server, config.Port, config.Address, config.Password)
	log.Printf("[SMTP INFO] Attempting to connect to SMTP server: %s:%d", config.Server, config.Port)

	// Отправляем email
	if err := d.DialAndSend(m); err != nil {
		log.Printf("[SMTP ERROR] Failed to send email to %s: %v", to, err)
		log.Printf("[SMTP ERROR] SMTP Details: Server=%s, Port=%d, Address=%s", config.Server, config.Port, config.Address)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("[SMTP SUCCESS] Email sent successfully to: %s", to)
	return nil
}

// SendWelcomeEmail отправляет приветственное письмо после подтверждения email
func SendWelcomeEmail(to string) error {
	log.Printf("[SMTP INFO] Attempting to send welcome email to: %s", to)

	// Check if we're in test mode
	if os.Getenv("TEST_MODE") == "true" {
		log.Printf("[SMTP TEST] Mocking welcome email to: %s", to)
		return nil
	}

	config, err := GetMailConfig()
	if err != nil {
		log.Printf("[SMTP ERROR] Failed to get mail config: %v", err)
		return fmt.Errorf("failed to get mail config: %v", err)
	}

	subject := "Добро пожаловать в Linka Type!"
	body := `
		<h2>🎉 Регистрация успешно завершена!</h2>
		<p>Добро пожаловать в Linka Type! Ваш аккаунт был успешно создан и подтвержден.</p>
		<p>Теперь вы можете:</p>
		<ul>
			<li>Создавать и управлять своими категориями</li>
			<li>Добавлять и отслеживать свои statements</li>
			<li>Получать уведомления в реальном времени</li>
		</ul>
		<p>Спасибо за регистрацию!</p>
	`

	m := gomail.NewMessage()
	m.SetHeader("From", config.Address)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(config.Server, config.Port, config.Address, config.Password)
	log.Printf("[SMTP INFO] Attempting to connect to SMTP server: %s:%d", config.Server, config.Port)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("[SMTP ERROR] Failed to send welcome email to %s: %v", to, err)
		log.Printf("[SMTP ERROR] SMTP Details: Server=%s, Port=%d, Address=%s", config.Server, config.Port, config.Address)
		return fmt.Errorf("failed to send welcome email: %v", err)
	}

	log.Printf("[SMTP SUCCESS] Welcome email sent successfully to: %s", to)
	return nil
}
