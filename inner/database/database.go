package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"idm/inner/common"
	"time"
)

// DB Временная переменная, которая будет ссылаться на подключение к базе данных.
// Позже мы от неё избавимся - // *sqlx.DB - указатель на объект базы данных
//var DB *sqlx.DB

// ConnectDb получить конфиг и подключиться с ним к базе данных
func ConnectDb() *sqlx.DB {
	cfg := common.GetConfig(".env")
	return ConnectDbWithCfg(cfg)
}

/**
* func ConnectDbWithCfg() подключиться к базе данных с переданным конфигом,
* Настройки ниже конфигурируют пулл подключений к базе данных.
* Их названия стандартны для большинства библиотек.
* Ознакомиться с их описанием можно на примере документации Hikari pool:
* https://github.com/brettwooldridge/HikariCP?tab=readme-ov-file#gear-configuration-knobs-baby
 */
func ConnectDbWithCfg(cfg common.Config) *sqlx.DB {
	db := sqlx.MustConnect(cfg.DbDriverName, cfg.Dsn)

	// Настраиваем пул соединений через глобальную переменную DB
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	return db
}
