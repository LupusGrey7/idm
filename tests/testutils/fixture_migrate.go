package testutils

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Драйвер PostgreSQL
	"github.com/pressly/goose/v3"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func LoadTestConfig() {
	// Получаем путь к директории текущего файла
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// Поднимаемся на 2 уровня вверх (из testutils в tests)
	baseDir := filepath.Join(dir, "../..")

	// Формируем путь к .env.test
	envPath := filepath.Join(baseDir, "tests/testdata", ".env.test")

	// Загружаем конфиг
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env.test file from %s: %v", envPath, err)
	}

	// Проверяем, что переменные загрузились
	if os.Getenv("DB_DSN") == "" {
		log.Fatal("DB_DSN not set in .env.test")
	}
}

func InitTestDB() *sqlx.DB {
	LoadTestConfig()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN not set in .env.test")
	}
	println("Test DB_DSN --->>: ", dsn)

	driver := os.Getenv("DB_DRIVER_NAME")
	println("Test DB_DRIVER --->>: ", driver)

	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := ApplyMigrations(db.DB); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	return db
}

func ApplyMigrations(db *sql.DB) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		log.Println("Error GooSe Dialect set data!")
		return err
	}
	// Получаем путь к директории текущего файла
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	// Поднимаемся на 2 уровня вверх (из testutils в tests)
	baseDir := filepath.Join(dir, "../..")

	// Формируем путь к .env.test
	migrationsPath := filepath.Join(baseDir, "migrations")

	//migrationsPath := filepath.Join(dir, "../../../migrations")

	if err := goose.Up(db, migrationsPath); err != nil {
		return fmt.Errorf("goose up failed: %w", err)
	}
	return nil
}

//// ApplyMigrations - применяет миграции к тестовой БД
//func ApplyMigrations(db *sql.DB) error {
//	goose.SetDialect("postgres")
//
//	// Получаем путь к миграциям
//	_, filename, _, _ := runtime.Caller(0)
//	dir := filepath.Dir(filename)
//	migrationsPath := filepath.Join(dir, "../../../migrations")
//
//	// Проверяем существование директории
//	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
//		return fmt.Errorf("migrations directory not found: %s", migrationsPath)
//	}
//
//	log.Printf("Applying migrations from: %s", migrationsPath)
//
//	// Применяем миграции
//	if err := goose.Up(db, migrationsPath); err != nil {
//		return fmt.Errorf("goose up failed: %w", err)
//	}
//
//	return nil
//}
//
//// InitTestDB создает подключение к БД и применяет миграции
//func InitTestDB() *sqlx.DB {
//	// Определяем путь к test.env независимо от рабочей директории
//	_, filename, _, _ := runtime.Caller(0)
//	dir := filepath.Dir(filename)
//	envPath := filepath.Join(dir, "../testdata/.env.test")
//
//	// Загружаем конфигурацию для тестов
//	if err := godotenv.Load(envPath); err != nil {
//		log.Fatalf("Error loading .env.test: %v", err)
//	}
//
//	dsn := os.Getenv("DB_DSN")
//	if dsn == "" {
//		log.Fatal("DB_DSN not set in .env.tets")
//	}
//
//	// Подключаемся к БД
//	db, err := sqlx.Connect("postgres", dsn)
//	if err != nil {
//		log.Fatalf("Failed to connect to test database: %v", err)
//	}
//
//	// Применяем миграции
//	if err := ApplyMigrations(db.DB); err != nil {
//		log.Fatalf("Migrations failed: %v", err)
//	}
//
//	return db
//}

//func ApplyMigrations(db *sql.DB) error {
//	goose.SetDialect("postgres")
//
//	// Получаем абсолютный путь к миграциям
//	wd, err := os.Getwd()
//	if err != nil {
//		log.Fatalf("Failed to get working directory: %v", err)
//	}
//
//	// Путь к миграциям относительно текущей директории тестов
//	migrationsDir := filepath.Join(wd, "migrations")
//	log.Printf("Applying migrations from: %s", migrationsDir)
//
//	// Получаем текущий статус миграций
//	if err := goose.Status(db, migrationsDir); err != nil {
//		log.Printf("Migration status error: %v", err)
//	}
//
//	// Применяем миграции
//	if err := goose.Up(db, migrationsDir); err != nil {
//		log.Fatalf("Failed to apply migrations: %v", err)
//	}
//
//	// Проверяем финальный статус
//	if err := goose.Status(db, migrationsDir); err != nil {
//		log.Printf("Post-migration status error: %v", err)
//	}
//	return err
//}
