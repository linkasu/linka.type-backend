# Документация проекта Linka Type Backend

**Сгенерировано:** 2025-08-31 00:57:45

## Содержание

- [API Endpoints](#api-endpoints)
- [websocket](#websocket)
- [bl](#bl)
- [fb](#fb)
- [auth](#auth)
- [db](#db)
- [handlers](#handlers)
- [utils](#utils)

## API Endpoints

### POST /api/register

Регистрация нового пользователя

**Handler:** handlers.Register
**Auth:** Нет

**Параметры:**
- `email` (string) *обязательный* - Email пользователя
- `password` (string) *обязательный* - Пароль пользователя

**Ответы:**
- `200` - Успешная регистрация
- `400` - Ошибка валидации
- `500` - Внутренняя ошибка сервера

---

### POST /api/login

Авторизация пользователя

**Handler:** handlers.Login
**Auth:** Нет

**Параметры:**
- `email` (string) *обязательный* - Email пользователя
- `password` (string) *обязательный* - Пароль пользователя

**Ответы:**
- `200` - Успешная авторизация
- `401` - Неверные учетные данные
- `500` - Внутренняя ошибка сервера

---

### GET /api/statements

Получение списка statements

**Handler:** handlers.GetStatements
**Auth:** Да

**Ответы:**
- `200` - Список statements
- `401` - Не авторизован
- `500` - Внутренняя ошибка сервера

---

### POST /api/statements

Создание нового statement

**Handler:** handlers.CreateStatement
**Auth:** Да

**Параметры:**
- `title` (string) *обязательный* - Заголовок statement
- `content` (string) *обязательный* - Содержание statement
- `category_id` (string) *обязательный* - ID категории

**Ответы:**
- `201` - Statement создан
- `400` - Ошибка валидации
- `401` - Не авторизован
- `500` - Внутренняя ошибка сервера

---

### GET /api/categories

Получение списка категорий

**Handler:** handlers.GetCategories
**Auth:** Да

**Ответы:**
- `200` - Список категорий
- `401` - Не авторизован
- `500` - Внутренняя ошибка сервера

---

### POST /api/categories

Создание новой категории

**Handler:** handlers.CreateCategory
**Auth:** Да

**Параметры:**
- `name` (string) *обязательный* - Название категории
- `description` (string)  - Описание категории

**Ответы:**
- `201` - Категория создана
- `400` - Ошибка валидации
- `401` - Не авторизован
- `500` - Внутренняя ошибка сервера

---

### GET /api/ws

WebSocket подключение для real-time уведомлений

**Handler:** handlers.HandleWebSocket
**Auth:** Да

**Ответы:**
- `101` - WebSocket upgrade успешен
- `401` - Не авторизован

---

## fb

### Функции

#### CheckPassword

**Параметры:**
- `email` (string) *обязательный* - 
- `password` (string) *обязательный* - 

**Возвращает:**
- *FirebaseAuthResponse
- error

**Файл:** fb/CheckPassword.go:29

---

#### GetCategories

**Параметры:**
- `user` (*auth.UserRecord) *обязательный* - 

**Возвращает:**
- []*Category
- error

**Файл:** fb/GetCategories.go:15

---

#### GetStatements

**Возвращает:**
- []*Statement
- error

**Файл:** fb/GetStatements.go:21

---

#### GetUser

**Параметры:**
- `email` (string) *обязательный* - 

**Возвращает:**
- *auth.UserRecord
- error

**Файл:** fb/GetUser.go:9

---

#### init

**Файл:** fb/fb.go:14

---

### Структуры

#### FirebaseAuthResponse

FirebaseAuthResponse represents the response from Firebase Authentication API

**Поля:**
- `IDToken` (string) - 
- `Email` (string) - 
- `RefreshToken` (string) - 
- `ExpiresIn` (string) - 
- `LocalID` (string) - 
- `Registered` (bool) - 

**Файл:** fb/CheckPassword.go:13

---

#### FirebaseAuthRequest

FirebaseAuthRequest represents the request to Firebase Authentication API

**Поля:**
- `Email` (string) - 
- `Password` (string) - 
- `ReturnSecureToken` (bool) - 

**Файл:** fb/CheckPassword.go:23

---

#### Category

**Поля:**
- `ID` (string) - 
- `Label` (string) - 
- `UserID` (string) - 

**Файл:** fb/GetCategories.go:9

---

#### Statement

**Поля:**
- `ID` (string) - 
- `CreatedAt` (time.Time) - 
- `Text` (string) - 
- `UserID` (string) - 
- `CategoryID` (string) - 

**Файл:** fb/GetStatements.go:8

---

#### fbStatementRaw

**Поля:**
- `CreatedAt` (int64) - 

**Файл:** fb/GetStatements.go:16

---

## auth

### Функции

#### getJWTSecret

**Возвращает:**
- []byte
- error

**Файл:** auth/jwt.go:15

---

#### GenerateToken

GenerateToken генерирует JWT токен для пользователя

**Параметры:**
- `userID` (string) *обязательный* - 
- `email` (string) *обязательный* - 

**Возвращает:**
- string
- error

**Файл:** auth/jwt.go:34

---

#### ValidateToken

ValidateToken валидирует JWT токен и возвращает claims

**Параметры:**
- `tokenString` (string) *обязательный* - 

**Возвращает:**
- *Claims
- error

**Файл:** auth/jwt.go:66

---

#### SetJWTSecret

SetJWTSecret устанавливает секретный ключ для JWT (для конфигурации)

**Параметры:**
- `secret` (string) *обязательный* - 

**Файл:** auth/jwt.go:98

---

#### TestGenerateAndValidateToken

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** auth/jwt_test.go:7

---

#### TestValidateInvalidToken

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** auth/jwt_test.go:37

---

#### TestTokenExpiration

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** auth/jwt_test.go:45

---

#### TestSetJWTSecret

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** auth/jwt_test.go:66

---

#### AuthMiddleware

AuthMiddleware проверяет JWT токен

**Возвращает:**
- gin.HandlerFunc

**Файл:** auth/middleware.go:13

---

#### EmailVerifiedMiddleware

EmailVerifiedMiddleware проверяет, что email пользователя верифицирован

**Возвращает:**
- gin.HandlerFunc

**Файл:** auth/middleware.go:46

---

#### GetUserIDFromContext

GetUserIDFromContext получает user_id из контекста Gin

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Возвращает:**
- string

**Файл:** auth/middleware.go:76

---

#### GetEmailFromContext

GetEmailFromContext получает email из контекста Gin

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Возвращает:**
- string

**Файл:** auth/middleware.go:84

---

#### JWTAuthMiddleware

JWTAuthMiddleware является алиасом для AuthMiddleware

**Возвращает:**
- gin.HandlerFunc

**Файл:** auth/middleware.go:92

---

### Структуры

#### Claims

Claims структура для JWT claims

**Поля:**
- `UserID` (string) - 
- `Email` (string) - 

**Файл:** auth/jwt.go:27

---

## db

### Функции

#### CreateCategory

CreateCategory creates a new category

**Параметры:**
- `category` (*Category) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/category_crud.go:13

---

#### GetCategoryByID

GetCategoryByID retrieves a category by ID

**Параметры:**
- `id` (string) *обязательный* - 

**Возвращает:**
- *Category
- error

**Файл:** db/category_crud.go:29

---

#### GetCategoriesByUserID

GetCategoriesByUserID retrieves all categories for a specific user

**Параметры:**
- `userID` (string) *обязательный* - 

**Возвращает:**
- []*Category
- error

**Файл:** db/category_crud.go:45

---

#### GetAllCategories

GetAllCategories retrieves all categories

**Возвращает:**
- []*Category
- error

**Файл:** db/category_crud.go:71

---

#### UpdateCategory

UpdateCategory updates an existing category

**Параметры:**
- `category` (*Category) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/category_crud.go:97

---

#### DeleteCategory

DeleteCategory deletes a category by ID

**Параметры:**
- `id` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/category_crud.go:123

---

#### DeleteCategoriesByUserID

DeleteCategoriesByUserID deletes all categories for a specific user

**Параметры:**
- `userID` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/category_crud.go:144

---

#### CategoryExists

CategoryExists checks if a category exists by ID

**Параметры:**
- `id` (string) *обязательный* - 

**Возвращает:**
- bool
- error

**Файл:** db/category_crud.go:165

---

#### InitDB

InitDB initializes the database connection

**Возвращает:**
- error

**Файл:** db/connection.go:16

---

#### CloseDB

CloseDB closes the database connection

**Возвращает:**
- error

**Файл:** db/connection.go:59

---

#### createTables

createTables creates the necessary tables if they don't exist

**Возвращает:**
- error

**Файл:** db/connection.go:67

---

#### getEnv

getEnv gets an environment variable or returns a default value

**Параметры:**
- `key` (string) *обязательный* - 
- `defaultValue` (string) *обязательный* - 

**Возвращает:**
- string

**Файл:** db/connection.go:144

---

#### ExampleCRUDOperations

ExampleCRUDOperations demonstrates all CRUD operations

**Файл:** db/examples.go:11

---

#### CreateMigrationLogsTable

CreateMigrationLogsTable создает таблицу для отслеживания миграций

**Возвращает:**
- error

**Файл:** db/migration_tracker.go:26

---

#### LogMigration

LogMigration записывает информацию о миграции

**Параметры:**
- `entityType` (string) *обязательный* - 
- `entityID` (string) *обязательный* - 
- `userID` (string) *обязательный* - 
- `action` (string) *обязательный* - 
- `status` (string) *обязательный* - 
- `errorMsg` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/migration_tracker.go:50

---

#### GetLastMigrationStatus

GetLastMigrationStatus получает статус последней миграции для сущности

**Параметры:**
- `entityType` (string) *обязательный* - 
- `entityID` (string) *обязательный* - 
- `userID` (string) *обязательный* - 

**Возвращает:**
- *MigrationLog
- error

**Файл:** db/migration_tracker.go:72

---

#### GetMigrationStats

GetMigrationStats получает статистику миграций

**Параметры:**
- `entityType` (string) *обязательный* - 

**Возвращает:**
- *ast.MapType
- error

**Файл:** db/migration_tracker.go:97

---

#### ClearMigrationLogs

ClearMigrationLogs очищает логи миграций (для тестирования)

**Параметры:**
- `entityType` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/migration_tracker.go:125

---

#### CreateOTP

CreateOTP создает новый OTP код

**Параметры:**
- `otpCode` (*OTPCode) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/otp_crud.go:13

---

#### GetOTPByEmailAndType

GetOTPByEmailAndType получает активный OTP код по email и типу

**Параметры:**
- `email` (string) *обязательный* - 
- `otpType` (string) *обязательный* - 

**Возвращает:**
- *OTPCode
- error

**Файл:** db/otp_crud.go:37

---

#### MarkOTPAsUsed

MarkOTPAsUsed помечает OTP код как использованный

**Параметры:**
- `id` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/otp_crud.go:68

---

#### DeleteExpiredOTP

DeleteExpiredOTP удаляет истекшие OTP коды

**Возвращает:**
- error

**Файл:** db/otp_crud.go:80

---

#### DeleteOTPByEmail

DeleteOTPByEmail удаляет все OTP коды для указанного email

**Параметры:**
- `email` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/otp_crud.go:92

---

#### GetOTPByCode

GetOTPByCode получает OTP код по самому коду

**Параметры:**
- `code` (string) *обязательный* - 
- `email` (string) *обязательный* - 
- `otpType` (string) *обязательный* - 

**Возвращает:**
- *OTPCode
- error

**Файл:** db/otp_crud.go:104

---

#### GetOTPByCodeAnyStatus

GetOTPByCodeAnyStatus получает OTP код по коду независимо от статуса used/expired

**Параметры:**
- `code` (string) *обязательный* - 
- `email` (string) *обязательный* - 
- `otpType` (string) *обязательный* - 

**Возвращает:**
- *OTPCode
- error

**Файл:** db/otp_crud.go:135

---

#### UpdateOTPExpiration

UpdateOTPExpiration обновляет время истечения OTP кода

**Параметры:**
- `id` (string) *обязательный* - 
- `expiresAt` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/otp_crud.go:166

---

#### CreateStatement

CreateStatement creates a new statement

**Параметры:**
- `statement` (*Statement) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/statement_crud.go:13

---

#### GetStatementByID

GetStatementByID retrieves a statement by ID

**Параметры:**
- `id` (string) *обязательный* - 

**Возвращает:**
- *Statement
- error

**Файл:** db/statement_crud.go:29

---

#### GetStatementsByUserID

GetStatementsByUserID retrieves all statements for a specific user

**Параметры:**
- `userID` (string) *обязательный* - 

**Возвращает:**
- []*Statement
- error

**Файл:** db/statement_crud.go:45

---

#### GetStatementsByCategoryID

GetStatementsByCategoryID retrieves all statements for a specific category

**Параметры:**
- `categoryID` (string) *обязательный* - 

**Возвращает:**
- []*Statement
- error

**Файл:** db/statement_crud.go:71

---

#### GetAllStatements

GetAllStatements retrieves all statements

**Возвращает:**
- []*Statement
- error

**Файл:** db/statement_crud.go:97

---

#### UpdateStatement

UpdateStatement updates an existing statement

**Параметры:**
- `statement` (*Statement) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/statement_crud.go:123

---

#### DeleteStatement

DeleteStatement deletes a statement by ID

**Параметры:**
- `id` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/statement_crud.go:149

---

#### DeleteStatementsByUserID

DeleteStatementsByUserID deletes all statements for a specific user

**Параметры:**
- `userID` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/statement_crud.go:170

---

#### DeleteStatementsByCategoryID

DeleteStatementsByCategoryID deletes all statements for a specific category

**Параметры:**
- `categoryID` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/statement_crud.go:191

---

#### StatementExists

StatementExists checks if a statement exists by ID

**Параметры:**
- `id` (string) *обязательный* - 

**Возвращает:**
- bool
- error

**Файл:** db/statement_crud.go:212

---

#### CreateUser

CreateUser creates a new user

**Параметры:**
- `user` (*User) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/user_crud.go:13

---

#### GetUserByID

GetUserByID retrieves a user by ID

**Параметры:**
- `id` (string) *обязательный* - 

**Возвращает:**
- *User
- error

**Файл:** db/user_crud.go:29

---

#### GetUserByEmail

GetUserByEmail retrieves a user by email

**Параметры:**
- `email` (string) *обязательный* - 

**Возвращает:**
- *User
- error

**Файл:** db/user_crud.go:45

---

#### GetAllUsers

GetAllUsers retrieves all users

**Возвращает:**
- []*User
- error

**Файл:** db/user_crud.go:61

---

#### UpdateUser

UpdateUser updates an existing user

**Параметры:**
- `user` (*User) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/user_crud.go:87

---

#### DeleteUser

DeleteUser deletes a user by ID

**Параметры:**
- `id` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/user_crud.go:113

---

#### UserExists

UserExists checks if a user exists by email

**Параметры:**
- `email` (string) *обязательный* - 

**Возвращает:**
- bool
- error

**Файл:** db/user_crud.go:134

---

#### VerifyUserEmail

VerifyUserEmail marks user email as verified

**Параметры:**
- `userID` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/user_crud.go:147

---

#### UpdateUserPassword

UpdateUserPassword updates user password

**Параметры:**
- `userID` (string) *обязательный* - 
- `newPassword` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** db/user_crud.go:169

---

### Структуры

#### CategoryCRUD

CategoryCRUD provides CRUD operations for Category entity

**Файл:** db/category_crud.go:10

---

#### MigrationLog

MigrationLog представляет запись о миграции

**Поля:**
- `ID` (int) - 
- `EntityType` (string) - 
- `EntityID` (string) - 
- `UserID` (string) - 
- `Action` (string) - 
- `Status` (string) - 
- `Error` (string) - 
- `CreatedAt` (time.Time) - 
- `UpdatedAt` (time.Time) - 

**Файл:** db/migration_tracker.go:10

---

#### MigrationTracker

MigrationTracker предоставляет методы для отслеживания миграций

**Файл:** db/migration_tracker.go:23

---

#### OTPCRUD

OTPCRUD предоставляет методы для работы с OTP кодами

**Файл:** db/otp_crud.go:10

---

#### StatementCRUD

StatementCRUD provides CRUD operations for Statement entity

**Файл:** db/statement_crud.go:10

---

#### Statement

**Поля:**
- `ID` (string) - 
- `Title` (string) - 
- `UserID` (string) - 
- `CategoryID` (string) - 

**Файл:** db/types.go:3

---

#### Category

**Поля:**
- `ID` (string) - 
- `Title` (string) - 
- `UserID` (string) - 

**Файл:** db/types.go:10

---

#### User

**Поля:**
- `ID` (string) - 
- `Email` (string) - 
- `Password` (string) - 
- `EmailVerified` (bool) - 
- `CreatedAt` (string) - 
- `UpdatedAt` (string) - 

**Файл:** db/types.go:16

---

#### OTPCode

OTPCode представляет OTP код в базе данных

**Поля:**
- `ID` (string) - 
- `Email` (string) - 
- `Code` (string) - 
- `Type` (string) - 
- `ExpiresAt` (string) - 
- `Used` (bool) - 
- `CreatedAt` (string) - 

**Файл:** db/types.go:26

---

#### UserCRUD

UserCRUD provides CRUD operations for User entity

**Файл:** db/user_crud.go:10

---

## handlers

### Функции

#### RegisterDirect

RegisterDirect обрабатывает прямую регистрацию пользователя без OTP

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/auth.go:67

---

#### Register

Register обрабатывает регистрацию пользователя

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/auth.go:135

---

#### Login

Login обрабатывает логин пользователя

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/auth.go:233

---

#### VerifyEmail

VerifyEmail обрабатывает верификацию email

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/auth.go:365

---

#### RequestPasswordReset

RequestPasswordReset обрабатывает запрос на сброс пароля

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/auth.go:456

---

#### VerifyPasswordResetOTP

VerifyPasswordResetOTP верифицирует OTP для сброса пароля

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/auth.go:513

---

#### ConfirmPasswordReset

ConfirmPasswordReset подтверждает сброс пароля

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/auth.go:569

---

#### GetStatements

GetStatements получает все statements пользователя

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/data.go:37

---

#### GetStatement

GetStatement получает конкретный statement

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/data.go:55

---

#### CreateStatement

CreateStatement создает новый statement

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/data.go:76

---

#### UpdateStatement

UpdateStatement обновляет statement

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/data.go:128

---

#### DeleteStatement

DeleteStatement удаляет statement

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/data.go:184

---

#### GetCategories

GetCategories получает все категории пользователя

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/data.go:208

---

#### GetCategory

GetCategory получает конкретную категорию

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/data.go:226

---

#### CreateCategory

CreateCategory создает новую категорию

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/data.go:247

---

#### UpdateCategory

UpdateCategory обновляет категорию

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/data.go:285

---

#### DeleteCategory

DeleteCategory удаляет категорию

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/data.go:327

---

#### InitWebSocketManager

InitWebSocketManager инициализирует WebSocket менеджер

**Файл:** handlers/websocket.go:17

---

#### HandleWebSocket

HandleWebSocket обрабатывает WebSocket подключения

**Параметры:**
- `c` (*gin.Context) *обязательный* - 

**Файл:** handlers/websocket.go:22

---

#### NotifyCategoryUpdate

NotifyCategoryUpdate отправляет уведомление об обновлении категории

**Параметры:**
- `userID` (string) *обязательный* - 
- `category` (interface{}) *обязательный* - 
- `action` (string) *обязательный* - 

**Файл:** handlers/websocket.go:34

---

#### NotifyCategoryCreated

NotifyCategoryCreated отправляет уведомление о создании категории

**Параметры:**
- `userID` (string) *обязательный* - 
- `category` (interface{}) *обязательный* - 

**Файл:** handlers/websocket.go:48

---

#### NotifyCategoryUpdated

NotifyCategoryUpdated отправляет уведомление об обновлении категории

**Параметры:**
- `userID` (string) *обязательный* - 
- `category` (interface{}) *обязательный* - 

**Файл:** handlers/websocket.go:53

---

#### NotifyCategoryDeleted

NotifyCategoryDeleted отправляет уведомление об удалении категории

**Параметры:**
- `userID` (string) *обязательный* - 
- `categoryID` (string) *обязательный* - 

**Файл:** handlers/websocket.go:58

---

#### NotifyStatementUpdate

NotifyStatementUpdate отправляет уведомление об обновлении statement

**Параметры:**
- `userID` (string) *обязательный* - 
- `statement` (interface{}) *обязательный* - 
- `action` (string) *обязательный* - 

**Файл:** handlers/websocket.go:69

---

#### NotifyStatementCreated

NotifyStatementCreated отправляет уведомление о создании statement

**Параметры:**
- `userID` (string) *обязательный* - 
- `statement` (interface{}) *обязательный* - 

**Файл:** handlers/websocket.go:83

---

#### NotifyStatementUpdated

NotifyStatementUpdated отправляет уведомление об обновлении statement

**Параметры:**
- `userID` (string) *обязательный* - 
- `statement` (interface{}) *обязательный* - 

**Файл:** handlers/websocket.go:88

---

#### NotifyStatementDeleted

NotifyStatementDeleted отправляет уведомление об удалении statement

**Параметры:**
- `userID` (string) *обязательный* - 
- `statementID` (string) *обязательный* - 

**Файл:** handlers/websocket.go:93

---

### Структуры

#### LoginRequest

LoginRequest структура для запроса логина

**Поля:**
- `Email` (string) - 
- `Password` (string) - 

**Файл:** handlers/auth.go:22

---

#### RegisterRequest

RegisterRequest структура для запроса регистрации

**Поля:**
- `Email` (string) - 
- `Password` (string) - 

**Файл:** handlers/auth.go:28

---

#### VerifyEmailRequest

VerifyEmailRequest структура для запроса верификации email

**Поля:**
- `Email` (string) - 
- `Code` (string) - 

**Файл:** handlers/auth.go:34

---

#### ResetPasswordRequest

ResetPasswordRequest структура для запроса сброса пароля

**Поля:**
- `Email` (string) - 

**Файл:** handlers/auth.go:40

---

#### ResetPasswordVerifyRequest

ResetPasswordVerifyRequest структура для верификации OTP при сбросе пароля

**Поля:**
- `Email` (string) - 
- `Code` (string) - 

**Файл:** handlers/auth.go:45

---

#### ResetPasswordConfirmRequest

ResetPasswordConfirmRequest структура для подтверждения нового пароля

**Поля:**
- `Email` (string) - 
- `Code` (string) - 
- `Password` (string) - 

**Файл:** handlers/auth.go:51

---

#### LoginResponse

LoginResponse структура для ответа при логине

**Поля:**
- `Token` (string) - 
- `User` (*ast.StructType) - 

**Файл:** handlers/auth.go:58

---

#### CreateStatementRequest

CreateStatementRequest структура для создания statement

**Поля:**
- `Title` (string) - 
- `CategoryID` (string) - 

**Файл:** handlers/data.go:15

---

#### UpdateStatementRequest

UpdateStatementRequest структура для обновления statement

**Поля:**
- `Title` (string) - 
- `CategoryID` (string) - 

**Файл:** handlers/data.go:21

---

#### CreateCategoryRequest

CreateCategoryRequest структура для создания категории

**Поля:**
- `Title` (string) - 

**Файл:** handlers/data.go:27

---

#### UpdateCategoryRequest

UpdateCategoryRequest структура для обновления категории

**Поля:**
- `Title` (string) - 

**Файл:** handlers/data.go:32

---

## utils

### Функции

#### GenerateID

GenerateID генерирует простой уникальный ID

**Возвращает:**
- string

**Файл:** utils/id.go:9

---

#### NewPasswordHasher

NewPasswordHasher создает новый экземпляр PasswordHasher

**Возвращает:**
- *PasswordHasher

**Файл:** utils/password.go:14

---

#### HashPassword

HashPassword хеширует пароль с использованием bcrypt

**Параметры:**
- `password` (string) *обязательный* - 

**Возвращает:**
- string
- error

**Файл:** utils/password.go:19

---

#### CheckPassword

CheckPassword проверяет пароль против хеша

**Параметры:**
- `hashedPassword` (string) *обязательный* - 
- `password` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** utils/password.go:30

---

#### GenerateRandomPassword

GenerateRandomPassword генерирует случайный пароль заданной длины

**Параметры:**
- `length` (int) *обязательный* - 

**Возвращает:**
- string
- error

**Файл:** utils/password.go:40

---

#### ValidatePasswordStrength

ValidatePasswordStrength проверяет сложность пароля

**Параметры:**
- `password` (string) *обязательный* - 

**Возвращает:**
- bool
- []string

**Файл:** utils/password.go:65

---

#### GetPasswordStrength

GetPasswordStrength возвращает оценку сложности пароля (0-100)

**Параметры:**
- `password` (string) *обязательный* - 

**Возвращает:**
- int

**Файл:** utils/password.go:119

---

#### GetPasswordStrengthText

GetPasswordStrengthText возвращает текстовое описание сложности пароля

**Параметры:**
- `strength` (int) *обязательный* - 

**Возвращает:**
- string

**Файл:** utils/password.go:173

---

#### TestPasswordHasher_HashPassword

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** utils/password_test.go:8

---

#### TestPasswordHasher_CheckPassword

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** utils/password_test.go:53

---

#### TestPasswordHasher_GenerateRandomPassword

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** utils/password_test.go:105

---

#### TestPasswordHasher_ValidatePasswordStrength

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** utils/password_test.go:184

---

#### TestPasswordHasher_GetPasswordStrength

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** utils/password_test.go:264

---

#### TestPasswordHasher_GetPasswordStrengthText

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** utils/password_test.go:316

---

#### BenchmarkPasswordHasher_HashPassword

Benchmark тесты для производительности

**Параметры:**
- `b` (*testing.B) *обязательный* - 

**Файл:** utils/password_test.go:362

---

#### BenchmarkPasswordHasher_CheckPassword

**Параметры:**
- `b` (*testing.B) *обязательный* - 

**Файл:** utils/password_test.go:375

---

#### BenchmarkPasswordHasher_GenerateRandomPassword

**Параметры:**
- `b` (*testing.B) *обязательный* - 

**Файл:** utils/password_test.go:392

---

### Структуры

#### PasswordHasher

PasswordHasher предоставляет методы для работы с паролями

**Файл:** utils/password.go:11

---

## websocket

### Функции

#### NewManager

NewManager создает новый WebSocket менеджер

**Возвращает:**
- *Manager

**Файл:** websocket/manager.go:39

---

#### HandleWebSocket

HandleWebSocket обрабатывает WebSocket подключения

**Параметры:**
- `w` (http.ResponseWriter) *обязательный* - 
- `r` (*http.Request) *обязательный* - 
- `userID` (string) *обязательный* - 

**Файл:** websocket/manager.go:52

---

#### registerClient

registerClient регистрирует клиента

**Параметры:**
- `client` (*Client) *обязательный* - 

**Файл:** websocket/manager.go:73

---

#### unregisterClient

unregisterClient удаляет клиента

**Параметры:**
- `client` (*Client) *обязательный* - 

**Файл:** websocket/manager.go:84

---

#### BroadcastToUser

BroadcastToUser отправляет сообщение всем клиентам пользователя

**Параметры:**
- `userID` (string) *обязательный* - 
- `messageType` (string) *обязательный* - 
- `payload` (interface{}) *обязательный* - 

**Файл:** websocket/manager.go:111

---

#### readPump

readPump читает сообщения от клиента

**Файл:** websocket/manager.go:148

---

#### writePump

writePump отправляет сообщения клиенту

**Файл:** websocket/manager.go:193

---

#### generateClientID

generateClientID генерирует уникальный ID для клиента

**Возвращает:**
- string

**Файл:** websocket/manager.go:239

---

### Структуры

#### Message

Message структура для WebSocket сообщений

**Поля:**
- `Type` (string) - 
- `Payload` (interface{}) - 
- `UserID` (string) - 

**Файл:** websocket/manager.go:15

---

#### Client

Client представляет WebSocket клиента

**Поля:**
- `ID` (string) - 
- `UserID` (string) - 
- `Conn` (*websocket.Conn) - 
- `Send` (*ast.ChanType) - 
- `Manager` (*Manager) - 

**Файл:** websocket/manager.go:22

---

#### Manager

Manager управляет WebSocket подключениями

**Поля:**
- `clients` (*ast.MapType) - 
- `userClients` (*ast.MapType) - 
- `mu` (sync.RWMutex) - 
- `upgrader` (websocket.Upgrader) - 

**Файл:** websocket/manager.go:31

---

## bl

### Функции

#### ImportCategories

ImportCategories импортирует категории пользователя из Firebase в PostgreSQL Поддерживает многократные запуски с инкрементальным обновлением

**Параметры:**
- `login` (string) *обязательный* - 
- `password` (string) *обязательный* - 

**Возвращает:**
- *ImportCategoriesResult
- error

**Файл:** bl/importCategories.go:33

---

#### determineAction

determineAction определяет действие на основе статуса миграции и существования категории

**Параметры:**
- `lastMigration` (*db.MigrationLog) *обязательный* - 
- `existsInPostgres` (bool) *обязательный* - 

**Возвращает:**
- string

**Файл:** bl/importCategories.go:153

---

#### importNewCategory

importNewCategory импортирует новую категорию

**Параметры:**
- `fbCategory` (*fb.Category) *обязательный* - 
- `categoryCRUD` (*db.CategoryCRUD) *обязательный* - 
- `migrationTracker` (*db.MigrationTracker) *обязательный* - 

**Возвращает:**
- error

**Файл:** bl/importCategories.go:181

---

#### updateExistingCategory

updateExistingCategory обновляет существующую категорию

**Параметры:**
- `fbCategory` (*fb.Category) *обязательный* - 
- `categoryCRUD` (*db.CategoryCRUD) *обязательный* - 
- `migrationTracker` (*db.MigrationTracker) *обязательный* - 

**Возвращает:**
- error

**Файл:** bl/importCategories.go:204

---

#### ImportCategoriesForAllUsers

ImportCategoriesForAllUsers импортирует категории для всех пользователей Полезно для массовой миграции

**Возвращает:**
- *ImportCategoriesResult
- error

**Файл:** bl/importCategories.go:228

---

#### GetImportStatus

GetImportStatus получает статус импорта для пользователя

**Параметры:**
- `userID` (string) *обязательный* - 

**Возвращает:**
- *ast.MapType
- error

**Файл:** bl/importCategories.go:235

---

#### ImportStatements

ImportStatements импортирует statements пользователя из Firebase в PostgreSQL Поддерживает многократные запуски с инкрементальным обновлением

**Параметры:**
- `login` (string) *обязательный* - 
- `password` (string) *обязательный* - 

**Возвращает:**
- *ImportStatementsResult
- error

**Файл:** bl/importStatements.go:26

---

#### importNewStatement

importNewStatement импортирует новое statement

**Параметры:**
- `fbStatement` (*fb.Statement) *обязательный* - 
- `statementCRUD` (*db.StatementCRUD) *обязательный* - 
- `migrationTracker` (*db.MigrationTracker) *обязательный* - 

**Возвращает:**
- error

**Файл:** bl/importStatements.go:160

---

#### updateExistingStatement

updateExistingStatement обновляет существующее statement

**Параметры:**
- `fbStatement` (*fb.Statement) *обязательный* - 
- `statementCRUD` (*db.StatementCRUD) *обязательный* - 
- `migrationTracker` (*db.MigrationTracker) *обязательный* - 

**Возвращает:**
- error

**Файл:** bl/importStatements.go:184

---

#### ImportAllData

ImportAllData импортирует пользователя, категории и statements

**Параметры:**
- `login` (string) *обязательный* - 
- `password` (string) *обязательный* - 

**Возвращает:**
- *ImportAllDataResult
- error

**Файл:** bl/importStatements.go:208

---

#### ImportUser

ImportUser импортирует пользователя и его категории из Firebase в PostgreSQL

**Параметры:**
- `login` (string) *обязательный* - 
- `password` (string) *обязательный* - 

**Возвращает:**
- error

**Файл:** bl/importUser.go:13

---

#### GetUser

**Параметры:**
- `email` (string) *обязательный* - 

**Возвращает:**
- *MockUser
- error

**Файл:** bl/import_test.go:23

---

#### CheckPassword

**Параметры:**
- `email` (string) *обязательный* - 
- `password` (string) *обязательный* - 

**Возвращает:**
- bool
- error

**Файл:** bl/import_test.go:30

---

#### GetCategories

**Параметры:**
- `user` (*MockUser) *обязательный* - 

**Возвращает:**
- []*fb.Category
- error

**Файл:** bl/import_test.go:38

---

#### TestImportCategories

TestImportCategories тестирует основную логику импорта

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** bl/import_test.go:46

---

#### TestMigrationTracker

TestMigrationTracker тестирует трекер миграций

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** bl/import_test.go:133

---

#### TestDetermineAction

TestDetermineAction тестирует логику определения действия

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** bl/import_test.go:180

---

#### clearTestData

**Параметры:**
- `t` (*testing.T) *обязательный* - 

**Файл:** bl/import_test.go:239

---

#### importCategoriesWithMock

**Параметры:**
- `_` (*MockFirebaseClient) *обязательный* - 
- `_` (string) *обязательный* - 
- `_` (string) *обязательный* - 

**Возвращает:**
- *ImportCategoriesResult
- error

**Файл:** bl/import_test.go:254

---

#### BenchmarkImportCategories

BenchmarkImportCategories тестирует производительность импорта

**Параметры:**
- `b` (*testing.B) *обязательный* - 

**Файл:** bl/import_test.go:262

---

#### clearTestDataForBenchmark

**Параметры:**
- `b` (*testing.B) *обязательный* - 

**Файл:** bl/import_test.go:283

---

### Структуры

#### ImportCategoriesResult

ImportCategoriesResult содержит результат импорта категорий

**Поля:**
- `TotalProcessed` (int) - 
- `Imported` (int) - 
- `Updated` (int) - 
- `Skipped` (int) - 
- `Failed` (int) - 
- `Errors` ([]ImportError) - 
- `Stats` (*ast.MapType) - 
- `Duration` (time.Duration) - 

**Файл:** bl/importCategories.go:13

---

#### ImportError

ImportError содержит информацию об ошибке импорта

**Поля:**
- `CategoryID` (string) - 
- `UserID` (string) - 
- `Error` (string) - 

**Файл:** bl/importCategories.go:25

---

#### ImportStatementsResult

ImportStatementsResult содержит результат импорта statements

**Поля:**
- `TotalProcessed` (int) - 
- `Imported` (int) - 
- `Updated` (int) - 
- `Skipped` (int) - 
- `Failed` (int) - 
- `Errors` ([]ImportError) - 
- `Stats` (*ast.MapType) - 
- `Duration` (time.Duration) - 

**Файл:** bl/importStatements.go:13

---

#### ImportAllDataResult

ImportAllDataResult содержит результат полного импорта данных

**Поля:**
- `StatementsResult` (*ImportStatementsResult) - 
- `Duration` (time.Duration) - 
- `StartTime` (time.Time) - 
- `EndTime` (time.Time) - 

**Файл:** bl/importStatements.go:255

---

#### MockFirebaseClient

MockFirebaseClient для тестирования

**Поля:**
- `users` (*ast.MapType) - 
- `categories` (*ast.MapType) - 

**Файл:** bl/import_test.go:13

---

#### MockUser

**Поля:**
- `UID` (string) - 
- `Email` (string) - 

**Файл:** bl/import_test.go:18

---

