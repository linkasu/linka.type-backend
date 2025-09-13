package repositories

import (
	"linka.type-backend/models"
)

// EventCRUD provides CRUD operations for Event entity
// This is a wrapper around EventRepository for backward compatibility
type EventCRUD struct {
	repo *EventRepository
}

// NewEventCRUD creates a new EventCRUD
func NewEventCRUD() *EventCRUD {
	return &EventCRUD{
		repo: NewEventRepository(),
	}
}

// CreateEvent creates a new event
func (e *EventCRUD) CreateEvent(event *models.Event) error {
	return e.repo.CreateEvent(event)
}

// GetEventByID retrieves an event by ID
func (e *EventCRUD) GetEventByID(id string) (*models.Event, error) {
	return e.repo.GetEventByID(id)
}

// GetEventsByUserID retrieves all events for a specific user
func (e *EventCRUD) GetEventsByUserID(userID string) ([]*models.Event, error) {
	return e.repo.GetEventsByUserID(userID)
}

// GetAllEvents retrieves all events
func (e *EventCRUD) GetAllEvents() ([]*models.Event, error) {
	return e.repo.GetAllEvents()
}

// DeleteEvent deletes an event by ID
func (e *EventCRUD) DeleteEvent(id string) error {
	return e.repo.DeleteEvent(id)
}

// DeleteEventsByUserID deletes all events for a specific user
func (e *EventCRUD) DeleteEventsByUserID(userID string) error {
	return e.repo.DeleteEventsByUserID(userID)
}
