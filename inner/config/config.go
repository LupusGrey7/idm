package config

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"log"
	"os"
)

// Config - общая конфигурация всего приложения для БД
type Config struct {
	DbDriverName string `validate:"required"`
	Dsn          string `validate:"required"`
	AppName      string `validate:"required"` // Название приложения
	AppVersion   string `validate:"required"` // Версия приложения
}

//GetConfig
/**
* GetConfig - получение конфигурации из .env файла или переменных окружения
* from _ = godotenv.Load(envFile) // где символ _ игнорирует возвращаемое значение
 */
func GetConfig(envFile string) Config {
	var err = godotenv.Load(envFile) // где символ _ игнорирует возвращаемое значение
	// Тихая загрузка .env без логирования ошибок
	if err != nil {
		// если нет файла, то залогируем это и попробуем получить конфиг из переменных окружения
		fmt.Printf("Error loading .env file: %v\n", err)
		// Не логируем ошибку "файл не найден"
		if !os.IsNotExist(err) {
			log.Printf("Error loading .env file: %v", err)
		}
	}
	var cfg = Config{ // значения переменных окружения могут быть получены из .env файла или переменных окружения
		DbDriverName: os.Getenv("DB_DRIVER_NAME"),
		Dsn:          os.Getenv("DB_DSN"),
		AppName:      os.Getenv("APP_NAME"),
		AppVersion:   os.Getenv("APP_VERSION"), //for example, see = .env file APP_VERSION
	}
	err = validator.New().Struct(cfg)
	if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			// если конфиг не прошел валидацию, то паникуем
			panic(fmt.Sprintf("config validation error: %v", err))
		}
	}
	return cfg
}
