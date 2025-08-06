package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"linka.type-backend/otp"
)

// PasswordResetCLI интерактивный CLI для сброса пароля
type PasswordResetCLI struct {
	baseURL string
	reader  *bufio.Reader
}

// NewPasswordResetCLI создает новый CLI экземпляр
func NewPasswordResetCLI(baseURL string) *PasswordResetCLI {
	return &PasswordResetCLI{
		baseURL: baseURL,
		reader:  bufio.NewReader(os.Stdin),
	}
}

// readLine читает строку из консоли
func (cli *PasswordResetCLI) readLine(prompt string) string {
	fmt.Print(prompt)
	input, _ := cli.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// readPassword читает пароль из консоли (скрывая ввод)
func (cli *PasswordResetCLI) readPassword(prompt string) string {
	fmt.Print(prompt)
	input, _ := cli.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// makeRequest выполняет HTTP запрос
func (cli *PasswordResetCLI) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
	}

	url := cli.baseURL + endpoint
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// requestPasswordReset запрашивает сброс пароля
func (cli *PasswordResetCLI) requestPasswordReset(email string) error {
	fmt.Println("\n🔄 Запрашиваем сброс пароля...")

	reqBody := map[string]string{
		"email": email,
	}

	resp, err := cli.makeRequest("POST", "/api/auth/reset-password", reqBody)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	fmt.Println("✅ Код сброса пароля отправлен на ваш email")
	return nil
}

// getOTPFromDatabase получает OTP код из базы данных (для тестирования)
func (cli *PasswordResetCLI) getOTPFromDatabase(email string) (string, error) {
	// Используем Docker для получения OTP
	cmd := exec.Command("docker", "compose", "exec", "-T", "db", "psql", "-U", "postgres", "-d", "linkatype", "-c",
		fmt.Sprintf("SELECT code FROM otp_codes WHERE email = '%s' AND type = 'reset_password' ORDER BY created_at DESC LIMIT 1;", email))

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get OTP from database: %v", err)
	}

	// Парсим результат
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "code") && !strings.Contains(line, "----") && !strings.Contains(line, "(") {
			return line, nil
		}
	}

	return "", fmt.Errorf("no active OTP found for email: %s", email)
}

// verifyOTP верифицирует OTP код
func (cli *PasswordResetCLI) verifyOTP(email, code string) error {
	fmt.Println("\n🔍 Верифицируем код...")

	reqBody := map[string]string{
		"email": email,
		"code":  code,
	}

	resp, err := cli.makeRequest("POST", "/api/auth/reset-password/verify", reqBody)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("verification failed with status: %d", resp.StatusCode)
	}

	fmt.Println("✅ Код верифицирован успешно")
	return nil
}

// confirmPasswordReset подтверждает сброс пароля
func (cli *PasswordResetCLI) confirmPasswordReset(email, code, newPassword string) error {
	fmt.Println("\n🔐 Устанавливаем новый пароль...")

	reqBody := map[string]string{
		"email":    email,
		"code":     code,
		"password": newPassword,
	}

	resp, err := cli.makeRequest("POST", "/api/auth/reset-password/confirm", reqBody)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("password reset failed with status: %d", resp.StatusCode)
	}

	fmt.Println("✅ Пароль успешно изменен!")
	return nil
}

// validateEmail проверяет корректность email
func (cli *PasswordResetCLI) validateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// validatePassword проверяет корректность пароля
func (cli *PasswordResetCLI) validatePassword(password string) (bool, string) {
	if len(password) < 6 {
		return false, "Пароль должен содержать минимум 6 символов"
	}
	return true, ""
}

// validateOTP проверяет корректность OTP кода
func (cli *PasswordResetCLI) validateOTP(code string) bool {
	return otp.ValidateOTPCode(code)
}

// Run запускает интерактивный процесс сброса пароля
func (cli *PasswordResetCLI) Run() {
	fmt.Println("🔐 Интерактивный сброс пароля")
	fmt.Println("================================")

	// Шаг 1: Ввод email
	var email string
	for {
		email = cli.readLine("Введите ваш email: ")
		if cli.validateEmail(email) {
			break
		}
		fmt.Println("❌ Некорректный email. Попробуйте еще раз.")
	}

	// Шаг 2: Запрос сброса пароля
	if err := cli.requestPasswordReset(email); err != nil {
		fmt.Printf("❌ Ошибка при запросе сброса пароля: %v\n", err)
		return
	}

	// Шаг 3: Получение OTP кода
	fmt.Println("\n📧 Проверьте ваш email и введите код подтверждения.")
	fmt.Println("💡 Для тестирования код можно получить из базы данных.")

	var code string
	for {
		code = cli.readLine("Введите 6-значный код: ")
		if cli.validateOTP(code) {
			break
		}
		fmt.Println("❌ Некорректный код. Введите 6 цифр.")
	}

	// Шаг 4: Верификация OTP
	if err := cli.verifyOTP(email, code); err != nil {
		fmt.Printf("❌ Ошибка при верификации кода: %v\n", err)
		return
	}

	// Шаг 5: Ввод нового пароля
	var newPassword string
	for {
		newPassword = cli.readPassword("Введите новый пароль: ")
		if valid, message := cli.validatePassword(newPassword); valid {
			break
		} else {
			fmt.Printf("❌ %s\n", message)
		}
	}

	// Подтверждение пароля
	confirmPassword := cli.readPassword("Подтвердите новый пароль: ")
	if newPassword != confirmPassword {
		fmt.Println("❌ Пароли не совпадают.")
		return
	}

	// Шаг 6: Подтверждение сброса пароля
	if err := cli.confirmPasswordReset(email, code, newPassword); err != nil {
		fmt.Printf("❌ Ошибка при сбросе пароля: %v\n", err)
		return
	}

	fmt.Println("\n🎉 Сброс пароля завершен успешно!")
	fmt.Println("Теперь вы можете войти в систему с новым паролем.")
}

// TestMode запускает тестовый режим с автоматическим получением OTP из базы
func (cli *PasswordResetCLI) TestMode() {
	fmt.Println("🧪 Тестовый режим сброса пароля")
	fmt.Println("==================================")

	// Шаг 1: Ввод email
	var email string
	for {
		email = cli.readLine("Введите ваш email: ")
		if cli.validateEmail(email) {
			break
		}
		fmt.Println("❌ Некорректный email. Попробуйте еще раз.")
	}

	// Шаг 2: Запрос сброса пароля
	if err := cli.requestPasswordReset(email); err != nil {
		fmt.Printf("❌ Ошибка при запросе сброса пароля: %v\n", err)
		return
	}

	// Шаг 3: Автоматическое получение OTP из базы данных
	fmt.Println("\n🔍 Получаем OTP код из базы данных...")
	code, err := cli.getOTPFromDatabase(email)
	if err != nil {
		fmt.Printf("❌ Ошибка при получении OTP: %v\n", err)
		return
	}

	fmt.Printf("✅ Получен код: %s\n", code)

	// Шаг 4: Верификация OTP
	if err := cli.verifyOTP(email, code); err != nil {
		fmt.Printf("❌ Ошибка при верификации кода: %v\n", err)
		return
	}

	// Шаг 5: Ввод нового пароля
	var newPassword string
	for {
		newPassword = cli.readPassword("Введите новый пароль: ")
		if valid, message := cli.validatePassword(newPassword); valid {
			break
		} else {
			fmt.Printf("❌ %s\n", message)
		}
	}

	// Подтверждение пароля
	confirmPassword := cli.readPassword("Подтвердите новый пароль: ")
	if newPassword != confirmPassword {
		fmt.Println("❌ Пароли не совпадают.")
		return
	}

	// Шаг 6: Подтверждение сброса пароля
	if err := cli.confirmPasswordReset(email, code, newPassword); err != nil {
		fmt.Printf("❌ Ошибка при сбросе пароля: %v\n", err)
		return
	}

	fmt.Println("\n🎉 Сброс пароля завершен успешно!")
	fmt.Println("Теперь вы можете войти в систему с новым паролем.")
}

func main() {
	// Настройка базового URL
	baseURL := "http://localhost:8080"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	cli := NewPasswordResetCLI(baseURL)

	// Проверяем аргументы командной строки
	if len(os.Args) > 2 && os.Args[2] == "--test" {
		cli.TestMode()
	} else {
		fmt.Println("💡 Для тестового режима используйте: go run scripts/password_reset_cli.go [baseURL] --test")
		cli.Run()
	}
}
