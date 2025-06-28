package role

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"idm/inner/common"
	"idm/inner/config"
	"idm/inner/web"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRole_Controller(t *testing.T) {
	var a = assert.New(t) // Создаём экземпляр объекта с ассерт-функциями

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

	// 1. Инициализация
	app := fiber.New()

	server := &web.Server{
		App:            app,
		GroupApiV1:     app.Group("/api/v1"),
		GroupEmployees: app.Group("/api/v1/roles"),
	}

	mockService := new(MockRoleService)
	ctrl := NewController(server, mockService, logger)

	server.GroupEmployees.Get("/", ctrl.FindAll)
	server.GroupEmployees.Get("/ids", ctrl.FindAllByIds) // Сначала специфичный маршрут
	server.GroupEmployees.Get("/:id", ctrl.FindById)     // Потом общий
	server.GroupEmployees.Post("/", ctrl.CreateRole)
	server.GroupEmployees.Put("/:id", ctrl.UpdateRole)
	server.GroupEmployees.Delete("/ids", ctrl.DeleteByIds) // Сначала специфичный маршрут
	server.GroupEmployees.Delete("/:id", ctrl.DeleteById)  // Потом общий

	testID := int64(1)
	testName := "ADMIN"
	now := time.Now().UTC().Truncate(time.Second)

	// 6. Тестовый случай
	t.Run("should return role by ID", func(t *testing.T) {

		expectedData := Response{
			Id:         testID,
			Name:       testName,
			EmployeeID: nil,
			CreateAt:   now,
			UpdateAt:   now,
		}

		// 3. Настройка мока
		mockService.On("FindById", testID).Return(expectedData, nil)

		// 4. Выполнение запроса
		req := httptest.NewRequest("GET", "/api/v1/roles/1", nil)

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
		a.Nil(err)
		assert.Equal(t, testID, actualResponse.Id)
		assert.Equal(t, testName, actualResponse.Name)
		assert.WithinDuration(t, now, actualResponse.CreateAt, time.Second)
		assert.WithinDuration(t, now, actualResponse.UpdateAt, time.Second)

		mockService.AssertExpectations(t)

	})
	// create
	t.Run("should return created role", func(t *testing.T) {

		expectedData := Response{
			Id:         testID,
			Name:       testName,
			EmployeeID: nil,
			CreateAt:   now,
			UpdateAt:   now,
		}

		createRequest := CreateRequest{
			Name: testName,
		}

		// Готовим тестовое окружение
		var body = strings.NewReader("{\"name\": \"ADMIN\"}")
		var req = httptest.NewRequest(fiber.MethodPost, "/api/v1/roles/", body) // 4. Выполнение запроса
		req.Header.Set("Content-Type", "application/json")

		// Настраиваем поведение мока в тесте
		mockService.On("CreateRole", createRequest).Return(expectedData, nil)
		// Отправляем тестовый запрос на веб сервер
		resp, err := app.Test(req)

		// Выполняем проверки полученных данных
		a.Nil(err)
		a.NotEmpty(resp)
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
	//update by id
	t.Run("should return role when update by ID", func(t *testing.T) {
		ID := int64(1)
		expectedData := Response{
			Id:         testID,
			Name:       testName,
			EmployeeID: &ID,
			CreateAt:   now,
			UpdateAt:   now,
		}

		requestEmployee := UpdateRequest{
			Id:         int64(0),
			EmployeeID: &ID,
			Name:       testName,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		// 1. Сериализуем структуру в JSON
		requestBody, err := json.Marshal(requestEmployee)
		if err != nil {
			t.Fatal(err)
		}

		// 2. Создаем запрос с телом
		req := httptest.NewRequest("PUT", "/api/v1/roles/1", bytes.NewBuffer(requestBody))

		// 3. Устанавливаем заголовки
		req.Header.Set("Content-Type", "application/json")

		// 4. Настройка мока (убедитесь, что ожидаете правильные параметры)
		mockService.On("UpdateRole", testID, mock.AnythingOfType("UpdateRequest")).Return(expectedData, nil)
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
		a.Nil(err)
		actualResponse := responseWrapper.Data
		assert.Equal(t, testID, actualResponse.Id)
		assert.Equal(t, testName, actualResponse.Name)
		assert.WithinDuration(t, now, actualResponse.CreateAt, time.Second)
		assert.WithinDuration(t, now, actualResponse.UpdateAt, time.Second)

		mockService.AssertExpectations(t)
		mockService.AssertNumberOfCalls(t, "UpdateRole", 1)
	})
	// update by id error
	t.Run("should return error when update by ID role", func(t *testing.T) {
		ID := int64(1)
		expectedData := Response{
			Id:         testID,
			Name:       testName,
			EmployeeID: &ID,
			CreateAt:   now,
			UpdateAt:   now,
		}

		requestEmployee := UpdateRequest{
			Id:         int64(0),
			EmployeeID: &ID,
			Name:       testName,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		idParam := "abc"
		// 1. Сериализуем структуру в JSON
		requestBody, err := json.Marshal(requestEmployee)
		if err != nil {
			t.Fatal(err)
		}

		// 2. Создаем запрос с телом
		req := httptest.NewRequest("PUT", "/api/v1/roles/"+idParam, bytes.NewBuffer(requestBody))

		// 3. Устанавливаем заголовки
		req.Header.Set("Content-Type", "application/json")

		// 4. Настройка мока (убедитесь, что ожидаете правильные параметры)
		mockService.On("UpdateRole", testID, mock.AnythingOfType("UpdateRequest")).Return(expectedData, nil)
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
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Декодируем в структуру-обёртку
		var responseWrapper struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}

		err = json.Unmarshal(body, &responseWrapper)
		require.NoError(t, err)

		// Проверяем обёртку
		assert.False(t, responseWrapper.Success)

		// Проверяем данные
		a.Empty(err)
		assert.False(t, responseWrapper.Success)
		assert.Contains(t, responseWrapper.Error, "Invalid ID format")

	})
	// Тест на успешное получение по IDs
	t.Run("should return roles when found by IDs", func(t *testing.T) {
		// 1. Подготовка данных
		requestIDs := []int64{1, 2, 3}
		idParam := "1,2,3"
		expectedData := []Response{
			{Id: 1, Name: "ADMIN"},
			{Id: 2, Name: "CLIENT"},
			{Id: 3, Name: "CEO"},
		}

		// 2. Настройка мока
		mockService.On("FindAllByIds", requestIDs).Return(expectedData, nil).Once()

		// 3. Создание запроса
		req := httptest.NewRequest("GET", "/api/v1/roles/ids?ids="+idParam, nil)

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
		req := httptest.NewRequest("GET", "/api/v1/roles/ids", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Missing ids parameter")
	})
	// Тест на неверный формат ID для FindAllByIds
	t.Run("should return error when invalid ID format for FindAllByIds", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/roles/ids?ids=1,abc,3", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Invalid ID format")
	})
	// Тест на пустой результат для FindAllByIds
	t.Run("should return empty array when no roles found", func(t *testing.T) {
		// 1. Подготовка данных
		requestIDs := []int64{1, 2, 3}
		idParam := "1,2,3"

		// 2. Настройка мока
		mockService.On("FindAllByIds", requestIDs).Return([]Response{}, nil).Once()

		// 3. Создание запроса
		req := httptest.NewRequest("GET", "/api/v1/roles/ids?ids="+idParam, nil)

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
	t.Run("should return success when delete role by ID", func(t *testing.T) {

		response := Response{}

		// 1. Сериализуем структуру в JSON

		// 2. Создаем запрос с телом
		req := httptest.NewRequest("DELETE", "/api/v1/roles/1", nil)

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
	t.Run("should return error when delete role by ID", func(t *testing.T) {
		// 1. Сброс предыдущих ожиданий мока
		mockService.ExpectedCalls = nil

		// 2. Подготовка данных
		roleID := int64(1)
		expectedError := errors.New("database error")

		// 3. Настройка мока (с явным указанием .Once())- (вызовется ровно 1 раз)
		mockService.On("DeleteById", roleID).Return(Response{}, expectedError).Once()

		// 4. Создание запроса
		req := httptest.NewRequest("DELETE", "/api/v1/roles/1", nil)
		req.Header.Set("Content-Type", "application/json")

		// 5. Выполнение запроса
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, resp.Body.Close())
		}()

		// 6. Проверка статуса
		if resp.StatusCode != fiber.StatusInternalServerError {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 500, got %d. Response body: %s", resp.StatusCode, string(body))
		} else {
			require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		}

		// 7. Проверка тела ответа
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var response struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(body, &response))

		assert.False(t, response.Success)
		assert.Equal(t, expectedError.Error(), response.Error)

		// 8. Проверка, что мок, был вызван
		mockService.AssertExpectations(t)
	})
	// Тест на неверный формат ID
	t.Run("should return error when invalid role ID format", func(t *testing.T) {
		// 1. Подготовка данных с невалидным ID
		idParam := "abc"

		// 2. Создаем запрос
		req := httptest.NewRequest("DELETE", "/api/v1/roles/"+idParam, nil)
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
	t.Run("should return success when delete ALl roles by Ids", func(t *testing.T) {
		// Подготовка тестовых данных
		requestIDs := []int64{1, 2, 3}
		idParam := "1,2,3" // ID как строка с разделителем-запятой

		// 1. Создаем запрос с query-параметром
		// 3. Создание КОРРЕКТНОГО запроса (с /ids в пути)
		req := httptest.NewRequest(
			"DELETE",
			"/api/v1/roles/ids?ids="+idParam, // Добавлен /ids перед ?
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
		defer closeBody(t, resp.Body)

		// 4. Проверка статуса
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
		// 1. Создаем запрос БЕЗ параметра ids
		req := httptest.NewRequest("DELETE", "/api/v1/roles/ids", nil)
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
	t.Run("should return error when invalid role IDs format", func(t *testing.T) {
		// 1. Подготовка данных с невалидным ID
		idParam := "1,abc,3"

		// 2. Создаем запрос
		req := httptest.NewRequest("DELETE", "/api/v1/roles/ids?ids="+idParam, nil)
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
		// 1. Подготовка данных
		requestIDs := []int64{1, 2, 3}
		idParam := "1,2,3"
		expectedError := errors.New("database error")

		// 2. Настройка мока на возврат ошибки
		mockService.On("DeleteByIds", requestIDs).Return(Response{}, expectedError).Once()

		// 3. Создание запроса
		req := httptest.NewRequest("DELETE", "/api/v1/roles/ids?ids="+idParam, nil)
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
		t.Errorf("failed to close body: %v", err)
	}
}
