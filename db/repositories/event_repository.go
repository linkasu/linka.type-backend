package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"linka.type-backend/db"
	"linka.type-backend/models"
)

// EventRepository provides CRUD operations for Event entity
type EventRepository struct{}

// NewEventRepository creates a new EventRepository
func NewEventRepository() *EventRepository {
	return &EventRepository{}
}

// CreateEvent creates a new event
func (e *EventRepository) CreateEvent(event *models.Event) error {
	query := `
		INSERT INTO events (id, user_id, event, data, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()
	_, err := db.DB.Exec(query, event.ID, event.UserID, event.Event, event.Data, now)
	if err != nil {
		return fmt.Errorf("error creating event: %v", err)
	}

	return nil
}

// GetEventByID retrieves an event by ID
func (e *EventRepository) GetEventByID(id string) (*models.Event, error) {
	query := `SELECT id, user_id, event, data, created_at FROM events WHERE id = $1`

	var event models.Event
	err := db.DB.QueryRow(query, id).Scan(&event.ID, &event.UserID, &event.Event, &event.Data, &event.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("error getting event: %v", err)
	}

	return &event, nil
}

// GetEventsByUserID retrieves all events for a specific user
func (e *EventRepository) GetEventsByUserID(userID string) ([]*models.Event, error) {
	query := `SELECT id, user_id, event, data, created_at FROM events WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %v", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.UserID, &event.Event, &event.Data, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning event: %v", err)
		}
		events = append(events, &event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %v", err)
	}

	return events, nil
}

// GetAllEvents retrieves all events
func (e *EventRepository) GetAllEvents() ([]*models.Event, error) {
	query := `SELECT id, user_id, event, data, created_at FROM events ORDER BY created_at DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %v", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.UserID, &event.Event, &event.Data, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning event: %v", err)
		}
		events = append(events, &event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %v", err)
	}

	return events, nil
}

// DeleteEvent deletes an event by ID
func (e *EventRepository) DeleteEvent(id string) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := db.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting event: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

// DeleteEventsByUserID deletes all events for a specific user
func (e *EventRepository) DeleteEventsByUserID(userID string) error {
	query := `DELETE FROM events WHERE user_id = $1`

	_, err := db.DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("error deleting events: %v", err)
	}

	return nil
}
