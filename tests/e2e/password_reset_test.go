package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"linka.type-backend/db"
	"linka.type-backend/handlers"
	"linka.type-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PasswordResetTestSuite содержит тесты для сброса пароля
type PasswordResetTestSuite struct {
	router       *gin.Engine
	userID       string
	email        string
	otpCode      string
	resetOTPCode string
}

// NewPasswordResetTestSuite создает новый набор тестов для сброса пароля
func NewPasswordResetTestSuite() *PasswordResetTestSuite {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Добавляем CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Группа для публичных маршрутов
	public := router.Group("/api/auth")
	{
		public.POST("/register", handlers.Register)
		public.POST("/login", handlers.Login)
		public.POST("/verify-email", handlers.VerifyEmail)
		public.POST("/reset-password", handlers.RequestPasswordReset)
		public.POST("/reset-password/verify", handlers.VerifyPasswordResetOTP)
		public.POST("/reset-password/confirm", handlers.ConfirmPasswordReset)
	}

	return &PasswordResetTestSuite{
		router: router,
		userID: utils.GenerateID(),
		email:  "test-reset@example.com",
	}
}

// TestPasswordResetFlow тестирует полный процесс сброса пароля
func TestPasswordResetFlow(t *testing.T) {
	suite := NewPasswordResetTestSuite()

	// Шаг 1: Регистрация пользователя
	t.Run("Register User", func(t *testing.T) {
		reqBody := handlers.RegisterRequest{
			Email:    suite.email,
			Password: "oldpassword123",
		}

		w := suite.makeRequest("POST", "/api/auth/register", reqBody, "")
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "message")
		assert.Contains(t, response, "user_id")
	})

	// Шаг 2: Получение OTP кода из базы данных (имитация получения email)
	t.Run("Get OTP Code from Database", func(t *testing.T) {
		otpCRUD := &db.OTPCRUD{}
		otpRecord, err := otpCRUD.GetOTPByEmailAndType(suite.email, "registration")
		require.NoError(t, err)
		require.NotNil(t, otpRecord)
		
		// Сохраняем код для использования в следующих тестах
		suite.otpCode = otpRecord.Code
	})

	// Шаг 3: Верификация email
	t.Run("Verify Email", func(t *testing.T) {
		reqBody := handlers.VerifyEmailRequest{
			Email: suite.email,
			Code:  suite.otpCode,
		}

		w := suite.makeRequest("POST", "/api/auth/verify-email", reqBody, "")
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "message")
		assert.Contains(t, response, "token")
		assert.Contains(t, response, "user")
	})

	// Шаг 4: Запрос сброса пароля
	t.Run("Request Password Reset", func(t *testing.T) {
		reqBody := handlers.ResetPasswordRequest{
			Email: suite.email,
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password", reqBody, "")
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "message")
	})

	// Шаг 5: Получение OTP кода для сброса пароля
	t.Run("Get Reset OTP Code", func(t *testing.T) {
		otpCRUD := &db.OTPCRUD{}
		otpRecord, err := otpCRUD.GetOTPByEmailAndType(suite.email, "reset_password")
		require.NoError(t, err)
		require.NotNil(t, otpRecord)
		
		// Сохраняем код для сброса пароля
		suite.resetOTPCode = otpRecord.Code
	})

	// Шаг 6: Верификация OTP для сброса пароля
	t.Run("Verify Reset OTP", func(t *testing.T) {
		reqBody := handlers.ResetPasswordVerifyRequest{
			Email: suite.email,
			Code:  suite.resetOTPCode,
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password/verify", reqBody, "")
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "message")
		assert.Contains(t, response, "otp_id")
	})

	// Шаг 7: Подтверждение сброса пароля
	t.Run("Confirm Password Reset", func(t *testing.T) {
		reqBody := handlers.ResetPasswordConfirmRequest{
			Email:    suite.email,
			Code:     suite.resetOTPCode,
			Password: "newpassword123",
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password/confirm", reqBody, "")
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "message")
	})

	// Шаг 8: Проверка, что новый пароль работает
	t.Run("Login with New Password", func(t *testing.T) {
		reqBody := handlers.LoginRequest{
			Email:    suite.email,
			Password: "newpassword123",
		}

		w := suite.makeRequest("POST", "/api/auth/login", reqBody, "")
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "token")
		assert.Contains(t, response, "user")
	})

	// Шаг 9: Проверка, что старый пароль больше не работает
	t.Run("Login with Old Password Should Fail", func(t *testing.T) {
		reqBody := handlers.LoginRequest{
			Email:    suite.email,
			Password: "oldpassword123",
		}

		w := suite.makeRequest("POST", "/api/auth/login", reqBody, "")
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestPasswordResetInvalidScenarios тестирует некорректные сценарии сброса пароля
func TestPasswordResetInvalidScenarios(t *testing.T) {
	suite := NewPasswordResetTestSuite()

	t.Run("Request Reset for Non-existent User", func(t *testing.T) {
		reqBody := handlers.ResetPasswordRequest{
			Email: "nonexistent@example.com",
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password", reqBody, "")
		
		// Должен вернуть 200, но не отправлять email
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "message")
	})

	t.Run("Verify Invalid OTP Code", func(t *testing.T) {
		reqBody := handlers.ResetPasswordVerifyRequest{
			Email: suite.email,
			Code:  "000000",
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password/verify", reqBody, "")
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "error")
	})

	t.Run("Confirm Reset with Invalid OTP", func(t *testing.T) {
		reqBody := handlers.ResetPasswordConfirmRequest{
			Email:    suite.email,
			Code:     "000000",
			Password: "newpassword123",
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password/confirm", reqBody, "")
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "error")
	})

	t.Run("Confirm Reset with Weak Password", func(t *testing.T) {
		reqBody := handlers.ResetPasswordConfirmRequest{
			Email:    suite.email,
			Code:     "123456",
			Password: "123", // Слишком короткий пароль
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password/confirm", reqBody, "")
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "error")
	})
}

// TestOTPExpiration тестирует истечение срока действия OTP
func TestOTPExpiration(t *testing.T) {
	suite := NewPasswordResetTestSuite()

	// Создаем пользователя
	t.Run("Setup User", func(t *testing.T) {
		reqBody := handlers.RegisterRequest{
			Email:    suite.email,
			Password: "password123",
		}

		w := suite.makeRequest("POST", "/api/auth/register", reqBody, "")
		assert.Equal(t, http.StatusCreated, w.Code)

		// Получаем OTP код
		otpCRUD := &db.OTPCRUD{}
		otpRecord, err := otpCRUD.GetOTPByEmailAndType(suite.email, "registration")
		require.NoError(t, err)
		require.NotNil(t, otpRecord)
		suite.otpCode = otpRecord.Code

		// Верифицируем email
		verifyReq := handlers.VerifyEmailRequest{
			Email: suite.email,
			Code:  suite.otpCode,
		}
		w = suite.makeRequest("POST", "/api/auth/verify-email", verifyReq, "")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Запрашиваем сброс пароля
	t.Run("Request Password Reset", func(t *testing.T) {
		reqBody := handlers.ResetPasswordRequest{
			Email: suite.email,
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password", reqBody, "")
		assert.Equal(t, http.StatusOK, w.Code)

		// Получаем OTP код
		otpCRUD := &db.OTPCRUD{}
		otpRecord, err := otpCRUD.GetOTPByEmailAndType(suite.email, "reset_password")
		require.NoError(t, err)
		require.NotNil(t, otpRecord)
		suite.resetOTPCode = otpRecord.Code
	})

	// Имитируем истечение срока действия OTP
	t.Run("Simulate OTP Expiration", func(t *testing.T) {
		// Обновляем время истечения в базе данных на прошлое время
		otpCRUD := &db.OTPCRUD{}
		otpRecord, err := otpCRUD.GetOTPByCode(suite.resetOTPCode, suite.email, "reset_password")
		require.NoError(t, err)
		require.NotNil(t, otpRecord)

		// Обновляем время истечения на прошлое время
		expiredTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		err = otpCRUD.UpdateOTPExpiration(otpRecord.ID, expiredTime)
		require.NoError(t, err)
	})

	// Пытаемся использовать истекший OTP
	t.Run("Use Expired OTP", func(t *testing.T) {
		reqBody := handlers.ResetPasswordVerifyRequest{
			Email: suite.email,
			Code:  suite.resetOTPCode,
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password/verify", reqBody, "")
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"], "expired")
	})
}

// TestOTPReuse тестирует предотвращение повторного использования OTP
func TestOTPReuse(t *testing.T) {
	suite := NewPasswordResetTestSuite()

	// Создаем пользователя и верифицируем email
	t.Run("Setup User", func(t *testing.T) {
		reqBody := handlers.RegisterRequest{
			Email:    suite.email,
			Password: "password123",
		}

		w := suite.makeRequest("POST", "/api/auth/register", reqBody, "")
		assert.Equal(t, http.StatusCreated, w.Code)

		// Получаем и используем OTP для верификации
		otpCRUD := &db.OTPCRUD{}
		otpRecord, err := otpCRUD.GetOTPByEmailAndType(suite.email, "registration")
		require.NoError(t, err)
		require.NotNil(t, otpRecord)

		verifyReq := handlers.VerifyEmailRequest{
			Email: suite.email,
			Code:  otpRecord.Code,
		}
		w = suite.makeRequest("POST", "/api/auth/verify-email", verifyReq, "")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Запрашиваем сброс пароля
	t.Run("Request Password Reset", func(t *testing.T) {
		reqBody := handlers.ResetPasswordRequest{
			Email: suite.email,
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password", reqBody, "")
		assert.Equal(t, http.StatusOK, w.Code)

		// Получаем OTP код
		otpCRUD := &db.OTPCRUD{}
		otpRecord, err := otpCRUD.GetOTPByEmailAndType(suite.email, "reset_password")
		require.NoError(t, err)
		require.NotNil(t, otpRecord)
		suite.resetOTPCode = otpRecord.Code
	})

	// Используем OTP первый раз
	t.Run("Use OTP First Time", func(t *testing.T) {
		reqBody := handlers.ResetPasswordVerifyRequest{
			Email: suite.email,
			Code:  suite.resetOTPCode,
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password/verify", reqBody, "")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Пытаемся использовать тот же OTP второй раз
	t.Run("Reuse OTP Should Fail", func(t *testing.T) {
		reqBody := handlers.ResetPasswordVerifyRequest{
			Email: suite.email,
			Code:  suite.resetOTPCode,
		}

		w := suite.makeRequest("POST", "/api/auth/reset-password/verify", reqBody, "")
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "error")
	})
}

// makeRequest вспомогательная функция для выполнения HTTP запросов
func (suite *PasswordResetTestSuite) makeRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			panic(err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

 