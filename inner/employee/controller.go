package employee

import (
	"errors"
	"github.com/gofiber/fiber/v2" // Версия 2 - позволяет выводить ошибку
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/domain"
	"idm/inner/http"
	"idm/inner/web"
	"strconv"
	"strings"
)

const (
	invalidRequestFormat = "Invalid request format"
	validationFailed     = "Validation failed"
	internalServerError  = "Internal server error"
	invalidIDFormat      = "Invalid ID format"
	invalidRequestBody   = "Invalid request body"
)

type Controller struct {
	server          *web.Server
	employeeService Svc
	logger          *common.Logger
}

// Svc - интерфейс сервиса employee.Service
type Svc interface {
	FindAll() ([]Response, error)
	FindById(id int64) (Response, error)
	FindAllByIds(ids []int64) ([]Response, error)
	CreateEmployee(request CreateRequest) (Response, error)
	CreateEmployeeTx(request CreateRequest) (int64, error)
	UpdateEmployee(id int64, request UpdateRequest) (Response, error)
	DeleteById(id int64) (Response, error)
	DeleteByIds(ids []int64) (Response, error)
	FindEmployeeByNameTx(name string) (bool, error)
	CloseTx(*sqlx.Tx, error, string)
}

// NewController - функция-конструктор
func NewController(
	server *web.Server,
	employeeService Svc,
	logger *common.Logger,
) *Controller {
	return &Controller{
		server:          server,
		employeeService: employeeService,
		logger:          logger,
	}
}

// RegisterRoutes - функция для регистрации маршрутов
func (c *Controller) RegisterRoutes() {
	// полный маршрут получится "/transport/v1/employees"
	c.server.GroupEmployees.Get("/", c.FindAll)
	c.server.GroupEmployees.Get("/ids", c.FindAllByIds)
	c.server.GroupEmployees.Get("/:id", c.FindById)
	c.server.GroupEmployees.Post("/", c.CreateEmployee)
	c.server.GroupEmployees.Post("/employee", c.CreateEmployeeTx)
	c.server.GroupEmployees.Put("/:id", c.Update)
	c.server.GroupEmployees.Delete("/ids", c.DeleteByIds)
	c.server.GroupEmployees.Delete("/:id", c.DeleteById)
}

// -- функции-хендлеры, которые будут вызываться при POST\GET... запросе по маршруту "/transport/v1/employees" --//

func (c *Controller) CreateEmployee(ctx *fiber.Ctx) error {
	var request CreateRequest

	// Парсинг тела запроса
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("CreateEmployee: body parse error", zap.Error(err))
		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidRequestFormat)

	}
	// логируем тело запроса
	c.logger.Debug("create employee: received request", zap.Any("request", request))

	// Вызов сервиса
	newEmployee, err := c.employeeService.CreateEmployee(request)
	if err != nil {

		c.logger.Error("When create employee was error", zap.Error(err)) // логируем ошибку
		switch {                                                         // Обработка ошибок с использованием ваших функций
		case errors.Is(err, domain.ErrValidation):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, validationFailed)

		case errors.Is(err, domain.ErrConflict):
			return http.ErrResponse(ctx, fiber.StatusConflict, "Employee already exists")

		default:
			c.logger.Error("CreateEmployee service error: %s", zap.Error(err))
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, internalServerError)
		}
	}

	// Успешный ответ с использованием ваших функций
	return http.CreatedResponse(ctx, newEmployee)
}

func (c *Controller) FindById(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id") // Анмаршалим path var запроса
	employeeID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	response, err := c.employeeService.FindById(employeeID)
	if err != nil {
		switch {
		case errors.As(err, &domain.RequestValidationError{}), errors.As(err, &domain.AlreadyExistsError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			c.logger.Error("When Employee FindById was error: %s", zap.Error(err))
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, internalServerError)
		}
	}

	return http.OkResponse(ctx, map[string]interface{}{
		"id":       response.Id,
		"name":     response.Name,
		"createAt": response.CreateAt,
		"updateAt": response.UpdateAt,
	})
}

func (c *Controller) FindAll(ctx *fiber.Ctx) error {
	response, err := c.employeeService.FindAll()
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrFindAllFailed):
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, "Failed to find all employees")
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return http.OkResponse(ctx, response)
}

func (c *Controller) FindAllByIds(ctx *fiber.Ctx) error {
	idsParam := ctx.Query("ids")
	if idsParam == "" {
		return http.ErrResponse(ctx, fiber.StatusBadRequest, "Missing ids parameter")
	}

	var ids []int64
	for _, idStr := range strings.Split(idsParam, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat+idStr)
		}
		ids = append(ids, id)
	}

	response, err := c.employeeService.FindAllByIds(ids)
	if err != nil {
		switch {
		case errors.As(err, &domain.RequestValidationError{}), errors.As(err, &domain.AlreadyExistsError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	return http.OkResponse(ctx, response)
}

func (c *Controller) Update(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	employeeID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	var request UpdateRequest
	if err := ctx.BodyParser(&request); err != nil {
		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidRequestBody)
	}

	updatedEmployee, err := c.employeeService.UpdateEmployee(employeeID, request)
	if err != nil {
		switch {
		case errors.As(err, &domain.RequestValidationError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	return http.OkResponse(ctx, updatedEmployee)
}

func (c *Controller) DeleteById(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	employeeID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	response, err := c.employeeService.DeleteById(employeeID)
	if err != nil {
		switch {
		case errors.As(err, &domain.RequestValidationError{}), errors.As(err, &domain.AlreadyExistsError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	return http.OkResponse(ctx, response)
}

func (c *Controller) DeleteByIds(ctx *fiber.Ctx) error {
	idsParam := ctx.Query("ids")
	if idsParam == "" {
		return http.ErrResponse(ctx, fiber.StatusBadRequest, "Missing ids parameter")
	}

	var ids []int64
	for _, idStr := range strings.Split(idsParam, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return http.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid ID format"+idStr)
		}
		ids = append(ids, id)
	}

	response, err := c.employeeService.DeleteByIds(ids)
	if err != nil {
		switch {
		case errors.As(err, &domain.RequestValidationError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	return http.OkResponse(ctx, response)
}

func (c *Controller) CreateEmployeeTx(ctx *fiber.Ctx) error {
	var request CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidRequestBody)
	}

	response, err := c.employeeService.CreateEmployeeTx(request)
	if err != nil {
		switch {
		case errors.As(err, &domain.RequestValidationError{}), errors.As(err, &domain.AlreadyExistsError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	return http.OkResponse(ctx, response)
}
