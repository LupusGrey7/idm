package employee

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"idm/inner/common"
	"idm/inner/config"
	"idm/inner/domain"
	"idm/inner/web"
	"idm/inner/web/middleware"
	"io"
	"os"

	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http/httptest"
	"testing"
	"time"
)

// ... аналогично для других методов

func TestEmployee_Controller(t *testing.T) {
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
	cfg := config.GetConfig(envFile)
	var logger = common.NewLogger(cfg) // Создаем логгер

	var a = assert.New(t) // Создаём экземпляр объекта с ассерт-функциями
	// заглушка для тестов - .env file

	// 1. Инициализация
	app := fiber.New()
	middleware.RegisterMiddleware(app, logger) //middleware

	server := &web.Server{
		App:            app,
		GroupApiV1:     app.Group("/api/v1"),
		GroupEmployees: app.Group("/api/v1/employees"),
	}

	mockService := new(MockEmployeeService)
	ctrl := NewController(server, mockService, logger)

	server.GroupEmployees.Get("/", ctrl.FindAll)
	server.GroupEmployees.Get("/ids", ctrl.FindAllByIds)
	server.GroupEmployees.Get("/:id", ctrl.FindById)
	server.GroupEmployees.Post("/", ctrl.CreateEmployee)
	server.GroupEmployees.Post("/employee", ctrl.CreateEmployeeTx)
	server.GroupEmployees.Put("/:id", ctrl.Update)
	server.GroupEmployees.Delete("/ids", ctrl.DeleteByIds) // Сначала специфичный маршрут
	server.GroupEmployees.Delete("/:id", ctrl.DeleteById)  // Потом общий

	testID := int64(1)
	testName := "John Sena"
	now := time.Now().UTC().Truncate(time.Second)
	//Сброс моков перед каждым тестом (рекомендуется)
	t.Cleanup(func() {
		mockService.AssertExpectations(t)
	})

	// 6. Тестовые случаи
	t.Run("should return employee by Id", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом

		expectedData := Response{
			Id:       testID,
			Name:     testName,
			CreateAt: now,
			UpdateAt: now,
		}

		// 3. Настройка мока
		mockService.On("FindById", testID).Return(expectedData, nil).Once()

		// 4. Выполнение запроса
		req := httptest.NewRequest("GET", "/api/v1/employees/1", nil)

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

		// 6. Выполняем проверки полученных данных
		a.Nil(err)
		a.NotEmpty(resp)
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Декодируем в структуру-обёртку
		var responseWrapper struct {
			Success bool     `json:"success"`
			Data    Response `json:"data"`
		}

		err = json.Unmarshal(body, &responseWrapper)
		require.NoError(t, err)

		// Проверяем обёртку
		assert.True(t, responseWrapper.Success)

		// Проверяем данные
		actualResponse := responseWrapper.Data
		assert.Equal(t, testID, actualResponse.Id)
		assert.Equal(t, testName, actualResponse.Name)
		assert.WithinDuration(t, now, actualResponse.CreateAt, time.Second)
		assert.WithinDuration(t, now, actualResponse.UpdateAt, time.Second)

		mockService.AssertExpectations(t)

	})
	// create entity
	t.Run("should return created employee", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом

		expectedData := Response{
			Id:       testID,
			Name:     testName,
			CreateAt: now,
			UpdateAt: now,
		}

		createRequest := CreateRequest{
			Name: testName,
		}

		// Готовим тестовое окружение
		var body = strings.NewReader("{\"name\": \"John Sena\"}")
		var req = httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/", body) // 4. Выполнение запроса
		req.Header.Set("Content-Type", "application/json")

		// Настраиваем поведение мока в тесте
		mockService.On("CreateEmployee", createRequest).Return(expectedData, nil)
		// Отправляем тестовый запрос на веб сервер
		resp, err := app.Test(req)

		// Выполняем проверки полученных данных
		a.Nil(err)
		a.NotEmpty(resp)
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")
		a.Equal(http.StatusCreated, resp.StatusCode)

		bytesData, err := io.ReadAll(resp.Body)
		a.Nil(err)

		t.Log(string(bytesData)) // Логируем JSON для отладки

		// Декодируем в структуру-обёртку
		var responseBody struct {
			Success bool     `json:"success"`
			Data    Response `json:"data"`
		}
		err = json.Unmarshal(bytesData, &responseBody)

		a.Nil(err)
		a.True(responseBody.Success)
		//a.Empty(responseBody.Message)
		a.Equal(testID, responseBody.Data.Id)
		a.Equal(testName, responseBody.Data.Name)
		a.Equal(now, responseBody.Data.CreateAt)
		a.Equal(now, responseBody.Data.UpdateAt)
	})
	//create error by name
	t.Run("when create employee then should return error", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		testName := "J"                 // Используем то же имя, что и в теле запроса

		createRequest := CreateRequest{
			Name: testName,
		}
		//Важно!- Ошибка валидации должна быть типа domain.RequestValidationError
		expectError := domain.RequestValidationError{Message: "validate name error"}

		// Готовим тестовое окружение
		var body = strings.NewReader("{\"name\": \"J\"}")
		var req = httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/", body) // 4. Выполнение запроса
		req.Header.Set("Content-Type", "application/json")

		// Настраиваем поведение мока в тесте
		mockService.On("CreateEmployee", createRequest).Return(Response{}, expectError).Once()
		// Отправляем тестовый запрос на веб сервер
		resp, err := app.Test(req)
		require.NoError(t, err) // Здесь не должно быть ошибок на уровне HTTP
		defer closeBody(t, resp.Body)

		// 4. Проверяем ответ
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Выполняем проверки полученных данных - Проверка тела ответа
		var errorResponse struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)
		require.Contains(t, errorResponse.Error, "validate name error")

		// 6. Проверка вызовов мока
		mockService.AssertExpectations(t)
	})
	//create error by server error
	t.Run("when create employee then should return error", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		testName := "J"                 // Используем то же имя, что и в теле запроса

		createRequest := CreateRequest{
			Name: testName,
		}
		//Важно!- Ошибка валидации должна быть типа domain.RequestValidationError
		expectError := fmt.Errorf("error creating Role with name %s: %w", createRequest.Name, errors.New("server error"))

		// Готовим тестовое окружение
		var body = strings.NewReader("{\"name\": \"J\"}")
		var req = httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/", body) // 4. Выполнение запроса
		req.Header.Set("Content-Type", "application/json")

		// Настраиваем поведение мока в тесте
		mockService.On("CreateEmployee", createRequest).Return(Response{}, expectError).Once()
		// Отправляем тестовый запрос на веб сервер
		resp, err := app.Test(req)
		require.NoError(t, err) // Здесь не должно быть ошибок на уровне HTTP
		defer closeBody(t, resp.Body)

		// 4. Проверяем ответ
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Выполняем проверки полученных данных - Проверка тела ответа
		var errorResponse struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)
		require.Contains(t, errorResponse.Error, "Internal server error")

		// 6. Проверка вызовов мока
		mockService.AssertExpectations(t)
	})
	//update by id
	t.Run("should return employee when update by Id", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом

		expectedData := Response{
			Id:       testID,
			Name:     testName,
			CreateAt: now,
			UpdateAt: now,
		}

		requestEmployee := UpdateRequest{
			Id:        int64(0),
			Name:      testName,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// 1. Сериализуем структуру в JSON
		requestBody, err := json.Marshal(requestEmployee)
		if err != nil {
			t.Fatal(err)
		}

		// 2. Создаем запрос с телом
		req := httptest.NewRequest("PUT", "/api/v1/employees/1", bytes.NewBuffer(requestBody))

		// 3. Устанавливаем заголовки
		req.Header.Set("Content-Type", "application/json")

		// 4. Настройка мока (убедитесь, что ожидаете правильные параметры)
		mockService.On("UpdateEmployee", testID, mock.AnythingOfType("UpdateRequest")).Return(expectedData, nil).Once()
		// 5. Выполнение запроса
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

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

		// 6. Проверки
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Декодируем в структуру-обёртку
		var responseWrapper struct {
			Success bool     `json:"success"`
			Data    Response `json:"data"`
		}

		err = json.Unmarshal(body, &responseWrapper)
		require.NoError(t, err)

		// Проверяем обёртку
		assert.True(t, responseWrapper.Success)

		// Проверяем данные
		actualResponse := responseWrapper.Data
		assert.Equal(t, testID, actualResponse.Id)
		assert.Equal(t, testName, actualResponse.Name)
		assert.WithinDuration(t, now, actualResponse.CreateAt, time.Second)
		assert.WithinDuration(t, now, actualResponse.UpdateAt, time.Second)

		mockService.AssertExpectations(t)

	})
	//when update by ID error
	t.Run("should return error when update by Id", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		requestEmployee := UpdateRequest{
			Id:        int64(0),
			Name:      testName,
			CreatedAt: now,
			UpdatedAt: now,
		}
		//Важно!- Ошибка валидации должна быть типа domain.RequestValidationError
		expectError := domain.RequestValidationError{Message: "validate name error"}
		// 1. Сериализуем структуру в JSON
		requestBody, err := json.Marshal(requestEmployee)
		if err != nil {
			t.Fatal(err)
		}

		// 2. Создаем запрос с телом
		req := httptest.NewRequest("PUT", "/api/v1/employees/1", bytes.NewBuffer(requestBody))

		// 3. Устанавливаем заголовки
		req.Header.Set("Content-Type", "application/json")

		// 4. Настройка мока (убедитесь, что ожидаете правильные параметры)
		mockService.On("UpdateEmployee", testID, mock.AnythingOfType("UpdateRequest")).Return(Response{}, expectError).Once()
		// 5. Выполнение запроса
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		require.NoError(t, err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				if err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}
		}(resp.Body)

		// 6. Проверки
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var errorResponse struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)
		require.Contains(t, errorResponse.Error, "validate name error")

		// 6. Проверка вызовов мока
		mockService.AssertExpectations(t)
	})
	//when ID param update error
	t.Run("should return error when update by Id", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		idParam := "abc"

		requestEmployee := UpdateRequest{
			Id:        int64(0),
			Name:      testName,
			CreatedAt: now,
			UpdatedAt: now,
		}
		//Важно!- Ошибка валидации должна быть типа domain.RequestValidationError
		//expectError := domain.RequestValidationError{Message: "Invalid ID format"}
		// 1. Сериализуем структуру в JSON
		requestBody, err := json.Marshal(requestEmployee)
		if err != nil {
			t.Fatal(err)
		}

		// 2. Создаем запрос с телом
		req := httptest.NewRequest("PUT", "/api/v1/employees/"+idParam, bytes.NewBuffer(requestBody))

		// 3. Устанавливаем заголовки
		req.Header.Set("Content-Type", "application/json")

		// 4. Настройка мока (убедитесь, что ожидаете правильные параметры) - в данном случае мок не вызывается
		//mockService.On(
		//	"UpdateEmployee",
		//	testID,
		//	mock.AnythingOfType("UpdateRequest"),
		//).Return(expectedData, expectError).Once()

		// 5. Выполнение запроса
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		require.NoError(t, err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				if err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}
		}(resp.Body)

		// 6. Проверки
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var errorResponse struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)
		require.Contains(t, errorResponse.Error, "Invalid ID format")
		assert.Equal(t, "Invalid ID format", errorResponse.Error, "expected error message to match")

		// 6. Проверка вызовов мока
		mockService.AssertNotCalled(t, "UpdateEmployee") // Проверяем, что UpdateEmployee НЕ был вызван
	})
	// Тест на успешное получение по IDs
	t.Run("should return employees when found by IDs", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		// 1. Подготовка данных
		requestIDs := []int64{1, 2, 3}
		idParam := "1,2,3"
		expectedData := []Response{
			{Id: 1, Name: "John"},
			{Id: 2, Name: "Jane"},
			{Id: 3, Name: "Doe"},
		}

		// 2. Настройка мока
		mockService.On("FindAllByIds", requestIDs).Return(expectedData, nil).Once()

		// 3. Создание запроса
		req := httptest.NewRequest("GET", "/api/v1/employees/ids?ids="+idParam, nil)

		// 4. Выполнение запроса
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { //Явная проверка ошибки (рекомендуется)
			err := resp.Body.Close()
			if err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		// 5. Проверки
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response struct {
			Success bool       `json:"success"`
			Data    []Response `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))

		assert.True(t, response.Success)
		assert.Equal(t, expectedData, response.Data)
	})
	// Тест на отсутствие параметра ids для FindAllByIds
	t.Run("should return error when ids parameter is missing for FindAllByIds", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом

		req := httptest.NewRequest("GET", "/api/v1/employees/ids", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		//
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Missing ids parameter")
	})

	// Тест на неверный формат ID для FindAllByIds
	t.Run("should return error when invalid ID format for FindAllByIds", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		req := httptest.NewRequest("GET", "/api/v1/employees/ids?ids=1,abc,3", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Invalid ID format")
	})
	// Тест на пустой результат для FindAllByIds
	t.Run("should return empty array when no employees found", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		// 1. Подготовка данных
		requestIDs := []int64{1, 2, 3}
		idParam := "1,2,3"

		// 2. Настройка мока
		mockService.On("FindAllByIds", requestIDs).Return([]Response{}, nil).Once()

		// 3. Создание запроса
		req := httptest.NewRequest("GET", "/api/v1/employees/ids?ids="+idParam, nil)

		// 4. Выполнение запроса
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		// 5. Проверки
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response struct {
			Success bool       `json:"success"`
			Data    []Response `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))

		assert.True(t, response.Success)
		assert.Empty(t, response.Data)
	})
	//delete by id
	t.Run("should return success when delete by Id", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		response := Response{}

		// 1. Сериализуем структуру в JSON

		// 2. Создаем запрос с телом
		req := httptest.NewRequest("DELETE", "/api/v1/employees/1", nil)

		// 3. Устанавливаем заголовки
		req.Header.Set("Content-Type", "application/json")

		// 4. Настройка мока (убедитесь, что ожидаете правильные параметры)
		mockService.On("DeleteById", testID).Return(response, nil)
		// 5. Выполнение запроса
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

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

		// 6. Проверки
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Декодируем в структуру-обёртку
		var responseWrapper struct {
			Success bool     `json:"success"`
			Data    Response `json:"data"`
		}

		err = json.Unmarshal(body, &responseWrapper)
		require.NoError(t, err)

		// Проверяем обёртку
		assert.True(t, responseWrapper.Success)

		// Проверяем данные
		actualResponse := responseWrapper.Data
		assert.Equal(t, "", actualResponse.Name)

		mockService.AssertExpectations(t)
		mockService.AssertNumberOfCalls(t, "DeleteById", 1)

	})
	// delete by id error
	t.Run("should return error when delete employee by ID", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		// 1. Подготовка данных
		requestIDs := "abc" // ID роли
		expectedError := errors.New("Invalid ID format")

		// 2. Настройка мока на возврат ошибки
		mockService.On("DeleteById", requestIDs).Return(Response{}, expectedError).Once()

		// 3. Создание запроса
		req := httptest.NewRequest("DELETE", "/api/v1/employees/"+requestIDs, nil)
		req.Header.Set("Content-Type", "application/json")

		// 4. Выполнение запроса
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		// 5. Проверки
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var response struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(body, &response))

		assert.False(t, response.Success)
		assert.Equal(t, expectedError.Error(), response.Error)
	})
	// Тест на неверный формат ID
	t.Run("should return error when invalid employee ID format", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		// 1. Подготовка данных с невалидным ID
		idParam := "abc"

		// 2. Создаем запрос
		req := httptest.NewRequest("DELETE", "/api/v1/employees/"+idParam, nil)
		req.Header.Set("Content-Type", "application/json")

		// 3. Выполнение запроса
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		// 4. Проверки
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var response struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(body, &response))

		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "Invalid ID format")
	})
	//delete all by ids
	t.Run("should return success when delete All Employees by IDs", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом

		// Подготовка тестовых данных
		requestIDs := []int64{1, 2, 3}
		idParam := "1,2,3" // ID как строка с разделителем-запятой

		// 1. Создаем запрос с query-параметром
		// 3. Создание КОРРЕКТНОГО запроса (с /ids в пути)
		req := httptest.NewRequest(
			"DELETE",
			"/api/v1/employees/ids?ids="+idParam, // Добавлен /ids перед ?
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		// 2. Настройка мока (возвращаем nil, так как это DELETE)
		mockService.On("DeleteByIds", requestIDs).Return(Response{}, nil).Once()

		// Перед app.Test(req)
		routes := app.GetRoutes()
		for _, route := range routes {
			t.Logf("Route: %s %s", route.Method, route.Path)
		}

		// 3. Выполнение запроса
		resp, err := app.Test(req)
		require.NoError(t, err)
		a.Nil(err)
		defer closeBody(t, resp.Body)

		// 4. Проверка статуса
		requestId := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestId, "expected non-empty request ID")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		// 5. Проверка тела ответа
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var response struct {
			Success bool        `json:"success"`
			Error   string      `json:"error"`
			Data    interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(body, &response))

		assert.True(t, response.Success)
		assert.Empty(t, response.Error)

		// 6. Проверка вызова мока
		mockService.AssertNumberOfCalls(t, "DeleteByIds", 1)
	})
	// Тест на отсутствие параметра ids
	t.Run("should return error when ids parameter is missing", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		// 1. Создаем запрос БЕЗ параметра ids
		req := httptest.NewRequest("DELETE", "/api/v1/employees/ids", nil)
		req.Header.Set("Content-Type", "application/json")

		// 2. Выполнение запроса
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		// 3. Проверки
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var response struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(body, &response))

		assert.False(t, response.Success)
		assert.Equal(t, "Missing ids parameter", response.Error)
	})
	// Тест на неверный формат ID
	t.Run("should return error when invalid ID format", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		// 1. Подготовка данных с невалидным ID
		idParam := "1,abc,3"

		// 2. Создаем запрос
		req := httptest.NewRequest("DELETE", "/api/v1/employees/ids?ids="+idParam, nil)
		req.Header.Set("Content-Type", "application/json")

		// 3. Выполнение запроса
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		// 4. Проверки
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var response struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(body, &response))

		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "Invalid ID format")
	})
	// Тест на ошибку сервиса при удалении
	t.Run("should return error when service fails", func(t *testing.T) {
		mockService.ExpectedCalls = nil // Сбрасываем моки перед тестом
		// 1. Подготовка данных
		requestIDs := []int64{1, 2, 3}
		idParam := "1,2,3"
		expectedError := errors.New("database error")

		// 2. Настройка мока на возврат ошибки
		mockService.On("DeleteByIds", requestIDs).Return(Response{}, expectedError).Once()

		// 3. Создание запроса
		req := httptest.NewRequest("DELETE", "/api/v1/employees/ids?ids="+idParam, nil)
		req.Header.Set("Content-Type", "application/json")

		// 4. Выполнение запроса
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		// 5. Проверки
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var response struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(body, &response))

		assert.False(t, response.Success)
		assert.Equal(t, expectedError.Error(), response.Error)
	})
}
func closeBody(t *testing.T, body io.ReadCloser) {
	if err := body.Close(); err != nil {
		t.Errorf("Error closing response body: %v", err)
	}
}
