package common

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

// Config общая конфигурация всего приложения
type Config struct {
	DbDriverName string `validate:"required"`
	Dsn          string `validate:"required"`
}

//GetConfig
/**
* GetConfig - получение конфигурации из .env файла или переменных окружения
* _ = godotenv.Load(envFile) // где символ _ игнорирует возвращаемое значение
 */
func GetConfig(envFile string) Config {
	// Тихая загрузка .env без логирования ошибок
	if err := godotenv.Load(envFile); err != nil {
		// Не логируем ошибку "файл не найден"
		if !os.IsNotExist(err) {
			log.Printf("Error loading .env file: %v", err)
		}
	}
	var cfg = Config{ // значения переменных окружения могут быть получены
		DbDriverName: os.Getenv("DB_DRIVER_NAME"),
		Dsn:          os.Getenv("DB_DSN"),
	}
	return cfg
}
