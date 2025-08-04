package db

import (
	"database/sql"
	"fmt"
	"time"
)

// MigrationLog представляет запись о миграции
type MigrationLog struct {
	ID         int       `json:"id"`
	EntityType string    `json:"entityType"` // "category", "user", "statement"
	EntityID   string    `json:"entityId"`
	UserID     string    `json:"userId"`
	Action     string    `json:"action"` // "import", "update", "delete"
	Status     string    `json:"status"` // "success", "failed", "skipped"
	Error      string    `json:"error"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// MigrationTracker предоставляет методы для отслеживания миграций
type MigrationTracker struct{}

// CreateMigrationLogsTable создает таблицу для отслеживания миграций
func (mt *MigrationTracker) CreateMigrationLogsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS migration_logs (
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
	);`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating migration_logs table: %v", err)
	}

	return nil
}

// LogMigration записывает информацию о миграции
func (mt *MigrationTracker) LogMigration(entityType, entityID, userID, action, status, errorMsg string) error {
	query := `
	INSERT INTO migration_logs (entity_type, entity_id, user_id, action, status, error, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (entity_type, entity_id, user_id) 
	DO UPDATE SET 
		action = EXCLUDED.action,
		status = EXCLUDED.status,
		error = EXCLUDED.error,
		updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := DB.Exec(query, entityType, entityID, userID, action, status, errorMsg, now)
	if err != nil {
		return fmt.Errorf("error logging migration: %v", err)
	}

	return nil
}

// GetLastMigrationStatus получает статус последней миграции для сущности
func (mt *MigrationTracker) GetLastMigrationStatus(entityType, entityID, userID string) (*MigrationLog, error) {
	query := `
	SELECT id, entity_type, entity_id, user_id, action, status, error, created_at, updated_at
	FROM migration_logs 
	WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3
	ORDER BY updated_at DESC 
	LIMIT 1
	`

	var log MigrationLog
	err := DB.QueryRow(query, entityType, entityID, userID).Scan(
		&log.ID, &log.EntityType, &log.EntityID, &log.UserID,
		&log.Action, &log.Status, &log.Error, &log.CreatedAt, &log.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Нет записи о миграции
		}
		return nil, fmt.Errorf("error getting migration status: %v", err)
	}

	return &log, nil
}

// GetMigrationStats получает статистику миграций
func (mt *MigrationTracker) GetMigrationStats(entityType string) (map[string]int, error) {
	query := `
	SELECT status, COUNT(*) as count
	FROM migration_logs 
	WHERE entity_type = $1
	GROUP BY status
	`

	rows, err := DB.Query(query, entityType)
	if err != nil {
		return nil, fmt.Errorf("error getting migration stats: %v", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("error scanning migration stats: %v", err)
		}
		stats[status] = count
	}

	return stats, nil
}

// ClearMigrationLogs очищает логи миграций (для тестирования)
func (mt *MigrationTracker) ClearMigrationLogs(entityType string) error {
	query := `DELETE FROM migration_logs WHERE entity_type = $1`

	_, err := DB.Exec(query, entityType)
	if err != nil {
		return fmt.Errorf("error clearing migration logs: %v", err)
	}

	return nil
}
