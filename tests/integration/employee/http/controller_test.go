package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	var db = testutils.InitTestDB() //var db = database.ConnectDb()
	defer db.Close()

	// Проверяем соединение
	require.NoError(t, db.Ping())

	// 2. Подготавливаем БД
	//_, err = db.Exec("TRUNCATE TABLE employees RESTART IDENTITY CASCADE")
	//require.NoError(t, err)

	// 3. Создаем тестовые данные
	repo := employee.NewRepository(db) // Ваш реальный репозиторий
	for i := 1; i <= 3; i++ {
		var request = employee.CreateRequest{Name: fmt.Sprintf("Employee %d", i)}
		var toEntity = request.ToEntity()
		_, err := repo.CreateEmployee(
			context.Background(),
			toEntity,
		)
		require.NoError(t, err)
	}
	//Сброс моков перед каждым тестом (рекомендуется)
	//t.Cleanup(func() {
	//	db.Exec("DELETE FROM employees WHERE id IN (SELECT id FROM employees WHERE name LIKE 'Employee %')")
	//})

	//1. считывание конфигурации
	// Получаем путь к директории текущего файла
	//_, filename, _, _ := runtime.Caller(0)
	//dir := filepath.Dir(filename)
	// Поднимаемся на 1-н уровень вверх (из testutils в tests)
	//baseDir := filepath.Join(dir, "../..")

	// Формируем путь к .env.test
	//envPath := filepath.Join(baseDir, "tests/testdata", ".env.test")
	cfg := config.GetConfig(".env.test")
	var logger = common.NewLogger(cfg) // Создаем логгер
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

		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=1&pageSize=3", nil)

		// Перед app.Test(req)
		routes := app.GetRoutes()
		for _, route := range routes {
			t.Logf("Route: %s %s", route.Method, route.Path) //- Route: DELETE /api/v1/employees/ids
		}

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

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
		assert.Equal(t, int64(3), response.Data.PageNumber)
		assert.Equal(t, int64(1), response.Data.PageSize)
		assert.Equal(t, int64(5), response.Data.Total)

		//Очистка После тестом (не в Cleanup)
		_, err1 := db.Exec("DELETE FROM employees WHERE name LIKE 'Employee %'")
		require.NoError(t, err1)
	})

	t.Run("second page with 3 items", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=2&pageSize=3", nil)
		resp, _ := app.Test(req)
		defer resp.Body.Close()

		var response struct {
			Data struct {
				Result []employee.Response `json:"result"`
			} `json:"data"`
		}

		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &response)

		assert.Len(t, response.Data.Result, 2)
	})

	t.Run("third page with 3 items", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=3&pageSize=3", nil)
		resp, _ := app.Test(req)
		defer resp.Body.Close()

		var response struct {
			Data struct {
				Result []employee.Response `json:"result"`
			} `json:"data"`
		}

		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &response)

		assert.Len(t, response.Data.Result, 0)
	})

	t.Run("invalid request parameters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=abc&pageSize=10", nil)
		resp, _ := app.Test(req)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var errorResp map[string]interface{}
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &errorResp)

		assert.Contains(t, errorResp["error"], "invalid page values format")
	})

	t.Run("missing pageNumber", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageSize=3", nil)
		resp, _ := app.Test(req)
		defer resp.Body.Close()

		var response struct {
			Data struct {
				PageNumber int64 `json:"page_number"`
			} `json:"data"`
		}

		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &response)

		// Ожидаем дефолтное значение pageNumber = 1
		assert.Equal(t, int64(1), response.Data.PageNumber)
	})

	t.Run("missing pageSize", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=1", nil)
		resp, _ := app.Test(req)
		defer resp.Body.Close()

		var response struct {
			Data struct {
				PageSize int64 `json:"page_size"`
			} `json:"data"`
		}

		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &response)

		// Ожидаем дефолтное значение pageSize = 10
		assert.Equal(t, int64(10), response.Data.PageSize)
	})
}
