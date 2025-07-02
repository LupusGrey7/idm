package web

import (
	"github.com/gofiber/fiber/v2"
	"idm/inner/common"
	"idm/inner/web/middleware"
)

const (
	APIPrefix     = "/api"
	APIVersion    = "/v1"
	EmployeesPath = "/employees"
	RolesPath     = "/roles"
	InternalPath  = "/internal"
)

// Server - Cтруктура веб-сервера
type Server struct {
	App            *fiber.App
	GroupApiV1     fiber.Router
	GroupEmployees fiber.Router
	GroupRoles     fiber.Router
	GroupInternal  fiber.Router // Группа непубличного API
}

// NewServer - функция-конструктор
func NewServer(logger *common.Logger) *Server {

	// создаём новый web-сервер
	app := fiber.New()
	middleware.RegisterMiddleware(app, logger) // регистрация middleware, передаем logger

	groupInternal := app.Group(InternalPath)          // Группа непубличного API "/internal"
	groupApi := app.Group(APIPrefix)                  // создаём группу "/api" - Group is used for Routes
	groupApiV1 := groupApi.Group(APIVersion)          // создаём подгруппу "api/v1"
	groupEmployees := groupApiV1.Group(EmployeesPath) // создаём подгруппу "/employees"
	groupRoles := groupApiV1.Group(RolesPath)         // создаём подгруппу "/roles"

	return &Server{
		App:            app,
		GroupApiV1:     groupApiV1,
		GroupEmployees: groupEmployees,
		GroupRoles:     groupRoles,
		GroupInternal:  groupInternal,
	}
}
