package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
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

	//в тесте проверяется- Отсутствие .env файла
	t.Run("No .env file", func(t *testing.T) {
		// Убедимся, что .env файла нет
		os.Clearenv()       //Удаляет ВСЕ переменные окружения текущего процесса
		defer os.Clearenv() //"отложенного выполнения" -> Поставили в очередь очистку окружения

		cfg := GetConfig("non_existent.env")

		//assert - проверяет условия, но не останавливает тест при ошибке
		assert.Empty(t, cfg.DbDriverName) // Проверяем, что структура пустая
		assert.Empty(t, cfg.Dsn)          // Проверяем, что структура пустая
	})

	//в тесте проверяется- Пустой .env и переменные окружения
	t.Run("Empty .env and no env vars", func(t *testing.T) {
		// Создаем временный пустой .env
		dir := t.TempDir()                    //t.TempDir() - создает временную директорию, которая автоматически очищается после теста
		envPath := filepath.Join(dir, ".env") // Формируем путь к файлу .env внутри временной директории -> Например: /tmp/gotest123456/.env

		err := os.WriteFile(envPath, []byte(""), 0644) // Создаем пустой файл .env (содержимое - пустой байтовый срез) -> 0644 - права доступа (rw-r--r--)
		require.NoError(t, err, "Failed to create empty .env file")

		cfg := GetConfig(envPath) // Вызываем функцию GetConfig, функцию получения конфигурации передавая путь к пустому .env файлу

		assert.Empty(t, cfg.DbDriverName) // Проверяем, что DbDriverName пустая строка
		assert.Empty(t, cfg.Dsn)          // Проверяем, что Dsn пустая строка
	})

	//в тесте проверяется-Переменные окружения переопределяют .env // Создаем .env без переменных
	t.Run("Env vars override .env", func(t *testing.T) {

		dir := t.TempDir()                    //t.TempDir() - создает временную директорию, которая автоматически очищается после теста
		envPath := filepath.Join(dir, ".env") // Создаем путь к .env

		err := os.WriteFile(envPath, []byte(""), 0644) // Создаем пустой файл .env (содержимое - пустой байтовый срез) -> 0644 - права доступа (rw-r--r--)
		require.NoError(t, err, "Failed to Env vars override .env file")

		//t.Setenv() - временно устанавливает переменные окружения (восстанавливает оригинальные значения после теста)
		t.Setenv("DB_DRIVER_NAME", "postgres")                       // Устанавливаем переменные окружения
		t.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/db") // Устанавливаем переменные окружения

		cfg := GetConfig(envPath) // Читаем конфиг

		//assert - проверяет условия, но не останавливает тест при ошибке
		assert.Equal(t, "postgres", cfg.DbDriverName) //Проверяет, что значение равно ожидаемому
		assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.Dsn)
	})

	//в тесте проверяется-Корректный .env
	t.Run("Valid .env file", func(t *testing.T) {
		// Создаем корректный .env
		dir := t.TempDir() //t.TempDir() - создает временную директорию, которая автоматически очищается после теста
		envPath := filepath.Join(dir, ".env")

		err := os.WriteFile(envPath, []byte(`
DB_DRIVER_NAME=postgres
DB_DSN=postgres://user:pass@localhost:5432/db
`), 0644) // Создаем пустой файл .env (содержимое - пустой байтовый срез), заполняем данными -> 0644 - права доступа (rw-r--r--)
		require.NoError(t, err, "Failed to create .env file")

		cfg := GetConfig(envPath)

		//assert - проверяет условия, но не останавливает тест при ошибке
		assert.Equal(t, "postgres", cfg.DbDriverName)
		assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.Dsn)
	})

	//в тесте проверяется-Конфликт .env и переменных окружения
	t.Run("Env vars override .env", func(t *testing.T) {
		// Создаем .env с одними значениями
		dir := t.TempDir() //t.TempDir() - создает временную директорию, которая автоматически очищается после теста
		envPath := filepath.Join(dir, ".env")

		err := os.WriteFile(envPath, []byte(`
DB_DRIVER_NAME=mysql
DB_DSN=mysql://user:pass@localhost:3306/db
`), 0644)
		require.NoError(t, err, "Failed to Env vars override .env file")

		//t.Setenv() - временно устанавливает переменные окружения (восстанавливает оригинальные значения после теста)
		t.Setenv("DB_DRIVER_NAME", "postgres")                       // Устанавливаем другие значения через env
		t.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/db") // Устанавливаем другие значения через env

		cfg := GetConfig(envPath)

		// assert - проверяет условия, но не останавливает тест при ошибке
		// Проверяем, что приоритет у переменных окружения
		assert.Equal(t, "postgres", cfg.DbDriverName)
		assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.Dsn)
	})
}
