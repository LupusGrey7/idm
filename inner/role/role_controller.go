package role

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"
	"strings"
)

type Controller struct {
	server      *web.Server
	roleService Svc
}

// NewController - функция-конструктор
func NewController(server *web.Server, roleService Svc) *Controller {
	return &Controller{
		server:      server,
		roleService: roleService,
	}
}

type Svc interface {
	FindById(id int64) (Response, error)
	CreateRole(request CreateRequest) (Response, error)
	UpdateRole(id int64, request UpdateRequest) (Response, error)
	FindAll() ([]Response, error)
	FindAllByIds(ids []int64) ([]Response, error)
	DeleteById(id int64) (Response, error)
	DeleteByIds(ids []int64) (Response, error)
}

// RegisterRoutes - функция для регистрации маршрутов
func (c *Controller) RegisterRoutes() {
	// полный маршрут получится "/api/v1/roles"
	c.server.GroupEmployees.Get("/", c.FindAll)
	c.server.GroupEmployees.Get("/ids", c.FindAllByIds)
	c.server.GroupEmployees.Get("/:id", c.FindById)
	c.server.GroupEmployees.Post("/", c.CreateRole)
	c.server.GroupEmployees.Put("/:id", c.UpdateRole)
	c.server.GroupEmployees.Delete("/ids", c.DeleteByIds)
	c.server.GroupEmployees.Delete("/:id", c.DeleteById)
}

// -- функции-хендлеры, которые будут вызываться при POST\GET... запросе по маршруту "/api/v1/employees" --//

func (c *Controller) FindAll(ctx *fiber.Ctx) {

	// вызываем сервис role.Service
	var response, err = c.roleService.FindAll()
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
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning find All roles")
		return
	}

	return
}

func (c *Controller) FindById(ctx *fiber.Ctx) {
	// Анмаршалим path var запроса
	idStr := ctx.Params("id")
	roleID, errConv := strconv.ParseInt(idStr, 10, 64)
	if errConv != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, errConv.Error())
		return
	}

	var response, err = c.roleService.FindById(roleID)
	if err != nil {
		switch {
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())

		default:
			_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
		return
	}

	if err = common.OkResponse(ctx, response); err != nil {

		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning employee by ID")
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

	// Конвертируем в массив чисел
	var ids []int64
	for _, idStr := range strings.Split(idsParam, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid ID format: "+idStr)
			return
		}
		ids = append(ids, id)
	}

	var response, err = c.roleService.FindAllByIds(ids)
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
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning find All roles")
		return
	}

	return
}

// CreateRole - функция-хендлер, которая будет вызываться при POST запросе по маршруту "/api/v1/roles"
func (c *Controller) CreateRole(ctx *fiber.Ctx) {

	// Анмаршалим JSON body запроса в структуру CreateRequest
	var request CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		return
	}

	// вызываем метод сервиса role.Service
	var newRoleId, err = c.roleService.CreateRole(request)
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
	if err = common.OkResponse(ctx, newRoleId); err != nil {

		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning created role id")
		return
	}
}

func (c *Controller) UpdateRole(ctx *fiber.Ctx) {

	idStr := ctx.Params("id")
	employeeID, errConv := strconv.ParseInt(idStr, 10, 64)
	if errConv != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, errConv.Error())
		return
	}

	var request UpdateRequest
	if err := ctx.BodyParser(&request); err != nil { // Анмаршалим JSON body запроса в структуру CreateRequest
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		return
	}

	var updatedEmployee, err = c.roleService.UpdateRole(employeeID, request)
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
		gotAnsw := fmt.Sprintf("error returning updated role with id %s", idStr)
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, gotAnsw)
		return
	}
	return
}

func (c *Controller) DeleteById(ctx *fiber.Ctx) {
	idStr := ctx.Params("id")
	roleID, errConv := strconv.ParseInt(idStr, 10, 64)
	if errConv != nil {
		//ctx.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, errConv.Error())
		return
	}

	var response, err = c.roleService.DeleteById(roleID)
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
		gotAnsw := fmt.Sprintf("error deleting role by id %s", idStr)
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

	// Конвертируем в массив чисел
	var ids []int64
	for _, idStr := range strings.Split(idsParam, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			_ = common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid ID format: "+idStr)
			return
		}
		ids = append(ids, id)
	}

	var response, err = c.roleService.DeleteByIds(ids)
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
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning deleting All roles")
		return
	}

	return
}
