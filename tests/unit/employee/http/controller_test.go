package http

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"idm/inner/web"
	"io"

	"github.com/gofiber/fiber"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"idm/inner/employee"
	"idm/tests/unit/mocks"
	"net/http/httptest"
	"testing"
	"time"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) FindAll() ([]employee.Response, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockService) FindAllByIds(ids []int64) ([]employee.Response, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockService) CreateEmployeeTx(request employee.Entity) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockService) UpdateEmployee(id int64, request employee.UpdateRequest) (employee.Response, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockService) DeleteById(id int64) (employee.Response, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockService) DeleteByIds(ids []int64) (employee.Response, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockService) FindEmployeeByNameTx(name string) (bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockService) closeTx(tx *sqlx.Tx, err error, s string) {
	//TODO implement me
	panic("implement me")
}

func (m *MockService) FindById(id int64) (employee.Response, error) {
	args := m.Called(id)
	return args.Get(0).(employee.Response), args.Error(1)
}

func (m *MockService) CreateEmployee(req employee.CreateRequest) (employee.Response, error) {
	args := m.Called(req)
	return args.Get(0).(employee.Response), args.Error(1)
}

// ... аналогично для других методов

func TestEmployeeController_FindById(t *testing.T) {

	// 1. Инициализация
	app := fiber.New()

	server := &web.Server{
		App:            app,
		GroupApiV1:     app.Group("/api/v1"),
		GroupEmployees: app.Group("/api/v1/employees"),
	}

	mockService := new(mocks.MockEmployeeService)
	ctrl := employee.NewController(server, mockService)
	server.GroupEmployees.Get("/:id", ctrl.FindById)

	// 6. Тестовый случай
	t.Run("success", func(t *testing.T) {
		testID := int64(1)
		testName := "John Sena"
		now := time.Now().UTC().Truncate(time.Second)

		expectedData := employee.Response{
			Id:       testID,
			Name:     testName,
			CreateAt: now,
			UpdateAt: now,
		}

		// 3. Настройка мока
		mockService.On("FindById", testID).Return(expectedData, nil)

		// 4. Выполнение запроса
		req := httptest.NewRequest("GET", "/api/v1/employees/1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// 5. Проверка сырого ответа
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Raw response: %s", string(body))

		// 6. Проверки
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Декодируем в структуру-обёртку
		var responseWrapper struct {
			Success bool              `json:"success"`
			Data    employee.Response `json:"data"`
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
}

//	t.Run("validation error", func(t *testing.T) {
//		// Arrange
//		validationErr := errors.New("ID must be positive")
//		validator.ExpectValidate(employee.FindByIDRequest{ID: 0}, validationErr)
//
//		// Act
//		req := httptest.NewRequest("GET", "/employees/0", nil)
//		resp, _ := app.Test(req)
//
//		// Assert
//		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
//
//		var errorResponse = common.ErrorResponse
//		err := json.NewDecoder(resp.Body).Decode(&errorResponse)
//		assert.NoError(t, err)
//		assert.Contains(t, errorResponse.Message, "ID must be positive")
//
//		service.AssertNotCalled(t, "FindById")
//		validator.AssertExpectations(t)
//	})
//}
//
//func TestEmployeeController_CreateEmployee(t *testing.T) {
//	app := fiber.New()
//	service := new(MockService)
//	validator := new(mocks.MockValidator)
//	ctrl := employee.NewController(service, validator)
//
//	app.Post("/employees", ctrl.CreateEmployee)
//
//	t.Run("success", func(t *testing.T) {
//		// Arrange
//		requestBody := employee.CreateRequest{Name: "John Doe"}
//		response := employee.Response{Id: 1, Name: "John Doe"}
//
//		service.On("CreateEmployee", requestBody).Return(response, nil)
//		validator.ExpectValidate(requestBody, nil)
//
//		body, _ := json.Marshal(requestBody)
//		req := httptest.NewRequest("POST", "/employees", bytes.NewReader(body))
//		req.Header.Set("Content-Type", "application/json")
//
//		// Act
//		resp, _ := app.Test(req)
//
//		// Assert
//		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
//
//		var createdResponse employee.Response
//		err := json.NewDecoder(resp.Body).Decode(&createdResponse)
//		assert.NoError(t, err)
//		assert.Equal(t, response, createdResponse)
//
//		service.AssertExpectations(t)
//		validator.AssertExpectations(t)
//	})
//
//	t.Run("invalid JSON", func(t *testing.T) {
//		// Arrange
//		req := httptest.NewRequest("POST", "/employees", bytes.NewReader([]byte(`{"name": 123}`)))
//		req.Header.Set("Content-Type", "application/json")
//
//		// Act
//		resp, _ := app.Test(req)
//
//		// Assert
//		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
//		service.AssertNotCalled(t, "CreateEmployee")
//		validator.AssertNotCalled(t, "Validate")
//	})
//
//	t.Run("validation error", func(t *testing.T) {
//		// Arrange
//		requestBody := employee.CreateRequest{Name: ""}
//		validationErr := errors.New("Name is required")
//
//		body, _ := json.Marshal(requestBody)
//		req := httptest.NewRequest("POST", "/employees", bytes.NewReader(body))
//		req.Header.Set("Content-Type", "application/json")
//
//		validator.ExpectValidate(requestBody, validationErr)
//
//		// Act
//		resp, _ := app.Test(req)
//
//		// Assert
//		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
//
//		var errorResponse common.ErrorResponse
//		err := json.NewDecoder(resp.Body).Decode(&errorResponse)
//		assert.NoError(t, err)
//		assert.Contains(t, errorResponse.Message, "Name is required")
//
//		service.AssertNotCalled(t, "CreateEmployee")
//		validator.AssertExpectations(t)
//	})
//}

// Аналогично для других методов контроллера: Update, DeleteById, FindAllByIds и т.д.
