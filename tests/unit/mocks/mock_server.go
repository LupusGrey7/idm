package mocks

import (
	"github.com/gofiber/fiber"
	"github.com/stretchr/testify/mock"
	"idm/inner/web"
)

type MockWebServer struct {
	mock.Mock
	web.Server // Встраиваем оригинальную структуру
}

func NewMockWebServer() *MockWebServer {
	app := fiber.New()
	groupApi := app.Group("/api")
	groupApiV1 := groupApi.Group("/v1")
	groupEmployees := groupApiV1.Group("/employees")

	return &MockWebServer{
		Server: web.Server{
			App:            app,
			GroupApiV1:     groupApiV1,
			GroupEmployees: groupEmployees,
		},
	}
}
