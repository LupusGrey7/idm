package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

/**
* Config_test.go
* os - стандартный пакет Go для работы с ОС
* - Стандартный пакет Go для работы с операционной системой
* - Содержит функции для работы с файлами, переменными окружения и т.д.
* t - контекст теста для привязки проверки
* require - останавливает тест сразу при ошибке
* assert - проверяет условия, но не останавливает тест при ошибке
 */
func TestGetConfig(t *testing.T) {
	// Сохраняем оригинальные переменные окружения
	originalEnv := make(map[string]string)
	for _, key := range []string{"DB_DRIVER_NAME", "DB_DSN", "APP_NAME", "APP_VERSION"} {
		if val, exists := os.LookupEnv(key); exists {
			originalEnv[key] = val
		}
	}

	// Восстановление окружения после тестов
	defer func() {
		for key, val := range originalEnv {
			err := os.Setenv(key, val)
			if err != nil {
				t.Errorf("failed to restore env var %s: %v", key, err)
			}
		}
	}()

	t.Run("Valid config from .env file", func(t *testing.T) {
		// Подготовка тестового .env файла
		envContent := `
DB_DRIVER_NAME=postgres
DB_DSN=host=localhost user=test dbname=test
APP_NAME=TestApp
APP_VERSION=1.0.0
SSL_SERT=certs/ssl.cert
SSL_KEY=certs/ssl.key
`
		envFile := ".test.env"
		err := os.WriteFile(envFile, []byte(envContent), 0644)
		require.NoError(t, err)
		defer func() {
			err := os.Remove(envFile)
			if err != nil {
				t.Errorf("failed to remove test env file: %v", err)
			}
		}()

		// Тестируем
		cfg := GetConfig(envFile)

		// Проверки
		assert.Equal(t, "postgres", cfg.DbDriverName)
		assert.Equal(t, "host=localhost user=test dbname=test", cfg.Dsn)
		assert.Equal(t, "TestApp", cfg.AppName)
		assert.Equal(t, "1.0.0", cfg.AppVersion)

		// Проверка валидации
		err = validator.New().Struct(cfg)
		assert.NoError(t, err, "Config should be valid")
	})

	t.Run("Valid config from environment variables", func(t *testing.T) {
		// Устанавливаем переменные окружения
		err := os.Setenv("DB_DRIVER_NAME", "mysql")
		require.NoError(t, err)
		err = os.Setenv("DB_DSN", "host=127.0.0.1 user=root")
		require.NoError(t, err)
		err = os.Setenv("APP_NAME", "TestApp2")
		require.NoError(t, err)
		err = os.Setenv("APP_VERSION", "2.0.0")
		require.NoError(t, err)
		err = os.Setenv("SSL_SERT", "certs/ssl.cert")
		require.NoError(t, err)
		err = os.Setenv("SSL_KEY", "certs/ssl.key")
		require.NoError(t, err)

		// Тестируем с несуществующим .env файлом
		cfg := GetConfig("non_existent.env")

		// Проверки
		assert.Equal(t, "mysql", cfg.DbDriverName)
		assert.Equal(t, "host=127.0.0.1 user=root", cfg.Dsn)
		assert.Equal(t, "TestApp2", cfg.AppName)
		assert.Equal(t, "2.0.0", cfg.AppVersion)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		// Очищаем все переменные окружения
		for _, key := range []string{"DB_DRIVER_NAME", "DB_DSN", "APP_NAME", "APP_VERSION"} {
			err := os.Unsetenv(key)
			if err != nil {
				t.Errorf("failed to unset env var %s: %v", key, err)
			}
		}

		// Проверяем панику при отсутствии конфигурации
		assert.Panics(t, func() {
			GetConfig("non_existent.env")
		})
	})

	t.Run("Invalid .env file but valid env vars", func(t *testing.T) {
		// Создаем неполный .env файл
		envContent := `DB_DRIVER_NAME=postgres`
		envFile := ".test_partial.env"
		err := os.WriteFile(envFile, []byte(envContent), 0644)
		require.NoError(t, err)
		defer func() {
			err := os.Remove(envFile)
			if err != nil {
				t.Errorf("failed to remove test env file: %v", err)
			}
		}()

		// Устанавливаем полные переменные окружения
		err = os.Setenv("DB_DSN", "host=127.0.0.1")
		require.NoError(t, err)
		err = os.Setenv("APP_NAME", "TestApp3")
		require.NoError(t, err)
		err = os.Setenv("APP_VERSION", "3.0.0")
		require.NoError(t, err)
		err = os.Setenv("SSL_SERT", "certs/ssl.cert")
		require.NoError(t, err)
		err = os.Setenv("SSL_KEY", "certs/ssl.key")
		require.NoError(t, err)

		// Тестируем
		cfg := GetConfig(envFile)

		// Проверки
		assert.Equal(t, "postgres", cfg.DbDriverName)
		assert.Equal(t, "host=127.0.0.1", cfg.Dsn)
		assert.Equal(t, "TestApp3", cfg.AppName)
		assert.Equal(t, "3.0.0", cfg.AppVersion)

		// Проверка валидации
		err = validator.New().Struct(cfg)
		assert.NoError(t, err, "Config should be valid")
	})
	// 1. Тест на валидную конфигурацию
	t.Run("Valid config with SSL", func(t *testing.T) {
		validCfg := Config{
			DbDriverName: "postgres",
			Dsn:          "host=localhost",
			AppName:      "TestApp",
			AppVersion:   "1.0.0",
			SslSert:      "certs/ssl.cert",
			SslKey:       "certs/ssl.key",
		}

		// Проверка валидации
		err := validator.New().Struct(validCfg)
		assert.NoError(t, err)
	})

	// 2. Тест на отсутствие SSL-сертификатов
	t.Run("Missing SSL paths", func(t *testing.T) {
		invalidCfg := Config{
			DbDriverName: "postgres",
			Dsn:          "host=localhost",
			AppName:      "TestApp",
			AppVersion:   "1.0.0",
			// SslSert и SslKey отсутствуют
		}

		// Проверка валидации
		err := validator.New().Struct(invalidCfg)
		assert.Error(t, err)

		// Проверяем конкретные ошибки
		validationErrors := err.(validator.ValidationErrors)
		assert.Len(t, validationErrors, 2)

		for _, e := range validationErrors {
			assert.Contains(t, []string{"SslSert", "SslKey"}, e.Field())
			assert.Equal(t, "required", e.Tag())
		}
	})

	// 3. Тест с пустыми путями SSL
	t.Run("Empty SSL paths", func(t *testing.T) {
		emptyPathsCfg := Config{
			DbDriverName: "postgres",
			Dsn:          "host=localhost",
			AppName:      "TestApp",
			AppVersion:   "1.0.0",
			SslSert:      "", // Пустая строка
			SslKey:       "", // Пустая строка
		}

		err := validator.New().Struct(emptyPathsCfg)
		assert.Error(t, err)
	})

	// 4. Тест с частичной конфигурацией
	t.Run("Only SSL cert provided", func(t *testing.T) {
		partialCfg := Config{
			DbDriverName: "postgres",
			Dsn:          "host=localhost",
			AppName:      "TestApp",
			AppVersion:   "1.0.0",
			SslSert:      "certs/ssl.cert",
			// SslKey отсутствует
		}

		err := validator.New().Struct(partialCfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SslKey")
	}) // тест для несуществующих ssl files
	t.Run("Non-existent SSL files", func(t *testing.T) {
		cfg := Config{
			DbDriverName: "postgres",
			Dsn:          "host=localhost",
			AppName:      "TestApp",
			AppVersion:   "1.0.0",
			SslSert:      "/nonexistent/cert.crt",
			SslKey:       "/nonexistent/key.key",
		}

		// Проверяем только валидацию структуры
		err := validator.New().Struct(cfg)
		assert.NoError(t, err) // Пути могут не существовать, но структура валидна

		// Дополнительная проверка существования файлов (если нужно)
		_, err = os.Stat(cfg.SslSert)
		assert.True(t, os.IsNotExist(err))

		_, err = os.Stat(cfg.SslKey)
		assert.True(t, os.IsNotExist(err))
	})
}
