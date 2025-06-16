package employee

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber"
	"github.com/jmoiron/sqlx"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"
	"strings"
)

type Controller struct {
	server          *web.Server
	employeeService Svc
}

// Svc - интерфейс сервиса employee.Service
type Svc interface {
	FindAll() ([]Response, error)
	FindById(id int64) (Response, error)
	FindAllByIds(ids []int64) ([]Response, error)
	CreateEmployee(request CreateRequest) (int64, error)
	CreateEmployeeTx(request Entity) (int64, error)
	Update(entity *Entity) (Response, error)
	DeleteById(id int64) (Response, error)
	DeleteByIds(ids []int64) (Response, error)
	FindEmployeeByNameTx(name string) (bool, err error)
	closeTx(*sqlx.Tx, error, string)
}

// NewController - функция-конструктор
func NewController(server *web.Server, employeeService Svc) *Controller {
	return &Controller{
		server:          server,
		employeeService: employeeService,
	}
}

// RegisterRoutes - функция для регистрации маршрутов
func (c *Controller) RegisterRoutes() {
	// полный маршрут получится "/api/v1/employees"
	c.server.GroupEmployees.Get("/", c.FindAll)
	c.server.GroupEmployees.Get("/ids", c.FindAllByIds)
	c.server.GroupEmployees.Get("/:id", c.FindById)
	c.server.GroupEmployees.Post("/", c.CreateEmployee)
	c.server.GroupEmployees.Post("/employees", c.CreateEmployeeTx)
	c.server.GroupEmployees.Put("/:id", c.Update)
	c.server.GroupEmployees.Delete("/ids", c.DeleteByIds)
	c.server.GroupEmployees.Delete("/:id", c.DeleteById)
}

// -- функции-хендлеры, которые будут вызываться при POST\GET... запросе по маршруту "/api/v1/employees" --//

// CreateEmployee - функция-хендлер, которая будет вызываться при POST запросе по маршруту "/api/v1/employees"
func (c *Controller) CreateEmployee(ctx *fiber.Ctx) {

	// Анмаршалим JSON body запроса в структуру CreateRequest
	var request CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		return
	}

	// вызываем метод CreateEmployee сервиса employee.Service
	var newEmployeeId, err = c.employeeService.CreateEmployee(request)
	if err != nil {
		switch {

		// если сервис возвращает ошибку RequestValidationError или AlreadyExistsError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())

		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, newEmployeeId); err != nil {

		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning created employee id")
		return
	}
}

func (c *Controller) FindById(ctx *fiber.Ctx) {
	// Анмаршалим path var запроса
	idStr := ctx.Params("id")
	employeeID, errConv := strconv.ParseInt(idStr, 10, 64)
	if errConv != nil {
		//ctx.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, errConv.Error())
		return
	}

	// вызываем метод CreateEmployee сервиса employee.Service
	var response, err = c.employeeService.FindById(employeeID)
	if err != nil {
		switch {

		// если сервис возвращает ошибку RequestValidationError или AlreadyExistsError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())

		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, response); err != nil {

		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning employee by ID")
		return
	}

	return
}

func (c *Controller) FindAll(ctx *fiber.Ctx) {

	// вызываем сервис employee.Service
	var response, err = c.employeeService.FindAll()
	if err != nil {
		switch {

		// если сервис возвращает ошибку RequestValidationError или AlreadyExistsError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())

		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, response); err != nil {

		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning find All employees")
		return
	}

	return
}

func (c *Controller) FindAllByIds(ctx *fiber.Ctx) {
	idsParam := ctx.Query("ids") //query параметр ?ids=1,2,3
	if idsParam == "" {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "Missing ids parameter")
		return
	}

	// Конвертируем в массив чисел - TODO duplicate
	var ids []int64
	for _, idStr := range strings.Split(idsParam, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid ID format: "+idStr)
			return
		}
		ids = append(ids, id)
	}

	var response, err = c.employeeService.FindAllByIds(ids)
	if err != nil {
		switch {

		// если сервис возвращает ошибку RequestValidationError или AlreadyExistsError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())

		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, response); err != nil {

		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning find All employees")
		return
	}

	return
}

func (c *Controller) Update(ctx *fiber.Ctx) {

	idStr := ctx.Params("id")
	employeeID, errConv := strconv.ParseInt(idStr, 10, 64)
	if errConv != nil {
		//ctx.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, errConv.Error())
		return
	}

	var request UpdateRequest
	if err := ctx.BodyParser(&request); err != nil { // Анмаршалим JSON body запроса в структуру CreateRequest
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		return
	}

	var employeeEntity = request.ToEntity()
	employeeEntity.Id = employeeID
	// вызываем метод CreateEmployee сервиса employee.Service
	var updatedEmployee, err = c.employeeService.Update(employeeEntity)
	if err != nil {
		switch {

		// если сервис возвращает ошибку RequestValidationError или AlreadyExistsError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())

		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, updatedEmployee); err != nil {

		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		gotAnsw := fmt.Sprintf("error returning updated employee with id %s", idStr)
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, gotAnsw)
		return
	}
	return
}
func (c *Controller) DeleteById(ctx *fiber.Ctx) {
	idStr := ctx.Params("id")
	employeeID, errConv := strconv.ParseInt(idStr, 10, 64)
	if errConv != nil {
		//ctx.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, errConv.Error())
		return
	}
	var response, err = c.employeeService.DeleteById(employeeID)
	if err != nil {
		switch {

		// если сервис возвращает ошибку RequestValidationError или AlreadyExistsError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())

		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, response); err != nil {

		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		gotAnsw := fmt.Sprintf("error deleting employee by id %s", idStr)
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, gotAnsw)
		return
	}

	return
}
func (c *Controller) DeleteByIds(ctx *fiber.Ctx) {
	idsParam := ctx.Query("ids") //query параметр ?ids=1,2,3
	if idsParam == "" {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "Missing ids parameter")
		return
	}

	// Конвертируем в массив чисел - FIXME duplicate
	var ids []int64
	for _, idStr := range strings.Split(idsParam, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid ID format: "+idStr)
			return
		}
		ids = append(ids, id)
	}

	var response, err = c.employeeService.DeleteByIds(ids)
	if err != nil {
		switch {

		// если сервис возвращает ошибку RequestValidationError или AlreadyExistsError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())

		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, response); err != nil {

		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning deleting All employees")
		return
	}

	return
}

// TODO
func (c *Controller) CreateEmployeeTx(ctx *fiber.Ctx) {
	return
}
func (c *Controller) FindEmployeeByNameTx(ctx *fiber.Ctx) {
	return
}
