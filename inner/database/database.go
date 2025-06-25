package database

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"idm/inner/config"
	"log"
	"time"
)

// DB Временная переменная, которая будет ссылаться на подключение к базе данных.
// Позже мы от неё избавимся - // *sqlx.DB - указатель на объект базы данных
//var DB *sqlx.DB

// ConnectDb получить конфиг и подключиться с ним к базе данных
func ConnectDb() *sqlx.DB {
	cfg := config.GetConfig(".env")
	log.Printf("cfn env file %v", cfg.Dsn)

	return ConnectDbWithCfg(cfg)
}

/**
* func ConnectDbWithCfg() подключиться к базе данных с переданным конфигом,
* Настройки ниже конфигурируют пулл подключений к базе данных.
* Их названия стандартны для большинства библиотек.
* Ознакомиться с их описанием можно на примере документации Hikari pool:
* https://github.com/brettwooldridge/HikariCP?tab=readme-ov-file#gear-configuration-knobs-baby
 */
func ConnectDbWithCfg(cfg config.Config) *sqlx.DB {
	db := sqlx.MustConnect(cfg.DbDriverName, cfg.Dsn)

	// Настраиваем пул соединений через глобальную переменную DB
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	return db
}

func RunMigrations(db *sql.DB) error {
	goose.SetTableName("goose_db_version") // явно задаём имя таблицы

	// 1. Проверяем, существует ли таблица миграций
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_name = 'goose_db_version'
        )`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check migrations table: %w", err)
	}

	// 2. Если таблицы нет - создаём
	if !exists {
		if err := goose.Up(db, "./migrations"); err != nil {
			return fmt.Errorf("failed to apply initial migrations: %w", err)
		}
	}

	// 3. Проверяем наличие новых миграций
	if err := goose.Status(db, "./migrations"); err != nil {
		return fmt.Errorf("migration status check failed: %w", err)
	}

	return nil
}
