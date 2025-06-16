package web

import "github.com/gofiber/fiber"

const (
	APIPrefix     = "/api"
	APIVersion    = "/v1"
	EmployeesPath = "/employees"
)

// Server - Cтруктура веб-сервера
type Server struct {
	App            *fiber.App
	GroupApiV1     fiber.Router
	GroupEmployees fiber.Router
}

// NewServer - функция-конструктор
func NewServer() *Server {

	// создаём новый web-сервер
	app := fiber.New()

	// создаём группу "/api" - Group is used for Routes
	groupApi := app.Group(APIPrefix)

	// создаём подгруппу "api/v1"
	groupApiV1 := groupApi.Group(APIVersion)
	// создаём подгруппу "/employees"
	groupEmployees := groupApiV1.Group(EmployeesPath)

	return &Server{
		App:            app,
		GroupApiV1:     groupApiV1,
		GroupEmployees: groupEmployees,
	}
}
