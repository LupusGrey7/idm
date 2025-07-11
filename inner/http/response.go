package http

import (
	"github.com/gofiber/fiber/v2"
)

// Response model info
// Response is a common response structure
// @Description Common API response format
type Response struct {
	Success bool        `json:"success" example:"false"`
	Message string      `json:"error" example:"Invalid request format"`
	Data    interface{} `json:"data"`
}

// PageResponse model info
// PageResponse is a common pageResponse structure
// @Description Common API pageResponse format
type PageResponse struct {
	Success bool        `json:"success" example:"false"`
	Message string      `json:"error" example:"Invalid request format"`
	Data    interface{} `json:"data"`
}

// ErrResponse model info
// @Description ErrResponse Controller response information
// @Description with status, Response
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

// OkResponse
// @Description ErrResponse Controller response information
// @Description with status 200, Response
func OkResponse(
	c *fiber.Ctx,
	data interface{},
) error {
	return c.Status(200).JSON(&Response{
		Success: true,
		Data:    data,
	})
}

// OkPageResponse
// @Description ErrResponse Controller response information
// @Description with status 200, PageResponse
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
// @Description ErrResponse Controller response information
// @Description with status 201, Response
func CreatedResponse(
	c *fiber.Ctx,
	data interface{},
) error {
	return c.Status(fiber.StatusCreated).JSON(&Response{
		Success: true,
		Data:    data,
	})
}
