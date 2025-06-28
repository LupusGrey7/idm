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
	cfg := config.GetConfig(".env")    //1. считывание конфигурации
	var logger = common.NewLogger(cfg) // Создаем логгер
	logger.Debug("-->> Start Go App : IDM Project ")

	db, err := initializeDatabase(cfg, logger) // 2. Инициализация БД и миграций
	if err != nil {
		logger.Fatal("Database initialization failed:", zap.Error(err))
	}
	//9. отложенная функция - закрываем соединение с базой данных после выхода из функции main
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("error closing db: %s", zap.Error(err))
		}
	}()

	var server = build(db, cfg, logger) //3. создание сервера
	go func() {                         //4. Запускаем сервер в отдельной горутине
		var err = server.App.Listen(":8080")
		if err != nil {
			logger.Panic("HTTP server error: %s", zap.Error(err)) // паникуем через метод логгера (custom common. Logger)
		}
	}()

	//5. Создаем группу для ожидания сигнала завершения работы сервера
	var wg = &sync.WaitGroup{}
	wg.Add(1)

	go gracefulShutdown(server, wg, logger) //7. Запускаем gracefulShutdown в отдельной горутине

	wg.Wait()                                 //8. Ожидаем сигнал от горутины gracefulShutdown, что сервер завершил работу
	logger.Info("graceful shutdown complete") // все события логируем через общий логгер
}

// Выносим инициализацию БД в отдельную функцию

func initializeDatabase(cfg config.Config, logger *common.Logger) (*sqlx.DB, error) {
	db := database.ConnectDbWithCfg(cfg) // Инициализация БД

	// Проверяем соединение (аналог Ping)
	if err := db.Ping(); err != nil {
		if err := db.Close(); err != nil { // Закрываем соединение при ошибке
			logger.Error("failed to close database connection: %s", zap.Error(err))
		}
		return nil, fmt.Errorf("failed to ping DB: %w", err) //fixme
	}

	// Затем запускаем миграции// передаем *sql.DB если нужно
	if err := database.RunMigrations(db.DB); err != nil {
		if err := db.Close(); err != nil { // Закрываем соединение при ошибке
			logger.Error("failed to close database connection: %s", zap.Error(err))
		}
		return nil, fmt.Errorf("migrations failed: %w", err) //fixme
	}

	return db, nil
}

// Функция "элегантного" завершения работы сервера по сигналу от операционной системы
func gracefulShutdown(
	server *web.Server,
	wg *sync.WaitGroup,
	logger *common.Logger,
) {
	// Уведомить основную горутину о завершении работы
	defer wg.Done()
	// Создаём контекст, который слушает сигналы прерывания от операционной системы
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)
	defer stop()
	// Слушаем сигнал прерывания от операционной системы
	<-ctx.Done()
	logger.Info("shutting down gracefully, press Ctrl+C again to force")
	// Контекст используется для информирования веб-сервера о том,
	// что у него есть 5 секунд на выполнение запроса, который он обрабатывает в данный момент
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.App.ShutdownWithContext(ctx); err != nil {
		logger.Error("Server forced to shutdown with error", zap.Error(err))
	}
	logger.Info("Server exiting")
}

// Build - функция, конструирующая наш веб-сервер
func build(
	dbase *sqlx.DB,
	cfg config.Config,
	logger *common.Logger,
) *web.Server {
	var server = web.NewServer()       // создаём веб-сервер
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
