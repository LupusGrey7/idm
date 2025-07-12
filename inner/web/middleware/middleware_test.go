package middleware

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"idm/inner/common"
	"idm/inner/config"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMiddleware(t *testing.T) {
	// Создаём логгер для тестов
	// Подготовка тестового .env файла
	envContent := `
DB_DRIVER_NAME=postgres
DB_DSN=host=127.0.0.1 user=test dbname=idm_tests
APP_NAME=TestIdm
APP_VERSION=1.0.0
LOG_LEVEL=DEBUG
LOG_DEVELOP_MODE=true
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
	cfg := config.GetConfig(envFile)
	var logger = common.NewLogger(cfg) // Создаем логгер

	app := fiber.New()
	RegisterMiddleware(app, logger) // middleware func

	// Вспомогательная функция для закрытия тела ответа
	closeBody := func(t *testing.T, body io.ReadCloser) {
		if err := body.Close(); err != nil {
			t.Errorf("Error closing response body: %v", err)
		}
	}

	var a = assert.New(t) // Создаём экземпляр объекта с ассерт-функциями
	t.Run("should recover from panic and return 500", func(t *testing.T) {
		// Arrange

		app.Get("/panic", func(c *fiber.Ctx) error {
			panic("simulated panic")

		})

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := app.Test(req)
		a.Nil(err)
		require.NoError(t, err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				if err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}
		}(resp.Body)

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		// Логируем заголовки и статус
		t.Logf("Status: %d", resp.StatusCode)
		t.Logf("Content-Type: %s", resp.Header.Get("Content-Type"))

		// Читаем тело ответа
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "expected no error reading response body")
		t.Logf("Raw response body: %s", string(body))

		// Проверяем JSON
		var errResp struct {
			Error string `json:"error"`
		}
		err = json.Unmarshal(body, &errResp)
		require.NoError(t, err, "expected no error decoding JSON response, got body: %s", body)
		assert.Equal(t, "Internal server error", errResp.Error, "expected error message")
	})

	t.Run("should add request ID to response and context", func(t *testing.T) {
		// Arrange
		app := fiber.New()
		RegisterMiddleware(app, logger)

		// Роут для проверки request_id
		app.Get("/test", func(c *fiber.Ctx) error {
			requestId := c.Locals("request_id").(string)
			return c.JSON(fiber.Map{"request_id": requestId})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err, "expected no HTTP error")
		defer closeBody(t, resp.Body)

		// Assert
		assert.Equal(t, fiber.StatusOK, resp.StatusCode, "expected status 200")

		// Проверяем заголовок X-Request-Id
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")

		// Проверяем тело ответа
		var response struct {
			RequestID string `json:"request_id"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "expected no error decoding response")
		assert.Equal(t, requestId, response.RequestID, "expected request ID in response to match header")
	})
}
