package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jmoiron/sqlx"
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
	fmt.Println("Hello, Go.-->> Start App")

	//1. считывание конфигурации
	cfg := config.GetConfig(".env")

	// 2. Инициализация БД и миграций
	db, err := initializeDatabase(cfg)
	if err != nil {
		log.Fatal("Database initialization failed:", err)
	}
	//9. отложенная функция - закрываем соединение с базой данных после выхода из функции main
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("error closing db: %v", err)
		}
	}()

	var server = build(db, cfg) //3. создание сервера
	go func() {                 //4. Запускаем сервер в отдельной горутине
		var err = server.App.Listen(":8080")
		if err != nil {
			log.Debug("HTTP server error: %v", err) // Логируем вместо panic при желании
			panic(fmt.Sprintf("http server error: %s", err))
		}
	}()

	//5. Создаем группу для ожидания сигнала завершения работы сервера
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	//7. Запускаем gracefulShutdown в отдельной горутине
	go gracefulShutdown(server, wg)
	//8. Ожидаем сигнал от горутины gracefulShutdown, что сервер завершил работу
	wg.Wait()
	fmt.Println("Graceful shutdown complete.")
}

// Выносим инициализацию БД в отдельную функцию

func initializeDatabase(cfg config.Config) (*sqlx.DB, error) {
	db := database.ConnectDbWithCfg(cfg) // Инициализация БД

	// Проверяем соединение (аналог Ping)
	if err := db.Ping(); err != nil {
		if err := db.Close(); err != nil { // Закрываем соединение при ошибке
			log.Error("failed to close database connection: %v", err)
		}
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	// Затем запускаем миграции// передаем *sql.DB если нужно
	if err := database.RunMigrations(db.DB); err != nil {
		if err := db.Close(); err != nil { // Закрываем соединение при ошибке
			log.Error("failed to close database connection: %v", err)
		}
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	return db, nil
}

// Функция "элегантного" завершения работы сервера по сигналу от операционной системы
func gracefulShutdown(server *web.Server, wg *sync.WaitGroup) {
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
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")
	// Контекст используется для информирования веб-сервера о том,
	// что у него есть 5 секунд на выполнение запроса, который он обрабатывает в данный момент
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.App.ShutdownWithContext(ctx); err != nil {
		fmt.Printf("Server forced to shutdown with error: %v\n", err)
	}
	fmt.Println("Server exiting")
}

// Build - функция, конструирующая наш веб-сервер
func build(dbase *sqlx.DB, cfg config.Config) *web.Server {
	var server = web.NewServer()       // создаём веб-сервер
	var vld = validator.NewValidator() // создаём валидатор
	//routing
	var employeeRepo = employee.NewRepository(dbase)                         // создаём репозиторий
	var employeeService = employee.NewService(employeeRepo, vld)             // создаём сервис
	var employeeController = employee.NewController(server, employeeService) // создаём контроллер
	employeeController.RegisterRoutes()

	var roleRepo = role.NewRepository(dbase)
	var roleService = role.NewService(roleRepo, vld)
	var roleController = role.NewController(server, roleService)
	roleController.RegisterRoutes()

	var healthService = info.NewService(dbase)
	var infoController = info.NewController(server, cfg, healthService)
	infoController.RegisterRoutes()

	return server
}
