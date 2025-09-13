package auth

import (
	"log"
	"net/http"

	"linka.type-backend/bl"
	"linka.type-backend/db"
	"linka.type-backend/fb"

	"github.com/gin-gonic/gin"
)

// importFirebaseData импортирует данные из Firebase
func importFirebaseData(email, password string, user *db.User, userCRUD *db.UserCRUD, c *gin.Context) {
	fbUser, err := fb.GetUser(email)
	if err == nil && fbUser != nil {
		importResult, err := bl.ImportAllData(email, password)
		if err != nil {
			log.Printf("Failed to import data from Firebase: %v", err)
		} else {
			log.Printf("Successfully imported data from Firebase: %+v", importResult)
		}

		if user == nil {
			user, err = userCRUD.GetUserByEmail(email)
			if err != nil {
				log.Printf("Failed to get imported user from PostgreSQL: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user after import"})
				return
			}
			log.Printf("Successfully retrieved imported user: %s", user.ID)
		}
	}
}