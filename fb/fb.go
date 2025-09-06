package fb

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var fb *firebase.App

func init() {
	var err error
	conf := &firebase.Config{
		DatabaseURL: "https://distypepro-android.firebaseio.com",
	}
	
	// Определяем путь к firebase.json
	firebasePath := os.Getenv("FIREBASE_CONFIG_PATH")
	if firebasePath == "" {
		// Проверяем разные возможные пути
		possiblePaths := []string{"/app/firebase.json", "./firebase.json", "firebase.json"}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				firebasePath = path
				break
			}
		}
	}
	
	// Если файл не найден, логируем предупреждение, но не падаем
	if firebasePath == "" {
		log.Printf("Warning: firebase.json file not found, Firebase features will be disabled")
		return
	}
	
	opt := option.WithCredentialsFile(firebasePath)
	fb, err = firebase.NewApp(context.Background(), conf, opt)
	if err != nil {
		log.Printf("Warning: error initializing Firebase app: %v", err)
	}
	log.Println("Successfully initialized Firebase app")
}

// GetFirebaseApp возвращает экземпляр Firebase приложения
func GetFirebaseApp() *firebase.App {
	return fb
}

// IsFirebaseInitialized проверяет, инициализирован ли Firebase
func IsFirebaseInitialized() bool {
	return fb != nil
}
