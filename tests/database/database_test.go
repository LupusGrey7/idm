package database

import (
	"fmt"
	"idm/inner/config"
	"idm/inner/database"
	_ "os"
	"testing"

	"github.com/stretchr/testify/assert"
)

/**
* Database.go Test
* defer db.Close() - гарантирует закрытие подключения после завершения теста
* t - контекст теста для привязки проверки
* assert - проверяет условия, но не останавливает тест при ошибке
* запустить тесты - go test -v ./tests/
 */
func TestDatabaseConnection(t *testing.T) {

	// Тест для некорректного подключения с отловом паники + Некорректный конфиг БД
	t.Run("Invalid connection config", func(t *testing.T) {
		cfg := config.Config{
			DbDriverName: "postgres",
			Dsn:          "postgres://invalid:invalid@localhost:9999/nonexistent?sslmode=disable", // используем заведомо ложные данные
		}

		// Ожидаем панику и обрабатываем её
		defer func() {
			if r := recover(); r != nil {
				// Проверяем, что паника содержит ожидаемую ошибку
				errMsg, ok := r.(error)
				if ok {
					assert.Contains(t, errMsg.Error(), "dial tcp", "Должна быть ошибка подключения")
				} else {
					t.Errorf("Unexpected panic type: %v", r)
				}
			} else {
				t.Error("Expected panic, but no panic occurred")
			}
		}()

		// Этот вызов должен вызвать панику
		db := database.ConnectDbWithCfg(cfg)
		//после
		defer func() {
			if err := db.Close(); err != nil {
				t.Logf("failed to close database connection: %v", err)
			}
		}()
	})

	// Тест для корректного подключения без пароля + Корректный конфиг БД
	t.Run("Valid connection config", func(t *testing.T) {
		// Используем тестовую БД в Docker
		cfg := config.Config{
			DbDriverName: "postgres",
			Dsn:          "postgres://test:test@localhost:5434/idm_tests?sslmode=disable", // Параметры из docker-compose
		}

		db := database.ConnectDbWithCfg(cfg)
		//после завершения теста
		defer func() {
			if err := db.Close(); err != nil {
				t.Logf("failed to close database connection: %v", err)
			}
		}()

		// Проверяем подключение
		err := db.Ping()
		fmt.Println("Ping value -->> ", err)
		assert.NoError(t, err, "Должно быть успешное подключение")
	})
}
