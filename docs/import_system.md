# Система импорта данных из Firebase в PostgreSQL

## Обзор

Система обеспечивает безопасную миграцию данных из Firebase в PostgreSQL с поддержкой многократных запусков. Система автоматически отслеживает прогресс миграции и предотвращает дублирование данных.

## Порядок импорта

### 1. Инициализация базы данных
```go
// Создание таблиц в PostgreSQL
- users (пользователи)
- categories (категории) 
- statements (высказывания)
- migration_logs (логи миграций)
```

### 2. Импорт пользователя
```go
// Шаг 1: Аутентификация в Firebase
user, err := fb.GetUser(login)
_, err = fb.CheckPassword(login, password)

// Шаг 2: Хеширование пароля
hashedPassword, err := passwordHasher.HashPassword(password)

// Шаг 3: Создание/обновление пользователя в PostgreSQL
pgUser := &db.User{
    ID:       user.UID,
    Email:    user.Email,
    Password: hashedPassword,
}
```

### 3. Импорт категорий
```go
// Шаг 1: Получение категорий из Firebase
fbCategories, err := fb.GetCategories(user)

// Шаг 2: Обработка каждой категории
for _, fbCategory := range fbCategories {
    // Проверка статуса миграции
    lastMigration, err := migrationTracker.GetLastMigrationStatus("category", fbCategory.ID, user.UID)
    
    // Определение действия
    action := determineAction(lastMigration, existingCategory != nil)
    
    switch action {
    case "import":
        // Создание новой категории
    case "update":
        // Обновление существующей категории
    case "skip":
        // Пропуск (уже актуальна)
    }
}
```

### 4. Импорт statements
```go
// Шаг 1: Получение категорий пользователя
fbCategories, err := fb.GetCategories(user)

// Шаг 2: Обработка statements для каждой категории
for _, fbCategory := range fbCategories {
    // Получение statements из Firebase
    fbStatements, err := fbCategory.GetStatements()
    
    // Обработка каждого statement
    for _, fbStatement := range fbStatements {
        // Установка правильного UserId
        fbStatement.UserId = fbCategory.UserId
        
        // Проверка статуса миграции
        lastMigration, err := migrationTracker.GetLastMigrationStatus("statement", fbStatement.ID, user.UID)
        
        // Определение действия
        action := determineAction(lastMigration, existingStatement != nil)
        
        switch action {
        case "import":
            // Создание нового statement
        case "update":
            // Обновление существующего statement
        case "skip":
            // Пропуск (уже актуально)
        }
    }
}
```

### 5. Логирование миграций
```go
// Каждая операция логируется в migration_logs
migrationTracker.LogMigration(entityType, entityID, userID, action, status, errorMsg)
```

## Алгоритм определения действий

| Статус миграции | Существует в PG | Действие |
|-----------------|-----------------|----------|
| Нет записи | - | import |
| success | true | skip |
| success | false | import |
| failed | true | update |
| failed | false | import |

## Архитектура

### Компоненты системы

1. **MigrationTracker** (`db/migration_tracker.go`)
   - Отслеживает статус миграции всех сущностей
   - Хранит логи миграций в таблице `migration_logs`
   - Предоставляет статистику миграций

2. **ImportUser** (`bl/importUser.go`)
   - Импортирует пользователя с хешированием пароля
   - Создает/обновляет пользователя в PostgreSQL

3. **ImportCategories** (`bl/importCategories.go`)
   - Импортирует категории пользователя
   - Поддерживает инкрементальные обновления

4. **ImportStatements** (`bl/importStatements.go`)
   - Импортирует statements для каждой категории
   - Обрабатывает иерархическую структуру данных

5. **ImportAllData** (`bl/importStatements.go`)
   - Выполняет полный импорт всех данных
   - Обеспечивает правильный порядок операций

### Таблица миграций

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

## Логика миграции

### Определение действия

Система определяет необходимое действие на основе:
1. **Статуса последней миграции** (success/failed)
2. **Существования сущности в PostgreSQL**

Применяется ко всем типам данных: пользователи, категории, statements.

### Обработка ошибок

- **Ошибки аутентификации**: Прерывают весь процесс
- **Ошибки получения данных**: Логируются, процесс продолжается
- **Ошибки базы данных**: Логируются в `migration_logs`, сущность помечается как failed
- **Ошибки внешних ключей**: Проверяется существование связанных сущностей

## Использование

### Полный импорт всех данных (рекомендуется)

```go
result, err := bl.ImportAllData("user@example.com", "password")
if err != nil {
    log.Fatalf("Complete import failed: %v", err)
}

fmt.Printf("Complete import finished in %v\n", result.Duration)
if result.StatementsResult != nil {
    fmt.Printf("Statements: %d imported, %d updated, %d skipped, %d failed\n",
        result.StatementsResult.Imported,
        result.StatementsResult.Updated,
        result.StatementsResult.Skipped,
        result.StatementsResult.Failed)
}
```

### Отдельные импорты

```go
// Импорт только пользователя
err := bl.ImportUser("user@example.com", "password")

// Импорт только категорий
result, err := bl.ImportCategories("user@example.com", "password")

// Импорт только statements
result, err := bl.ImportStatements("user@example.com", "password")
```

```go
// Импорт всех данных для пользователя
result, err := bl.ImportAllData("user@example.com", "password")
if err != nil {
    log.Fatalf("Complete import failed: %v", err)
}

fmt.Printf("Complete import finished in %v\n", result.Duration)
```

### Получение статуса импорта

```go
status, err := bl.GetImportStatus("user_id")
if err != nil {
    log.Printf("Failed to get status: %v", err)
} else {
    fmt.Printf("Migration stats: %+v\n", status)
}
```

## Примеры использования

### Запуск приложения

```bash
# Запуск основного приложения (импорт всех данных)
docker compose run --rm playground

# Или полный запуск
docker compose up --build
```

## Мониторинг и отладка

### Логи

Система выводит подробные логи:
- Количество найденных категорий и statements
- Статус обработки каждого элемента
- Итоговая статистика импорта
- Время выполнения операций

### Статистика миграций

```go
stats, err := migrationTracker.GetMigrationStats("category")
// Возвращает map[string]int с количеством по статусам
```

## Безопасность

### Аутентификация

- Проверка учетных данных через Firebase Auth
- Валидация пользователя перед импортом
- Безопасное хеширование паролей с bcrypt

### Целостность данных

- Уникальные ограничения в таблице миграций
- Транзакционная обработка ошибок
- Проверка существования данных перед импортом
- Проверка внешних ключей для statements

## Производительность

### Оптимизации

1. **Инкрементальный импорт**: Импортируются только новые/измененные данные
2. **Пропуск существующих**: Сущности с успешной миграцией пропускаются
3. **Логирование статуса**: Избегает повторной обработки
4. **Иерархическая обработка**: Statements импортируются после категорий

### Масштабирование

Для больших объемов данных рекомендуется:
1. Параллельная обработка пользователей
2. Пакетная обработка категорий и statements
3. Настройка таймаутов для Firebase API

## Расширение системы

### Добавление новых типов сущностей

1. Добавить новый `entity_type` в `migration_logs`
2. Создать соответствующие CRUD операции
3. Реализовать логику импорта в `bl` пакете
4. Обновить `ImportAllData` для включения новой сущности

### Интеграция с API

```go
// Пример HTTP endpoint для полного импорта
func ImportAllDataHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    result, err := bl.ImportAllData(req.Email, req.Password)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(result)
}
```

## Устранение неполадок

### Частые проблемы

1. **Ошибки подключения к Firebase**
   - Проверьте конфигурацию Firebase
   - Убедитесь в правильности учетных данных

2. **Ошибки базы данных**
   - Проверьте подключение к PostgreSQL
   - Убедитесь в существовании таблиц

3. **Ошибки внешних ключей**
   - Убедитесь, что пользователь импортирован перед statements
   - Проверьте существование категорий перед импортом statements

4. **Дублирование данных**
   - Проверьте логи миграций
   - Очистите таблицу `migration_logs` при необходимости

### Отладка

```go
// Включение подробных логов
log.SetLevel(log.DebugLevel)

// Проверка статуса миграции
status, err := migrationTracker.GetLastMigrationStatus("statement", "statement_id", "user_id")
if err != nil {
    log.Printf("Error: %v", err)
} else if status != nil {
    log.Printf("Last migration: %+v", status)
}
```

## Планы развития

1. **Параллельная обработка**: Поддержка одновременного импорта нескольких пользователей
2. **Веб-интерфейс**: Dashboard для мониторинга миграций
3. **Уведомления**: Email/Slack уведомления о завершении миграций
4. **Откат изменений**: Возможность отката неудачных миграций
5. **Валидация данных**: Проверка целостности данных перед импортом
6. **Инкрементальная синхронизация**: Автоматическая синхронизация изменений 