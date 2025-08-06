package mail

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

// MailConfig содержит конфигурацию SMTP
type MailConfig struct {
	Server   string
	Port     int
	Address  string
	Password string
}

// GetMailConfig получает конфигурацию из переменных окружения
func GetMailConfig() (*MailConfig, error) {
	server := os.Getenv("MAIL_SERVER")
	if server == "" {
		return nil, fmt.Errorf("MAIL_SERVER environment variable is required")
	}

	address := os.Getenv("MAIL_ADRESS")
	if address == "" {
		return nil, fmt.Errorf("MAIL_ADRESS environment variable is required")
	}

	password := os.Getenv("MAIL_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("MAIL_PASSWORD environment variable is required")
	}

	// Порт по умолчанию для Yandex SMTP
	port := 587
	if portStr := os.Getenv("MAIL_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	return &MailConfig{
		Server:   server,
		Port:     port,
		Address:  address,
		Password: password,
	}, nil
}

// SendOTPEmail отправляет email с OTP кодом
func SendOTPEmail(to, otpCode, otpType string) error {
	config, err := GetMailConfig()
	if err != nil {
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

	// Отправляем email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// SendWelcomeEmail отправляет приветственное письмо после подтверждения email
func SendWelcomeEmail(to string) error {
	config, err := GetMailConfig()
	if err != nil {
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

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send welcome email: %v", err)
	}

	return nil
}
