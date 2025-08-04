package utils

import (
	"fmt"
	"time"
)

// GenerateID генерирует простой уникальный ID
func GenerateID() string {
	// В реальном проекте лучше использовать UUID
	return fmt.Sprintf("id_%d", time.Now().UnixNano())
}
