package config

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"os"
)

// Config - общая конфигурация всего приложения для БД
type Config struct {
	DbDriverName   string `validate:"required"`
	Dsn            string `validate:"required"`
	AppName        string `validate:"required"` // Название приложения
	AppVersion     string `validate:"required"` // Версия приложения
	LogLevel       string
	LogDevelopMode bool
}

//GetConfig
/**
* GetConfig - получение конфигурации из .env файла или переменных окружения
* from _ = godotenv.Load(envFile) // где символ _ игнорирует возвращаемое значение
 */
func GetConfig(envFile string) Config {
	var err = godotenv.Load(envFile) // Тихая загрузка .env без логирования ошибок
	if err != nil {
		// если нет файла, то залогируем это и попробуем получить конфиг из переменных окружения
		log.Error("Error loading .env file: %v\n", zap.Error(err))
		//  логируем ошибку "файл не найден"
		if !os.IsNotExist(err) {
			log.Error("Error loading .env file: %v", err)
		}
	}
	// значения переменных окружения могут быть получены из .env файла или переменных окружения
	var cfg = Config{
		DbDriverName:   os.Getenv("DB_DRIVER_NAME"),
		Dsn:            os.Getenv("DB_DSN"),
		AppName:        os.Getenv("APP_NAME"),
		AppVersion:     os.Getenv("APP_VERSION"), //for example, see = .env file APP_VERSION
		LogLevel:       os.Getenv("LOG_LEVEL"),
		LogDevelopMode: os.Getenv("LOG_DEVELOP_MODE") == "true",
	}
	log.Infof("GetConfig DB dsn: %v", cfg.Dsn)

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
