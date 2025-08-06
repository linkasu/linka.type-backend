package bl

import (
	"fmt"
	"testing"

	"linka.type-backend/db"
	"linka.type-backend/fb"
)

// MockFirebaseClient для тестирования
type MockFirebaseClient struct {
	users      map[string]*MockUser
	categories map[string][]*fb.FBCategory
}

type MockUser struct {
	UID   string `json:"uid"`
	Email string `json:"email"`
}

func (m *MockFirebaseClient) GetUser(email string) (*MockUser, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockFirebaseClient) CheckPassword(email, password string) (bool, error) {
	if _, exists := m.users[email]; exists {
		// Простая проверка для тестов
		return password == "correct_password", nil
	}
	return false, fmt.Errorf("authentication failed")
}

func (m *MockFirebaseClient) GetCategories(user *MockUser) ([]*fb.FBCategory, error) {
	if categories, exists := m.categories[user.UID]; exists {
		return categories, nil
	}
	return []*fb.FBCategory{}, nil
}

// TestImportCategories тестирует основную логику импорта
func TestImportCategories(t *testing.T) {
	// Настройка тестовой базы данных
	if err := db.InitDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
	}()

	// Очищаем таблицы перед тестом
	clearTestData(t)

	// Создаем тестовые данные
	mockClient := &MockFirebaseClient{
		users: map[string]*MockUser{
			"test@example.com": {
				UID:   "user123",
				Email: "test@example.com",
			},
		},
		categories: map[string][]*fb.FBCategory{
			"user123": {
				{ID: "cat1", Label: "Category 1", UserId: "user123"},
				{ID: "cat2", Label: "Category 2", UserId: "user123"},
			},
		},
	}

	// Тест 1: Первый импорт
	t.Run("First Import", func(t *testing.T) {
		result, err := importCategoriesWithMock(mockClient, "test@example.com", "correct_password")
		if err != nil {
			t.Fatalf("Import failed: %v", err)
		}

		expectedImported := 2
		if result.Imported != expectedImported {
			t.Errorf("Expected %d imported categories, got %d", expectedImported, result.Imported)
		}

		if result.Failed > 0 {
			t.Errorf("Expected no failed imports, got %d", result.Failed)
		}
	})

	// Тест 2: Повторный импорт (должен пропустить существующие)
	t.Run("Second Import", func(t *testing.T) {
		result, err := importCategoriesWithMock(mockClient, "test@example.com", "correct_password")
		if err != nil {
			t.Fatalf("Second import failed: %v", err)
		}

		expectedSkipped := 2
		if result.Skipped != expectedSkipped {
			t.Errorf("Expected %d skipped categories, got %d", expectedSkipped, result.Skipped)
		}

		if result.Imported > 0 {
			t.Errorf("Expected no new imports, got %d", result.Imported)
		}
	})

	// Тест 3: Импорт с новой категорией
	t.Run("Import with new category", func(t *testing.T) {
		// Добавляем новую категорию
		mockClient.categories["user123"] = append(mockClient.categories["user123"],
			&fb.FBCategory{ID: "cat3", Label: "Category 3", UserId: "user123"})

		result, err := importCategoriesWithMock(mockClient, "test@example.com", "correct_password")
		if err != nil {
			t.Fatalf("Import with new category failed: %v", err)
		}

		expectedImported := 1
		expectedSkipped := 2
		if result.Imported != expectedImported {
			t.Errorf("Expected %d imported categories, got %d", expectedImported, result.Imported)
		}
		if result.Skipped != expectedSkipped {
			t.Errorf("Expected %d skipped categories, got %d", expectedSkipped, result.Skipped)
		}
	})
}

// TestMigrationTracker тестирует трекер миграций
func TestMigrationTracker(t *testing.T) {
	if err := db.InitDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
	}()

	tracker := &db.MigrationTracker{}

	t.Run("Log Migration", func(t *testing.T) {
		err := tracker.LogMigration("category", "test_id", "user123", "import", "success", "")
		if err != nil {
			t.Fatalf("Failed to log migration: %v", err)
		}
	})

	t.Run("Get Migration Status", func(t *testing.T) {
		status, err := tracker.GetLastMigrationStatus("category", "test_id", "user123")
		if err != nil {
			t.Fatalf("Failed to get migration status: %v", err)
		}

		if status == nil {
			t.Fatal("Expected migration status, got nil")
		}

		if status.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", status.Status)
		}
	})

	t.Run("Get Migration Stats", func(t *testing.T) {
		stats, err := tracker.GetMigrationStats("category")
		if err != nil {
			t.Fatalf("Failed to get migration stats: %v", err)
		}

		if stats["success"] < 1 {
			t.Errorf("Expected at least 1 successful migration, got %d", stats["success"])
		}
	})
}

// TestDetermineAction тестирует логику определения действия
func TestDetermineAction(t *testing.T) {
	tests := []struct {
		name           string
		lastMigration  *db.MigrationLog
		existsInPG     bool
		expectedAction string
	}{
		{
			name:           "No migration record",
			lastMigration:  nil,
			existsInPG:     false,
			expectedAction: "import",
		},
		{
			name: "Successful migration, exists in PG",
			lastMigration: &db.MigrationLog{
				Status: "success",
			},
			existsInPG:     true,
			expectedAction: "skip",
		},
		{
			name: "Successful migration, not exists in PG",
			lastMigration: &db.MigrationLog{
				Status: "success",
			},
			existsInPG:     false,
			expectedAction: "import",
		},
		{
			name: "Failed migration, exists in PG",
			lastMigration: &db.MigrationLog{
				Status: "failed",
			},
			existsInPG:     true,
			expectedAction: "update",
		},
		{
			name: "Failed migration, not exists in PG",
			lastMigration: &db.MigrationLog{
				Status: "failed",
			},
			existsInPG:     false,
			expectedAction: "import",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := determineAction(tt.lastMigration, tt.existsInPG)
			if action != tt.expectedAction {
				t.Errorf("Expected action '%s', got '%s'", tt.expectedAction, action)
			}
		})
	}
}

// Вспомогательные функции

func clearTestData(t *testing.T) {
	// Очищаем таблицы для чистого теста
	queries := []string{
		"DELETE FROM migration_logs",
		"DELETE FROM categories",
		"DELETE FROM users",
	}

	for _, query := range queries {
		if _, err := db.DB.Exec(query); err != nil {
			t.Fatalf("Failed to clear test data: %v", err)
		}
	}
}

func importCategoriesWithMock(mockClient *MockFirebaseClient, email, password string) (*ImportCategoriesResult, error) {
	// Эта функция должна быть реализована для работы с mock клиентом
	// В реальной реализации нужно будет модифицировать ImportCategories
	// для поддержки dependency injection
	return nil, fmt.Errorf("not implemented")
}

// BenchmarkImportCategories тестирует производительность импорта
func BenchmarkImportCategories(b *testing.B) {
	if err := db.InitDB(); err != nil {
		b.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.CloseDB()

	// Подготовка тестовых данных
	clearTestDataForBenchmark(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Здесь должен быть вызов ImportCategories с тестовыми данными
		// Реализация зависит от конкретных требований к бенчмарку
	}
}

func clearTestDataForBenchmark(b *testing.B) {
	queries := []string{
		"DELETE FROM migration_logs",
		"DELETE FROM categories",
		"DELETE FROM users",
	}

	for _, query := range queries {
		if _, err := db.DB.Exec(query); err != nil {
			b.Fatalf("Failed to clear test data: %v", err)
		}
	}
}
