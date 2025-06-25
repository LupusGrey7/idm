package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"idm/inner/role"
	"idm/inner/validator"

	"github.com/jmoiron/sqlx"
	"idm/inner/config"
	"idm/inner/database"

	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/web"
)

func main() {
	fmt.Println("Hello, Go.-->> Start App")
	// считывание конфигурации
	cfg := config.GetConfig(".env")

	// Сначала подключаемся к БД
	var db, err = sql.Open(cfg.DbDriverName, cfg.Dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	// Затем запускаем миграции
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Migrations failed:", err)
	}

	// Потом создаём основное соединение и сервер
	var dbase = database.ConnectDbWithCfg(cfg)
	// отложенная функция - закрываем соединение с базой данных после выхода из функции main
	defer func() {
		if err := dbase.Close(); err != nil {
			fmt.Printf("error closing db: %v", err)
		}
	}()
	var server = build(dbase, cfg)
	err = server.App.Listen(":8080")
	if err != nil {
		panic(fmt.Sprintf("http server error: %s", err))
	}
}

// Buil - функция, конструирующая наш веб-сервер
func build(database *sqlx.DB, cfg config.Config) *web.Server {
	// создаём веб-сервер
	var server = web.NewServer()

	// создаём репозиторий
	var employeeRepo = employee.NewRepository(database)
	var roleRepo = role.NewRepository(database)

	// создаём валидатор
	var vld = validator.NewValidator()

	// создаём сервисы
	var employeeService = employee.NewService(employeeRepo, vld)
	var roleService = role.NewService(roleRepo, vld)
	var healthService = info.NewService(database)

	// создаём контроллеры
	var employeeController = employee.NewController(server, employeeService)
	employeeController.RegisterRoutes()
	var roleController = role.NewController(server, roleService)
	roleController.RegisterRoutes()
	var infoController = info.NewController(server, cfg, healthService)
	infoController.RegisterRoutes()

	return server
}
