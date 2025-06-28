package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"idm/inner/common"

	"github.com/gofiber/fiber/v2/middleware/recover"

	"go.uber.org/zap"
)

// RegisterMiddleware - функция регистрации middleware
func RegisterMiddleware(app *fiber.App, logger *common.Logger) {
	//app.Use(recover.New()) // middleware для восстановления после паники
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true, // Включаем стек-трейс для отладки
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) { // Логируем панику с request_id и путём
			logger.Error("Panic recovered",
				zap.Any("error", e),
				zap.String("path", c.Path()),
				zap.String("request_id", c.Locals("request_id").(string)),
			)
			// Кастомный ответ для паники
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		},
	}))
	app.Use(requestid.New(requestid.Config{ // Middleware для генерации requestId
		Header: "X-Request-Id", // Заголовок для request_id
		Generator: func() string {
			return uuid.New().String() // Генерируем UUID
		},
		ContextKey: "request_id", // Ключ для сохранения в ctx.Locals
	})) // middleware для генерации уникального id запроса
}
