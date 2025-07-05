package http

import (
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"error"`
	Data    interface{} `json:"data"`
}
type PageResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"error"`
	Data    interface{} `json:"data"`
}

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
