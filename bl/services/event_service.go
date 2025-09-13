package services

import (
	"linka.type-backend/db/repositories"
	"linka.type-backend/models"
	"linka.type-backend/utils"
)

// EventService provides event business logic
type EventService struct {
	eventRepo *repositories.EventRepository
}

// NewEventService creates a new EventService
func NewEventService() *EventService {
	return &EventService{
		eventRepo: repositories.NewEventRepository(),
	}
}

// CreateEvent creates a new event
func (s *EventService) CreateEvent(userID, event, data string) (*models.Event, error) {
	eventID := utils.GenerateID()

	newEvent := &models.Event{
		ID:     eventID,
		UserID: userID,
		Event:  event,
		Data:   data,
	}

	err := s.eventRepo.CreateEvent(newEvent)
	if err != nil {
		return nil, err
	}

	return newEvent, nil
}

// GetEventByID gets an event by ID
func (s *EventService) GetEventByID(id string) (*models.Event, error) {
	return s.eventRepo.GetEventByID(id)
}

// GetEventsByUserID gets all events for a specific user
func (s *EventService) GetEventsByUserID(userID string) ([]*models.Event, error) {
	return s.eventRepo.GetEventsByUserID(userID)
}

// GetAllEvents gets all events
func (s *EventService) GetAllEvents() ([]*models.Event, error) {
	return s.eventRepo.GetAllEvents()
}

// DeleteEvent deletes an event by ID
func (s *EventService) DeleteEvent(id string) error {
	return s.eventRepo.DeleteEvent(id)
}

// DeleteEventsByUserID deletes all events for a specific user
func (s *EventService) DeleteEventsByUserID(userID string) error {
	return s.eventRepo.DeleteEventsByUserID(userID)
}
