# Документация по аутентификации

## Обзор

API использует JWT (JSON Web Token) токены для аутентификации пользователей. Система поддерживает регистрацию, вход, верификацию email и сброс пароля с использованием OTP (One-Time Password) кодов.

## Базовый URL

```
https://type-backend.linka.su/api
```

## Аутентификация

Все защищенные endpoints требуют JWT токен в заголовке `Authorization`:

```
Authorization: Bearer <jwt_token>
```

## Endpoints аутентификации

### 1. Регистрация пользователя

**POST** `/register`

Создает нового пользователя и отправляет OTP код на email для верификации.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "StrongPassword123!"
}
```

**Требования к паролю:**
- Минимум 8 символов
- Должен содержать буквы (a-z, A-Z)
- Должен содержать цифры (0-9)
- Должен содержать специальные символы (!@#$%^&*)

**Response (201 Created):**
```json
{
  "message": "User registered successfully. Please check your email for verification code.",
  "user_id": "id_1757168658740968212",
  "email": "user@example.com"
}
```

**Ошибки:**
- `400 Bad Request` - Неверный формат email или слабый пароль
- `409 Conflict` - Пользователь уже существует

### 2. Верификация email

**POST** `/verify-email`

Подтверждает email пользователя с помощью OTP кода.

**Request Body:**
```json
{
  "email": "user@example.com",
  "code": "123456"
}
```

**Response (200 OK):**
```json
{
  "message": "Email verified successfully",
  "user_id": "id_1757168658740968212"
}
```

**Ошибки:**
- `400 Bad Request` - Неверный или истекший OTP код
- `404 Not Found` - Пользователь не найден

### 3. Вход в систему

**POST** `/login`

Аутентифицирует пользователя и возвращает JWT токен.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "StrongPassword123!"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "id_1757168658740968212",
    "email": "user@example.com"
  }
}
```

**Ошибки:**
- `400 Bad Request` - Неверный формат email
- `401 Unauthorized` - Неверные учетные данные

### 4. Запрос сброса пароля

**POST** `/reset-password`

Отправляет OTP код для сброса пароля на email пользователя.

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Response (200 OK):**
```json
{
  "message": "Password reset code sent to your email"
}
```

**Примечание:** Код отправляется даже если пользователь не существует (для безопасности).

### 5. Верификация OTP для сброса пароля

**POST** `/reset-password/verify`

Подтверждает OTP код для сброса пароля.

**Request Body:**
```json
{
  "email": "user@example.com",
  "code": "123456"
}
```

**Response (200 OK):**
```json
{
  "message": "OTP verified successfully"
}
```

**Ошибки:**
- `400 Bad Request` - Неверный или истекший OTP код
- `404 Not Found` - Пользователь не найден

### 6. Подтверждение сброса пароля

**POST** `/reset-password/confirm`

Устанавливает новый пароль после верификации OTP.

**Request Body:**
```json
{
  "email": "user@example.com",
  "code": "123456",
  "new_password": "NewStrongPassword123!"
}
```

**Response (200 OK):**
```json
{
  "message": "Password reset successfully"
}
```

**Ошибки:**
- `400 Bad Request` - Неверный OTP код или слабый пароль
- `404 Not Found` - Пользователь не найден

### 7. Обновление токена

**POST** `/refresh-token`

Обновляет JWT токен (если токен еще действителен).

**Headers:**
```
Authorization: Bearer <current_jwt_token>
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "id_1757168658740968212",
    "email": "user@example.com"
  }
}
```

## Защищенные endpoints

### Получение профиля пользователя

**GET** `/profile`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (200 OK):**
```json
{
  "user_id": "id_1757168658740968212",
  "email": "user@example.com"
}
```

## OTP коды

### Характеристики OTP:
- 6-значные числовые коды
- Время жизни: 10 минут
- Одноразовые (после использования становятся недействительными)
- Отправляются на email пользователя

### Типы OTP:
1. **Регистрация** - для верификации email при регистрации
2. **Сброс пароля** - для подтверждения сброса пароля

## JWT токены

### Характеристики:
- Время жизни: настраивается через переменные окружения
- Содержат: user_id, email, время выдачи и истечения
- Подписываются секретным ключом

### Структура payload:
```json
{
  "user_id": "id_1757168658740968212",
  "email": "user@example.com",
  "iat": 1757168658,
  "exp": 1757172258,
  "iss": "linka.type-backend",
  "aud": "linka.type-users"
}
```

## Примеры использования

### Полный цикл регистрации и входа

```javascript
// 1. Регистрация
const registerResponse = await fetch('/api/register', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'StrongPassword123!'
  })
});

const registerData = await registerResponse.json();
console.log('User registered:', registerData.user_id);

// 2. Верификация email (код приходит на email)
const verifyResponse = await fetch('/api/verify-email', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    code: '123456' // код из email
  })
});

// 3. Вход в систему
const loginResponse = await fetch('/api/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'StrongPassword123!'
  })
});

const loginData = await loginResponse.json();
const token = loginData.token;

// 4. Использование токена для защищенных запросов
const profileResponse = await fetch('/api/profile', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

const profile = await profileResponse.json();
console.log('User profile:', profile);
```

### Сброс пароля

```javascript
// 1. Запрос сброса пароля
await fetch('/api/reset-password', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com'
  })
});

// 2. Верификация OTP (код приходит на email)
await fetch('/api/reset-password/verify', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    code: '123456' // код из email
  })
});

// 3. Установка нового пароля
await fetch('/api/reset-password/confirm', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    code: '123456',
    new_password: 'NewStrongPassword123!'
  })
});
```

## Коды ошибок

| Код | Описание |
|-----|----------|
| 200 | Успешный запрос |
| 201 | Ресурс создан |
| 400 | Неверный формат запроса |
| 401 | Неавторизованный доступ |
| 403 | Доступ запрещен |
| 404 | Ресурс не найден |
| 409 | Конфликт (пользователь уже существует) |
| 500 | Внутренняя ошибка сервера |

## Безопасность

### Рекомендации для клиента:

1. **Хранение токенов:**
   - Не храните JWT токены в localStorage (уязвимо для XSS)
   - Используйте httpOnly cookies или secure storage
   - Реализуйте автоматическое обновление токенов

2. **Пароли:**
   - Используйте сильные пароли (8+ символов, буквы, цифры, спецсимволы)
   - Не передавайте пароли в URL или логах
   - Реализуйте проверку силы пароля на клиенте

3. **HTTPS:**
   - Всегда используйте HTTPS в продакшене
   - Не передавайте токены по незащищенным соединениям

4. **Обработка ошибок:**
   - Не показывайте детальные ошибки пользователю
   - Логируйте ошибки для отладки
   - Реализуйте graceful fallback для сетевых ошибок

## Переменные окружения

Для настройки аутентификации используются следующие переменные:

```bash
# JWT настройки
JWT_SECRET=your-secret-key
JWT_ISSUER=linka.type-backend
JWT_AUDIENCE=linka.type-users
JWT_EXPIRATION=1h

# Email настройки
MAIL_SERVER=smtp.gmail.com
MAIL_PORT=587
MAIL_ADDRESS=your-email@gmail.com
MAIL_PASSWORD=your-app-password

# Firebase (опционально)
FIREBASE_API_KEY=your-firebase-api-key
FIREBASE_CONFIG_PATH=./firebase.json
```

## Поддержка

При возникновении проблем с аутентификацией:

1. Проверьте формат запросов
2. Убедитесь в правильности email и пароля
3. Проверьте срок действия OTP кодов
4. Убедитесь в наличии интернет-соединения
5. Проверьте настройки CORS для вашего домена



