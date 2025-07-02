package role

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
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
	invalidParseIDs      = "When the parse request parameter an FindAll Role By IDs ended with an error"
)

type Controller struct {
	server      *web.Server
	roleService Svc
	logger      *common.Logger
}

// NewController - функция-конструктор
func NewController(
	server *web.Server,
	roleService Svc,
	logger *common.Logger,
) *Controller {
	return &Controller{
		server:      server,
		roleService: roleService,
		logger:      logger,
	}
}

type Svc interface {
	FindById(ctx context.Context, id int64) (Response, error)
	CreateRole(ctx context.Context, request CreateRequest) (Response, error)
	UpdateRole(ctx context.Context, id int64, request UpdateRequest) (Response, error)
	FindAll(ctx context.Context) ([]Response, error)
	FindAllByIds(ctx context.Context, ids []int64) ([]Response, error)
	DeleteById(ctx context.Context, id int64) (Response, error)
	DeleteByIds(ctx context.Context, ids []int64) (Response, error)
}

// RegisterRoutes - функция для регистрации маршрутов
func (c *Controller) RegisterRoutes() {
	// полный маршрут получится "/api/v1/roles"
	c.server.GroupRoles.Get("/", c.FindAll)
	c.server.GroupRoles.Get("/ids", c.FindAllByIds)
	c.server.GroupRoles.Get("/:id", c.FindById)
	c.server.GroupRoles.Post("/", c.CreateRole)
	c.server.GroupRoles.Put("/:id", c.UpdateRole)
	c.server.GroupRoles.Delete("/ids", c.DeleteByIds)
	c.server.GroupRoles.Delete("/:id", c.DeleteById)
}

// -- функции-хендлеры, которые будут вызываться при POST\GET... запросе по маршруту "/transport/v1/employees" --//

func (c *Controller) FindAll(ctx *fiber.Ctx) error {
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	response, err := c.roleService.FindAll(appContext)
	if err != nil {
		c.logger.Error(
			"FindAll ended with error",           // Сообщение без форматирования - zap сам обработает
			zap.Error(err),                       // Ошибка
			zap.String("request_id", requestId),  // Добавляем request_id в лог
			zap.Int("roles_size", len(response)), // Более явное название поля
		)

		switch {
		case errors.As(err, &domain.RequestValidationError{}), errors.As(err, &domain.AlreadyExistsError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	c.logger.Debug(
		"Get All Roles have size",
		zap.String("request_id", requestId),
		zap.Int("roles_size", len(response)),
	)

	return http.OkResponse(ctx, response)
}

func (c *Controller) FindById(ctx *fiber.Ctx) error {
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	idStr := ctx.Params("id")
	roleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.logger.Error(
			"ID parse error when get role",
			zap.Error(err),
			zap.String("id", idStr),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	response, err := c.roleService.FindById(appContext, roleID)
	if err != nil {
		c.logger.Error(
			"Failed to get roles By ID",
			zap.Error(err),
			zap.Int64("id", roleID),
			zap.String("request_id", requestId),
		)

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
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	idsParam := ctx.Query("ids")
	if idsParam == "" {
		c.logger.Error(
			invalidParseIDs,
			zap.String("ids", idsParam),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, "Missing ids parameter")
	}

	var ids []int64
	for _, idStr := range strings.Split(idsParam, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.logger.Error(invalidParseIDs,
				zap.Int("ids", len(idsParam)),
				zap.String("request_id", requestId),
			)

			return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat+idStr)
		}
		ids = append(ids, id)
	}

	response, err := c.roleService.FindAllByIds(appContext, ids)
	if err != nil {
		c.logger.Error("When the parse request parameter an FindAll Role By IDs ended with an error",
			zap.Int("ids", len(idsParam)),
			zap.String("request_id", requestId),
		)
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
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	var request CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error(
			"body parse error when create role",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidRequestBody)
	}

	newRoleId, err := c.roleService.CreateRole(appContext, request)
	if err != nil {
		c.logger.Error(
			"When the create role ended with an error",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

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
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	idStr := ctx.Params("id")
	roleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.logger.Error("ID parse error when Update Role",
			zap.Error(err),
			zap.String("id", idStr),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	var request UpdateRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error(
			"body parse error when update role",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidRequestBody)
	}

	updatedRole, err := c.roleService.UpdateRole(appContext, roleID, request)
	if err != nil {
		c.logger.Error("When the update role ended with an error",
			zap.Error(err),
			zap.Int64("id", roleID),
			zap.String("request_id", requestId),
		)

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
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	idStr := ctx.Params("id")
	roleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.logger.Error("ID parse error when Delete Role",
			zap.Error(err),
			zap.String("id", idStr),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	response, err := c.roleService.DeleteById(appContext, roleID)
	if err != nil {
		c.logger.Error("When the delete role ended with an error",
			zap.Error(err),
			zap.Int64("id", roleID),
		)
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
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	idsParam := ctx.Query("ids")
	if idsParam == "" {
		c.logger.Error("When the parse request parameter an Delete Role By Ids ended with an error",
			zap.Int("ids", len(idsParam)),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, "Missing ids parameter")
	}

	var ids []int64
	for _, idStr := range strings.Split(idsParam, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.logger.Error("When the parse request parameter an Delete Role By Ids ended with an error",
				zap.Int("ids", len(idsParam)),
				zap.String("request_id", requestId),
			)

			return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat+idStr)
		}
		ids = append(ids, id)
	}

	response, err := c.roleService.DeleteByIds(appContext, ids)
	if err != nil {
		c.logger.Error("When the delete role ended with an error",
			zap.Error(err),
			zap.Int("ids", len(ids)),
			zap.String("request_id", requestId),
		)

		switch {
		case errors.As(err, &domain.RequestValidationError{}), errors.As(err, &domain.AlreadyExistsError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	return http.OkResponse(ctx, response)
}
