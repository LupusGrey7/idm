package info

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"idm/inner/config"
	"idm/inner/domain"
	"idm/inner/web"
	"io"
	"net/http/httptest"
	"testing"
)

// MockDatabase - мок для интерфейса Database

func TestController_GetInfo_Error(t *testing.T) {
	var a = assert.New(t) // Создаём экземпляр объекта с ассерт-функциями
	// 1. Подготовка
	app := fiber.New()
	mockCfg := config.Config{
		AppName:    "TestApp",
		AppVersion: "1.0.0",
	}

	// Создаем контроллер с моком сервера
	mockService := new(MockHealthService)
	server := &web.Server{
		App:           app,
		GroupInternal: app.Group("/internal"), // Группа непубличного API
	}

	ctrl := NewController(server, mockCfg, mockService)

	server.GroupInternal.Get("/info", ctrl.GetInfo)
	server.GroupInternal.Get("/health", ctrl.GetHealth)

	response := InfoResponse{
		Name:    "TestApp",
		Version: "1.0.0",
		Status:  "OK",
	}

	t.Run("should return app info successfully", func(t *testing.T) {
		mockService.Test(t) // Важно для корректного отслеживания вызовов - (вызовется ровно 1 раз)
		mockService.On("CheckDB").Return(nil).Once()
		// 4. Выполнение запроса
		req := httptest.NewRequest("GET", "/internal/info", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				if err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}
		}(resp.Body)

		// 5. Проверка сырого ответа
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Raw response: %s", string(body))

		// 6.  Проверка статуса
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		// Декодируем в структуру-обёртку
		var responseWrapper InfoResponse
		err = json.Unmarshal(body, &responseWrapper)
		require.NoError(t, err)

		// Проверяем данные
		a.Nil(err)
		assert.Equal(t, response.Name, responseWrapper.Name)
		assert.Equal(t, response.Version, responseWrapper.Version)
		assert.Equal(t, response.Status, responseWrapper.Status)

	})

	// 1. Info -failure
	t.Run("should return Info check failure", func(t *testing.T) {
		// 1. Сброс предыдущих ожиданий мока
		mockService.ExpectedCalls = nil
		mockService.Test(t) // Важно для корректного отслеживания вызовов

		// Подготовка мока: CheckDB() возвращает ошибку -(вызовется ровно 1 раз)
		mockService.On("CheckDB").Return(errors.New("db connection failed")).Once()
		req := httptest.NewRequest("GET", "/internal/info", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				if err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}
		}(resp.Body)

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Raw response: %s", string(body))
		// Проверка статуса
		require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		var apiErr domain.APIError
		err = json.Unmarshal(body, &apiErr)
		require.NoError(t, err)

		// Проверка тела ответа
		assert.Equal(t, "Database service unavailable", apiErr.Message)
		assert.Equal(t, "db connection failed", apiErr.Details)
	})
	//GetHealth -error
	t.Run("should return Health check failure", func(t *testing.T) {
		// 1. Сброс предыдущих ожиданий мока
		mockService.ExpectedCalls = nil
		mockService.Test(t) // Важно для корректного отслеживания вызовов

		// Подготовка мока: CheckDB() возвращает ошибку - (вызовется ровно 1 раз)
		mockService.On("CheckDB").Return(errors.New("db connection failed")).Once()

		// Тестовый сервер
		app := fiber.New()
		app.Get("/info", ctrl.GetInfo)

		// Запрос
		req := httptest.NewRequest("GET", "/info", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				if err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}
		}(resp.Body)

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Raw response: %s", string(body))

		// Проверки
		require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		var apiErr domain.APIError
		err = json.Unmarshal(body, &apiErr)
		require.NoError(t, err)

		assert.Equal(t, "Database service unavailable", apiErr.Message)
		mockService.AssertExpectations(t)
		mockService.AssertCalled(t, "CheckDB")
	})

	//GetHealth-  Успешный случай
	t.Run("should return Health check success", func(t *testing.T) {
		// 1. Сброс предыдущих ожиданий мока
		mockService.ExpectedCalls = nil
		mockService.Test(t) // Важно для корректного отслеживания вызовов
		// Подготовка мока: CheckDB() возвращает ошибку - (вызовется ровно 1 раз)
		mockService.On("CheckDB").Return(nil).Once()

		app := fiber.New()
		app.Get("/health", ctrl.GetHealth)

		req := httptest.NewRequest("GET", "/health", nil)
		resp, _ := app.Test(req)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				if err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}
		}(resp.Body)

		assert.Equal(t, 200, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}
