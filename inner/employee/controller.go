package employee

import (
	"context"
	"errors"
	"fmt"
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
	invalidRequestFormat    = "Invalid request format"
	validationFailed        = "validate name error"
	internalServerError     = "Internal server error"
	invalidIDFormat         = "Invalid ID format"
	invalidRequestBody      = "Invalid request body"
	invalidPageValuesFormat = "Invalid Page Values format"
)

// Controller (transport layer):
type Controller struct {
	server          *web.Server
	employeeService Svc
	logger          *common.Logger
}

// Svc - интерфейс сервиса Service class
type Svc interface {
	FindAll(ctx context.Context) ([]Response, error)
	FindById(ctx context.Context, id int64) (Response, error)
	FindAllByIds(ctx context.Context, ids []int64) ([]Response, error)
	GetAllByPage(ctx context.Context, req PageRequest) (PageResponse, error)
	CreateEmployee(ctx context.Context, request CreateRequest) (Response, error)
	CreateEmployeeTx(ctx context.Context, request CreateRequest) (int64, error)
	UpdateEmployee(ctx context.Context, id int64, request UpdateRequest) (Response, error)
	DeleteById(ctx context.Context, id int64) (Response, error)
	DeleteByIds(ctx context.Context, ids []int64) (Response, error)
	FindEmployeeByNameTx(ctx context.Context, name string) (bool, error)
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
	c.server.GroupEmployees.Get("/page", c.GetAllPages)
	c.server.GroupEmployees.Delete("/ids", c.DeleteByIds)
	c.server.GroupEmployees.Post("/", c.CreateEmployee)
	c.server.GroupEmployees.Post("/tx", c.CreateEmployeeTx)
	c.server.GroupEmployees.Get("/:id", c.FindById)
	c.server.GroupEmployees.Put("/:id", c.Update)
	c.server.GroupEmployees.Delete("/:id", c.DeleteById)
}

// -- функции-хендлеры, которые будут вызываться при POST\GET... запросе по маршруту "/api/v1/employees" --//

// CreateEmployee godoc
// @Summary      create a new employee
// @Description  Create a new employee
// @Tags 		 employee
// @Accept 		 json
// @Produce 	 json
// @Param 		 request body 	employee.CreateRequest true "Employee creation details"
// @Success 	 200  {object}  employee.Response	"Employee response"
// @Failure      400  {object}  http.ErrResponse	"Bad request"
// @Failure      500  {object}  http.ErrResponse	"Bad request"
// @Router 		 /employees/ 	[post]
func (c *Controller) CreateEmployee(ctx *fiber.Ctx) error {
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func
	var request CreateRequest

	// Парсинг тела запроса
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error(
			"When the body parse an CreateEmployee ended with an error: %s",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidRequestFormat)

	}
	// логируем тело запроса
	c.logger.Debug(
		"When the body parse an CreateEmployee was: received request",
		zap.Any("request", request),
		zap.String("request_id", requestId),
	)

	// Вызов сервиса
	newEmployee, err := c.employeeService.CreateEmployee(appContext, request)
	if err != nil {
		c.logger.Error( // логируем ошибку
			"When the create employee ended with an error:",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		switch { // Обработка ошибок с использованием ваших функций
		case errors.As(err, &domain.RequestValidationError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, validationFailed)

		case errors.Is(err, domain.ErrConflict):
			return http.ErrResponse(ctx, fiber.StatusConflict, "Employee already exists")

		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, internalServerError)
		}
	}

	// Успешный ответ с использованием ваших функций
	return http.CreatedResponse(ctx, newEmployee)
}

// FindById 	 godoc
// @Description  Find by ID employee
// @Summary 	 find by ID employee
// @Tags 		 employee
// @Accept  	 json
// @Produce 	 json
// @Param 		 id   path      	int  true  "Employee ID"
// @Success 	 200  {object}  	employee.Response	"Employee response"
// @Failure      400  {object}  	http.ErrResponse	"Bad request"
// @Failure      500  {object}  	http.ErrResponse	"Bad request"
// @Router 		 /employees/{id} 	[get]
func (c *Controller) FindById(ctx *fiber.Ctx) error {
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func
	pathUrl := ctx.Path()                          // Получаем путь запроса
	idStr := ctx.Params("id")                      // Анмаршалим path var запроса

	employeeID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.logger.Error(
			"When the body parse an UpdateEmployee ended with an error:",
			zap.Error(err),
			zap.String("path", pathUrl),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	response, err := c.employeeService.FindById(appContext, employeeID)
	if err != nil {
		c.logger.Error(
			"When the get Employee ended with an error:",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		switch {
		case errors.As(err, &domain.RequestValidationError{}), errors.As(err, &domain.AlreadyExistsError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
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

// GetAllPages   godoc
// @Description  Find all Employees by page
// @Summary		 get all employees by page
// @Tags 		 employee
// @Accept  	 json
// @Produce 	 json
// @Param   	 pageNumber 		query     	string     			false  "string valid"       minlength(1)  maxlength(10)
// @Param   	 pageSize 			query     	string     			false  "string valid"       minlength(1)  maxlength(155)
// @Param   	 textFilter 		query     	string     			false  "string valid"       minlength(0)  maxlength(10)
// @Success 	 200  				{object}  	employee.Response	"Employee request"
// @Failure      400  				{object}  	http.ErrResponse	"Bad request"
// @Failure      404  				{object}  	http.ErrResponse	"Bad request"
// @Failure      500  				{object}  	http.ErrResponse	"Bad request"
// @Router 		 /employees/page 	[get]
func (c *Controller) GetAllPages(ctx *fiber.Ctx) error {
	appContext := ctx.UserContext()                // получаем контекст приложения из запроса (задаем ранее в App main())
	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	var pageValues []int64
	var textFilter string

	pageValues, textFilter, err := c.parsePageValues(ctx, requestId)
	if err != nil {
		c.logger.Error(
			"Invalid parse page request values, error:",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidPageValuesFormat)
	}

	req := PageRequest{
		PageNumber: pageValues[0],
		PageSize:   pageValues[1],
		TextFilter: textFilter,
	}

	response, err := c.employeeService.GetAllByPage(appContext, req)
	if err != nil {
		c.logger.Error(
			"When the get All Employees by Page ended with an error: %s",
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

	return http.OkPageResponse(ctx, response)
}

// Логирование подозрительных запросов
func (c *Controller) checkForInjectionAttempt(input string, requestId string) {
	if strings.ContainsAny(input, ";'\"\\--") {
		c.logger.Warn(
			"Potential SQL injection attempt detected",
			zap.String("input", input),
			zap.String("request_id", requestId),
		)
		//for example
		//{"level":"warn","timestamp":"2025-07-10 19:01:06.294012","caller":"employee/controller.go:216","msg":"Potential SQL injection attempt detected","input":"test';DROP TABLE employees--","request_id":"80b67040-688b-4ee0-a3b9-f61213ffa76a"}
		// Можно добавить отправку алерта
	}
}

func (c *Controller) parsePageValues(ctx *fiber.Ctx, requestId string) ([]int64, string, error) {
	pageNumber := ctx.Query("pageNumber", "1")
	pageSize := ctx.Query("pageSize", "10")
	textFilter := ctx.Query("textFilter", "")

	c.logger.Debug(
		"GetAllPages request",
		zap.String("request_id", requestId),
		zap.String("method", "GET"),
		zap.String("url", ctx.OriginalURL()),
		zap.String("pageNumber ", pageNumber),
		zap.String("pageSize ", pageSize),
		zap.String("textFilter ", textFilter),
	)
	c.checkForInjectionAttempt(textFilter, requestId)
	_, _, err := c.checkNotNullRequestParam(pageNumber, pageSize, requestId)
	if err != nil {
		return nil, "", err
	}

	var pageList = []string{pageNumber, pageSize} // Declares and initializes with values, for example {10, 2}

	var pageValues []int64

	for _, idStr := range pageList {
		value, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page values format")
		}
		pageValues = append(pageValues, value)
	}
	// Двойная проверка (на случай если валидатор пропустит)
	if pageValues[0] < 1 || pageValues[1] < 1 {
		return nil, "", fmt.Errorf("invalid pagination parameters")
	}

	c.logger.Debug(
		"GetAllPages request",
		zap.String("request_id", requestId),
		zap.String("method", "GET"),
		zap.Int64("pageNumber", pageValues[0]),
		zap.Int64("pageSize", pageValues[1]),
	)

	return pageValues, textFilter, nil
}

func (c *Controller) checkNotNullRequestParam(
	pageNumber string,
	pageSize string,
	requestId string,
) ([]int64, string, error) {
	if pageNumber == "" {
		c.logger.Error(
			"When the parse pageNumber request param ended with an error:",
			zap.Error(nil),
			zap.String("request_id", requestId),
		)
		return nil, "", fmt.Errorf("missing pageNumber parameter")
	}

	if pageSize == "" {
		c.logger.Error(
			"When the parse pageSize request param ended with an error:",
			zap.Error(nil),
			zap.String("request_id", requestId),
		)
		return nil, "", fmt.Errorf("missing pageSize parameter")
	}

	return nil, "", nil
}

// FindAll   	 godoc
// @Description  Find all Employees
// @Summary		 get all employees
// @Tags 		 employee
// @Accept  	 json
// @Produce 	 json
// @Success 	 200  				{array}  	employee.Response	"Employee response"
// @Failure      400  				{object}  	http.ErrResponse	"Bad request"
// @Failure      500  				{object}  	http.ErrResponse	"Bad request"
// @Router 		 /employees/ 		[get]
func (c *Controller) FindAll(ctx *fiber.Ctx) error {
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	response, err := c.employeeService.FindAll(appContext)
	if err != nil {
		c.logger.Error(
			"When the find for ALl Employees ended with an error:",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		switch {
		case errors.Is(err, domain.ErrFindAllFailed):
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, "Failed to find all employees")
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return http.OkResponse(ctx, response)
}

// FindAllByIds  godoc
// @Description  Find all Employees by IDs
// @Summary		 get all employees by IDs
// @Tags 		 employee
// @Accept  	 json
// @Produce 	 json
// @Param   	 ids 				query     	string     			true  "Employees ids string values"       minlength(1)
// @Success 	 200  				{array}  	employee.Response	"Employee response"
// @Failure      400  				{object}  	http.ErrResponse	"Bad request"
// @Failure      500  				{object}  	http.ErrResponse	"Bad request"
// @Router 		 /employees/ 		[get]
func (c *Controller) FindAllByIds(ctx *fiber.Ctx) error {
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	idsParam := ctx.Query("ids")
	if idsParam == "" {
		c.logger.Error(
			"When the parse an Find All Employees By IDs request param ended with an error:",
			zap.Error(nil),
			zap.String("request_id", requestId),
		)
		return http.ErrResponse(ctx, fiber.StatusBadRequest, "Missing ids parameter")
	}

	var ids []int64
	for _, idStr := range strings.Split(idsParam, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.logger.Error(
				"When the parse an Find All Employees By IDs request param ended with an error:",
				zap.Error(err),
				zap.String("request_id", requestId),
			)

			return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat+idStr)
		}
		ids = append(ids, id)
	}

	response, err := c.employeeService.FindAllByIds(appContext, ids)
	if err != nil {
		c.logger.Error(
			"When the search for all employees by identifiers ended with an error: %s",
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

	return http.OkResponse(ctx, response)
}

// Update   	 godoc
// @Summary		 update employee by IDs
// @Description  Update Employee by IDs
// @Tags 		 employee
// @Accept  	 json
// @Produce 	 json
// @Param        id   				path      	int  					true  "Employee ID" minlength(1)
// @Param   	 request 			body     	employee.UpdateRequest	true  "Employee updated details"
// @Success 	 200  				{object}  	employee.Response		"Employee response"
// @Failure      400  				{object}  	http.ErrResponse		"Bad request"
// @Failure      500  				{object}  	http.ErrResponse		"Bad request"
// @Router 		 /employees/{id} 	[post]
func (c *Controller) Update(ctx *fiber.Ctx) error {
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	idStr := ctx.Params("id")
	employeeID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.logger.Error(
			"When the parse an Update Employee request ended with an error: %s",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	var request UpdateRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error(
			"When the parse an Update Employee request ended with an error: %s",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidRequestBody)
	}

	updatedEmployee, err := c.employeeService.UpdateEmployee(appContext, employeeID, request)
	if err != nil {
		c.logger.Error(
			"When the update for employee ended with an error: %s",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		switch {
		case errors.As(err, &domain.RequestValidationError{}):
			return http.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return http.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	return http.OkResponse(ctx, updatedEmployee)
}

// DeleteById  godoc
// @Description  Delete Employee by ID
// @Summary		 delete employee by ID
// @Tags 		 employee
// @Accept  	 json
// @Produce 	 json
// @Param        id   				path      	int  					true  "Employee ID" minlength(1)
// @Success 	 200  				{object}  	employee.Response		"Employee response"
// @Failure      400  				{object}  	http.ErrResponse		"Bad request"
// @Failure      500  				{object}  	http.ErrResponse		"Bad request"
// @Router 		 /employees/{id} 	[delete]
func (c *Controller) DeleteById(ctx *fiber.Ctx) error {
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	idStr := ctx.Params("id")
	employeeID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.logger.Error(
			"When the parse an Delete Employee By Id request param ended with an error: %s",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidIDFormat)
	}

	response, err := c.employeeService.DeleteById(appContext, employeeID)
	if err != nil {
		c.logger.Error(
			"When the delete an Employee By ID request ended with an error: %s",
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

	return http.OkResponse(ctx, response)
}

// DeleteByIds  godoc
// @Description  Find all Employees by IDs
// @Summary		 get all employees by IDs
// @Tags 		 employee
// @Accept  	 json
// @Produce 	 json
// @Param   	 ids 				query     	string     			true  "Employees ids string values"       minlength(1)
// @Success 	 200  				{array}  	employee.Response	"Employee array"
// @Failure      400  				{object}  	http.ErrResponse	"Bad request"
// @Failure      500  				{object}  	http.ErrResponse	"Bad request"
// @Router 		 /employees/ 		[delete]
func (c *Controller) DeleteByIds(ctx *fiber.Ctx) error {
	appContext := ctx.UserContext() // получаем контекст приложения из запроса (задаем ранее в App main())

	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	idsParam := ctx.Query("ids")
	if idsParam == "" {
		c.logger.Error(
			"When the parse an Delete Employees By IDs request param ended with an error:",
			zap.Error(nil),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, "Missing ids parameter")
	}

	var ids []int64
	for _, idStr := range strings.Split(idsParam, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.logger.Error(
				"parse request param ended with an error:",
				zap.Error(err),
				zap.String("request_id", requestId),
				zap.String("url", ctx.OriginalURL()),
				zap.String("method", ctx.Method()),
			)

			return http.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid ID format"+idStr)
		}
		ids = append(ids, id)
	}

	response, err := c.employeeService.DeleteByIds(appContext, ids)
	if err != nil {
		c.logger.Error(
			"delete Employees By Ids ended with an error: %s",
			zap.Error(err),
			zap.String("request_id", requestId),
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

// CreateEmployeeTx  godoc
// @Description  Create Employee by Transactional
// @Summary		 create employee by transactional
// @Tags 		 employee
// @Accept  	 json
// @Produce 	 json
// @Param   	 request			body    	employee.CreateRequest	true  "Employee creation details"       minlength(1)
// @Success 	 200  				{array}  	employee.Response		"Bad request"
// @Failure      400  				{object}  	http.ErrResponse		"Bad request"
// @Failure      500  				{object}  	http.ErrResponse		"Bad request"
// @Router 		 /employees/tx		[get]
func (c *Controller) CreateEmployeeTx(ctx *fiber.Ctx) error {
	appContext := ctx.UserContext()                // получаем контекст приложения из запроса (задаем ранее в App main())
	requestId := ctx.Locals("request_id").(string) // Получаем request_id благодаря middleware func

	var request CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error(
			"When the parse an Create Employee request ended with an error: %s",
			zap.Error(err),
			zap.String("request_id", requestId),
		)

		return http.ErrResponse(ctx, fiber.StatusBadRequest, invalidRequestBody)
	}

	response, err := c.employeeService.CreateEmployeeTx(appContext, request)
	if err != nil {
		c.logger.Error(
			"create Employee by Tx ended with an error: %s",
			zap.Error(err),
			zap.String("request_id", requestId),
			zap.String("url", ctx.OriginalURL()),
			zap.String("method", ctx.Method()),
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
