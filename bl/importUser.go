package bl

import (
	"fmt"
	"log"

	"linka.type-backend/db"
	"linka.type-backend/fb"
	"linka.type-backend/utils"
)

// ImportUser импортирует пользователя и его категории из Firebase в PostgreSQL
func ImportUser(login, password string) error {
	// Получаем пользователя из Firebase
	user, err := fb.GetUser(login)
	if err != nil {
		return fmt.Errorf("failed to get user from Firebase: %v", err)
	}

	// Проверяем пароль
	_, err = fb.CheckPassword(login, password)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %v", err)
	}

	// Импортируем пользователя в PostgreSQL
	userCRUD := &db.UserCRUD{}

	// Хешируем пароль для безопасного хранения
	passwordHasher := utils.NewPasswordHasher()
	hashedPassword, err := passwordHasher.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	pgUser := &db.User{
		ID:       user.UID,
		Email:    user.Email,
		Password: hashedPassword,
	}

	// Проверяем, существует ли пользователь
	existingUser, err := userCRUD.GetUserByID(user.UID)
	if err != nil && err.Error() != "user not found" {
		return fmt.Errorf("failed to check existing user: %v", err)
	}

	if existingUser == nil {
		// Создаем нового пользователя
		err = userCRUD.CreateUser(pgUser)
		if err != nil {
			return fmt.Errorf("failed to create user: %v", err)
		}
		log.Printf("Successfully created user %s", user.UID)
	} else {
		// Обновляем существующего пользователя
		err = userCRUD.UpdateUser(pgUser)
		if err != nil {
			return fmt.Errorf("failed to update user: %v", err)
		}
		log.Printf("Successfully updated user %s", user.UID)
	}

	// Импортируем категории пользователя
	result, err := ImportCategories(login, password)
	if err != nil {
		return fmt.Errorf("failed to import categories: %v", err)
	}

	log.Printf("User import completed. Categories: %d imported, %d updated, %d skipped, %d failed",
		result.Imported, result.Updated, result.Skipped, result.Failed)

	return nil
}
