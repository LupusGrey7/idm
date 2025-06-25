package info

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/config"
	"idm/inner/domain"
	"idm/inner/web"
)

type Controller struct {
	server *web.Server
	cfg    config.Config
	svc    Svc
	logger *common.Logger
}

type Svc interface {
	CheckDB() error
}

func NewController(
	server *web.Server,
	cfg config.Config,
	svc Svc,
	logger *common.Logger,
) *Controller {
	return &Controller{
		server: server,
		cfg:    cfg,
		svc:    svc,
		logger: logger,
	}
}

func (c *Controller) RegisterRoutes() {
	c.server.GroupInternal.Get("/info", c.GetInfo)     // полный путь будет "/internal/info"
	c.server.GroupInternal.Get("/health", c.GetHealth) // полный путь будет "/internal/health"
}

// GetInfo получение информации о приложении
func (c *Controller) GetInfo(ctx *fiber.Ctx) error {
	if err := c.svc.CheckDB(); err != nil {
		return ctx.Status(fiber.StatusServiceUnavailable).
			JSON(domain.NewDBUnavailableError(err.Error()))
	}

	response := Response{
		Name:    c.cfg.AppName,
		Version: c.cfg.AppVersion,
		Status:  "OK",
	}

	if err := ctx.Status(fiber.StatusOK).JSON(response); err != nil {
		c.logger.Error("Failed to encode response %s", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).
			JSON(domain.NewInternalServerError("response serialization failed"))
	}

	return nil
}

// GetHealth проверка работоспособности приложения
func (c *Controller) GetHealth(ctx *fiber.Ctx) error {
	if err := c.svc.CheckDB(); err != nil {
		c.logger.Error("Failed to Check DB Health %s", zap.Error(err))
		return ctx.Status(503).SendString("DB unavailable")
	}

	return ctx.Status(200).SendString("OK")
}
