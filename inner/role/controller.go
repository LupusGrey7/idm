package role

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"idm/inner/domain"

	"idm/inner/http"
	"idm/inner/web"
	"log"
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

// -- функции-хендлеры, которые будут вызываться при POST\GET... запросе по маршруту "/transport/v1/employees" --//

func (c *Controller) FindAll(ctx *fiber.Ctx) error {
	response, err := c.roleService.FindAll()
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

func (c *Controller) FindById(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	roleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	response, err := c.roleService.FindById(roleID)
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

	response, err := c.roleService.FindAllByIds(ids)
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

func (c *Controller) CreateRole(ctx *fiber.Ctx) error {
	var request CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		log.Printf("CreateRole: body parse error: %v", err)
		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidRequestBody)
	}

	newRoleId, err := c.roleService.CreateRole(request)
	if err != nil {
		switch {
		case errors.As(err, &domain.RequestValidationError{}), errors.As(err, &domain.AlreadyExistsError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	return http.CreatedResponse(ctx, newRoleId)
}

func (c *Controller) UpdateRole(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	roleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	var request UpdateRequest
	if err := ctx.BodyParser(&request); err != nil {
		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidRequestBody)
	}

	updatedRole, err := c.roleService.UpdateRole(roleID, request)
	if err != nil {
		switch {
		case errors.As(err, &domain.RequestValidationError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	return http.OkResponse(ctx, updatedRole)
}

func (c *Controller) DeleteById(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	roleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	response, err := c.roleService.DeleteById(roleID)
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

func (c *Controller) DeleteByIds(ctx *fiber.Ctx) error {
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

	response, err := c.roleService.DeleteByIds(ids)
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
