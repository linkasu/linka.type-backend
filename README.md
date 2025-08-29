# Linka Type Backend - PostgreSQL CRUD Implementation with JWT API

This project implements a complete CRUD (Create, Read, Update, Delete) system for PostgreSQL database using Go with Firebase data migration capabilities and a RESTful API with JWT authentication.

## Features

- **PostgreSQL Database**: Uses PostgreSQL 17 with Docker
- **Complete CRUD Operations**: Full CRUD support for Users, Categories, and Statements
- **RESTful API**: Gin-based HTTP API with JWT authentication
- **JWT Authentication**: Secure token-based authentication system
- **WebSocket Support**: Real-time notifications for category and statement updates
- **Firebase Migration System**: Robust migration system from Firebase to PostgreSQL
- **Incremental Import**: Supports multiple runs without data duplication
- **Migration Tracking**: Comprehensive logging and tracking of migration progress
- **Docker Environment**: Configured with Docker Compose for easy deployment
- **Environment Variables**: Uses Docker environment variables for database configuration
- **Documentation Generator**: Automatic documentation generation from code comments

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Categories Table
```sql
CREATE TABLE categories (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Migration Logs Table
```sql
CREATE TABLE migration_logs (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(entity_type, entity_id, user_id)
);
```

### Statements Table
```sql
CREATE TABLE statements (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    category_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);
```

## CI/CD Pipeline

Проект включает полный CI/CD pipeline с GitHub Actions:

### Workflows

1. **CI (`ci.yml`)**
   - Линтинг кода с golangci-lint
   - Unit тесты
   - Integration тесты с PostgreSQL
   - E2E тесты с Docker Compose
   - Сборка для разных платформ
   - Проверка безопасности
   - Покрытие кода тестами

2. **Tests (`test.yml`)**
   - Раздельные jobs для unit, integration и e2e тестов
   - Линтинг с автоисправлением
   - Параллельное выполнение тестов

3. **E2E Tests (`e2e.yml`)**
   - Отдельный workflow для E2E тестов
   - Использует существующие Docker E2E тесты
   - Полное тестирование API endpoints
   - Тестирование аутентификации и авторизации
   - Тестирование CRUD операций
   - Тестирование WebSocket соединений
   - Генерирует отчеты в JUnit формате

4. **Security (`security.yml`)**
   - Сканирование уязвимостей (govulncheck, gosec)
   - Проверка зависимостей
   - Сканирование Docker образов (Trivy)
   - Запускается еженедельно

5. **Deploy (`deploy.yml`)**
   - Автоматический деплой в staging при push в main
   - Ручной деплой в production
   - Тесты перед деплоем

### Локальное тестирование

```bash
# Запуск всех тестов
make test

# Только unit тесты
make test-unit

# Только integration тесты
make test-integration

# E2E тесты с Docker
make test-e2e

# E2E тесты без пересборки
make test-e2e-only

# Тесты с покрытием
make test-coverage

# Линтинг
make lint

# Линтинг с автоисправлением
make lint-fix
```

### E2E тесты

Проект включает готовые E2E тесты в `e2e-tests/`:

- **Categories Tests** - 15 тестов CRUD операций
- **Statements Tests** - 15 тестов CRUD операций  
- **WebSocket Tests** - 12 тестов real-time уведомлений
- **Authentication Tests** - тесты аутентификации
- **Integration Tests** - комплексные тесты

E2E тесты запускаются в Docker с полной изоляцией и тестируют:
- API endpoints
- JWT аутентификацию
- Авторизацию и изоляцию данных
- WebSocket соединения
- Real-time уведомления
- Обработку ошибок

### Переменные окружения для тестов

Все тесты используют безопасные переменные окружения:

```bash
JWT_SECRET=test-jwt-secret-for-ci
JWT_ISSUER=test-issuer
JWT_AUDIENCE=test-audience
TEST_MODE=true
```

## Environment Variables

The application uses the following environment variables (configured in `docker-compose.yml`):

- `POSTGRES_HOST`: Database host (default: `db`)
- `POSTGRES_PORT`: Database port (default: `5432`)
- `POSTGRES_USER`: Database user (default: `postgres`)
- `POSTGRES_PASSWORD`: Database password (default: `postgres`)
- `POSTGRES_DB`: Database name (default: `linkatype`)
- `MAIL_SERVER`: SMTP server (default: `smtp.gmail.com`)
- `MAIL_PORT`: SMTP port (default: `587`)
- `MAIL_ADRESS`: Email address for sending (default: `test@example.com`)
- `MAIL_PASSWORD`: Email password (default: `test-password`)
- `JWT_SECRET`: Secret key for JWT tokens (default: `test-secret-key-for-development`)
- `JWT_ISSUER`: JWT issuer claim (default: `linka-backend`)
- `JWT_AUDIENCE`: JWT audience claim (default: `linka-users`)
- `CORS_ORIGINS`: Comma-separated list of allowed origins (default: `http://localhost:3000,http://localhost:8080`)
- `CORS_METHODS`: Comma-separated list of allowed HTTP methods (default: `GET, POST, PUT, DELETE, OPTIONS`)
- `CORS_HEADERS`: Comma-separated list of allowed headers (default: `Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization`)

## Migration System

### Import Categories from Firebase

```go
// Import categories for a specific user
result, err := bl.ImportCategories("user@example.com", "password")
if err != nil {
    log.Fatalf("Import failed: %v", err)
}

fmt.Printf("Imported: %d, Updated: %d, Skipped: %d, Failed: %d\n",
    result.Imported, result.Updated, result.Skipped, result.Failed)
```

### Import User with Categories

```go
// Import user and all their categories
err := bl.ImportUser("user@example.com", "password")
if err != nil {
    log.Fatalf("User import failed: %v", err)
}
```

### Import Statements

```go
// Import statements for a specific user
result, err := bl.ImportStatements("user@example.com", "password")
if err != nil {
    log.Fatalf("Statements import failed: %v", err)
}

fmt.Printf("Statements: %d imported, %d updated, %d skipped, %d failed\n",
    result.Imported, result.Updated, result.Skipped, result.Failed)
```

### Import All Data (User, Categories, Statements)

```go
// Import all data for a user
result, err := bl.ImportAllData("user@example.com", "password")
if err != nil {
    log.Fatalf("Complete import failed: %v", err)
}

fmt.Printf("Complete import finished in %v\n", result.Duration)
```

### Migration Status

```go
// Get migration statistics
status, err := bl.GetImportStatus("user_id")
if err != nil {
    log.Printf("Failed to get status: %v", err)
} else {
    fmt.Printf("Migration stats: %+v\n", status)
}
```

## CRUD Operations

### User CRUD Operations

```go
userCRUD := &db.UserCRUD{}

// Create a user
user := &db.User{
    ID:       uuid.New().String(),
    Email:    "user@example.com",
    Password: "hashed_password",
}
err := userCRUD.CreateUser(user)

// Get user by email
user, err := userCRUD.GetUserByEmail("user@example.com")

// Get user by ID
user, err := userCRUD.GetUserByID(userID)

// Get all users
users, err := userCRUD.GetAllUsers()

// Update user
user.Email = "newemail@example.com"
err := userCRUD.UpdateUser(user)

// Delete user
err := userCRUD.DeleteUser(userID)

// Check if user exists
exists, err := userCRUD.UserExists("user@example.com")
```

### Category CRUD Operations

```go
categoryCRUD := &db.CategoryCRUD{}

// Create a category
category := &db.Category{
    ID:     uuid.New().String(),
    Title:  "Work",
    UserId: userID,
}
err := categoryCRUD.CreateCategory(category)

// Get category by ID
category, err := categoryCRUD.GetCategoryByID(categoryID)

// Get categories by user ID
categories, err := categoryCRUD.GetCategoriesByUserID(userID)

// Get all categories
categories, err := categoryCRUD.GetAllCategories()

// Update category
category.Title = "Updated Title"
err := categoryCRUD.UpdateCategory(category)

// Delete category
err := categoryCRUD.DeleteCategory(categoryID)

// Delete all categories for a user
err := categoryCRUD.DeleteCategoriesByUserID(userID)

// Check if category exists
exists, err := categoryCRUD.CategoryExists(categoryID)
```

### Statement CRUD Operations

```go
statementCRUD := &db.StatementCRUD{}

// Create a statement
statement := &db.Statement{
    ID:         uuid.New().String(),
    Title:      "Complete task",
    UserId:     userID,
    CategoryId: categoryID,
}
err := statementCRUD.CreateStatement(statement)

// Get statement by ID
statement, err := statementCRUD.GetStatementByID(statementID)

// Get statements by user ID
statements, err := statementCRUD.GetStatementsByUserID(userID)

// Get statements by category ID
statements, err := statementCRUD.GetStatementsByCategoryID(categoryID)

// Get all statements
statements, err := statementCRUD.GetAllStatements()

// Update statement
statement.Title = "Updated Task"
err := statementCRUD.UpdateStatement(statement)

// Delete statement
err := statementCRUD.DeleteStatement(statementID)

// Delete all statements for a user
err := statementCRUD.DeleteStatementsByUserID(userID)

// Delete all statements for a category
err := statementCRUD.DeleteStatementsByCategoryID(categoryID)

// Check if statement exists
exists, err := statementCRUD.StatementExists(statementID)
```

## Running the Application

### Using Docker Compose

1. Build and start the services:
```bash
docker compose up --build
```

2. The application will start:
   - **Playground** (port 8080): Data migration and CRUD examples
   - **API Server** (port 8081): RESTful API with JWT authentication
   - **PostgreSQL** database

### API Server

The API server provides a complete RESTful interface:

- **Base URL**: `http://localhost:8081/api`
- **Authentication**: JWT tokens required for protected endpoints
- **Documentation**: See `docs/api.md` for complete API documentation
- **Auto-generated Docs**: Run `make docs` to generate documentation from code

#### Quick Start with API

1. Register a new user:
```bash
curl -X POST http://localhost:8081/api/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'
```

2. Login to get a token:
```bash
curl -X POST http://localhost:8081/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'
```

3. Use the token for authenticated requests:
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8081/api/categories
```

4. Connect to WebSocket for real-time updates:
```javascript
const ws = new WebSocket('ws://localhost:8081/api/ws');
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Category update:', message);
};
```

### Manual Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL database with the schema

3. Set environment variables or use defaults

4. Run the playground (data migration):
```bash
go run cmd/playground/main.go
```

5. Run the API server:
```bash
go run cmd/server/main.go
```

## Testing

The project includes comprehensive testing:

### Unit Tests
```bash
make test-unit
```
Tests individual components without external dependencies.

### Integration Tests
```bash
make test-integration
```
Tests API endpoints and JWT functionality without database.

### All Tests
```bash
make test
```
Runs all tests (unit + integration).

## Documentation

The project includes automatic documentation generation:

### Generate Documentation
```bash
make docs
```
Generates documentation from code comments in multiple formats:
- `docs/generated.md` - Markdown format
- `docs/generated.html` - HTML format with styling
- `docs/generated.json` - JSON format for programmatic use

### View Documentation
```bash
make docs-serve
```
Starts a local server to view HTML documentation at http://localhost:8080

### Documentation Features
- **Auto-extraction**: Extracts comments, function signatures, and structures
- **API Documentation**: Complete API endpoints with parameters and responses
- **Code Documentation**: Functions, structs, and packages with descriptions
- **Multiple Formats**: Markdown, HTML, and JSON output
- **Tag Support**: Special tags like `@example`, `@deprecated`, `@version`

### Manual Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL database with the schema

3. Set environment variables or use defaults

4. Run the application:
```bash
go run cmd/playground/main.go
```

### Migration Examples

The main application (`cmd/playground/main.go`) demonstrates:
- User import with password hashing
- Category import with migration tracking
- Statement import with incremental updates
- Complete data migration from Firebase to PostgreSQL

## File Structure

```
linka.type-backend/
├── cmd/
│   ├── playground/
│   │   └── main.go              # Data migration playground
│   └── server/
│       └── main.go              # API server entry point
├── auth/
│   ├── jwt.go                   # JWT token generation and validation
│   └── middleware.go            # JWT authentication middleware
├── websocket/
│   └── manager.go               # WebSocket connection management
├── handlers/
│   ├── auth.go                  # Authentication handlers (login/register)
│   └── data.go                  # Data handlers (statements/categories)
├── utils/
│   ├── password.go              # Password hashing utilities
│   ├── password_test.go         # Password utilities tests
│   └── id.go                    # ID generation utilities
├── bl/
│   ├── importUser.go            # User import logic
│   ├── importCategories.go      # Category import system
│   ├── importStatements.go      # Statement import system
│   └── import_test.go           # Import system tests
├── db/
│   ├── types.go                 # Data models (User, Category, Statement)
│   ├── connection.go            # Database connection and table creation
│   ├── migration_tracker.go     # Migration tracking system
│   ├── user_crud.go            # User CRUD operations
│   ├── category_crud.go        # Category CRUD operations
│   └── statement_crud.go       # Statement CRUD operations
├── fb/                         # Firebase integration
├── docs/
│   ├── import_system.md        # Migration system documentation
│   ├── api.md                  # API documentation
│   ├── README.md               # Documentation guide
│   ├── generated.md            # Auto-generated Markdown docs
│   ├── generated.html          # Auto-generated HTML docs
│   └── generated.json          # Auto-generated JSON docs
├── examples/
│   └── websocket-demo.html     # WebSocket demo page
├── scripts/
│   ├── migrate.sh              # Database migration script
│   └── generate-docs.sh        # Documentation generation script
├── docker-compose.yml          # Docker services configuration
├── Dockerfile.playground       # Docker build for playground
├── Dockerfile.server           # Docker build for API server
├── go.mod                      # Go module dependencies
└── README.md                   # This file
```

## Password Security

The system includes a comprehensive password security module with the following features:

### Password Hashing
- Uses bcrypt algorithm for secure password hashing
- Automatic salt generation for each password
- Configurable cost factor for performance vs security balance

### Password Validation
- Minimum length requirements (8 characters)
- Complexity requirements (letters, digits, special characters)
- Strength scoring (0-100 scale)

### Password Utilities
```go
hasher := utils.NewPasswordHasher()

// Hash a password
hash, err := hasher.HashPassword("mypassword")

// Verify a password
err := hasher.CheckPassword(hash, "mypassword")

// Generate random password
password, err := hasher.GenerateRandomPassword(12)

// Validate password strength
isValid, errors := hasher.ValidatePasswordStrength("MyPass123!")

// Get password strength score
strength := hasher.GetPasswordStrength("MyPass123!")
```

## Dependencies

- `github.com/lib/pq`: PostgreSQL driver
- `github.com/gin-gonic/gin`: HTTP web framework
- `github.com/golang-jwt/jwt/v4`: JWT token handling
- `github.com/google/uuid`: UUID generation
- `github.com/joho/godotenv`: Environment variable loading
- `golang.org/x/crypto/bcrypt`: Password hashing

## Notes

- The application uses UUIDs for all primary keys
- Foreign key relationships are maintained with CASCADE delete
- Timestamps are automatically managed for created_at and updated_at
- All CRUD operations include proper error handling
- The example code demonstrates all available operations
- Delete operations are commented out in examples to preserve data 