package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"idm/inner/common"

	"go.uber.org/zap"
)

// RegisterMiddleware - функция регистрации middleware
func RegisterMiddleware(app *fiber.App, logger *common.Logger) {
	app.Use(requestid.New(requestid.Config{ // Middleware для генерации requestId
		Header: "X-Request-Id", // Заголовок для request_id
		Generator: func() string {
			return uuid.New().String() // Генерируем UUID
		},
		ContextKey: "request_id", // Ключ для сохранения в ctx.Locals
	})) // middleware для генерации уникального id запроса

	// Логирование всех запросов
	app.Use(func(c *fiber.Ctx) error {
		logger.Info("Request received",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("request_id", c.Locals("request_id").(string)),
		)
		return c.Next()
	})

	//app.Use(recover.New()) // middleware для восстановления после паники
	// Recover middleware - использовать в стабильных версиях Fiber после  v2.52.8(июнь 2025)  присутствует баг который не обойти
	//app.Use(recover.New(recover.Config{
	//	EnableStackTrace: true,
	//	StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
	//		requestID, _ := c.Locals("request_id").(string)
	//		logger.Info("StackTraceHandler called")
	//		logger.Error("Panic recovered",
	//			zap.Any("error", e),
	//			zap.String("path", c.Path()),
	//			zap.String("request_id", requestID),
	//		)
	//		// Очищаем контекст ответа, чтобы избежать дефолтного ответа Fiber
	//		c.Response().Reset()
	//		// Явно устанавливаем Content-Type
	//		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	//		// Отправляем JSON
	//		err := c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	//			"error": "Internal server error",
	//		})
	//		if err != nil {
	//			logger.Error("Failed to send JSON response", zap.Error(err))
	//		} else {
	//			logger.Info("JSON response sent successfully")
	//		}
	//	},
	//}))
	app.Use(CustomRecoverMiddleware(logger)) //app.Use(recover.New()) // middleware для восстановления после паники

}

// Кастомный middleware (в fiber v2 v2.52.8(июнь 2025)  присутс баг который не обойти)
// В Fiber v2.52.8 есть баг в middleware recover: даже если StackTraceHandler устанавливает JSON-ответ, Fiber может отправлять дефолтный ответ (text/plain с текстом паники),
// если ответ уже частично сформирован до вызова StackTraceHandler. Это подтверждается в GitHub issues Fiber.
// Fiber игнорирует изменения заголовков после начала формирования ответа.
// Согласно информации из GitHub Releases Fiber, баг в recover middleware, связанный с неправильной обработкой ответа в StackTraceHandler,
// может быть исправлен в более поздних версиях v2 или в ветке v3.
// Последняя стабильная версия на момент 30 июня 2025 года — v2.52.8,
func CustomRecoverMiddleware(logger *common.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := c.Locals("request_id").(string)
				logger.Info("Custom recover handler called")
				logger.Error("Panic recovered",
					zap.Any("error", r),
					zap.String("path", c.Path()),
					zap.String("request_id", requestID),
				)
				c.Response().Reset() // Очищаем ответ
				c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
				err := c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
				if err != nil {
					logger.Error("Failed to send JSON response", zap.Error(err))
				} else {
					logger.Info("JSON response sent successfully")
				}
			}
		}()
		return c.Next()
	}
}
