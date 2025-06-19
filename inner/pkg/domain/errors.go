package domain

import "errors"

var (
	ErrValidation   = errors.New("validation error")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
)

// Специфичные ошибки для сотрудников
var (
	ErrEmployeeAlreadyExists = errors.New("employee already exists")
	ErrInvalidEmployeeData   = errors.New("invalid employee data")
)
