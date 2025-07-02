package main

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/role"
	"idm/inner/validator"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"idm/inner/config"
	"idm/inner/database"

	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/web"
)

func main() {
	//1. считывание конфигурации
	cfg := config.GetConfig(".env")
	var logger = common.NewLogger(cfg) // Создаем логгер
	logger.Debug("-->> Start Go App : IDM Project ")

	// 2. Создаём `Kонтекст` с отменой для управления ресурсами
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. Инициализация БД и миграций
	db, err := initializeDatabase(ctx, cfg, logger)
	if err != nil {
		logger.Fatal(
			"Database initialization failed:",
			zap.Error(err),
		)
	}

	//4. создание сервера
	var server = build(ctx, db, cfg, logger)

	//5. Запускаем сервер в отдельной горутине
	go func() {
		var err = server.App.Listen(":8080")
		if err != nil {
			logger.Panic(
				"HTTP server error:",
				zap.Error(err),
			) // паникуем через метод логгера (custom common. Logger)
		}
	}()

	//6. Создаем группу для ожидания сигнала завершения работы сервера
	var wg = &sync.WaitGroup{}
	wg.Add(1)

	//7. Запускаем gracefulShutdown в отдельной горутине
	go gracefulShutdown(ctx, server, db, wg, logger)

	//8. Ожидаем сигнал от горутины gracefulShutdown, что сервер завершил работу
	wg.Wait()
	logger.Info("graceful shutdown complete") // все события логируем через общий логгер
}

// Функция инициализация БД и миграций (Выносим инициализацию БД в отдельную функцию)
func initializeDatabase(
	ctx context.Context,
	cfg config.Config,
	logger *common.Logger,
) (*sqlx.DB, error) {
	//1. Инициализация БД
	db := database.ConnectDbWithCfg(cfg)

	//2. Проверяем соединение (аналог Ping)
	if err := db.PingContext(ctx); err != nil { //Проверка соединения с БД с контекстом
		if err := db.Close(); err != nil { // Закрываем соединение при ошибке
			logger.Error(
				"failed to close database connection: %s",
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	//3. Затем запускаем миграции// передаем *sql.DB если нужно
	if err := database.RunMigrations(db.DB); err != nil {
		if err := db.Close(); err != nil { // Закрываем соединение при ошибке
			logger.Error(
				"failed to close database connection: %s",
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	return db, nil
}

// Build - функция, конструирующая наш веб-сервер( - иначе Создание сервера с контекстом)
func build(
	ctx context.Context,
	dbase *sqlx.DB,
	cfg config.Config,
	logger *common.Logger,
) *web.Server {
	var server = web.NewServer(logger) // создаём веб-сервер
	var vld = validator.NewValidator() // создаём валидатор

	//routing
	var employeeRepo = employee.NewRepository(dbase)                                 // создаём репозиторий
	var employeeService = employee.NewService(employeeRepo, vld)                     // создаём сервис
	var employeeController = employee.NewController(server, employeeService, logger) // создаём контроллер
	employeeController.RegisterRoutes()

	var roleRepo = role.NewRepository(dbase)
	var roleService = role.NewService(roleRepo, vld)
	var roleController = role.NewController(server, roleService, logger)
	roleController.RegisterRoutes()

	var healthService = info.NewService(dbase, logger)
	var infoController = info.NewController(server, cfg, healthService, logger)
	infoController.RegisterRoutes()

	return server
}

// Функция "элегантного" завершения работы сервера по сигналу от операционной системы
func gracefulShutdown(
	ctx context.Context,
	server *web.Server,
	db *sqlx.DB,
	wg *sync.WaitGroup,
	logger *common.Logger,
) {
	// Уведомить основную горутину о завершении работы
	defer wg.Done()

	// Создаём контекст, который слушает сигналы прерывания от операционной системы
	ctx, stop := signal.NotifyContext(
		ctx, //было context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)
	defer stop()

	// Слушаем сигнал прерывания от операционной системы
	<-ctx.Done()
	logger.Info("shutting down gracefully, press Ctrl+C again to force")

	// Создаём контекст с таймаутом для завершения
	// Контекст используется для информирования веб-сервера о том,
	//что у него есть 5 секунд на выполнение запроса, который он обрабатывает в данный момент.
	//Используется context.WithTimeout для ограничения времени завершения (5 секунд).
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second) //было ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Закрываем сервер
	if err := server.App.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error(
			"Server forced to shutdown with error",
			zap.Error(err),
		)
	} else {
		logger.Info("Server shut down successfully")
	}

	// Закрываем БД, чтобы координировать с завершением сервера.
	if err := db.Close(); err != nil {
		logger.Error(
			"Error closing database:",
			zap.Error(err),
		)
	} else {
		logger.Info("Database closed successfully")
	}
}
