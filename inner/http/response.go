package http

import (
	"github.com/gofiber/fiber/v2"
)

// Response model info
// @Description Employee Controller response information
// @Description with success, error, date
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"error"`
	Data    interface{} `json:"data"`
}

// PageResponse model info
// @Description Employee Controller response information
// @Description with success, error, date
type PageResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"error"`
	Data    interface{} `json:"data"`
}

// fixme -ParseComment error in file /Users/kenobi/Projects/Home/Go/idm/inner/employee/controller.go :cannot find type definition: http.ErrResponse
// ErrResponse model info
// @Description ErrResponse Controller response information
// @Description with Response
func ErrResponse(
	c *fiber.Ctx,
	code int,
	message string,
) error {
	return c.Status(code).JSON(&Response{
		Success: false,
		Message: message,
		Data:    nil,
	})
}

func OkResponse(
	c *fiber.Ctx,
	data interface{},
) error {
	return c.Status(200).JSON(&Response{
		Success: true,
		Data:    data,
	})
}

func OkPageResponse(
	c *fiber.Ctx,
	data interface{},
) error {
	return c.Status(fiber.StatusOK).JSON(&PageResponse{
		Success: true,
		Data:    data,
	})
}

// CreatedResponse - Дополнительная функция для 201 Created
func CreatedResponse(
	c *fiber.Ctx,
	data interface{},
) error {
	return c.Status(fiber.StatusCreated).JSON(&Response{
		Success: true,
		Data:    data,
	})
}
