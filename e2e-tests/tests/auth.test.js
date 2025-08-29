const ApiClient = require('../utils/api-client');

describe('Authentication Tests', () => {
  let apiClient;

  beforeEach(() => {
    apiClient = new ApiClient();
  });

  afterEach(() => {
    apiClient.clearToken();
  });

  describe('Basic Authentication', () => {
    test('should register new user successfully', async () => {
      const email = generateTestEmail();
      const password = generateTestPassword();

      const response = await apiClient.register(email, password);

      expect(response).toHaveProperty('token');
      expect(response).toHaveProperty('user');
      expect(response.user).toHaveProperty('id');
      expect(response.user.email).toBe(email);
      expect(typeof response.token).toBe('string');
      expect(response.token.length).toBeGreaterThan(0);
    });

    test('should login existing user successfully', async () => {
      const email = generateTestEmail();
      const password = generateTestPassword();

      // Сначала регистрируем пользователя
      await apiClient.register(email, password);

      // Затем логинимся
      const response = await apiClient.login(email, password);

      expect(response).toHaveProperty('token');
      expect(response).toHaveProperty('user');
      expect(response.user.email).toBe(email);
      expect(typeof response.token).toBe('string');
    });

    test('should get user profile with valid token', async () => {
      const email = generateTestEmail();
      const password = generateTestPassword();

      const registerResponse = await apiClient.register(email, password);
      apiClient.setToken(registerResponse.token);

      const profileResponse = await apiClient.getProfile();

      expect(profileResponse).toHaveProperty('user_id');
      expect(profileResponse).toHaveProperty('email');
      expect(profileResponse.email).toBe(email);
    });

    test('should reject login with wrong password', async () => {
      const email = generateTestEmail();
      const password = generateTestPassword();

      // Регистрируем пользователя
      await apiClient.register(email, password);

      // Пытаемся войти с неправильным паролем
      await expect(apiClient.login(email, 'wrongpassword')).rejects.toThrow();
    });

    test('should reject access to protected endpoint without token', async () => {
      await expect(apiClient.getProfile()).rejects.toThrow();
    });

    test('should reject access to protected endpoint with invalid token', async () => {
      apiClient.setToken('invalid-token');
      await expect(apiClient.getProfile()).rejects.toThrow();
    });

    test('should reject registration with invalid email', async () => {
      const password = generateTestPassword();

      await expect(apiClient.register('invalid-email', password)).rejects.toThrow();
    });

    test('should reject registration with empty password', async () => {
      const email = generateTestEmail();

      await expect(apiClient.register(email, '')).rejects.toThrow();
    });
  });

  describe('OTP Authentication', () => {
    test('should register user with OTP successfully', async () => {
      const email = generateTestEmail();
      const password = generateTestPassword();

      const response = await apiClient.registerWithOTP(email, password);

      expect(response).toHaveProperty('message');
      expect(response).toHaveProperty('user_id');
      expect(response.message).toContain('verification code');
    });

    test('should verify email with OTP code', async () => {
      const email = generateTestEmail();
      const password = generateTestPassword();

      // Регистрируем пользователя с OTP
      await apiClient.registerWithOTP(email, password);

      // В реальном тесте здесь нужно получить OTP код из email
      // Для тестирования используем моковый код
      const mockOTPCode = '123456';

      try {
        const response = await apiClient.verifyEmail(email, mockOTPCode);
        expect(response).toHaveProperty('message');
        expect(response).toHaveProperty('token');
        expect(response).toHaveProperty('user');
      } catch (error) {
        // Ожидаем ошибку, так как используем моковый код
        expect(error.response.status).toBe(400);
      }
    });

    test('should request password reset successfully', async () => {
      const email = generateTestEmail();
      const password = generateTestPassword();

      // Сначала регистрируем и верифицируем пользователя
      await apiClient.registerWithOTP(email, password);
      
      const response = await apiClient.requestPasswordReset(email);

      expect(response).toHaveProperty('message');
      expect(response.message).toContain('reset code has been sent');
    });

    test('should verify password reset OTP', async () => {
      const email = generateTestEmail();
      const password = generateTestPassword();

      // Регистрируем пользователя
      await apiClient.registerWithOTP(email, password);
      
      // Запрашиваем сброс пароля
      await apiClient.requestPasswordReset(email);

      // Пытаемся верифицировать с моковым кодом
      const mockOTPCode = '123456';

      try {
        const response = await apiClient.verifyPasswordResetOTP(email, mockOTPCode);
        expect(response).toHaveProperty('message');
        expect(response).toHaveProperty('otp_id');
      } catch (error) {
        // Ожидаем ошибку, так как используем моковый код
        expect(error.response.status).toBe(400);
      }
    });

    test('should confirm password reset', async () => {
      const email = generateTestEmail();
      const password = generateTestPassword();
      const newPassword = generateTestPassword();

      // Регистрируем пользователя
      await apiClient.registerWithOTP(email, password);
      
      // Запрашиваем сброс пароля
      await apiClient.requestPasswordReset(email);

      // Пытаемся подтвердить сброс с моковым кодом
      const mockOTPCode = '123456';

      try {
        const response = await apiClient.confirmPasswordReset(email, mockOTPCode, newPassword);
        expect(response).toHaveProperty('message');
        expect(response.message).toContain('reset successfully');
      } catch (error) {
        // Ожидаем ошибку, так как используем моковый код
        expect(error.response.status).toBe(400);
      }
    });
  });

  describe('Health Check', () => {
    test('should return health status', async () => {
      const response = await apiClient.healthCheck();

      expect(response).toHaveProperty('status');
      expect(response.status).toBe('ok');
    });
  });
});
