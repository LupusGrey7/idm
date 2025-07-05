package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/config"
	"idm/inner/employee"
	"idm/inner/validator"
	"idm/inner/web"
	"idm/inner/web/middleware"
	"idm/tests/testutils"
	"io"
	"net/http/httptest"
	"testing"
)

func TestEmployeePaginationIntegration(t *testing.T) {
	// 1. Подключаемся к тестовой БД
	a := assert.New(t)

	// 2. Подготавливаем БД
	var db = testutils.InitTestDB()
	defer func() { //var db = database.ConnectDb()
		if err := db.Close(); err != nil { // Закрываем соединение при ошибке
			log.Error(
				"failed to close database connection: %s",
				zap.Error(err),
			)
		}
	}()
	// Проверяем соединение
	require.NoError(t, db.Ping())

	// 3. Создаем тестовые данные
	repo := employee.NewRepository(db) // Ваш реальный репозиторий

	// config setUp
	cfg := config.GetConfig(".env.test")
	// Создаем Logger
	var logger = common.NewLogger(cfg)
	logger.Debug("-->> Start  Integration Test")

	app := fiber.New()
	// Инициализируем валидатор (если используется)
	validator := validator.NewValidator() // Замените на ваш валидатор

	// Инициализируем сервис и контроллер
	service := employee.NewService(repo, validator)
	// 1. Инициализация
	middleware.RegisterMiddleware(app, logger) //middleware

	server := &web.Server{
		App:            app,
		GroupApiV1:     app.Group("/api/v1"),
		GroupEmployees: app.Group("/api/v1/employees"),
	}

	ctrl := employee.NewController(server, service, logger)

	// Регистрируем роуты
	server.GroupEmployees.Get("/page", ctrl.GetAllPages)

	// 5. Тестовые случаи
	t.Run("first page with 3 items", func(t *testing.T) {
		initTestData(repo, t)
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=1&pageSize=3", nil)

		// Перед app.Test(req)
		routes := app.GetRoutes()
		for _, route := range routes {
			t.Logf("Route: %s %s", route.Method, route.Path) //- Route: DELETE /api/v1/employees/ids
		}

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"error"`
			Data    struct {
				Result     []employee.Response `json:"result"`
				PageSize   int64               `json:"page_size"`
				PageNumber int64               `json:"page_number"`
				Total      int64               `json:"total"`
			} `json:"data"`
		}
		// 5. Проверка сырого ответа
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Raw response: %s", string(body))

		a.Nil(err)
		a.NotEmpty(resp)

		errUnMrh := json.Unmarshal(body, &response)
		require.NoError(t, errUnMrh)

		assert.Len(t, response.Data.Result, 3)
		assert.Equal(t, int64(1), response.Data.PageNumber) //номер страницы (начиная с 0).
		assert.Equal(t, int64(3), response.Data.PageSize)   //количество записей на странице.
		assert.Equal(t, int64(5), response.Data.Total)

		//Очистка После тестом (не в Cleanup)
		_, err1 := db.Exec("DELETE FROM employees WHERE name LIKE 'Employee %'")
		require.NoError(t, err1)
	})

	t.Run("second page with 2 items", func(t *testing.T) {
		initTestData(repo, t)
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=2&pageSize=3", nil)

		resp, _ := app.Test(req)
		defer closeBody(t, resp.Body)

		// Перед app.Test(req)
		routes := app.GetRoutes()
		for _, route := range routes {
			t.Logf("Route: %s %s", route.Method, route.Path) //- Route: DELETE /api/v1/employees/ids
		}

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer closeBody(t, resp.Body)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"error"`
			Data    struct {
				Result     []employee.Response `json:"result"`
				PageSize   int64               `json:"page_size"`
				PageNumber int64               `json:"page_number"`
				Total      int64               `json:"total"`
			} `json:"data"`
		}
		// 5. Проверка сырого ответа
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Raw response: %s", string(body))

		errUnMrh := json.Unmarshal(body, &response)
		require.NoError(t, errUnMrh)

		assert.True(t, response.Success)
		assert.Empty(t, response.Message)

		assert.Len(t, response.Data.Result, 2)

		assert.Equal(t, int64(2), response.Data.PageNumber) //номер страницы (начиная с 0).
		assert.Equal(t, int64(3), response.Data.PageSize)   //количество записей на странице.
		assert.Equal(t, int64(5), response.Data.Total)

		//Очистка После тестом (не в Cleanup)
		_, err1 := db.Exec("DELETE FROM employees WHERE name LIKE 'Employee %'")
		require.NoError(t, err1)
	})

	t.Run("third page with 0 items", func(t *testing.T) {
		initTestData(repo, t)
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=3&pageSize=3", nil)

		resp, _ := app.Test(req)
		defer closeBody(t, resp.Body)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Перед app.Test(req)
		routes := app.GetRoutes()
		for _, route := range routes {
			t.Logf("Route: %s %s", route.Method, route.Path) //- Route: DELETE /api/v1/employees/ids
		}

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"error"`
			Data    struct {
				Result     []employee.Response `json:"result"`
				PageSize   int64               `json:"page_size"`
				PageNumber int64               `json:"page_number"`
				Total      int64               `json:"total"`
			} `json:"data"`
		}
		// 5. Проверка сырого ответа
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Raw response: %s", string(body))

		errUnMrh := json.Unmarshal(body, &response)
		require.NoError(t, errUnMrh)

		assert.True(t, response.Success)
		assert.Empty(t, response.Message)

		assert.Len(t, response.Data.Result, 0) //проверяем, что в ответе 0 записей

		assert.Equal(t, int64(3), response.Data.PageNumber) //номер страницы (начиная с 0).
		assert.Equal(t, int64(0), response.Data.PageSize)   //количество записей на странице.
		assert.Equal(t, int64(5), response.Data.Total)

		//Очистка После тестом (не в Cleanup)
		_, err1 := db.Exec("DELETE FROM employees WHERE name LIKE 'Employee %'")
		require.NoError(t, err1)
	})

	t.Run("invalid request parameters", func(t *testing.T) {
		initTestData(repo, t)

		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=abc&pageSize=10", nil)

		resp, _ := app.Test(req)
		defer closeBody(t, resp.Body)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// 5. Проверка сырого ответа
		var errorResp map[string]interface{}
		body, _ := io.ReadAll(resp.Body)
		errMrh := json.Unmarshal(body, &errorResp)
		require.NoError(t, errMrh)
		t.Logf("Raw response: %s", string(body))

		assert.Contains(t, errorResp["error"], "Invalid Page Values format")
		//Очистка После тестом (не в Cleanup)
		_, err1 := db.Exec("DELETE FROM employees WHERE name LIKE 'Employee %'")
		require.NoError(t, err1)
	})

	t.Run("missing pageNumber", func(t *testing.T) {
		initTestData(repo, t)

		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageSize=3", nil)

		resp, _ := app.Test(req)
		defer closeBody(t, resp.Body)

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"error"`
			Data    struct {
				Result     []employee.Response `json:"result"`
				PageSize   int64               `json:"page_size"`
				PageNumber int64               `json:"page_number"`
				Total      int64               `json:"total"`
			} `json:"data"`
		}

		body, _ := io.ReadAll(resp.Body)
		errMrh := json.Unmarshal(body, &response)
		require.NoError(t, errMrh)
		t.Logf("Raw response: %s", string(body))

		// Ожидаем дефолтное значение pageNumber = 1
		assert.Equal(t, int64(1), response.Data.PageNumber) //номер страницы (начиная с 0).
		assert.Equal(t, int64(3), response.Data.PageSize)   //количество записей на странице.
		assert.Equal(t, int64(5), response.Data.Total)

		//Очистка После тестом (не в Cleanup)
		_, err1 := db.Exec("DELETE FROM employees WHERE name LIKE 'Employee %'")
		require.NoError(t, err1)
	})

	t.Run("missing pageSize", func(t *testing.T) {
		initTestData(repo, t)

		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=1", nil)

		resp, _ := app.Test(req)
		defer closeBody(t, resp.Body)

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"error"`
			Data    struct {
				Result     []employee.Response `json:"result"`
				PageSize   int64               `json:"page_size"`
				PageNumber int64               `json:"page_number"`
				Total      int64               `json:"total"`
			} `json:"data"`
		}

		body, _ := io.ReadAll(resp.Body)
		errMrh := json.Unmarshal(body, &response)
		require.NoError(t, errMrh)
		t.Logf("Raw response: %s", string(body))

		// Ожидаем дефолтное значение pageSize = 10
		assert.Equal(t, int64(1), response.Data.PageNumber) //номер страницы (начиная с 0).
		// Ожидаем дефолтное значение pageSize = 10
		assert.Equal(t, int64(10), response.Data.PageSize) //количество записей на странице

		assert.Equal(t, int64(5), response.Data.Total)

		//Очистка После тестом (не в Cleanup)
		_, err1 := db.Exec("DELETE FROM employees WHERE name LIKE 'Employee %'")
		require.NoError(t, err1)
	})

	t.Run("missing pageNumber and pageSize", func(t *testing.T) {
		initTestData(repo, t)

		req := httptest.NewRequest("GET", "/api/v1/employees/page?", nil)
		resp, _ := app.Test(req)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				if err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}
		}(resp.Body)

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"error"`
			Data    struct {
				Result     []employee.Response `json:"result"`
				PageSize   int64               `json:"page_size"`
				PageNumber int64               `json:"page_number"`
				Total      int64               `json:"total"`
			} `json:"data"`
		}

		body, _ := io.ReadAll(resp.Body)
		errMrh := json.Unmarshal(body, &response)
		require.NoError(t, errMrh)
		t.Logf("Raw response: %s", string(body))

		// Ожидаем дефолтное значение pageSize = 10
		assert.Equal(t, int64(1), response.Data.PageNumber) //номер страницы (начиная с 0).
		// Ожидаем дефолтное значение pageSize = 10
		assert.Equal(t, int64(10), response.Data.PageSize) //количество записей на странице

		assert.Equal(t, int64(5), response.Data.Total)

		//Очистка После тестом (не в Cleanup)
		_, err1 := db.Exec("DELETE FROM employees WHERE name LIKE 'Employee %'")
		require.NoError(t, err1)
	})
}

func initTestData(repo *employee.Repository, t *testing.T) {
	for i := 1; i <= 3; i++ {
		var request = employee.CreateRequest{Name: fmt.Sprintf("Employee %d", i)}
		var toEntity = request.ToEntity()
		_, err := repo.CreateEmployee(
			context.Background(),
			toEntity,
		)
		require.NoError(t, err)
	}
}

func closeBody(t *testing.T, body io.ReadCloser) {
	if err := body.Close(); err != nil {
		t.Errorf("Error closing response body: %v", err)
	}
}
