package domain

import "github.com/gofiber/fiber/v2"

//Структуры кастомных ошибок - Доменные ошибки
//Доменные ошибки отдельно от транспортных

// RequestValidationError - ошибка валидации запроса
type RequestValidationError struct {
	Message string
}

func (err RequestValidationError) Error() string {
	return err.Message
}

// AlreadyExistsError - ошибка, когда объект уже существует
type AlreadyExistsError struct {
	Message string
}

func (err AlreadyExistsError) Error() string {
	return err.Message
}

// Унифицированная структура для API ошибок
type APIError struct {
	Code    int    `json:"code"`    // HTTP-статус код
	Message string `json:"message"` // Человекочитаемое сообщение
	Details string `json:"details"` // Технические детали (опционально)
}

// Реализуем error interface
func (e APIError) Error() string {
	return e.Message
}

// Конструкторы для конкретных ошибок
func NewDBUnavailableError(details string) APIError {
	return APIError{
		Code:    fiber.StatusServiceUnavailable,
		Message: "Database service unavailable",
		Details: details,
	}
}

func NewInternalServerError(details string) APIError {
	return APIError{
		Code:    fiber.StatusInternalServerError,
		Message: "Internal server error",
		Details: details,
	}
}

type NotFoundError struct {
	Message string
}

func (err NotFoundError) Error() string {
	return err.Message
}
