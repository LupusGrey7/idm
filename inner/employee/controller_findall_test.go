package employee

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"idm/inner/common"
	"idm/inner/config"
	"idm/inner/domain"
	"idm/inner/web"
	"idm/inner/web/middleware"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestEmployeeController_FindAll(t *testing.T) {
	// Подготовка тестового .env файла
	envContent := `DB_DRIVER_NAME=postgres
DB_DSN=host=127.0.0.1 user=test dbname=idm_tests
APP_NAME=TestIdm
APP_VERSION=1.0.0
LOG_LEVEL=DEBUG
LOG_DEVELOP_MODE=true`
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
	appContext := context.Background() // Создаем контекст

	app := fiber.New()
	middleware.RegisterMiddleware(app, logger) // middleware func
	mockService := new(MockEmployeeService)

	ctrl := NewController(
		&web.Server{
			App:            app,
			GroupEmployees: app.Group("/api/v1/employees"),
		},
		mockService,
		logger,
	)

	app.Get("/api/v1/employees/", ctrl.FindAll)

	// Тестовые данные
	testTime := time.Date(2025, 6, 20, 12, 0, 0, 0, time.UTC)
	testEmployee := Response{
		Id:       123,
		Name:     "John Sena",
		CreateAt: testTime,
		UpdateAt: testTime,
	}

	// 1. Успешный запрос с данными
	t.Run("SuccessWithData", func(t *testing.T) {
		mockService.On("FindAll", appContext).Return([]Response{testEmployee}, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/employees/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		//3. Проверки
		require.NoError(t, err)
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result struct {
			Success bool       `json:"success"`
			Error   string     `json:"error"`
			Data    []Response `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}

		if !result.Success || result.Error != "" || len(result.Data) != 1 || result.Data[0] != testEmployee {
			t.Errorf("Expected success with data, got %+v", result)
		}
	})

	// 2. Успешный запрос без данных
	t.Run("SuccessEmpty", func(t *testing.T) {
		mockService.On("FindAll", appContext).Return([]Response{}, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/employees/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		//3. Проверки
		require.NoError(t, err)
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result struct {
			Success bool       `json:"success"`
			Error   string     `json:"error"`
			Data    []Response `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}

		if !result.Success || result.Error != "" || len(result.Data) != 0 {
			t.Errorf("Expected success with empty data, got %+v", result)
		}
	})

	// 3. Ошибка поиска
	t.Run("FindAllFailed", func(t *testing.T) {
		mockService.On("FindAll", appContext).Return([]Response{}, domain.ErrFindAllFailed).Once()

		req := httptest.NewRequest("GET", "/api/v1/employees/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		//3. Проверки
		require.NoError(t, err)
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		//4. Проверка тела ответа
		var result struct {
			Success bool        `json:"success"`
			Error   string      `json:"error"`
			Data    interface{} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}

		expectedError := "Failed to find all employees" // Точное соответствие тексту ошибки
		if result.Success || result.Error != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, result.Error)
		}
	})

	// 4. Внутренняя ошибка
	t.Run("InternalError", func(t *testing.T) {
		mockService.On("FindAll", appContext).Return([]Response{}, errors.New("db error")).Once()

		req := httptest.NewRequest("GET", "/api/v1/employees/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		//3. Проверки
		require.NoError(t, err)
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		//4. Проверка тела ответа
		var result struct {
			Success bool        `json:"success"`
			Error   string      `json:"error"`
			Data    interface{} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}

		expectedError := "db error" // Точное соответствие тексту ошибки
		if result.Success || result.Error != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, result.Error)
		}
	})
}
