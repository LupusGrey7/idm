package employee

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"idm/inner/domain"
	"idm/inner/web"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEmployeeController_FindAll(t *testing.T) {
	app := fiber.New()
	mockService := new(MockEmployeeService)
	ctrl := NewController(&web.Server{
		App:            app,
		GroupEmployees: app.Group("/api/v1/employees"),
	}, mockService)

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
		mockService.On("FindAll").Return([]Response{testEmployee}, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/employees/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

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
		mockService.On("FindAll").Return([]Response{}, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/employees/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

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

	// 3. Ошибка поиска (ИСПРАВЛЕН ТЕКСТ ОШИБКИ)
	t.Run("FindAllFailed", func(t *testing.T) {
		mockService.On("FindAll").Return([]Response{}, domain.ErrFindAllFailed).Once()

		req := httptest.NewRequest("GET", "/api/v1/employees/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

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
		mockService.On("FindAll").Return([]Response{}, errors.New("db error")).Once()

		req := httptest.NewRequest("GET", "/api/v1/employees/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

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
