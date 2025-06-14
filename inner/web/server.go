package web

import "github.com/gofiber/fiber"

// Server - Cтруктура веб-сервера
type Server struct {
	App        *fiber.App
	GroupApiV1 fiber.Router
}

// NewServer - функция-конструктор
func NewServer() *Server {

	// создаём новый web-сервер
	app := fiber.New()

	// создаём группу "/api" - Group is used for Routes
	groupApi := app.Group("/api")

	// создаём подгруппу "api/v1"
	groupApiV1 := groupApi.Group("/v1")
	return &Server{
		App:        app,
		GroupApiV1: groupApiV1,
	}
}
