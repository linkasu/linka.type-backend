package data

import (
	"net/http"
	"strings"

	"linka.type-backend/auth"
	"linka.type-backend/bl/services"

	"github.com/gin-gonic/gin"
)

// CreateEventRequest структура для создания события
type CreateEventRequest struct {
	Event string `json:"event" binding:"required"`
	Data  string `json:"data"`
}

// CreateEvent создает новое событие
func CreateEvent(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Дополнительная валидация
	if strings.TrimSpace(req.Event) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event cannot be empty"})
		return
	}

	eventService := services.NewEventService()
	event, err := eventService.CreateEvent(userID, req.Event, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}
