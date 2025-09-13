package fb

import (
	"linka.type-backend/fb/auth"
	"linka.type-backend/fb/data"
	"firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
)

// Re-export functions from subpackages for backward compatibility

// CheckPassword проверяет пароль в Firebase
func CheckPassword(email, password string) (*auth.FirebaseAuthResponse, error) {
	return auth.CheckPassword(email, password)
}

// GetUser получает пользователя из Firebase по email
func GetUser(email string) (*fbauth.UserRecord, error) {
	return auth.GetUser(email)
}

// GetFirebaseApp возвращает экземпляр Firebase приложения
func GetFirebaseApp() *firebase.App {
	return auth.GetFirebaseApp()
}

// IsFirebaseInitialized проверяет, инициализирован ли Firebase
func IsFirebaseInitialized() bool {
	return auth.IsFirebaseInitialized()
}

// GetCategories получает категории пользователя из Firebase
func GetCategories(user *fbauth.UserRecord) ([]*data.Category, error) {
	return data.GetCategories(user)
}

// Category представляет категорию в Firebase
type Category = data.Category

// Statement представляет statement в Firebase
type Statement = data.Statement